package history

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"scout/internal/domain"
)

// fixture is a git repo with controlled author identity and dates.
type fixture struct {
	t    *testing.T
	dir  string
	last time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, dir: t.TempDir()}
	f.git(nil, "init", "-q", "-b", "main")
	return f
}

func (f *fixture) git(env []string, args ...string) string {
	f.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", f.dir, "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		f.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// commit writes files (path -> full content) and commits as name/email at when.
func (f *fixture) commit(when time.Time, name, email string, files map[string]string) {
	f.t.Helper()
	for p, c := range files {
		full := filepath.Join(f.dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			f.t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			f.t.Fatal(err)
		}
		f.git(nil, "add", "--", p)
	}
	ts := when.UTC().Format(time.RFC3339)
	env := []string{"GIT_AUTHOR_NAME=" + name, "GIT_AUTHOR_EMAIL=" + email, "GIT_AUTHOR_DATE=" + ts,
		"GIT_COMMITTER_NAME=" + name, "GIT_COMMITTER_EMAIL=" + email, "GIT_COMMITTER_DATE=" + ts}
	f.git(env, "commit", "-q", "--allow-empty", "-m", "c")
	f.last = when
}

func lines(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

// windowStart is the start of the k-th UTC calendar quarter since 1970-Q1.
func windowStart(k int) time.Time { return time.Date(1970, time.Month(1+3*k), 1, 0, 0, 0, 0, time.UTC) }

func scopesAB() []Scope {
	return []Scope{PrefixScope(domain.SignalScopeDirectory, "a", "a"), PrefixScope(domain.SignalScopeDirectory, "b", "b")}
}

func find(t *testing.T, sigs []domain.Signal, metric, anchor, window string) domain.Signal {
	t.Helper()
	for _, s := range sigs {
		if s.ID() == metric && s.ScopeAnchor == anchor && s.Window == window {
			return s
		}
	}
	t.Fatalf("no signal %s %s %s", metric, anchor, window)
	return domain.Signal{}
}

func near(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil || math.Abs(*got-want) > 1e-9 {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}

func winLabel(k int) string {
	s := windowStart(k)
	return fmt.Sprintf("%d-Q%d", s.Year(), (int(s.Month())-1)/3+1)
}

// 30 changes, one per day from the start of window k=200: alice 20, bob 10.
func buildMain(t *testing.T) (*fixture, time.Time) {
	f := newFixture(t)
	start := windowStart(200)
	acc := ""
	for i := 0; i < 30; i++ {
		name, email := "alice", "alice@example.com"
		if i%3 == 2 {
			name, email = "bob", "bob@example.com"
		}
		for j := 0; j <= i; j++ {
			acc += fmt.Sprintf("c%d l%d\n", i, j) // commit i adds i+1 unique lines
		}
		files := map[string]string{"a/f.go": acc}
		if i%2 == 1 {
			files["b/g.go"] = lines(i)
		}
		f.commit(start.Add(time.Duration(i)*24*time.Hour+time.Hour), name, email, files)
	}
	return f, windowStart(201)
}

func TestCommitLevelMeasuresOnHandComputableFixture(t *testing.T) {
	f, end := buildMain(t)
	commits, err := Load(context.Background(), f.dir)
	if err != nil {
		t.Fatal(err)
	}
	opts := Options{Scopes: scopesAB(), End: end.Add(-time.Second), Snapshot: "snap",
		Files: map[string]FileMetric{"a/f.go": {CC: 4, NLOC: 100}}}
	sigs := Compute(commits, opts)
	for _, s := range sigs {
		if err := s.Validate(); err != nil {
			t.Fatalf("invalid signal %s %s %s: %v", s.ID(), s.ScopeAnchor, s.Window, err)
		}
		if s.Claim != domain.ClaimDescriptive {
			t.Fatalf("claim %q", s.Claim)
		}
	}
	w := winLabel(200)

	ch := find(t, sigs, "scout.changes@1", "a", w)
	near(t, "changes", ch.Value, 30)
	if ch.Detail["authors"] != 2 {
		t.Errorf("authors = %v", ch.Detail["authors"])
	}

	churn := find(t, sigs, "scout.churn_per_change@1", "a", w) // added lines 1..30
	near(t, "churn p50", churn.Value, 15.5)
	near(t, "churn num", churn.Numerator, 465)
	near(t, "churn den", churn.Denominator, 30)
	if got := churn.Detail["p90"].(float64); math.Abs(got-27.1) > 1e-9 {
		t.Errorf("churn p90 = %v, want 27.1", got)
	}

	amp := find(t, sigs, "scout.amplification@1", "a", w) // 15 changes touch 1 file, 15 touch 2
	near(t, "amp p50 files", amp.Value, 1.5)
	if amp.Detail["p50_scopes"].(float64) != 1.5 {
		t.Errorf("p50 scopes = %v", amp.Detail["p50_scopes"])
	}

	fu := find(t, sigs, "scout.follow_up_rate@1", "a", w) // every change but the last is followed next day
	near(t, "follow num", fu.Numerator, 29)
	near(t, "follow den", fu.Denominator, 30)
	near(t, "follow value", fu.Value, 29.0/30.0)

	own := find(t, sigs, "scout.ownership@1", "a", w)
	near(t, "ownership", own.Value, 20.0/30.0)
	if own.Detail["authors"] != 2 || own.Detail["minor_contributors"] != 0 {
		t.Errorf("ownership detail %v", own.Detail)
	}
	key := own.Detail["top_author_key"].(string)
	if strings.Contains(key, "alice") || len(key) != 12 {
		t.Errorf("ownership key not pseudonymised: %q", key)
	}

	hot := find(t, sigs, "scout.hotspot@1", "a", w)
	near(t, "hotspot", hot.Value, 4*math.Log1p(30)/100)
	near(t, "hotspot den", hot.Denominator, 100)
	if hb := find(t, sigs, "scout.hotspot@1", "b", w); hb.Missing != domain.SignalUnmeasured {
		t.Errorf("hotspot without NLOC should be unmeasured, got %q", hb.Missing)
	}

	// Scope b: 15 changes (odd i) < 20 => insufficient, never zero.
	cb := find(t, sigs, "scout.churn_per_change@1", "b", w)
	if cb.Missing != domain.SignalInsufficient || cb.Value != nil {
		t.Errorf("scope b churn: %+v", cb)
	}
	if chb := find(t, sigs, "scout.changes@1", "b", w); chb.Value == nil || *chb.Value != 15 {
		t.Errorf("changes@1 for b should still be measured at 15: %+v", chb)
	}

	// PR-level measures are unmeasured without PR metadata.
	if pe := find(t, sigs, "scout.pr_effort@1", "a", w); pe.Missing != domain.SignalUnmeasured {
		t.Errorf("pr_effort = %q", pe.Missing)
	}
	d := find(t, sigs, "scout.dispersion@1", "a", w)
	near(t, "dispersion", d.Value, 27.1/15.5)
	if d.Detail["pr_part"] != "unmeasured" {
		t.Errorf("dispersion pr part: %v", d.Detail)
	}
}

func TestDeterminismAndPopulationDigest(t *testing.T) {
	f, end := buildMain(t)
	commits, _ := Load(context.Background(), f.dir)
	opts := Options{Scopes: scopesAB(), End: end, Snapshot: "s"}
	a := Compute(commits, opts)
	b := Compute(commits, opts)
	if len(a) != len(b) {
		t.Fatal("length differs")
	}
	for i := range a {
		if a[i].Key() != b[i].Key() || a[i].PopulationDigest != b[i].PopulationDigest {
			t.Fatalf("signal %d differs between runs", i)
		}
	}
	w := winLabel(200)
	pa := find(t, a, "scout.churn_per_change@1", "a", w).PopulationDigest
	pb := find(t, a, "scout.churn_per_change@1", "b", w).PopulationDigest
	if pa == pb {
		t.Error("different populations must have different digests")
	}
	// A bot commit in the window does not change the population.
	f.commit(windowStart(200).Add(100*time.Hour), "x[bot]", "x@example.com", map[string]string{"a/f.go": lines(99)})
	commits2, _ := Load(context.Background(), f.dir)
	c := Compute(commits2, opts)
	if find(t, c, "scout.churn_per_change@1", "a", w).PopulationDigest != pa {
		t.Error("population digest changed after adding an excluded bot commit")
	}
	// Configured bots change the config digest.
	o2 := opts
	o2.Bots = []string{"alice@example.com"}
	if Compute(commits, o2)[0].ConfigDigest == a[0].ConfigDigest {
		t.Error("config digest must reflect history.bots")
	}
}

func TestChangeFilters(t *testing.T) {
	f := newFixture(t)
	base := windowStart(300)
	at := func(i int) time.Time { return base.Add(time.Duration(i) * time.Hour) }
	n := 0
	add := func(name, email string, files map[string]string) {
		f.commit(at(n), name, email, files)
		n++
	}
	for i := 0; i < 20; i++ { // 20 ordinary changes
		add("dev", "dev@example.com", map[string]string{"a/f.go": lines(i + 1)})
	}
	// Name-based bot rules.
	add("renovate-bot", "r@example.com", map[string]string{"a/f.go": lines(30)})   // substring, NOT a bot
	add("dependabot", "d@example.com", map[string]string{"a/f.go": lines(31)})     // substring, NOT a bot
	add("renovate[bot]", "r2@example.com", map[string]string{"a/f.go": lines(32)}) // suffix: bot
	add("ci", "ci@example.com", map[string]string{"a/f.go": lines(33)})            // configured by e-mail: bot
	add("Release Robot", "x@example.com", map[string]string{"a/f.go": lines(34)})  // configured by name: bot
	add("[bot]prefix", "p@example.com", map[string]string{"a/f.go": lines(35)})    // prefix only, NOT a bot
	add("dev", "dev@example.com", map[string]string{"docs/readme.md": "x\n"})      // no eligible file
	big := map[string]string{}
	for i := 0; i < 51; i++ {
		big[fmt.Sprintf("a/big%d.go", i)] = "x\n"
	}
	add("dev", "dev@example.com", big) // 51 files: bulk
	f.git(nil, "checkout", "-q", "-b", "side")
	add("dev", "dev@example.com", map[string]string{"a/side.go": "s\n"})
	f.git(nil, "checkout", "-q", "main")
	add("dev", "dev@example.com", map[string]string{"a/main.go": "m\n"})
	f.git([]string{"GIT_AUTHOR_NAME=dev", "GIT_AUTHOR_EMAIL=dev@example.com", "GIT_COMMITTER_NAME=dev", "GIT_COMMITTER_EMAIL=dev@example.com",
		"GIT_AUTHOR_DATE=" + at(n).Format(time.RFC3339), "GIT_COMMITTER_DATE=" + at(n).Format(time.RFC3339)},
		"merge", "-q", "--no-ff", "-m", "merge", "side") // merge commit
	commits, err := Load(context.Background(), f.dir)
	if err != nil {
		t.Fatal(err)
	}
	opts := Options{
		Scopes:   []Scope{PrefixScope(domain.SignalScopeRepository, "", "")},
		Eligible: func(p string) bool { return strings.HasSuffix(p, ".go") },
		Bots:     []string{"ci@example.com", "Release Robot"},
		End:      at(n), Snapshot: "s",
	}
	changes, st := buildChanges(commits, opts)
	// 20 ordinary + renovate-bot + dependabot + [bot]prefix + side + main = 25
	if len(changes) != 25 {
		t.Errorf("changes = %d, want 25 (stats %+v)", len(changes), st)
	}
	if st.Bots != 3 || st.Bulk != 1 || st.NoEligible != 1 || st.Merges != 1 {
		t.Errorf("stats = %+v, want 3 bots, 1 bulk, 1 no-eligible, 1 merge", st)
	}
	if BotReason(Commit{AuthorName: "x[bot]"}, nil) != BotReasonSuffix || BotReason(Commit{AuthorName: "dev", AuthorEmail: "ci@example.com"}, opts.Bots) != BotReasonConfig {
		t.Error("bot_reason mismatch")
	}
	sigs := Compute(commits, opts)
	ch := find(t, sigs, "scout.changes@1", "", winLabel(300))
	near(t, "changes", ch.Value, 25)
	if ch.Detail["dropped_bots"] != 3 {
		t.Errorf("detail %v", ch.Detail)
	}
}

func TestWindowBucketingAndInsufficient(t *testing.T) {
	f := newFixture(t)
	k := 250
	// 4 changes in the last two days of window k, 3 in the first days of k+1.
	for i := 0; i < 4; i++ {
		f.commit(windowStart(k+1).Add(-time.Duration(48-i)*time.Hour), "dev", "dev@example.com", map[string]string{"a/f.go": lines(i + 1)})
	}
	for i := 0; i < 3; i++ {
		f.commit(windowStart(k+1).Add(time.Duration(i)*time.Hour), "dev", "dev@example.com", map[string]string{"a/f.go": lines(10 + i)})
	}
	commits, _ := Load(context.Background(), f.dir)
	end := windowStart(k + 1).Add(24 * time.Hour)
	sigs := Compute(commits, Options{Scopes: scopesAB()[:1], End: end, Snapshot: "s"})
	near(t, "window k", find(t, sigs, "scout.changes@1", "a", winLabel(k)).Value, 4)
	near(t, "window k+1", find(t, sigs, "scout.changes@1", "a", winLabel(k+1)).Value, 3)
	near(t, "trailing", find(t, sigs, "scout.changes@1", "a", "trailing-365d").Value, 7)
	for _, s := range sigs {
		if s.ID() == "scout.churn_per_change@1" && s.Window != "trailing-365d" && s.Missing != domain.SignalInsufficient {
			t.Errorf("%s: want insufficient with <20 changes, got %q", s.Window, s.Missing)
		}
	}
	// Boundary: the commit at exactly windowStart(k+1) belongs to k+1, which is immature.
	fixed, _ := Windows(windowStart(k), end, 0)
	if len(fixed) != 2 || fixed[1].Start != windowStart(k+1) || !fixed[1].Immature || fixed[0].Immature {
		t.Errorf("windows = %+v", fixed)
	}
	if im := find(t, sigs, "scout.changes@1", "a", winLabel(k+1)).Detail["immature"]; im != true {
		t.Errorf("current quarter must be flagged immature, got %v", im)
	}
	if im := find(t, sigs, "scout.changes@1", "a", winLabel(k)).Detail["immature"]; im != false {
		t.Errorf("complete quarter flagged immature: %v", im)
	}
}

func TestReworkBlameOrigin(t *testing.T) {
	f := newFixture(t)
	k := 260
	// Commit 0 sits in the previous window; commits 1..29 rewrite the single line
	// in window k, each removing the line the previous commit added.
	f.commit(windowStart(k).Add(-time.Hour), "dev", "dev@example.com", map[string]string{"a/f.go": "v0\n"})
	for i := 1; i < 30; i++ {
		f.commit(windowStart(k).Add(time.Duration(i)*24*time.Hour), "dev", "dev@example.com", map[string]string{"a/f.go": fmt.Sprintf("v%d\n", i)})
	}
	commits, _ := Load(context.Background(), f.dir)
	opts := Options{Scopes: scopesAB()[:1], End: windowStart(k + 1), Snapshot: "s"}
	sigs, err := Rework(context.Background(), f.dir, commits, opts)
	if err != nil {
		t.Fatal(err)
	}
	r := find(t, sigs, "scout.rework@1", "a", winLabel(k))
	// 29 removed lines; origins commit 0..28; commit 0 is outside the window.
	near(t, "rework num", r.Numerator, 28)
	near(t, "rework den", r.Denominator, 29)
	near(t, "rework", r.Value, 28.0/29.0)
	if err := r.Validate(); err != nil {
		t.Error(err)
	}
	// Previous window has a single creation commit and no removed lines: unmeasured, not 0.
	if p := find(t, sigs, "scout.rework@1", "a", winLabel(k-1)); p.Missing != domain.SignalUnmeasured || p.Value != nil {
		t.Errorf("previous window: %+v", p)
	}
}

func TestParseRemoved(t *testing.T) {
	diff := "--- a/x.go\n+++ b/x.go\n@@ -3,2 +3,0 @@\n-a\n-b\n@@ -10 +8,2 @@\n-c\n+d\n+e\n--- /dev/null\n+++ b/new.go\n@@ -0,0 +1,3 @@\n+1\n"
	got := parseRemoved(diff)
	if len(got) != 1 || len(got["x.go"]) != 2 || got["x.go"][0] != (removedRange{3, 2}) || got["x.go"][1] != (removedRange{10, 1}) {
		t.Errorf("parseRemoved = %v", got)
	}
}

func TestActiveHoursHandComputed(t *testing.T) {
	h := 3600.0
	cases := []struct {
		in   []float64
		want float64
	}{
		{nil, 0.5},
		{[]float64{100}, 0.5},
		{[]float64{0, 1 * h, 6 * h}, 0.5 + 1 + 0.5},
		{[]float64{6 * h, 0, 1 * h}, 0.5 + 1 + 0.5}, // order-independent
		{[]float64{0, 4 * h}, 4.5},                  // gap == 4 h counts as the same session
	}
	for _, c := range cases {
		if got := ActiveHours(c.in); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("ActiveHours(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestPRLevelMeasures(t *testing.T) {
	w := windowStart(270)
	mk := func(i int, mod func(*PR)) PR {
		m := w.Add(time.Duration(i) * 24 * time.Hour)
		p := PR{Number: i, AuthorLogin: "dev", Created: m.Add(-10 * time.Hour), Merged: m,
			Files: []FileChange{{Path: "a/f.go", Added: 3, Removed: 1}, {Path: "b/g.go", Added: 2}}, TotalFiles: 2,
			CommitTimes: []time.Time{m.Add(-3 * time.Hour), m.Add(-2 * time.Hour), m.Add(-time.Hour / 2)}, TotalCommits: 3, Reviews: 1, Comments: 1}
		if mod != nil {
			mod(&p)
		}
		return p
	}
	var prs []PR
	for i := 1; i <= 5; i++ {
		prs = append(prs, mk(i, nil)) // 3 commits: gaps 1 h and 1.5 h => 0.5+1+1.5 = 3.0 h
	}
	prs = append(prs,
		mk(6, func(p *PR) { p.AuthorType = "Bot" }),
		mk(7, func(p *PR) { p.AuthorLogin = "listed" }),
		mk(8, func(p *PR) { p.TotalFiles = 51 }),
		mk(9, func(p *PR) { p.TotalCommits = 4 }),
		mk(10, func(p *PR) { p.CommitTimes, p.TotalCommits = nil, 0 }),
	)
	end := w.Add(30 * 24 * time.Hour)
	sc := Scope{Kind: domain.SignalScopeDirectory, Anchor: "a", Contains: func(p string) bool { return strings.HasPrefix(p, "a/") }}
	one := newFixture(t)
	one.commit(w.Add(time.Hour), "dev", "d@example.com", map[string]string{"a/f.go": "x\n"})
	commits, _ := Load(context.Background(), one.dir)
	sigs := Compute(commits, Options{Scopes: []Scope{sc}, End: end, Snapshot: "s", PRs: prs, BotLogins: []string{"listed"}})
	win := winLabel(270)
	pe := find(t, sigs, "scout.pr_effort@1", "a", win)
	if err := pe.Validate(); err != nil {
		t.Fatal(err)
	}
	near(t, "pr_effort p50", pe.Value, 3.0)
	near(t, "pr_effort den", pe.Denominator, 5)
	near(t, "queue p50", func() *float64 { v := pe.Detail["queue_hours_p50"].(float64); return &v }(), 10)
	ex := pe.Detail["excluded"].(map[string]any)
	if ex["bot:api_bot_type"] != 1 || ex["bot:login_list"] != 1 || ex["incomplete:over_cap"] != 1 || ex["incomplete:list"] != 2 {
		t.Errorf("excluded = %v", ex)
	}
	bs := find(t, sigs, "scout.batch_size@1", "a", win)
	near(t, "batch p50 lines", bs.Value, 6)
	// Fewer than 5 linked PRs: unmeasured.
	few := Compute(commits, Options{Scopes: []Scope{sc}, End: end, Snapshot: "s", PRs: prs[:4]})
	if s := find(t, few, "scout.pr_effort@1", "a", win); s.Missing != domain.SignalUnmeasured {
		t.Errorf("4 PRs: %q", s.Missing)
	}
}
