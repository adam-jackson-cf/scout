// Command frontend exposes Scout's deterministic association front end as JSON.
// It does not invoke a graph, a model, or any semantic association decision.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"unicode/utf8"

	"scout/internal/association"
	"scout/internal/areas"
	"scout/internal/clawpatch"
)

type request struct {
	Operation string              `json:"operation,omitempty"`
	StateDir  string              `json:"state_dir,omitempty"`
	Repo      string              `json:"repo"`
	Base      string              `json:"base"`
	Head      string              `json:"head"`
}

type ownershipRecord struct {
	Revision string `json:"revision"`
	Paths []string `json:"paths"`
}

type member struct {
	Path  string `json:"path"`
	Kind  string `json:"kind"`
	Old   string `json:"old"`
	Empty bool   `json:"empty"`
}

type cluster struct {
	ID        string   `json:"id"`
	Directory string   `json:"directory"`
	Members   []member `json:"members"`
}

type merge struct {
	Commit   string         `json:"commit"`
	Parent   string         `json:"parent"`
	Changes  []member       `json:"changes"`
	Selected []cluster      `json:"selected"`
	Capped   []cluster      `json:"capped"`
	Counts   map[string]int `json:"counts"`
}

type response struct {
	Merges []merge `json:"merges"`
	OwnedDirectories []string `json:"owned_directories"`
}

// heuristicOwnership intentionally has no areas, pending pools or siblings.
// Its map is replaced for each exact merge SHA, never carried forward.
type heuristicOwnership map[string]struct{}

func (heuristicOwnership) AreaOwners(string) []string { return nil }
func (heuristicOwnership) PendingAreas(string) []string { return nil }
func (heuristicOwnership) SiblingFiles(association.SiblingKey) []string { return nil }
func (h heuristicOwnership) HeuristicOnly(path string) bool {
	_, ok := h[path]
	return ok
}

// loadHeuristic uses the product's active-record loader and agent filter.
// Owned paths need not exist in the current tree, exactly as in the product.
func loadHeuristic(stateDir string) ([]string, error) {
	features, err := clawpatch.LoadFeatures(stateDir)
	if err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	for _, feature := range features {
		if !areas.IsAgent(feature) {
			for _, ref := range feature.OwnedFiles {
				set[ref.Path] = struct{}{}
			}
		}
	}
	paths := make([]string, 0, len(set))
	for path := range set {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func exportClusters(clusters []association.Cluster) []cluster {
	out := make([]cluster, 0, len(clusters))
	for _, c := range clusters {
		item := cluster{ID: c.ID, Directory: c.Dir, Members: make([]member, 0, len(c.Members))}
		for _, m := range c.Members {
			item.Members = append(item.Members, member{Path: m.Path, Kind: m.Kind, Old: m.Old, Empty: m.Empty})
		}
		out = append(out, item)
	}
	return out
}

func collect(ctx context.Context, req request, next func(string) ([]string, error)) (response, error) {
	out := response{Merges: []merge{}, OwnedDirectories: []string{}}
	if req.Repo == "" || req.Base == "" || req.Head == "" {
		return out, fmt.Errorf("repo, base and head are required")
	}
	basePaths, err := next(req.Base)
	if err != nil {
		return out, err
	}
	dirs := map[string]struct{}{}
	for _, path := range basePaths {
		dirs[association.SiblingKeyOf(path).Dir] = struct{}{}
	}
	for dir := range dirs {
		out.OwnedDirectories = append(out.OwnedDirectories, dir)
	}
	sort.Strings(out.OwnedDirectories)
	merges, err := (association.Repo{Dir: req.Repo}).Merges(ctx, req.Base, req.Head)
	if err != nil {
		return out, err
	}
	for _, m := range merges {
		paths, err := next(m.SHA)
		if err != nil {
			return out, err
		}
		own := make(heuristicOwnership, len(paths))
		for _, path := range paths {
			own[path] = struct{}{}
		}
		item := merge{Commit: m.SHA, Parent: m.Parent, Changes: make([]member, 0, len(m.Changes))}
		for _, c := range m.Changes {
			if !utf8.ValidString(c.Path) || !utf8.ValidString(c.OldPath) {
				return out, fmt.Errorf("invalid UTF-8 git path at merge %s; refusing lossful JSON", m.SHA)
			}
			item.Changes = append(item.Changes, member{Path: c.Path, Kind: string(c.Kind), Old: c.OldPath, Empty: c.Empty})
		}
		// With no sibling owners, Classify cannot inherit from the parent tree.
		classified := association.Classify(m.Changes, own, func(string) bool { return false })
		live, _ := association.BuildClusters(m.SHA, classified.Residual)
		selected, capped := association.Cap(live)
		item.Selected, item.Capped = exportClusters(selected), exportClusters(capped)
		item.Counts = classified.Counts
		out.Merges = append(out.Merges, item)
	}
	return out, nil
}

func decodeStrict(decoder *json.Decoder, target any) error {
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	if !utf8.Valid(raw) {
		return fmt.Errorf("invalid UTF-8 JSON input")
	}
	strict := json.NewDecoder(bytes.NewReader(raw))
	strict.DisallowUnknownFields()
	return strict.Decode(target)
}

func run(ctx context.Context, input io.Reader, output io.Writer) error {
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	var req request
	if err := decodeStrict(decoder, &req); err != nil {
		return err
	}
	if req.Operation == "ownership" {
		if req.StateDir == "" || req.Repo != "" || req.Base != "" || req.Head != "" {
			return fmt.Errorf("ownership operation requires only state_dir")
		}
		paths, err := loadHeuristic(req.StateDir)
		if err != nil {
			return err
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return fmt.Errorf("unexpected ownership input")
		}
		return json.NewEncoder(output).Encode(paths)
	}
	if req.Operation != "" || req.StateDir != "" {
		return fmt.Errorf("unknown operation or unexpected state_dir")
	}
	next := func(sha string) ([]string, error) {
		var record ownershipRecord
		if err := decodeStrict(decoder, &record); err != nil {
			return nil, fmt.Errorf("ownership record for exact merge SHA %s: %w", sha, err)
		}
		if record.Revision != sha || record.Paths == nil || !sort.StringsAreSorted(record.Paths) {
			return nil, fmt.Errorf("missing or unordered ownership record for exact merge SHA %s", sha)
		}
		for i, path := range record.Paths {
			if !utf8.ValidString(path) || (i > 0 && path == record.Paths[i-1]) {
				return nil, fmt.Errorf("invalid ownership paths for %s", sha)
			}
		}
		return record.Paths, nil
	}
	result, err := collect(ctx, req, next)
	if err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("unexpected trailing ownership record")
	}
	// Marshal the whole result before writing: a bad path in any merge must
	// refuse the request without exposing a partial inventory on stdout.
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = output.Write(append(encoded, '\n'))
	return err
}

func main() {
	// Product git subprocesses inherit these model-free/offline safeguards.
	for key, value := range map[string]string{"GIT_NO_REPLACE_OBJECTS": "1", "GIT_NO_LAZY_FETCH": "1", "GIT_TERMINAL_PROMPT": "0", "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": "/dev/null"} {
		if err := os.Setenv(key, value); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := run(context.Background(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
