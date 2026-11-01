// Package association is the deterministic front end of merge-time area
// association (design §4, revision 2.8): for each first-parent merge of a push
// it finds the residual files and groups them into clusters. It calls no model
// and does not touch state; ownership comes in through the Ownership interface
// and results are pure values. The rules port the measured pilot front end
// (evaluation/x2d-postrun-tuning/axum/pilot/frontend.py `owned_classify`,
// `clusters`, variant-stem and normalised-directory functions;
// build_units.py:247-280 for the per-merge cap).
package association

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
)

// PerMergeCap is the number of clusters decided per merge (x2d.PR_CAP); the
// rest are unresolved (pr_cap).
const PerMergeCap = 10

var (
	variantRE = regexp.MustCompile(`(?:_an)?(?:_py3\d+)?(?:_pv[12])?$`)
	numRE     = regexp.MustCompile(`_\d{1,2}$`)
)

// SiblingKey identifies variant siblings: same normalised directory, variant
// stem and extension.
type SiblingKey struct{ Dir, Stem, Ext string }

func splitName(p string) (dir, stem, ext string) {
	dir, f := "", p
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		dir, f = p[:i], p[i+1:]
	}
	if i := strings.LastIndexByte(f, '.'); i >= 0 {
		return dir, f[:i], f[i+1:]
	}
	return dir, f, ""
}

// VariantStem strips variant suffixes: tutorial003_an_py39 -> tutorial003,
// tutorial003_01 -> tutorial003.
func VariantStem(stem string) string {
	s := stem
	for {
		s2 := numRE.ReplaceAllString(variantRE.ReplaceAllString(s, ""), "")
		if s2 == s || s2 == "" {
			return s
		}
		s = s2
	}
}

// NormDir strips variant suffixes from every directory segment.
func NormDir(d string) string {
	segs := strings.Split(d, "/")
	for i, seg := range segs {
		if n := variantRE.ReplaceAllString(seg, ""); n != "" {
			segs[i] = n
		}
	}
	return strings.Join(segs, "/")
}

// SiblingKeyOf is the variant-sibling key of a path.
func SiblingKeyOf(p string) SiblingKey {
	d, stem, ext := splitName(p)
	return SiblingKey{NormDir(d), VariantStem(stem), ext}
}

// Member is one residual file.
type Member struct {
	Path string
	Kind string // "added", "renamed" or "modified"
	Old  string // renamed: the old path
	// HeuristicOnly: only a heuristic feature owns the file (rule (f)).
	HeuristicOnly bool
	// Empty: an added or renamed file with no content (never set for modified files).
	Empty bool
}

// Cluster is a group of residual files sharing a normalised parent directory
// and variant stem.
type Cluster struct {
	ID      string
	Dir     string
	Stem    string
	Files   []string // sorted
	Kinds   []string // sorted, distinct
	Members []Member // in residual order
}

// ClusterID is the pilot's cluster id: sha256(commit NUL dir NUL stem)[:16].
func ClusterID(commit, dir, stem string) string {
	h := sha256.Sum256([]byte(commit + "\x00" + dir + "\x00" + stem))
	return hex.EncodeToString(h[:])[:16]
}

// BuildClusters groups residual members by (normalised dir, variant stem) in
// first-appearance order. Clusters made only of empty files are inert.
func BuildClusters(commit string, residual []Member) (live, inert []Cluster) {
	idx := map[[2]string]int{}
	var groups []Cluster
	for _, m := range residual {
		d, stem, _ := splitName(m.Path)
		k := [2]string{NormDir(d), VariantStem(stem)}
		i, ok := idx[k]
		if !ok {
			i = len(groups)
			idx[k] = i
			groups = append(groups, Cluster{Dir: k[0], Stem: k[1]})
		}
		groups[i].Members = append(groups[i].Members, m)
	}
	for _, c := range groups {
		kinds := map[string]bool{}
		empty := true
		for _, m := range c.Members {
			c.Files = append(c.Files, m.Path)
			kinds[m.Kind] = true
			empty = empty && m.Empty
		}
		sort.Strings(c.Files)
		for k := range kinds {
			c.Kinds = append(c.Kinds, k)
		}
		sort.Strings(c.Kinds)
		c.ID = ClusterID(commit, c.Dir, c.Stem)
		if empty {
			inert = append(inert, c)
		} else {
			live = append(live, c)
		}
	}
	return live, inert
}

func capKey(id string) string {
	h := sha256.Sum256([]byte("x2d-cap\x00" + id))
	return hex.EncodeToString(h[:])
}

// Cap splits live clusters into the at most PerMergeCap decided ones (smallest
// sha256("x2d-cap\0"+id) first) and the rest, which are unresolved (pr_cap).
func Cap(live []Cluster) (decided, capped []Cluster) {
	order := append([]Cluster(nil), live...)
	sort.Slice(order, func(i, j int) bool { return capKey(order[i].ID) < capKey(order[j].ID) })
	if len(order) <= PerMergeCap {
		return order, nil
	}
	return order[:PerMergeCap], order[PerMergeCap:]
}

var mediaExt = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true, ".ico": true, ".webp": true,
	".woff": true, ".woff2": true, ".ttf": true, ".eot": true, ".pdf": true, ".mp4": true, ".webm": true}

// excludedReason is the measured `excluded_reason` without its translation
// rule: "markdown", "media" or "" (eligible).
func excludedReason(p string) string {
	low := strings.ToLower(p)
	if strings.HasSuffix(low, ".md") {
		return "markdown"
	}
	if i := strings.LastIndexByte(low, '.'); i != -1 && mediaExt[low[i:]] {
		return "media"
	}
	return ""
}
