package association

import "sort"

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
	r := Result{Counts: map[string]int{}}
	owned := func(p string) bool { return len(own.AreaOwners(p)) > 0 }
	heur := func(p string) bool { return !owned(p) && own.HeuristicOnly(p) }
	pooled := func(p string) bool { return !owned(p) && len(own.PendingAreas(p)) > 0 }
	residual := func(m Member) {
		if heur(m.Path) {
			r.Counts["heuristic_only"]++
			m.HeuristicOnly = true
		}
		r.Residual = append(r.Residual, m)
	}
	for _, c := range changes {
		if c.Kind == Deleted {
			r.Counts["deleted"]++
			r.Deleted = append(r.Deleted, c.Path)
			continue
		}
		if why := excludedReason(c.Path); why != "" {
			r.Counts["excluded:"+why]++
			continue
		}
		switch c.Kind {
		case Modified:
			switch {
			case owned(c.Path):
				r.Counts["modified_owned"]++
			case heur(c.Path): // rule (f): decided whether added, moved or modified
				residual(Member{Path: c.Path, Kind: string(Modified)})
			default:
				r.Counts["modified_unowned"]++
				if pooled(c.Path) {
					r.Counts["modified_unowned:pooled"]++
				} else {
					r.Counts["modified_unowned:never_owned"]++
				}
				r.CoverageGaps = append(r.CoverageGaps, c.Path)
			}
		case Renamed:
			switch {
			case c.Empty:
				r.Counts["renamed_empty_trap"]++
				residual(Member{Path: c.Path, Kind: string(Renamed), Old: c.OldPath, Empty: true})
			case owned(c.OldPath):
				r.Counts["renamed_inherit"]++
				r.Inherited = append(r.Inherited, Inherited{Path: c.Path, Rule: "rename", From: c.OldPath, Areas: sortedCopy(own.AreaOwners(c.OldPath))})
			case pooled(c.OldPath):
				r.Counts["renamed_pooled"]++
				r.PendingMoves = append(r.PendingMoves, PendingMove{Old: c.OldPath, New: c.Path, Pending: sortedCopy(own.PendingAreas(c.OldPath))})
			default:
				r.Counts["renamed_unowned"]++
				residual(Member{Path: c.Path, Kind: string(Renamed), Old: c.OldPath})
			}
		default: // Added
			if from, areas := variantOwners(c.Path, own, inTree); len(areas) > 0 {
				r.Counts["variant_inherit"]++
				r.Inherited = append(r.Inherited, Inherited{Path: c.Path, Rule: "variant", From: from, Areas: areas})
				continue
			}
			r.Counts["residual_added"]++
			residual(Member{Path: c.Path, Kind: string(Added), Empty: c.Empty})
		}
	}
	return r
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
