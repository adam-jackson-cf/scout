package state

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"scout/internal/areas"
	"scout/internal/domain"
)

// ExportDir is the repository-relative directory holding the exported ledger
// (design §7.2). Every key of an export map is a slash path beneath it.
const ExportDir = "state/v1"

// Export file names, relative to the repository root of the export.
const (
	ExportIssues          = ExportDir + "/issues.jsonl"
	ExportDispositions    = ExportDir + "/dispositions.jsonl"
	ExportWouldBlock      = ExportDir + "/would_block.jsonl"
	ExportCoverageSummary = ExportDir + "/coverage-summary.json"
	ExportManifest        = ExportDir + "/manifest.json"
	// ExportSignalsDir holds one signals/<metric@version>.jsonl per metric.
	ExportSignalsDir = ExportDir + "/signals/"
)

const (
	// ExportSchema identifies the export format in manifest.json.
	ExportSchema = "scout.state/v1"

	// WouldBlockPresent is the manifest value declaring that would_block.jsonl
	// carries the gate's shadow table.
	WouldBlockPresent = "present"

	metadataCoverageSummary = "coverage_summary"
	metadataExportBase      = "export_base"
)

// ErrImportNotEmpty reports an import into a store that already holds issues.
var ErrImportNotEmpty = errors.New("import requires a fresh store with no issues")

// dispositionKeys are the mutable, triage-owned fields of an issue. They are
// exported in dispositions.jsonl so concurrent publishers merge them per issue.
var dispositionKeys = []string{"state", "owner", "dispositionReason", "updatedAt", "verification", "history"}

// Export renders the durable ledger as deterministic state/v1 files. The result
// maps repository-relative paths to bytes: records are canonical JSON (sorted
// keys), one per line, sorted by issue ID, LF-terminated, with no timestamps
// beyond those stored in the ledger.
func (s *Store) Export(ctx context.Context) (map[string][]byte, error) {
	issues, err := s.ListIssues(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(issues, func(i, j int) bool { return issues[i].IssueID < issues[j].IssueID })

	var issueLines, dispositionLines [][]byte
	for _, issue := range issues {
		raw, err := json.Marshal(issue)
		if err != nil {
			return nil, fmt.Errorf("encode issue: %w", err)
		}
		row := map[string]any{}
		if err := unmarshalNumbers(raw, &row); err != nil {
			return nil, err
		}
		disposition := map[string]any{"issueId": issue.IssueID}
		for _, key := range dispositionKeys {
			if value, ok := row[key]; ok {
				disposition[key] = value
				delete(row, key)
			}
		}
		issueLine, err := CanonicalLine(row)
		if err != nil {
			return nil, err
		}
		dispositionLine, err := CanonicalLine(disposition)
		if err != nil {
			return nil, err
		}
		issueLines = append(issueLines, issueLine)
		dispositionLines = append(dispositionLines, dispositionLine)
	}

	wouldBlock, wouldBlockStatus, err := s.exportWouldBlock(ctx)
	if err != nil {
		return nil, err
	}
	summary, err := s.coverageSummary(ctx)
	if err != nil {
		return nil, err
	}
	meta, err := s.baseInfo(ctx)
	if err != nil {
		return nil, err
	}
	meta.WouldBlock = wouldBlockStatus
	meta.RepositoryID, err = s.metadataValue(ctx, "repository_id")
	if err != nil {
		return nil, err
	}

	files := map[string][]byte{
		ExportIssues:          joinLines(issueLines),
		ExportDispositions:    joinLines(dispositionLines),
		ExportWouldBlock:      joinLines(wouldBlock),
		ExportCoverageSummary: summary,
	}
	signalFiles, err := s.exportSignals(ctx)
	if err != nil {
		return nil, err
	}
	for path, data := range signalFiles {
		files[path] = data
	}
	areaFile, err := s.exportAreas(ctx)
	if err != nil {
		return nil, err
	}
	if areaFile != nil {
		files[ExportAreas] = areaFile
	}
	refreshFile, err := s.ExportMappingRefreshFile(ctx)
	if err != nil {
		return nil, err
	}
	if refreshFile != nil {
		files[ExportMappingRefresh] = refreshFile
	}
	files[ExportManifest] = RenderManifest(meta, files)
	return files, nil
}

// exportSignals renders the signals ledger as signals/<metric@version>.jsonl,
// one canonical record per line in identity order.
func (s *Store) exportSignals(ctx context.Context) (map[string][]byte, error) {
	signals, err := s.ListSignals(ctx, SignalFilter{})
	if err != nil {
		return nil, err
	}
	lines := map[string][][]byte{}
	for _, signal := range signals {
		raw, err := json.Marshal(signal)
		if err != nil {
			return nil, fmt.Errorf("encode signal: %w", err)
		}
		fields := map[string]any{}
		if err := unmarshalNumbers(raw, &fields); err != nil {
			return nil, err
		}
		line, err := CanonicalLine(fields)
		if err != nil {
			return nil, err
		}
		path := ExportSignalsDir + signal.ID() + ".jsonl"
		lines[path] = append(lines[path], line)
	}
	files := make(map[string][]byte, len(lines))
	for path, metricLines := range lines {
		files[path] = joinLines(metricLines)
	}
	return files, nil
}

// exportWouldBlock returns the gate's would_block rows as canonical lines
// sorted by row ID.
func (s *Store) exportWouldBlock(ctx context.Context) ([][]byte, string, error) {
	rows, err := s.ListWouldBlock(ctx)
	if err != nil {
		return nil, "", err
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	lines := make([][]byte, 0, len(rows))
	for _, row := range rows {
		raw, err := json.Marshal(row)
		if err != nil {
			return nil, "", fmt.Errorf("encode would_block: %w", err)
		}
		fields := map[string]any{}
		if err := unmarshalNumbers(raw, &fields); err != nil {
			return nil, "", err
		}
		line, err := CanonicalLine(fields)
		if err != nil {
			return nil, "", err
		}
		lines = append(lines, line)
	}
	return lines, WouldBlockPresent, nil
}

// ManifestMeta carries the manifest fields that are not derived from the files.
// Revision, ToolDigest and PolicyDigest come from the latest finalized complete
// run and its plan; they let a reader decide whether a baseline is comparable.
// Uninspected lists the scanner families that run did not fully inspect; an
// empty list means all were inspected. The exporter always writes it; nil only
// results from reading a manifest that lacks it.
type ManifestMeta struct {
	RepositoryID string              `json:"repositoryId"`
	WouldBlock   string              `json:"wouldBlock"`
	Revision     string              `json:"revision"`
	ToolDigest   string              `json:"toolDigest"`
	PolicyDigest string              `json:"policyDigest"`
	Uninspected  []UninspectedFamily `json:"uninspected"`
}

// RenderManifest renders manifest.json for the given export files (the
// manifest itself is excluded). It lists exactly what is present.
func RenderManifest(meta ManifestMeta, files map[string][]byte) []byte {
	type fileEntry struct {
		Path    string `json:"path"`
		Records int    `json:"records"`
		SHA256  string `json:"sha256"`
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		if path != ExportManifest {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	entries := make([]fileEntry, 0, len(paths))
	for _, path := range paths {
		sum := sha256.Sum256(files[path])
		records := bytes.Count(files[path], []byte{'\n'})
		if strings.HasSuffix(path, ".json") {
			records = 1
		}
		entries = append(entries, fileEntry{Path: path, Records: records, SHA256: hex.EncodeToString(sum[:])})
	}
	manifest := map[string]any{
		"schema":       ExportSchema,
		"repositoryId": meta.RepositoryID,
		"revision":     meta.Revision,
		"toolDigest":   meta.ToolDigest,
		"policyDigest": meta.PolicyDigest,
		"files":        entries,
		"absent":       absentFiles(paths),
		"wouldBlock":   meta.WouldBlock,
	}
	uninspected := meta.Uninspected
	if uninspected == nil {
		uninspected = []UninspectedFamily{}
	}
	manifest["uninspected"] = uninspected
	line, err := CanonicalLine(manifest)
	if err != nil { // map of plain values: cannot fail
		panic(err)
	}
	return line
}

// absentFiles lists the planned export files the manifest does not carry.
func absentFiles(paths []string) []string {
	var absent []string
	hasAreas := false
	for _, path := range paths {
		hasAreas = hasAreas || path == ExportAreas
	}
	if !hasAreas {
		absent = append(absent, ExportAreas)
	}
	for _, path := range paths {
		if strings.HasPrefix(path, ExportSignalsDir) {
			return absent
		}
	}
	return append(absent, ExportSignalsDir)
}

// UninspectedFamily is a scanner family the baseline run did not fully
// inspect: a coverage row of that family was blocked, incomplete, skipped or
// awaiting manual verification.
type UninspectedFamily struct {
	Family domain.Family `json:"family"`
	Reason string        `json:"reason,omitempty"`
}

// uninspectedFamilies lists, in family order, the scanner families (file
// coverage; model-review lanes are feature keyed) with an unaccounted row on run.
// The result is never nil: an empty list means every scanner family was inspected.
func uninspectedFamilies(ctx context.Context, db *sql.DB, runID string) ([]UninspectedFamily, error) {
	rows, err := db.QueryContext(ctx, `SELECT payload FROM coverage WHERE run_id = ? AND status NOT IN (?, ?, ?, ?, ?) ORDER BY coverage_key`, runID,
		domain.CoverageInspected, domain.CoverageReused, domain.CoverageNotApplicable, domain.CoverageCovered, domain.CoverageFlagged)
	if err != nil {
		return nil, fmt.Errorf("query base coverage gaps: %w", err)
	}
	defer rows.Close()
	seen := map[domain.Family]bool{}
	out := []UninspectedFamily{}
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan base coverage gap: %w", err)
		}
		var row domain.Coverage
		if err := decode(payload, &row); err != nil {
			return nil, err
		}
		if row.Key.FeatureID != "" || row.Key.Family == domain.FamilyClawpatch || seen[row.Key.Family] {
			continue
		}
		seen[row.Key.Family] = true
		reason := string(row.Status)
		if row.Reason != "" {
			reason += ": " + row.Reason
		}
		out = append(out, UninspectedFamily{Family: row.Key.Family, Reason: reason})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Family < out[j].Family })
	return out, nil
}

// ReadManifestMeta reads the non-derived manifest fields back.
func ReadManifestMeta(manifest []byte) (ManifestMeta, error) {
	var meta ManifestMeta
	if err := json.Unmarshal(manifest, &meta); err != nil {
		return ManifestMeta{}, fmt.Errorf("decode manifest: %w", err)
	}
	return meta, nil
}

// baseInfo reports the revision and digests of the latest finalized run. Runs
// are partial whenever optional lanes (contextual, clawpatch-review) are blocked,
// so partial runs are valid baseline sources; the gate checks required lanes.
func (s *Store) baseInfo(ctx context.Context) (ManifestMeta, error) {
	var meta ManifestMeta
	found := false
	err := s.withRead(ctx, func(db *sql.DB) error {
		var runPayload, planPayload []byte
		var runID string
		err := db.QueryRowContext(ctx, `SELECT r.run_id, r.payload, p.payload FROM runs r JOIN plans p ON p.plan_id = r.plan_id
			WHERE r.state IN (?, ?) AND r.finalized = 1 ORDER BY r.finished_at DESC, r.started_at DESC, r.run_id LIMIT 1`, domain.RunComplete, domain.RunPartial).Scan(&runID, &runPayload, &planPayload)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("query latest run: %w", err)
		}
		var run domain.Run
		var plan domain.Plan
		if err := decode(runPayload, &run); err != nil {
			return err
		}
		if err := decode(planPayload, &plan); err != nil {
			return err
		}
		meta.Revision, meta.ToolDigest, meta.PolicyDigest = run.Snapshot.Revision, plan.ToolDigest, plan.PolicyDigest
		meta.Uninspected, err = uninspectedFamilies(ctx, db, runID)
		if err != nil {
			return err
		}
		found = true
		return nil
	})
	if err != nil {
		return ManifestMeta{}, err
	}
	if !found {
		meta.Uninspected = []UninspectedFamily{}
		imported, err := s.metadataValue(ctx, metadataExportBase)
		if err != nil {
			return ManifestMeta{}, err
		}
		if imported != "" {
			var recorded ManifestMeta
			if err := json.Unmarshal([]byte(imported), &recorded); err != nil {
				return ManifestMeta{}, fmt.Errorf("decode recorded base info: %w", err)
			}
			// A recorded base without lane health counts as none: identifiers are
			// never re-emitted beside lane health they did not come with.
			if recorded.Uninspected != nil {
				return recorded, nil
			}
		}
	}
	return meta, nil
}

// coverageSummary reports the latest accounted coverage per unit key. A store
// restored from an export holds no runs, so it re-emits the summary it imported.
func (s *Store) coverageSummary(ctx context.Context) ([]byte, error) {
	type unit struct {
		Key    string `json:"key"`
		Status string `json:"status"`
		RunID  string `json:"runId"`
	}
	units := []unit{}
	err := s.withRead(ctx, func(db *sql.DB) error {
		statuses := domain.AccountedStatuses()
		args := []any{domain.RunComplete}
		for _, status := range statuses {
			args = append(args, status)
		}
		rows, err := db.QueryContext(ctx, `SELECT c.coverage_key, c.status, c.run_id
			FROM coverage c JOIN runs r ON r.run_id = c.run_id
			WHERE r.state = ? AND r.finalized = 1 AND c.status IN (`+strings.TrimSuffix(strings.Repeat("?,", len(statuses)), ",")+`)
			ORDER BY c.coverage_key, r.finished_at DESC, r.started_at DESC, c.run_id`, args...)
		if err != nil {
			return fmt.Errorf("query coverage summary: %w", err)
		}
		defer rows.Close()
		last := ""
		for rows.Next() {
			var u unit
			if err := rows.Scan(&u.Key, &u.Status, &u.RunID); err != nil {
				return fmt.Errorf("scan coverage summary: %w", err)
			}
			if u.Key != last {
				units = append(units, u)
				last = u.Key
			}
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	if len(units) == 0 {
		imported, err := s.metadataValue(ctx, metadataCoverageSummary)
		if err != nil {
			return nil, err
		}
		if imported != "" {
			return []byte(imported), nil
		}
	}
	return CanonicalLine(map[string]any{"schema": "scout.coverage-summary/v1", "units": units})
}

func (s *Store) metadataValue(ctx context.Context, key string) (string, error) {
	var value string
	err := s.withRead(ctx, func(db *sql.DB) error {
		err := db.QueryRowContext(ctx, `SELECT value FROM metadata WHERE key = ?`, key).Scan(&value)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	})
	if err != nil {
		return "", fmt.Errorf("read metadata %q: %w", key, err)
	}
	return value, nil
}

// Import rebuilds the ledger from state/v1 files into this fresh store. It
// preserves dispositions and audit history. Observations, runs and artifacts are
// not part of the export, so imported issues keep their observation IDs as
// opaque references.
func (s *Store) Import(ctx context.Context, files map[string][]byte) error {
	if s.readOnly {
		return errors.New("import requires a writable store")
	}
	if err := verifyManifest(files); err != nil {
		return err
	}
	meta, err := ReadManifestMeta(files[ExportManifest])
	if err != nil {
		return err
	}
	issues, err := DecodeIssues(files)
	if err != nil {
		return err
	}
	var blocks []domain.WouldBlock
	for number, line := range SplitLines(files[ExportWouldBlock]) {
		var row domain.WouldBlock
		if err := decode(line, &row); err != nil {
			return fmt.Errorf("would_block.jsonl line %d: %w", number+1, err)
		}
		if err := row.Validate(); err != nil {
			return fmt.Errorf("would_block.jsonl line %d: %w", number+1, err)
		}
		blocks = append(blocks, row)
	}
	signals, err := decodeSignals(files)
	if err != nil {
		return err
	}
	areaList, err := decodeAreas(files)
	if err != nil {
		return err
	}
	return s.importRows(ctx, meta, issues, blocks, signals, files[ExportCoverageSummary], areaList, files[ExportMappingRefresh])
}

// decodeSignals reads every signals/*.jsonl file of an export, validated, in
// identity order.
func decodeSignals(files map[string][]byte) ([]domain.Signal, error) {
	var paths []string
	for path := range files {
		if strings.HasPrefix(path, ExportSignalsDir) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	var signals []domain.Signal
	for _, path := range paths {
		for number, line := range SplitLines(files[path]) {
			var signal domain.Signal
			if err := decode(line, &signal); err != nil {
				return nil, fmt.Errorf("%s line %d: %w", path, number+1, err)
			}
			if err := signal.Validate(); err != nil {
				return nil, fmt.Errorf("%s line %d: %w", path, number+1, err)
			}
			signals = append(signals, signal)
		}
	}
	SortSignals(signals)
	return signals, nil
}

// DecodeIssues verifies an export and reassembles its issues (issue rows joined
// with their dispositions and audit history), validated, sorted by issue ID.
func DecodeIssues(files map[string][]byte) ([]domain.Issue, error) {
	if err := verifyManifest(files); err != nil {
		return nil, err
	}
	dispositions := map[string]map[string]any{}
	for number, line := range SplitLines(files[ExportDispositions]) {
		row := map[string]any{}
		if err := unmarshalNumbers(line, &row); err != nil {
			return nil, fmt.Errorf("dispositions.jsonl line %d: %w", number+1, err)
		}
		id, _ := row["issueId"].(string)
		if id == "" || dispositions[id] != nil {
			return nil, fmt.Errorf("dispositions.jsonl line %d: missing or duplicate issueId", number+1)
		}
		dispositions[id] = row
	}
	var issues []domain.Issue
	seen := map[string]bool{}
	for number, line := range SplitLines(files[ExportIssues]) {
		row := map[string]any{}
		if err := unmarshalNumbers(line, &row); err != nil {
			return nil, fmt.Errorf("issues.jsonl line %d: %w", number+1, err)
		}
		id, _ := row["issueId"].(string)
		disposition, ok := dispositions[id]
		if !ok || seen[id] {
			return nil, fmt.Errorf("issues.jsonl line %d: issue %q has no unique disposition", number+1, id)
		}
		seen[id] = true
		for key, value := range disposition {
			row[key] = value
		}
		merged, err := json.Marshal(row)
		if err != nil {
			return nil, fmt.Errorf("encode imported issue: %w", err)
		}
		var issue domain.Issue
		if err := decode(merged, &issue); err != nil {
			return nil, err
		}
		if err := issue.Validate(); err != nil {
			return nil, fmt.Errorf("imported issue %q: %w", id, err)
		}
		issues = append(issues, issue)
	}
	if len(seen) != len(dispositions) {
		return nil, errors.New("dispositions.jsonl has rows without an issue")
	}
	return issues, nil
}

func (s *Store) importRows(ctx context.Context, meta ManifestMeta, issues []domain.Issue, blocks []domain.WouldBlock, signals []domain.Signal, summary []byte, areaList []areas.Area, refresh []byte) error {
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM issues`).Scan(&count); err != nil {
			return fmt.Errorf("count issues: %w", err)
		}
		if count != 0 {
			return ErrImportNotEmpty
		}
		if meta.RepositoryID != "" {
			if err := s.ensureRepositoryIdentity(ctx, tx, meta.RepositoryID); err != nil {
				return err
			}
		}
		for _, issue := range issues {
			payload, err := encode(issue)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO issues(issue_id, state, updated_at, payload) VALUES (?, ?, ?, ?)`, issue.IssueID, issue.State, unixMillis(issue.UpdatedAt), payload); err != nil {
				return fmt.Errorf("import issue %q: %w", issue.IssueID, err)
			}
			if err := insertIssueHistoryTx(ctx, tx, issue); err != nil {
				return err
			}
		}
		for _, row := range blocks {
			var triaged any
			if row.TriagedAt != nil {
				triaged = unixMillis(*row.TriagedAt)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO would_block(id, head_revision, base_revision, issue_id, check_id, rule_version, path, attribution, profile, triage, triage_actor, triage_reason, recorded_at, triaged_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				row.ID, row.HeadRevision, row.BaseRevision, row.IssueID, row.CheckID, row.RuleVersion, row.Path, string(row.Attribution), row.Profile, row.Triage, row.TriageActor, row.TriageReason, unixMillis(row.RecordedAt), triaged); err != nil {
				return fmt.Errorf("import would_block %q: %w", row.ID, err)
			}
		}
		if err := insertSignals(ctx, tx, signals); err != nil {
			return err
		}
		if err := insertAreas(ctx, tx, areaList); err != nil {
			return err
		}
		if err := importMappingRefresh(ctx, tx, refresh); err != nil {
			return err
		}
		// The comparison identifiers travel with lane health: a manifest that
		// carries `uninspected` records both, any other manifest records none and
		// clears a base left by an earlier import.
		if meta.Uninspected != nil {
			baseJSON, err := json.Marshal(meta)
			if err != nil {
				return fmt.Errorf("encode imported base info: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO metadata(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, metadataExportBase, string(baseJSON)); err != nil {
				return fmt.Errorf("record imported base info: %w", err)
			}
		} else if _, err := tx.ExecContext(ctx, `DELETE FROM metadata WHERE key = ?`, metadataExportBase); err != nil {
			return fmt.Errorf("clear imported base info: %w", err)
		}
		if len(summary) != 0 {
			if _, err := tx.ExecContext(ctx, `INSERT INTO metadata(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, metadataCoverageSummary, string(summary)); err != nil {
				return fmt.Errorf("record imported coverage summary: %w", err)
			}
		}
		return nil
	})
}

// verifyManifest checks every listed file is present and matches its digest,
// and that no unlisted state/v1 file rides along.
func verifyManifest(files map[string][]byte) error {
	manifest, ok := files[ExportManifest]
	if !ok {
		return errors.New("export has no manifest.json")
	}
	var parsed struct {
		Schema string `json:"schema"`
		Files  []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"files"`
	}
	if err := json.Unmarshal(manifest, &parsed); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	if parsed.Schema != ExportSchema {
		return fmt.Errorf("unsupported export schema %q", parsed.Schema)
	}
	listed := map[string]bool{ExportManifest: true}
	for _, entry := range parsed.Files {
		data, ok := files[entry.Path]
		if !ok {
			return fmt.Errorf("export is missing %s", entry.Path)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != entry.SHA256 {
			return fmt.Errorf("export file %s does not match its manifest digest", entry.Path)
		}
		listed[entry.Path] = true
	}
	for path := range files {
		if !listed[path] {
			return fmt.Errorf("export file %s is not listed in the manifest", path)
		}
	}
	for _, required := range []string{ExportIssues, ExportDispositions} {
		if !listed[required] {
			return fmt.Errorf("export manifest does not list %s", required)
		}
	}
	return nil
}

// WriteExport writes files beneath dir, replacing any previous dir/state/v1.
func WriteExport(dir string, files map[string][]byte) error {
	if err := os.RemoveAll(filepath.Join(dir, filepath.FromSlash(ExportDir))); err != nil {
		return fmt.Errorf("clear export directory: %w", err)
	}
	for path, data := range files {
		target := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create export directory: %w", err)
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return fmt.Errorf("write export file: %w", err)
		}
	}
	return nil
}

// ReadExport reads every regular file beneath dir/state/v1.
func ReadExport(dir string) (map[string][]byte, error) {
	base := filepath.Join(dir, filepath.FromSlash(ExportDir))
	files := map[string][]byte{}
	err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("export contains non-regular file %s", path)
		}
		relative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = data
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read export: %w", err)
	}
	return files, nil
}

func unmarshalNumbers(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode record: %w", err)
	}
	return nil
}

// CanonicalLine renders value as compact JSON with sorted keys at every level
// (map keys marshal sorted), no HTML escaping, and one trailing LF.
func CanonicalLine(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, fmt.Errorf("encode export record: %w", err)
	}
	return buffer.Bytes(), nil
}

func joinLines(lines [][]byte) []byte {
	return bytes.Join(lines, nil)
}

// SplitLines splits export data into its non-blank lines.
func SplitLines(data []byte) [][]byte {
	var lines [][]byte
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) != 0 {
			lines = append(lines, line)
		}
	}
	return lines
}
