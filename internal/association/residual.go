package association

import (
	"sort"

	"scout/internal/domain"
)

// Kind is how a merge changed a file.
type Kind string

const (
	Added    Kind = "added"
	Modified Kind = "modified"
	Deleted  Kind = "deleted"
	Renamed  Kind = "renamed"
)

// Change is one file change of a merge against its first parent.
type Change struct {
	Kind    Kind
	Path    string // the new path (the old one for a deletion)
	OldPath string // renamed only
	Empty   bool   // the file has no content at the merge revision (added or renamed)
}

// Ownership is the read-only ownership view the rules need. An owner is an
// active area or an association record that assigns the file to an active area;
// a pending area takes the role of the measured pool.
type Ownership interface {
	// AreaOwners returns the area ids that own path ("owned" in the measured run).
	AreaOwners(path string) []string
	// PendingAreas returns the pending area ids whose file set holds path.
	PendingAreas(path string) []string
	// HeuristicOnly reports that a heuristic feature owns path; the rules apply
	// it only when no area owns the file.
	HeuristicOnly(path string) bool
	// SiblingFiles returns the owned files with this variant-sibling key.
	SiblingFiles(key SiblingKey) []string
}

// Inherited is a file that takes its owner by the rename or variant-sibling
// rule: an association record with outcome inherited, no model call.
type Inherited struct {
	Path  string
	Rule  string // "rename" or "variant"
	From  string // the old path, or the owned sibling
	Areas []string
}

// PendingMove is a renamed file whose old path is in a pending area and held by
// no owner: it takes the old path's place in that pending set. It gets no
// association record and is not decided.
type PendingMove struct {
	Old, New string
	Pending  []string
}

// Result is the classification of one merge's changes.
type Result struct {
	// Counts uses the pilot's names: modified_owned, modified_unowned (and
	// :pooled / :never_owned), deleted, renamed_inherit, renamed_pooled,
	// renamed_unowned, renamed_empty_trap, variant_inherit, variant_unowned,
	// residual_added, heuristic_only, excluded:<reason>.
	Counts       map[string]int
	Residual     []Member
	Inherited    []Inherited
	PendingMoves []PendingMove
	Deleted      []string // deleted paths: they leave ownership and pending sets
	CoverageGaps []string // modified, unowned, not heuristic-owned: counted, not decided
}

// Classify applies the measured residual rules (REQ-141, REQ-142) to one merge's
// changes. inTree reports whether a path exists in the parent tree (variant
// siblings must exist there).
func Classify(changes []Change, own Ownership, inTree func(string) bool) Result {
	c := classifier{r: Result{Counts: map[string]int{}}, own: own, inTree: inTree}
	for _, change := range changes {
		c.classify(change)
	}
	return c.r
}

type classifier struct {
	r      Result
	own    Ownership
	inTree func(string) bool
}

func (k *classifier) owned(p string) bool { return len(k.own.AreaOwners(p)) > 0 }

func (k *classifier) heur(p string) bool { return !k.owned(p) && k.own.HeuristicOnly(p) }

func (k *classifier) pooled(p string) bool { return !k.owned(p) && len(k.own.PendingAreas(p)) > 0 }

func (k *classifier) residual(m Member) {
	if k.heur(m.Path) {
		k.r.Counts["heuristic_only"]++
		m.HeuristicOnly = true
	}
	k.r.Residual = append(k.r.Residual, m)
}

func (k *classifier) classify(c Change) {
	if c.Kind == Deleted {
		k.r.Counts["deleted"]++
		k.r.Deleted = append(k.r.Deleted, c.Path)
		return
	}
	if why := excludedReason(c.Path); why != "" {
		k.r.Counts["excluded:"+why]++
		return
	}
	switch c.Kind {
	case Modified:
		k.modified(c)
	case Renamed:
		k.renamed(c)
	default: // Added
		k.added(c)
	}
}

func (k *classifier) modified(c Change) {
	switch {
	case k.owned(c.Path):
		k.r.Counts["modified_owned"]++
	case k.heur(c.Path):
		k.residual(Member{Path: c.Path, Kind: string(Modified)})
	default:
		k.r.Counts["modified_unowned"]++
		if k.pooled(c.Path) {
			k.r.Counts["modified_unowned:pooled"]++
		} else {
			k.r.Counts["modified_unowned:never_owned"]++
		}
		k.r.CoverageGaps = append(k.r.CoverageGaps, c.Path)
	}
}

func (k *classifier) renamed(c Change) {
	switch {
	case c.Empty:
		k.r.Counts["renamed_empty_trap"]++
		k.residual(Member{Path: c.Path, Kind: string(Renamed), Old: c.OldPath, Empty: true})
	case k.owned(c.OldPath):
		k.r.Counts["renamed_inherit"]++
		k.r.Inherited = append(k.r.Inherited, Inherited{Path: c.Path, Rule: domain.InheritedRename, From: c.OldPath, Areas: sortedCopy(k.own.AreaOwners(c.OldPath))})
	case k.pooled(c.OldPath):
		k.r.Counts["renamed_pooled"]++
		k.r.PendingMoves = append(k.r.PendingMoves, PendingMove{Old: c.OldPath, New: c.Path, Pending: sortedCopy(k.own.PendingAreas(c.OldPath))})
	default:
		k.r.Counts["renamed_unowned"]++
		k.residual(Member{Path: c.Path, Kind: string(Renamed), Old: c.OldPath})
	}
}

func (k *classifier) added(c Change) {
	if from, areas := variantOwners(c.Path, k.own, k.inTree); len(areas) > 0 {
		k.r.Counts["variant_inherit"]++
		k.r.Inherited = append(k.r.Inherited, Inherited{Path: c.Path, Rule: domain.InheritedVariant, From: from, Areas: areas})
		return
	}
	k.r.Counts["residual_added"]++
	k.residual(Member{Path: c.Path, Kind: string(Added), Empty: c.Empty})
}

// variantOwners returns the owners of owned files with the same variant-sibling
// key that exist in the parent tree, and the first such sibling (sorted).
func variantOwners(p string, own Ownership, inTree func(string) bool) (from string, areas []string) {
	set := map[string]bool{}
	sibs := sortedCopy(own.SiblingFiles(SiblingKeyOf(p)))
	for _, f := range sibs {
		if f == p || !inTree(f) {
			continue
		}
		if from == "" {
			from = f
		}
		for _, a := range own.AreaOwners(f) {
			set[a] = true
		}
	}
	for a := range set {
		areas = append(areas, a)
	}
	sort.Strings(areas)
	return from, areas
}

func sortedCopy(s []string) []string {
	out := append([]string(nil), s...)
	sort.Strings(out)
	return out
}
