package history

import (
	"fmt"
	"time"

	"scout/internal/domain"
)

// PR-level thresholds (design §8.3).
const (
	MaxPRFiles   = 50
	MaxPRCommits = 100
	MinPRs       = 5 // linked PRs required for pr_effort / batch_size
)

// PR is merged-PR metadata (GraphQL / pull_request payload). Files holds the
// returned file list; TotalFiles/TotalCommits are the API totals, used to
// detect an incomplete list.
type PR struct {
	Number       int
	AuthorLogin  string
	AuthorType   string // "Bot" for GitHub App identities
	Created      time.Time
	Merged       time.Time
	Files        []FileChange // path, additions, deletions
	TotalFiles   int
	CommitTimes  []time.Time // author timestamps of the PR's own commits
	TotalCommits int
	Reviews      int // review submissions
	Comments     int
}

// PRBotReason returns api_bot_type, login_list or "".
func PRBotReason(p PR, logins []string) string {
	if p.AuthorType == "Bot" {
		return "api_bot_type"
	}
	for _, l := range logins {
		if p.AuthorLogin == l {
			return "login_list"
		}
	}
	return ""
}

// PRLinkage classifies a PR: "" = linked, else the exclusion reason.
func PRLinkage(p PR, o Options) string {
	if r := PRBotReason(p, o.BotLogins); r != "" {
		return "bot:" + r
	}
	switch {
	case p.TotalFiles > MaxPRFiles || p.TotalCommits > MaxPRCommits:
		return "incomplete:over_cap"
	case len(p.Files) < p.TotalFiles || len(p.CommitTimes) < p.TotalCommits || len(p.CommitTimes) == 0:
		return "incomplete:list"
	}
	return ""
}

// prInScope returns the PRs merged in the window that touch the scope, split
// into linked ones and per-reason exclusion counts.
func (s *scopeWindow) prs() (linked []PR, excluded map[string]int) {
	excluded = map[string]int{}
	for _, p := range s.o.PRs {
		if !s.w.has(p.Merged) {
			continue
		}
		touches := false
		for _, f := range p.Files {
			if s.o.eligible(f.Path) && s.sc.Contains(f.Path) {
				touches = true
				break
			}
		}
		if !touches {
			continue
		}
		if r := PRLinkage(p, s.o); r != "" {
			excluded[r]++
			continue
		}
		linked = append(linked, p)
	}
	return linked, excluded
}

func prActiveHours(p PR) float64 {
	ts := make([]float64, len(p.CommitTimes))
	for i, t := range p.CommitTimes {
		ts[i] = float64(t.Unix())
	}
	return ActiveHours(ts)
}

func (s *scopeWindow) prHours() []float64 {
	linked, _ := s.prs()
	h := make([]float64, len(linked))
	for i, p := range linked {
		h[i] = prActiveHours(p)
	}
	return h
}

func (s *scopeWindow) prSignals() []domain.Signal {
	linked, excl := s.prs()
	pop := make([]string, len(linked))
	for i, p := range linked {
		pop[i] = fmt.Sprintf("pr#%d", p.Number)
	}
	excludedDetail := map[string]any{}
	for k, v := range excl {
		excludedDetail[k] = v
	}
	mk := func(metric string) domain.Signal {
		return mkSignal(s.sc, s.w, s.o, metric, pop)
	}
	effort, batch := mk("scout.pr_effort@1"), mk("scout.batch_size@1")
	if len(s.o.PRs) == 0 {
		why := "no PR metadata supplied (local history lacks PR commits)"
		return []domain.Signal{markMissing(effort, "unmeasured", why), markMissing(batch, "unmeasured", why)}
	}
	if len(linked) < MinPRs {
		why := fmt.Sprintf("%d linked PRs < %d", len(linked), MinPRs)
		effort = withDetail(markMissing(effort, "unmeasured", why), map[string]any{"excluded": excludedDetail})
		batch = withDetail(markMissing(batch, "unmeasured", why), map[string]any{"excluded": excludedDetail})
		return []domain.Signal{effort, batch}
	}
	n := float64(len(linked))
	var hours, commits, queue, lines, files, rv []float64
	single, totalH := 0, 0.0
	for _, p := range linked {
		h := prActiveHours(p)
		hours = append(hours, h)
		totalH += h
		commits = append(commits, float64(len(p.CommitTimes)))
		if len(p.CommitTimes) == 1 {
			single++
		}
		queue = append(queue, p.Merged.Sub(p.Created).Hours())
		rv = append(rv, float64(p.Reviews+p.Comments))
		l := 0
		for _, f := range p.Files {
			l += f.Added + f.Removed
		}
		lines = append(lines, float64(l))
		files = append(files, float64(len(p.Files)))
	}
	effort = setValue(effort, percentile(hours, 50), totalH, n)
	effort = withDetail(effort, map[string]any{
		"p50_hours": percentile(hours, 50), "p90_hours": percentile(hours, 90),
		"p50_commits": percentile(commits, 50), "single_commit_prs": single,
		"p50_review_events": percentile(rv, 50),
		"queue_hours_p50":   percentile(queue, 50), "queue_hours_p90": percentile(queue, 90),
		"queue_is_not_effort": true, "excluded": excludedDetail,
	})
	var totalL float64
	for _, x := range lines {
		totalL += x
	}
	batch = setValue(batch, percentile(lines, 50), totalL, n)
	batch = withDetail(batch, map[string]any{
		"p50_lines": percentile(lines, 50), "p90_lines": percentile(lines, 90),
		"p50_files": percentile(files, 50), "p90_files": percentile(files, 90), "excluded": excludedDetail,
	})
	return []domain.Signal{effort, batch}
}
