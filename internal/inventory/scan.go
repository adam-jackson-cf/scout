// Package inventory builds read-only, content-addressed repository snapshots.
package inventory

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"scout/internal/domain"
)

// Diagnostic describes an inventory limitation without converting a partial
// snapshot into a clean one.
type Diagnostic struct {
	Stage   domain.Stage `json:"stage"`
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Path    string       `json:"path,omitempty"`
}

// MappingGap records a missing or ambiguous canonical Clawpatch relationship.
type MappingGap struct {
	Path      string   `json:"path,omitempty"`
	FeatureID string   `json:"featureId,omitempty"`
	Reason    string   `json:"reason"`
	Related   []string `json:"related,omitempty"`
}

// ScanOptions controls a read-only inventory. Exclusions are slash-separated
// repository-relative patterns in addition to the mandatory .scout exclusion.
type ScanOptions struct {
	Scope      domain.Scope
	Exclusions []string
	Features   []domain.Feature
}

// ScanResult is the content-addressed input inventory for one selected scope.
// Removed contains paths selected by staged/range scopes that have no readable
// source in the target snapshot.
type ScanResult struct {
	Root        string            `json:"root"`
	Snapshot    domain.Snapshot   `json:"snapshot"`
	Scope       domain.Scope      `json:"scope"`
	ScopeDigest string            `json:"scopeDigest"`
	Files       []domain.File     `json:"files"`
	Removed     []domain.File     `json:"removed,omitempty"`
	Omissions   []domain.Omission `json:"omissions,omitempty"`
	Diagnostics []Diagnostic      `json:"diagnostics,omitempty"`
	MappingGaps []MappingGap      `json:"mappingGaps,omitempty"`
}

// Scan resolves the requested Git scope and hashes its selected source without
// writing to the repository. It returns every readable record gathered before
// an error and reports unreadable eligible inputs as omissions.
func Scan(ctx context.Context, root string, options ScanOptions) (ScanResult, error) {
	result := ScanResult{Scope: options.Scope}
	if err := options.Scope.Validate(); err != nil {
		return result, fmt.Errorf("validate scope: %w", err)
	}
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return result, fmt.Errorf("resolve root: %w", err)
	}
	resolvedRoot, err = filepath.EvalSymlinks(filepath.Clean(resolvedRoot))
	if err != nil {
		return result, fmt.Errorf("resolve root symlinks: %w", err)
	}
	result.Root = resolvedRoot

	gitRoot, err := gitOutput(ctx, resolvedRoot, "rev-parse", "--show-toplevel")
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, diagnostic("git_unavailable", err.Error(), ""))
		return result, fmt.Errorf("resolve Git repository: %w", err)
	}
	gitRootPath := filepath.Clean(strings.TrimSpace(string(gitRoot)))
	if gitRootPath != resolvedRoot {
		result.Diagnostics = append(result.Diagnostics, diagnostic("root_not_repository", "assessment root must be the Git repository root", ""))
		return result, fmt.Errorf("assessment root %q is not Git root %q", resolvedRoot, gitRootPath)
	}

	revision, revisionErr := gitOutput(ctx, resolvedRoot, "rev-parse", "--verify", "HEAD")
	if revisionErr != nil && unbornHEAD(ctx, resolvedRoot) {
		revisionErr = nil
		result.Diagnostics = append(result.Diagnostics, diagnostic("revision_unavailable", "repository has no initial commit", ""))
	}
	selectedRevision := strings.TrimSpace(string(revision))
	if options.Scope.Kind == domain.ScopeCommitRange {
		to, err := gitOutput(ctx, resolvedRoot, "rev-parse", options.Scope.ToRevision)
		if err != nil {
			result.Diagnostics = append(result.Diagnostics, diagnostic("range_revision_unavailable", err.Error(), ""))
			return result, fmt.Errorf("resolve range end: %w", err)
		}
		selectedRevision = strings.TrimSpace(string(to))
	}

	ignoredDigest, ignoreDiagnostics := ignoreRulesDigest(ctx, resolvedRoot)
	result.Diagnostics = append(result.Diagnostics, ignoreDiagnostics...)
	exclusionDigest := digestStrings(append([]string{scoutStatePrefix}, normalizedExclusions(options.Exclusions)...))
	result.ScopeDigest = scopeDigest(options.Scope, options.Exclusions)

	worktreeStatus, statusErr := gitOutput(ctx, resolvedRoot, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--", ".", ":(exclude).scout")
	dirty := len(worktreeStatus) != 0
	if statusErr != nil {
		result.Diagnostics = append(result.Diagnostics, diagnostic("worktree_status_unavailable", statusErr.Error(), ""))
	}

	candidates, removed, reader, listDiagnostics, err := scopeCandidates(ctx, resolvedRoot, options.Scope)
	result.Diagnostics = append(result.Diagnostics, listDiagnostics...)
	if err != nil {
		return result, fmt.Errorf("resolve scope: %w", err)
	}

	var inputErrors []error
	for _, path := range sortedUnique(candidates) {
		if err := validateRepositoryPath(path); err != nil {
			result.Diagnostics = append(result.Diagnostics, diagnostic("unsafe_git_path", "Git path is not a normalized repository-relative path", ""))
			inputErrors = append(inputErrors, err)
			continue
		}
		if IsScoutState(path) {
			continue
		}
		file, omission, readErr := scanPath(path, reader, options.Exclusions)
		if readErr != nil {
			code, message := readDiagnostic(readErr)
			result.Diagnostics = append(result.Diagnostics, diagnostic(code, message, path))
		}
		result.Files = append(result.Files, file)
		if omission != nil {
			result.Omissions = append(result.Omissions, *omission)
		}
	}
	for _, path := range sortedUnique(removed) {
		if err := validateRepositoryPath(path); err != nil {
			result.Diagnostics = append(result.Diagnostics, diagnostic("unsafe_git_path", "Git path is not a normalized repository-relative path", ""))
			inputErrors = append(inputErrors, err)
			continue
		}
		if IsScoutState(path) {
			continue
		}
		if excluded(path, options.Exclusions) {
			result.Removed = append(result.Removed, domain.File{Path: path, Excluded: true, ExclusionReason: exclusionReason(path, options.Exclusions)})
			continue
		}
		result.Removed = append(result.Removed, domain.File{Path: path, Eligible: candidateEligible(path)})
	}

	linked, gaps, linkErr := LinkFeatures(result.Files, options.Features)
	result.Files = linked
	result.MappingGaps = append(result.MappingGaps, gaps...)
	if linkErr != nil {
		result.Diagnostics = append(result.Diagnostics, diagnostic("feature_map_invalid", linkErr.Error(), ""))
	}

	result.Snapshot = domain.Snapshot{
		RepositoryID:      repositoryID(ctx, resolvedRoot),
		Revision:          selectedRevision,
		WorktreeDigest:    filesDigest(result.Files, result.Removed),
		Dirty:             dirty,
		IgnoreRulesDigest: ignoredDigest,
		ExclusionsDigest:  exclusionDigest,
	}
	if err := result.Snapshot.Validate(); err != nil {
		return result, fmt.Errorf("build snapshot: %w", err)
	}
	if revisionErr != nil || statusErr != nil || linkErr != nil || len(inputErrors) != 0 {
		return result, errors.Join(append(inputErrors, revisionErr, statusErr, linkErr)...)
	}
	return result, nil
}

// LinkFeatures attaches only canonical Clawpatch IDs to files. It neither
// derives new feature identities nor assigns ownership heuristically.
func LinkFeatures(files []domain.File, features []domain.Feature) ([]domain.File, []MappingGap, error) {
	byPath := make(map[string]int, len(files))
	for i, file := range files {
		if err := file.Validate(); err != nil {
			return files, nil, fmt.Errorf("validate file %q: %w", file.Path, err)
		}
		byPath[file.Path] = i
	}
	var errs []error
	valid := make([]domain.Feature, 0, len(features))
	for _, feature := range features {
		if err := feature.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("validate feature %q: %w", feature.FeatureID, err))
			continue
		}
		valid = append(valid, feature)
		for _, ref := range feature.OwnedFiles {
			if index, exists := byPath[ref.Path]; exists {
				files[index].OwnedFeatureIDs = append(files[index].OwnedFeatureIDs, feature.FeatureID)
			}
		}
		for _, ref := range feature.ContextFiles {
			if index, exists := byPath[ref.Path]; exists {
				files[index].ContextFeatureIDs = append(files[index].ContextFeatureIDs, feature.FeatureID)
			}
		}
	}
	for index := range files {
		files[index].OwnedFeatureIDs = sortedUnique(files[index].OwnedFeatureIDs)
		files[index].ContextFeatureIDs = sortedUnique(files[index].ContextFeatureIDs)
	}
	return files, mappingGaps(files, valid), errors.Join(errs...)
}

var (
	errPathContainsSymlink   = errors.New("repository path contains a symlink")
	errPathNotRegular        = errors.New("repository path is not a regular file")
	errInvalidRepositoryPath = errors.New("repository path is not a normalized relative path")
)

type contentReader func(string) (io.ReadCloser, error)

func scopeCandidates(ctx context.Context, root string, scope domain.Scope) ([]string, []string, contentReader, []Diagnostic, error) {
	switch scope.Kind {
	case domain.ScopeRepository, domain.ScopeSubpath, domain.ScopePaths:
		paths, err := workingPaths(ctx, root)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return filterScope(paths, scope), nil, workingTreeReader(root), nil, nil
	case domain.ScopeStaged:
		changes, err := nameStatus(ctx, root, "diff", "--cached", "--name-status", "-z", "--find-renames")
		if err != nil {
			return nil, nil, nil, nil, err
		}
		paths, removed, err := selectedChanges(changes)
		if err != nil {
			return nil, nil, nil, []Diagnostic{diagnostic("unsafe_git_path", "Git change path is not a normalized repository-relative path", "")}, err
		}
		return paths, removed, func(path string) (io.ReadCloser, error) {
			if err := validateRepositoryPath(path); err != nil {
				return nil, err
			}
			content, err := gitOutput(ctx, root, "show", ":"+path)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(bytes.NewReader(content)), nil
		}, nil, nil
	case domain.ScopeCommitRange:
		changes, err := nameStatus(ctx, root, "diff", "--name-status", "-z", "--find-renames", scope.FromRevision, scope.ToRevision)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		paths, removed, err := selectedChanges(changes)
		if err != nil {
			return nil, nil, nil, []Diagnostic{diagnostic("unsafe_git_path", "Git change path is not a normalized repository-relative path", "")}, err
		}
		return paths, removed, func(path string) (io.ReadCloser, error) {
			if err := validateRepositoryPath(path); err != nil {
				return nil, err
			}
			content, err := gitOutput(ctx, root, "show", scope.ToRevision+":"+path)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(bytes.NewReader(content)), nil
		}, nil, nil
	default:
		return nil, nil, nil, nil, fmt.Errorf("unsupported scope %q", scope.Kind)
	}
}

func workingPaths(ctx context.Context, root string) ([]string, error) {
	output, err := gitOutput(ctx, root, "ls-files", "-co", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	return nulPaths(output), nil
}

func filterScope(paths []string, scope domain.Scope) []string {
	if scope.Kind == domain.ScopeRepository {
		return paths
	}
	var selected []string
	for _, path := range paths {
		for _, selectedPath := range scope.Paths {
			if path == selectedPath || strings.HasPrefix(path, selectedPath+"/") {
				selected = append(selected, path)
				break
			}
		}
	}
	return selected
}

type nameStatusEntry struct {
	Status string
	Old    string
	New    string
}

func nameStatus(ctx context.Context, root string, args ...string) ([]nameStatusEntry, error) {
	output, err := gitOutput(ctx, root, args...)
	if err != nil {
		return nil, err
	}
	fields := nulPaths(output)
	var entries []nameStatusEntry
	for len(fields) > 0 {
		status := fields[0]
		fields = fields[1:]
		if status == "" || len(fields) == 0 {
			return nil, errors.New("malformed Git name-status output")
		}
		entry := nameStatusEntry{Status: status}
		if status[0] == 'R' || status[0] == 'C' {
			if len(fields) < 2 {
				return nil, errors.New("truncated Git rename status")
			}
			entry.Old, entry.New = fields[0], fields[1]
			fields = fields[2:]
		} else {
			entry.New = fields[0]
			fields = fields[1:]
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func selectedChanges(entries []nameStatusEntry) ([]string, []string, error) {
	var paths, removed []string
	for _, entry := range entries {
		if err := validateRepositoryPath(entry.New); err != nil {
			return nil, nil, err
		}
		if entry.Status[0] == 'R' || entry.Status[0] == 'C' {
			if err := validateRepositoryPath(entry.Old); err != nil {
				return nil, nil, err
			}
		}
		switch entry.Status[0] {
		case 'D':
			removed = append(removed, entry.New)
		case 'R':
			paths = append(paths, entry.New)
			removed = append(removed, entry.Old)
		default:
			paths = append(paths, entry.New)
		}
	}
	return paths, removed, nil
}

func workingTreeReader(root string) contentReader {
	return func(path string) (io.ReadCloser, error) {
		if err := validateRepositoryPath(path); err != nil {
			return nil, err
		}
		return openRegularFile(root, path)
	}
}

func openRegularFile(root, path string) (io.ReadCloser, error) {
	directory, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = unix.Close(directory) }()

	components := strings.Split(path, "/")
	for _, component := range components[:len(components)-1] {
		next, openErr := unix.Openat(directory, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if openErr != nil {
			if errors.Is(openErr, unix.ELOOP) {
				return nil, errPathContainsSymlink
			}
			return nil, openErr
		}
		_ = unix.Close(directory)
		directory = next
	}
	file, err := unix.Openat(directory, components[len(components)-1], unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		if errors.Is(err, unix.ELOOP) {
			return nil, errPathContainsSymlink
		}
		return nil, err
	}
	var stat unix.Stat_t
	if err := unix.Fstat(file, &stat); err != nil {
		_ = unix.Close(file)
		return nil, err
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		_ = unix.Close(file)
		return nil, errPathNotRegular
	}
	return os.NewFile(uintptr(file), filepath.Join(root, filepath.FromSlash(path))), nil
}

func validateRepositoryPath(path string) error {
	if path == "" || strings.ContainsRune(path, '\x00') || strings.Contains(path, "\\") || filepath.IsAbs(path) || path == "." || path == ".." || strings.HasPrefix(path, "../") || filepath.Clean(path) != path {
		return errInvalidRepositoryPath
	}
	return nil
}

func readDiagnostic(err error) (string, string) {
	if errors.Is(err, errPathContainsSymlink) {
		return "path_contains_symlink", "repository path contains a symlink"
	}
	if errors.Is(err, errPathNotRegular) {
		return "unsupported_file_type", "repository path is not a regular file"
	}
	if errors.Is(err, errInvalidRepositoryPath) {
		return "unsafe_git_path", "repository path is not a normalized relative path"
	}
	return "file_unreadable", err.Error()
}

func unbornHEAD(ctx context.Context, root string) bool {
	ref, err := gitOutput(ctx, root, "symbolic-ref", "-q", "HEAD")
	if err != nil {
		return false
	}
	refs, err := gitOutput(ctx, root, "for-each-ref", "--format=%(refname)", strings.TrimSpace(string(ref)))
	return err == nil && strings.TrimSpace(string(refs)) == ""
}

func scanPath(path string, reader contentReader, patterns []string) (domain.File, *domain.Omission, error) {
	if excluded(path, patterns) {
		return domain.File{Path: path, Excluded: true, ExclusionReason: exclusionReason(path, patterns)}, nil, nil
	}
	source, err := reader(path)
	var text bool
	var digest string
	if err == nil {
		defer source.Close()
		digest, text, err = digestText(source)
	}
	if err != nil {
		file := domain.File{Path: path, Language: languageFor(path), Eligible: candidateEligible(path)}
		if !file.Eligible {
			return file, nil, err
		}
		omission := domain.Omission{CoverageKey: domain.CoverageKey{FilePath: path, Family: domain.FamilySecurity, Category: "inventory"}, Reason: "eligible file could not be read: " + err.Error()}
		return file, &omission, err
	}
	return domain.File{
		Path:          path,
		Language:      languageFor(path),
		ContentDigest: digest,
		Eligible:      text && candidateEligible(path),
	}, nil, nil
}

func digestText(reader io.Reader) (string, bool, error) {
	hash := sha256.New()
	buffer := make([]byte, 32*1024)
	remainingInspection := 8192
	text := true
	for {
		count, err := reader.Read(buffer)
		if count != 0 {
			if _, writeErr := hash.Write(buffer[:count]); writeErr != nil {
				return "", false, writeErr
			}
			if remainingInspection != 0 {
				inspected := min(count, remainingInspection)
				if bytes.IndexByte(buffer[:inspected], 0) >= 0 {
					text = false
				}
				remainingInspection -= inspected
			}
		}
		if err == io.EOF {
			return hex.EncodeToString(hash.Sum(nil)), text, nil
		}
		if err != nil {
			return "", false, err
		}
	}
}

func candidateEligible(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".a", ".bmp", ".class", ".dll", ".dylib", ".exe", ".gif", ".ico", ".jar", ".jpeg", ".jpg", ".o", ".pdf", ".png", ".so", ".wasm", ".zip":
		return false
	default:
		return true
	}
}

func languageFor(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".py", ".pyi":
		return "python"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx", ".mts", ".cts":
		return "typescript"
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".cs":
		return "csharp"
	case ".sql":
		return "sql"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".toml":
		return "toml"
	case ".xml":
		return "xml"
	case ".sh":
		return "shell"
	default:
		return ""
	}
}

// IsScoutState reports whether a repository-relative path is Scout-managed
// state (`.scout/`), which every lane and trigger excludes.
func IsScoutState(path string) bool {
	return path == strings.TrimSuffix(scoutStatePrefix, "/") || strings.HasPrefix(path, scoutStatePrefix)
}

const scoutStatePrefix = ".scout/"

func excluded(path string, patterns []string) bool {
	if IsScoutState(path) {
		return true
	}
	for _, pattern := range normalizedExclusions(patterns) {
		if path == pattern || strings.HasPrefix(path, strings.TrimSuffix(pattern, "/")+"/") {
			return true
		}
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}
	return false
}

func exclusionReason(path string, patterns []string) string {
	if IsScoutState(path) {
		return "Scout-managed state is always excluded"
	}
	return "configured inventory exclusion"
}

func normalizedExclusions(patterns []string) []string {
	values := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(strings.ReplaceAll(pattern, "\\", "/"))
		if pattern != "" {
			values = append(values, strings.TrimPrefix(pattern, "./"))
		}
	}
	return sortedUnique(values)
}

func ignoreRulesDigest(ctx context.Context, root string) (string, []Diagnostic) {
	parts := []string{}
	var diagnostics []Diagnostic
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			diagnostics = append(diagnostics, diagnostic("ignore_rule_unreadable", err.Error(), path))
			return nil
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == ".scout") {
			return filepath.SkipDir
		}
		if !entry.IsDir() && entry.Name() == ".gitignore" {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				diagnostics = append(diagnostics, diagnostic("ignore_rule_unreadable", readErr.Error(), path))
				return nil
			}
			relative, _ := filepath.Rel(root, path)
			parts = append(parts, filepath.ToSlash(relative)+"\x00"+digestBytes(content))
		}
		return nil
	})
	infoExclude := filepath.Join(root, ".git", "info", "exclude")
	if content, err := os.ReadFile(infoExclude); err == nil {
		parts = append(parts, filepath.ToSlash(strings.TrimPrefix(infoExclude, root+string(filepath.Separator)))+"\x00"+digestBytes(content))
	} else if !errors.Is(err, os.ErrNotExist) {
		diagnostics = append(diagnostics, diagnostic("ignore_rule_unreadable", err.Error(), infoExclude))
	}
	global, err := gitOutput(ctx, root, "config", "--path", "--get", "core.excludesFile")
	if err == nil && strings.TrimSpace(string(global)) != "" {
		path := strings.TrimSpace(string(global))
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			diagnostics = append(diagnostics, diagnostic("global_ignore_unreadable", readErr.Error(), path))
		} else {
			parts = append(parts, "global\x00"+digestBytes(content))
		}
	}
	return digestStrings(parts), diagnostics
}

func repositoryID(ctx context.Context, root string) string {
	remote, err := gitOutput(ctx, root, "config", "--get", "remote.origin.url")
	identity := filepath.Clean(root)
	if err == nil && strings.TrimSpace(string(remote)) != "" {
		identity = strings.TrimSpace(string(remote))
	}
	return "repo-" + digestStrings([]string{identity})
}

func gitOutput(ctx context.Context, root string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	return output, nil
}

func filesDigest(files, removed []domain.File) string {
	parts := make([]string, 0, len(files)+len(removed))
	for _, file := range append(append([]domain.File{}, files...), removed...) {
		parts = append(parts, strings.Join([]string{file.Path, file.PreviousPath, file.ContentDigest, fmt.Sprintf("%t", file.Eligible), fmt.Sprintf("%t", file.Excluded), file.ExclusionReason}, "\x00"))
	}
	return digestStrings(parts)
}

func scopeDigest(scope domain.Scope, exclusions []string) string {
	paths := append([]string(nil), scope.Paths...)
	sort.Strings(paths)
	families := make([]string, len(scope.SelectedFamilies))
	for i, family := range scope.SelectedFamilies {
		families[i] = string(family)
	}
	sort.Strings(families)
	// Positional fields and the sorted set sections are one ordered tuple, so
	// swapping from/to or moving a value between sections changes the digest.
	exclusionSet := normalizedExclusions(exclusions)
	parts := []string{"kind", string(scope.Kind), "from", scope.FromRevision, "to", scope.ToRevision}
	parts = append(parts, "paths", strconv.Itoa(len(paths)))
	parts = append(parts, paths...)
	parts = append(parts, "families", strconv.Itoa(len(families)))
	parts = append(parts, families...)
	parts = append(parts, "exclusions", strconv.Itoa(len(exclusionSet)))
	parts = append(parts, exclusionSet...)
	return orderedDigest(parts...)
}

func digestBytes(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }

// orderedDigest hashes values as an ordered tuple; use it wherever position
// is meaningful. digestStrings hashes a set and ignores order.
func orderedDigest(values ...string) string {
	return digestStrings([]string{fingerprintTuple(values...)})
}

// digestStrings is order-insensitive: it hashes the sorted values.
func digestStrings(values []string) string {
	ordered := append([]string(nil), values...)
	sort.Strings(ordered)
	hash := sha256.New()
	for _, value := range ordered {
		_, _ = fmt.Fprintf(hash, "%d:", len(value))
		_, _ = io.WriteString(hash, value)
	}
	return hex.EncodeToString(hash.Sum(nil))
}
func nulPaths(value []byte) []string {
	return strings.FieldsFunc(string(value), func(r rune) bool { return r == '\x00' })
}

// sortedUnique returns the sorted distinct non-empty values in a new slice; the
// input is not modified.
func sortedUnique(values []string) []string {
	result := slices.DeleteFunc(slices.Sorted(slices.Values(values)), func(value string) bool { return value == "" })
	return slices.Compact(result)
}
func diagnostic(code, message, path string) Diagnostic {
	return Diagnostic{Stage: domain.StageInventory, Code: code, Message: message, Path: path}
}
func sortMappingGaps(gaps []MappingGap) {
	sort.Slice(gaps, func(i, j int) bool {
		if gaps[i].Path != gaps[j].Path {
			return gaps[i].Path < gaps[j].Path
		}
		if gaps[i].FeatureID != gaps[j].FeatureID {
			return gaps[i].FeatureID < gaps[j].FeatureID
		}
		return gaps[i].Reason < gaps[j].Reason
	})
}
