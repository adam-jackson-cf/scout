package history

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
)

// Tunables fixed by design §8.3.
const (
	MaxChangeFiles   = 50
	MinChanges       = 20 // below this a distributional measure is `insufficient`
	FollowUpDays     = 7
	TrailingDays     = 365
	MinorShare       = 0.05
	ToolName         = "scout/history"
	botSuffix        = "[bot]"
	BotReasonSuffix  = "app_suffix"
	BotReasonConfig  = "configured"
	defaultSalt      = "scout-history"
	pseudonymHexSize = 12
)

// Scope is a measurement scope: an anchor plus a membership predicate over
// repo-relative paths (a feature's file set, or a directory scope).
type Scope struct {
	Kind     string
	Anchor   string
	Contains func(path string) bool
}

// PrefixScope is a directory scope; an empty prefix contains every path.
func PrefixScope(kind, anchor, prefix string) Scope {
	p := strings.TrimSuffix(prefix, "/")
	return Scope{Kind: kind, Anchor: anchor, Contains: func(path string) bool {
		return p == "" || path == p || strings.HasPrefix(path, p+"/")
	}}
}

// FileSetScope is a scope over an explicit file set.
func FileSetScope(kind, anchor string, files []string) Scope {
	set := make(map[string]bool, len(files))
	for _, f := range files {
		set[f] = true
	}
	return Scope{Kind: kind, Anchor: anchor, Contains: func(path string) bool { return set[path] }}
}

// FileMetric is per-file complexity input for scout.hotspot@1 (from lizard).
type FileMetric struct {
	CC   float64 // Σ function cyclomatic complexity
	NLOC int
}

// Options configures a run. Zero values take the design defaults.
type Options struct {
	Scopes     []Scope
	Eligible   func(path string) bool // nil: every path is eligible
	Bots       []string               // history.bots: exact author name or e-mail
	WindowDays int                    // 0 = UTC calendar quarters (default); N>0 = N-day windows anchored at 1 January of each year, cut at the year end
	End        time.Time              // observation end; zero = latest commit time
	Snapshot   string                 // snapshot id recorded on every signal
	Salt       string                 // pseudonymisation salt for ownership keys
	Files      map[string]FileMetric  // per-file CC/NLOC; nil => hotspot unmeasured
	PRs        []PR                   // PR metadata; empty => PR-level measures unmeasured
	BotLogins  []string               // PR author logins treated as bots (E1b common.BOT_LOGINS)
	PRSource   *PRSource              // provenance of PRs; folded into the config digest and every PR-dependent signal's detail
	ToolInfo   string                 // appended to tool digest (e.g. git version)
}

// window returns the configured window length in days; 0 means calendar quarters.
func (o Options) window() int {
	if o.WindowDays < 0 {
		return 0
	}
	return o.WindowDays
}

func (o Options) salt() string {
	if o.Salt == "" {
		return defaultSalt
	}
	return o.Salt
}

func (o Options) eligible(p string) bool { return o.Eligible == nil || o.Eligible(p) }

// BotReason returns the bot_reason for a commit author, or "".
func BotReason(c Commit, bots []string) string {
	if strings.HasSuffix(c.AuthorName, botSuffix) {
		return BotReasonSuffix
	}
	for _, b := range bots {
		if b != "" && (c.AuthorName == b || c.AuthorEmail == b) {
			return BotReasonConfig
		}
	}
	return ""
}

// change is a commit that qualified as a change repo-wide.
type change struct {
	hash   string
	t      time.Time
	author string       // pseudonym
	files  []FileChange // eligible files only
}

// Stats counts what the change filter dropped (reported on changes@1 detail).
type Stats struct {
	Merges, Bots, Bulk, NoEligible int
}

func buildChanges(commits []Commit, o Options) ([]change, Stats) {
	var out []change
	var st Stats
	for _, c := range commits {
		switch {
		case c.Parents > 1:
			st.Merges++
			continue
		case BotReason(c, o.Bots) != "":
			st.Bots++
			continue
		case len(c.Files) > MaxChangeFiles:
			st.Bulk++
			continue
		}
		var el []FileChange
		for _, f := range c.Files {
			if o.eligible(f.Path) {
				el = append(el, f)
			}
		}
		if len(el) == 0 {
			st.NoEligible++
			continue
		}
		out = append(out, change{hash: c.Hash, t: c.Time, author: pseudonym(o.salt(), c), files: el})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].t.Equal(out[j].t) {
			return out[i].t.Before(out[j].t)
		}
		return out[i].hash < out[j].hash
	})
	return out, st
}

func pseudonym(salt string, c Commit) string {
	key := strings.ToLower(strings.TrimSpace(c.AuthorEmail))
	if key == "" {
		key = strings.ToLower(strings.TrimSpace(c.AuthorName))
	}
	h := sha256.Sum256([]byte(salt + "\x00" + key))
	return hex.EncodeToString(h[:])[:pseudonymHexSize]
}

// Window is a half-open [Start, End) interval.
type Window struct {
	Label      string
	Start, End time.Time
	Immature   bool // ends after the observation end: right-censored
}

// Windows returns the fixed windows covering [first, end] plus the
// trailing-365-day view ending at end.
//
// days == 0 (the default) gives UTC calendar quarters labelled YYYY-Qn. A
// non-zero days gives days-long windows anchored at 1 January of each year:
// window k of a year starts at Jan 1 + k*days, the last one is cut at the next
// 1 January, and the next year restarts at k=1; labels are YYYY-<days>d-<k>.
// A window whose end lies after the observation end (+1 s) is Immature:
// right-censored, to be shown but never compared with complete windows (§8.6).
func Windows(first, end time.Time, days int) (fixed []Window, trailing Window) {
	first, end = first.UTC(), end.UTC()
	for y := first.Year(); y <= end.Year(); y++ {
		jan1 := time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
		next := time.Date(y+1, 1, 1, 0, 0, 0, 0, time.UTC)
		add := func(label string, s, e time.Time) {
			if e.After(first) && !s.After(end) {
				fixed = append(fixed, Window{Label: label, Start: s, End: e, Immature: e.After(end.Add(time.Second))})
			}
		}
		if days <= 0 {
			for q := 0; q < 4; q++ {
				add(fmt.Sprintf("%d-Q%d", y, q+1), time.Date(y, time.Month(3*q+1), 1, 0, 0, 0, 0, time.UTC), time.Date(y, time.Month(3*q+4), 1, 0, 0, 0, 0, time.UTC))
			}
			continue
		}
		for k := 0; jan1.AddDate(0, 0, k*days).Before(next); k++ {
			s := jan1.AddDate(0, 0, k*days)
			e := jan1.AddDate(0, 0, (k+1)*days)
			if e.After(next) {
				e = next
			}
			add(fmt.Sprintf("%d-%dd-%d", y, days, k+1), s, e)
		}
	}
	te := end.Add(time.Second) // include a commit exactly at end
	trailing = Window{Label: fmt.Sprintf("trailing-%dd", TrailingDays), Start: te.AddDate(0, 0, -TrailingDays), End: te}
	return fixed, trailing
}

func (w Window) has(t time.Time) bool { return !t.Before(w.Start) && t.Before(w.End) }

type cfg struct {
	WindowDays int      `json:"window_days"`
	MaxFiles   int      `json:"max_files"`
	MinChanges int      `json:"min_changes"`
	FollowUp   int      `json:"follow_up_days"`
	Bots       []string `json:"bots"`
	BotLogins  []string `json:"bot_logins"`
}

func sha(s string) string { return domain.Digest(s) }

func (o Options) configDigest() string {
	b := append([]string(nil), o.Bots...)
	l := append([]string(nil), o.BotLogins...)
	sort.Strings(b)
	sort.Strings(l)
	j, _ := json.Marshal(cfg{o.window(), MaxChangeFiles, MinChanges, FollowUpDays, b, l})
	if o.PRSource != nil {
		return sha(string(j) + "|pr:" + o.PRSource.Digest)
	}
	return sha(string(j))
}

func (o Options) toolDigest(metric string) string {
	return sha(ToolName + "|" + metric + "|" + o.ToolInfo)
}

func popDigest(metric string, ids []string) string {
	s := append([]string(nil), ids...)
	sort.Strings(s)
	return sha(metric + "\n" + strings.Join(s, "\n"))
}

// Compute returns all commit-level (and, given PR metadata, PR-level) signals
// for every scope, for each fixed window and the trailing view. It does not
// run blame; call Rework separately (it needs the repo).
func Compute(commits []Commit, o Options) []domain.Signal {
	changes, st := buildChanges(commits, o)
	end := o.End
	if end.IsZero() && len(commits) > 0 {
		end = commits[len(commits)-1].Time
	}
	if len(commits) == 0 {
		return nil
	}
	first := commits[0].Time
	fixed, trailing := Windows(first, end, o.window())
	wins := append(append([]Window(nil), fixed...), trailing)

	var sigs []domain.Signal
	for _, sc := range o.Scopes {
		for _, w := range wins {
			ctx := newScopeWindow(sc, w, changes, o, end)
			sigs = append(sigs, ctx.changesSignal(st), ctx.churn(), ctx.amplification(),
				ctx.followUp(changes), ctx.ownership(), ctx.hotspot(), ctx.dispersion())
			sigs = append(sigs, ctx.prSignals()...)
		}
	}
	return sigs
}

// scopeWindow is the change population of one scope in one window.
type scopeWindow struct {
	sc    Scope
	w     Window
	o     Options
	end   time.Time
	items []swItem
	allIn []change // every repo-wide change in window (for amplification scope counts)
}

type swItem struct {
	c       change
	inScope []FileChange // eligible files inside the scope
}

func newScopeWindow(sc Scope, w Window, changes []change, o Options, end time.Time) *scopeWindow {
	s := &scopeWindow{sc: sc, w: w, o: o, end: end}
	for _, c := range changes {
		if !w.has(c.t) {
			continue
		}
		var in []FileChange
		for _, f := range c.files {
			if sc.Contains(f.Path) {
				in = append(in, f)
			}
		}
		if len(in) > 0 {
			s.items = append(s.items, swItem{c, in})
		}
	}
	return s
}

func (s *scopeWindow) ids() []string {
	ids := make([]string, len(s.items))
	for i, it := range s.items {
		ids[i] = it.c.hash
	}
	return ids
}

func (s *scopeWindow) base(metric string) domain.Signal {
	return mkSignal(s.sc, s.w, s.o, metric, s.ids())
}

func (s *scopeWindow) insufficient(sig domain.Signal, n int) (domain.Signal, bool) {
	if n < MinChanges {
		return markMissing(sig, "insufficient", fmt.Sprintf("%d changes < %d", n, MinChanges)), true
	}
	return sig, false
}

func (s *scopeWindow) changesSignal(st Stats) domain.Signal {
	sig := s.base("scout.changes@1")
	authors := map[string]bool{}
	for _, it := range s.items {
		authors[it.c.author] = true
	}
	n := float64(len(s.items))
	sig = setValue(sig, n, n, 0)
	sig = withDetail(sig, map[string]any{"authors": len(authors), "dropped_merges": st.Merges, "dropped_bots": st.Bots, "dropped_bulk": st.Bulk, "dropped_no_eligible": st.NoEligible})
	return sig
}

func (s *scopeWindow) churnValues() []float64 {
	v := make([]float64, len(s.items))
	for i, it := range s.items {
		for _, f := range it.inScope {
			v[i] += float64(f.Added + f.Removed)
		}
	}
	return v
}

func (s *scopeWindow) churn() domain.Signal {
	sig := s.base("scout.churn_per_change@1")
	if sig, bad := s.insufficient(sig, len(s.items)); bad {
		return sig
	}
	v := s.churnValues()
	total := 0.0
	for _, x := range v {
		total += x
	}
	sig = setValue(sig, percentile(v, 50), total, float64(len(v)))
	return withDetail(sig, map[string]any{"p50": percentile(v, 50), "p90": percentile(v, 90)})
}

func (s *scopeWindow) amplification() domain.Signal {
	sig := s.base("scout.amplification@1")
	if sig, bad := s.insufficient(sig, len(s.items)); bad {
		return sig
	}
	files := make([]float64, len(s.items))
	scopes := make([]float64, len(s.items))
	total := 0.0
	for i, it := range s.items {
		files[i] = float64(len(it.c.files))
		total += files[i]
		for _, o := range s.o.Scopes {
			for _, f := range it.c.files {
				if o.Contains(f.Path) {
					scopes[i]++
					break
				}
			}
		}
	}
	sig = setValue(sig, percentile(files, 50), total, float64(len(files)))
	return withDetail(sig, map[string]any{"p50_files": percentile(files, 50), "p50_scopes": percentile(scopes, 50)})
}

func (s *scopeWindow) followUp(all []change) domain.Signal {
	sig := s.base("scout.follow_up_rate@1")
	perFile := map[string][]time.Time{}
	for _, c := range all {
		for _, f := range c.files {
			perFile[f.Path] = append(perFile[f.Path], c.t)
		}
	}
	horizon := time.Duration(FollowUpDays) * 24 * time.Hour
	num, den := 0, 0
	for _, it := range s.items {
		if it.c.t.Add(horizon).After(s.end) {
			continue // incomplete 7-day observation
		}
		den++
		followed := false
	outer:
		for _, f := range it.inScope {
			for _, t := range perFile[f.Path] {
				if t.After(it.c.t) && !t.After(it.c.t.Add(horizon)) {
					followed = true
					break outer
				}
			}
		}
		if followed {
			num++
		}
	}
	if den < MinChanges {
		return markMissing(withDetail(sig, map[string]any{"followed": num, "observed": den}), "insufficient", fmt.Sprintf("%d changes with complete %d-day observation < %d", den, FollowUpDays, MinChanges))
	}
	return setValue(sig, float64(num)/float64(den), float64(num), float64(den))
}

func (s *scopeWindow) ownership() domain.Signal {
	sig := s.base("scout.ownership@1")
	if sig, bad := s.insufficient(sig, len(s.items)); bad {
		return sig
	}
	cnt := map[string]int{}
	for _, it := range s.items {
		cnt[it.c.author]++
	}
	keys := make([]string, 0, len(cnt))
	for k := range cnt {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if cnt[keys[i]] != cnt[keys[j]] {
			return cnt[keys[i]] > cnt[keys[j]]
		}
		return keys[i] < keys[j]
	})
	n := float64(len(s.items))
	minor := 0
	for _, k := range keys {
		if float64(cnt[k])/n < MinorShare {
			minor++
		}
	}
	sig = setValue(sig, float64(cnt[keys[0]])/n, float64(cnt[keys[0]]), n)
	return withDetail(sig, map[string]any{"top_author_key": keys[0], "authors": len(keys), "minor_contributors": minor})
}

func (s *scopeWindow) hotspot() domain.Signal {
	sig := s.base("scout.hotspot@1")
	if s.o.Files == nil {
		return markMissing(sig, "unmeasured", "no per-file complexity supplied")
	}
	touches := map[string]int{}
	for _, it := range s.items {
		for _, f := range it.inScope {
			touches[f.Path]++
		}
	}
	nloc, cc := 0, 0.0
	for p, m := range s.o.Files {
		if s.sc.Contains(p) && s.o.eligible(p) {
			nloc += m.NLOC
		}
	}
	if nloc == 0 {
		return markMissing(sig, "unmeasured", "scope has no measured NLOC")
	}
	paths := make([]string, 0, len(touches))
	for p := range touches {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	missing := 0
	for _, p := range paths {
		m, ok := s.o.Files[p]
		if !ok {
			missing++
			continue
		}
		cc += m.CC * math.Log1p(float64(touches[p]))
	}
	sig = setValue(sig, cc/float64(nloc), cc, float64(nloc))
	return withDetail(sig, map[string]any{"touched_files": len(paths), "touched_files_without_cc": missing})
}

// dispersion: git part is p90/p50 of churn; the PR part rides in prSignals'
// scout.pr_effort@1 detail and in this signal's Detail when measured.
func (s *scopeWindow) dispersion() domain.Signal {
	sig := s.base("scout.dispersion@1")
	if sig, bad := s.insufficient(sig, len(s.items)); bad {
		return sig
	}
	v := s.churnValues()
	p50, p90 := percentile(v, 50), percentile(v, 90)
	if p50 <= 0 {
		return markMissing(sig, "unmeasured", "churn p50 is 0")
	}
	sig = setValue(sig, p90/p50, p90, p50)
	d := map[string]any{"p50": p50, "p90": p90}
	hours := s.prHours()
	if len(hours) < MinPRs {
		d["pr_part"] = "unmeasured"
		d["pr_part_reason"] = prUnmeasuredReason(s.o, len(hours))
	} else if h50 := percentile(hours, 50); h50 > 0 {
		d["pr_p90_over_p50"] = percentile(hours, 90) / h50
		d["pr_provenance"] = s.o.PRSource.detail()
	} else {
		d["pr_part"] = "unmeasured"
		d["pr_part_reason"] = "PR active-hours p50 is 0"
	}
	return withDetail(sig, d)
}
