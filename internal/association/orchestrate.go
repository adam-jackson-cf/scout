package association

import (
	"context"
	"fmt"
	"sort"
)

// Decision is the caller's verdict on one decided cluster. The zero value
// leaves the files unowned (unresolved, none, or nothing sent).
type Decision struct {
	Owners  []string // assigned: the files join these active areas
	Pending string   // pending_new: the files join this pending area
}

// Decider decides the clusters of one merge (at most PerMergeCap, in cap-key
// order) and returns one Decision per cluster. It is where the Jev and fallback
// lanes plug in. It is called only for merges that have decided clusters.
type Decider func(ctx context.Context, m Merge, decided []Cluster) ([]Decision, error)

// MergeResult is everything one merge produced.
type MergeResult struct {
	Merge Merge
	Result
	Inert     []Cluster
	Decided   []Cluster
	Decisions []Decision // parallel to Decided
	Capped    []Cluster  // unresolved (pr_cap)
}

// Run classifies the first-parent merges of base..head, oldest first. Each
// merge is classified with its own changes against an ownership view that
// includes the inherited results, pending moves and decisions of the earlier
// merges of this push, so a file decided once is an owner (and a modification
// of it is not decided) afterwards. Clusters of earlier pushes are never
// revisited: own must already contain their outcomes.
func Run(ctx context.Context, repo Repo, base, head string, own Ownership, decide Decider) ([]MergeResult, error) {
	merges, err := repo.Merges(ctx, base, head)
	if err != nil {
		return nil, err
	}
	view := newOverlay(own)
	var out []MergeResult
	for _, m := range merges {
		var tree map[string]struct{}
		var treeErr error
		inTree := func(p string) bool {
			if tree == nil && treeErr == nil {
				tree, treeErr = repo.TreePaths(ctx, m.Parent)
			}
			_, ok := tree[p]
			return ok
		}
		mr := MergeResult{Merge: m, Result: Classify(m.Changes, view, inTree)}
		if treeErr != nil {
			return nil, treeErr
		}
		var live []Cluster
		live, mr.Inert = BuildClusters(m.SHA, mr.Residual)
		mr.Decided, mr.Capped = Cap(live)
		if decide != nil && len(mr.Decided) > 0 {
			if mr.Decisions, err = decide(ctx, m, mr.Decided); err != nil {
				return nil, err
			}
			if len(mr.Decisions) != len(mr.Decided) {
				return nil, fmt.Errorf("association: decider returned %d decisions for %d clusters of %s", len(mr.Decisions), len(mr.Decided), m.SHA)
			}
		}
		view.apply(m.Changes, mr)
		out = append(out, mr)
	}
	return out, nil
}

// overlay is an Ownership that layers the results of earlier merges over a base.
type overlay struct {
	base    Ownership
	owners  map[string][]string // path -> owners; a nil value means none
	pending map[string][]string
	sibs    map[SiblingKey][]string // paths the overlay made owned
}

func newOverlay(base Ownership) *overlay {
	return &overlay{base: base, owners: map[string][]string{}, pending: map[string][]string{}, sibs: map[SiblingKey][]string{}}
}

func (o *overlay) AreaOwners(p string) []string {
	if v, ok := o.owners[p]; ok {
		return v
	}
	return o.base.AreaOwners(p)
}

func (o *overlay) PendingAreas(p string) []string {
	if v, ok := o.pending[p]; ok {
		return v
	}
	return o.base.PendingAreas(p)
}

func (o *overlay) HeuristicOnly(p string) bool { return o.base.HeuristicOnly(p) }

func (o *overlay) SiblingFiles(k SiblingKey) []string {
	return append(append([]string(nil), o.base.SiblingFiles(k)...), o.sibs[k]...)
}

func (o *overlay) drop(p string) {
	o.owners[p] = nil
	o.pending[p] = nil
}

func (o *overlay) own(p string, areas []string) {
	o.owners[p] = areas
	o.pending[p] = nil
	k := SiblingKeyOf(p)
	o.sibs[k] = append(o.sibs[k], p)
}

// apply folds one merge's effects into the view (gold.py `_Layer.op`, per merge
// instead of per round): deletions and renamed-away paths leave, inherited
// files and decided clusters take their owners, pending moves carry the pending set.
func (o *overlay) apply(changes []Change, mr MergeResult) {
	for _, c := range changes {
		switch c.Kind {
		case Deleted:
			o.drop(c.Path)
		case Renamed:
			o.drop(c.OldPath)
		}
	}
	for _, in := range mr.Inherited {
		o.own(in.Path, in.Areas)
	}
	for _, pm := range mr.PendingMoves {
		o.owners[pm.New] = nil
		o.pending[pm.New] = pm.Pending
	}
	for i, c := range mr.Decided {
		if i >= len(mr.Decisions) {
			break
		}
		d := mr.Decisions[i]
		owners := append([]string(nil), d.Owners...)
		sort.Strings(owners)
		for _, f := range c.Files {
			switch {
			case len(owners) > 0:
				o.own(f, owners)
			case d.Pending != "":
				o.owners[f] = nil
				o.pending[f] = []string{d.Pending}
			}
		}
	}
}
