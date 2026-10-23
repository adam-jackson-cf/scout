package analyzers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"scout/internal/domain"
)

// Runner executes exactly one registered adapter contract.
type Runner struct {
	spec  Spec
	runID string

	// verifiedBinary is the resolved scanner path whose version was already
	// verified by this runner; one Runner is shared for a whole assessment run.
	verifyMu       sync.Mutex
	verifiedBinary string
}

func NewRunner(spec Spec, runID string) (*Runner, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	if !safeName(runID) {
		return nil, errors.New("runner requires a safe run ID")
	}
	return &Runner{spec: spec, runID: runID}, nil
}

type sourceInput struct {
	manifest ManifestInput
}

// Run scans selected inputs. A scanner that stages its batch and attests per
// file (Gitleaks) is retried one input per process for every input the batch
// failed to prove, so a batch never lowers coverage below the single-file runs.
func (r *Runner) Run(ctx context.Context, root string, selected []domain.File) (domain.AdapterResult, []domain.Observation, error) {
	result, found, err := r.runBatch(ctx, root, selected)
	if err == nil || len(selected) < 2 || !r.stagesInputs() {
		return result, found, err
	}
	return r.retryUnproven(ctx, root, selected, result, found, err)
}

func (r *Runner) stagesInputs() bool {
	for _, arg := range r.spec.Binary.Args {
		if arg == "{staged_input_dir}" {
			return true
		}
	}
	return false
}

// retryUnproven rescans, one input per process, each input the batch left
// incomplete for a proof reason, and replaces that input's coverage row and
// observations with the single-file outcome. Timeouts, missing tools and
// setup failures are not retried. The batch is completed only when every row
// ends inspected or not applicable.
func (r *Runner) retryUnproven(ctx context.Context, root string, selected []domain.File, result domain.AdapterResult, found []domain.Observation, batchErr error) (domain.AdapterResult, []domain.Observation, error) {
	retryable := map[string]struct{}{
		"declared input was not completely inspected":       {},
		"scanner output failed structured validation":       {},
		"scanner output exceeded the declared bound":        {},
		"scanner evidence failed current-source validation": {},
	}
	rows := make(map[string]int, len(result.Coverage))
	for index, row := range result.Coverage {
		if _, ok := retryable[row.Reason]; ok && row.Status == domain.CoverageIncomplete {
			rows[row.Key.FilePath] = index
		}
	}
	if len(rows) == 0 {
		return result, found, batchErr
	}
	retried := make(map[string]struct{}, len(rows))
	replaced := make(map[string][]domain.Observation, len(rows))
	for _, file := range selected {
		index, ok := rows[file.Path]
		if !ok {
			continue
		}
		if _, done := retried[file.Path]; done {
			continue
		}
		retried[file.Path] = struct{}{}
		single, observations, _ := r.runBatch(ctx, root, []domain.File{file})
		for _, row := range single.Coverage {
			if row.Key.FilePath == file.Path {
				row.AdapterResultIDs = []string{result.ResultID}
				result.Coverage[index] = row
			}
		}
		replaced[file.Path] = observations
	}
	kept := make([]domain.Observation, 0, len(found))
	for _, observation := range found {
		if len(observation.Evidence) > 0 {
			if _, ok := retried[observation.Evidence[0].Location.Path]; ok {
				continue
			}
		}
		kept = append(kept, observation)
	}
	for _, file := range selected {
		kept = append(kept, replaced[file.Path]...)
		delete(replaced, file.Path)
	}
	complete := true
	for _, row := range result.Coverage {
		if row.Status != domain.CoverageInspected && row.Status != domain.CoverageNotApplicable {
			complete = false
		}
	}
	for index := range result.Coverage {
		var ids []string
		for _, observation := range kept {
			if observation.Evidence[0].Location.Path == result.Coverage[index].Key.FilePath {
				ids = append(ids, observation.ObservationID)
			}
		}
		result.Coverage[index].ObservationIDs = ids
	}
	result.Observations = observationIDs(kept)
	finished := time.Now().UTC()
	result.FinishedAt = &finished
	if complete {
		result.Status, result.Failures = domain.AdapterCompleted, nil
		return result, kept, nil
	}
	return result, kept, batchErr
}

// runBatch returns all validated, sanitized observations even when the adapter is
// partial or failed. It never returns raw scanner output in its error value.
func (r *Runner) runBatch(ctx context.Context, root string, selected []domain.File) (domain.AdapterResult, []domain.Observation, error) {
	started := time.Now().UTC()
	resultID := digestID(r.runID, r.spec.ID, started.Format(time.RFC3339Nano))
	result := r.newResult(resultID, started)
	root, err := canonicalRoot(root)
	if err != nil {
		return r.finish(&result, domain.AdapterFailed, nil, "invalid_root", "assessment root is unavailable"), nil, errors.New("adapter root is unavailable")
	}

	accepted, coverage, acceptedByPath, eligibleCount, inputErr := r.selectInputs(root, selected, resultID)
	if len(accepted) == 0 {
		status, code, message := domain.AdapterNoApplicableInput, "no_applicable_input", "no applicable selected inputs"
		if inputErr {
			status, code, message = domain.AdapterPartial, "input_unavailable", "eligible input could not be safely read"
		} else if len(selected) != 0 && eligibleCount == 0 {
			status, code, message = domain.AdapterIneligible, "ineligible", "selected inputs are ineligible for this adapter"
		}
		if status == domain.AdapterNoApplicableInput || status == domain.AdapterIneligible {
			return r.finish(&result, status, coverage, "", ""), nil, nil
		}
		return r.finish(&result, status, coverage, code, message), nil, errors.New("adapter input is incomplete")
	}

	state, err := newScannerState()
	if err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "private scanner state could not be created")
		return r.finish(&result, domain.AdapterFailed, coverage, "state", "private scanner state could not be created"), nil, errors.New("adapter state setup failed")
	}
	defer os.RemoveAll(state)
	bundledRules, err := r.materializeBundledRules(state)
	if err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "bundled scanner rules could not be materialized")
		return r.finish(&result, domain.AdapterFailed, coverage, "bundled_rules", "bundled scanner rules could not be materialized"), nil, errors.New("bundled scanner rules unavailable")
	}

	footprint, err := r.footprint(root)
	if err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "declared project footprint is unavailable")
		return r.finish(&result, domain.AdapterIneligible, coverage, "project_footprint", "declared project footprint is unavailable"), nil, nil
	}
	manifest := InputManifest{Version: manifestVersion, Inputs: manifests(accepted), ProjectFootprint: footprint, RuleFiles: r.spec.Rules.Files, ConfigFiles: r.spec.Rules.ConfigFiles, BundledRules: bundledRules.digests}
	manifestPath, err := r.writeManifest(state, manifest)
	if err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "input manifest could not be safely created")
		return r.finish(&result, domain.AdapterFailed, coverage, "manifest", "input manifest could not be safely created"), nil, errors.New("adapter input manifest failed")
	}
	defer os.Remove(manifestPath)
	if _, err := sandboxBinary(); err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "required scanner sandbox is not installed")
		return r.finish(&result, domain.AdapterMissingTool, coverage, "missing_sandbox", "required scanner sandbox is not installed"), nil, errors.New("required scanner sandbox is not installed")
	}
	binary, err := exec.LookPath(r.spec.Binary.Path)
	if err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "required scanner binary is not installed")
		return r.finish(&result, domain.AdapterMissingTool, coverage, "missing_tool", "required scanner binary is not installed"), nil, errors.New("required scanner binary is not installed")
	}
	if err := externalScannerBinary(root, binary); err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "scanner launcher resolves inside the assessed repository")
		return r.finish(&result, domain.AdapterFailed, coverage, "project_script", "scanner launcher must not come from the assessed repository"), nil, errors.New("project scanner launcher is prohibited")
	}
	if err := r.verifyVersion(ctx, root, state, bundledRules.root, binary); err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "required scanner version is unavailable")
		return r.finish(&result, domain.AdapterFailed, coverage, "version", "required scanner version is unavailable"), nil, errors.New("required scanner version is unavailable")
	}

	attested, scanRoot, stageDir := acceptedByPath, root, ""
	if r.stagesInputs() {
		var stageErr error
		if stageDir, attested, stageErr = stageGitleaksInputs(root, state, acceptedByPath); stageErr != nil {
			coverage = markAcceptedIncomplete(coverage, acceptedByPath, "private input staging failed")
			return r.finish(&result, domain.AdapterFailed, coverage, "staging", "private input staging failed"), nil, errors.New("adapter input staging failed")
		}
		scanRoot = stageDir
	}
	args, err := expandArgs(r.spec.Binary.Args, root, manifestPath, manifests(accepted), bundledRules.paths, stageDir)
	if err != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "declared scanner invocation is invalid")
		return r.finish(&result, domain.AdapterFailed, coverage, "invocation", "declared scanner invocation is invalid"), nil, errors.New("adapter invocation is invalid")
	}
	runContext, cancel := context.WithTimeout(ctx, r.spec.Timeout)
	defer cancel()
	stdout, stderr, runErr := r.runCommand(runContext, root, state, bundledRules.root, binary, args)
	if runContext.Err() == context.DeadlineExceeded {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "scanner timed out")
		return r.finish(&result, domain.AdapterTimedOut, coverage, "timed_out", "scanner timed out"), nil, errors.New("scanner timed out")
	}
	if stdout.exceeded || stderr.exceeded {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "scanner output exceeded the declared bound")
		return r.finish(&result, domain.AdapterFailed, coverage, "output_limit", "scanner output exceeded the declared bound"), nil, errors.New("scanner output exceeded the declared bound")
	}

	parsed, parseErr := parseOutput(r.spec, stdout.Bytes(), scanRoot)
	if parseErr != nil {
		coverage = markAcceptedIncomplete(coverage, acceptedByPath, "scanner output failed structured validation")
		return r.finish(&result, domain.AdapterFailed, coverage, "output_schema", "scanner output failed structured validation"), nil, errors.New("scanner output failed structured validation")
	}
	if r.spec.ID == "security.gitleaks.default-credentials" {
		attestGitleaksInput(&parsed, attested, stderr.stderr.Bytes(), stageDir)
	}
	if r.spec.Binary.Path == "ruff" {
		r.attestRuffInputs(ctx, root, state, bundledRules.root, binary, args, acceptedByPath, &parsed)
	}
	observations, rejected := r.observations(root, acceptedByPath, parsed.findings, started)
	for path := range rejected {
		coverage = markPathIncomplete(coverage, path, "scanner evidence failed current-source validation")
	}
	status, code, message := r.resultStatus(runErr, parsed, acceptedByPath, rejected, inputErr)
	coverage = completeCoverage(coverage, acceptedByPath, parsed, status)
	byPath := make(map[string][]string, len(observations))
	for _, observation := range observations {
		path := observation.Evidence[0].Location.Path
		byPath[path] = append(byPath[path], observation.ObservationID)
	}
	for index := range coverage {
		coverage[index].ObservationIDs = byPath[coverage[index].Key.FilePath]
	}
	result = r.finish(&result, status, coverage, code, message)
	result.Observations = observationIDs(observations)
	if status == domain.AdapterCompleted || status == domain.AdapterNoApplicableInput || status == domain.AdapterIneligible {
		return result, observations, nil
	}
	return result, observations, errors.New("adapter completed with incomplete coverage")
}

func (r *Runner) newResult(id string, started time.Time) domain.AdapterResult {
	return domain.AdapterResult{ResultID: id, RunID: r.runID, Adapter: r.spec.ID, AdapterVersion: r.spec.Binary.Version, RuleVersion: r.spec.Rules.Version, Command: filepath.Base(r.spec.Binary.Path), StartedAt: started}
}

func (r *Runner) finish(result *domain.AdapterResult, status domain.AdapterStatus, coverage []domain.Coverage, code, message string) domain.AdapterResult {
	finished := time.Now().UTC()
	result.Status, result.FinishedAt, result.Coverage = status, &finished, coverage
	if message != "" {
		result.Failures = []domain.Failure{{Stage: domain.StageFileAnalysis, Code: code, Message: message, Retryable: status == domain.AdapterTimedOut || status == domain.AdapterMissingTool}}
	}
	return *result
}

func externalScannerBinary(root, binary string) error {
	resolved, err := filepath.EvalSymlinks(binary)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(root, resolved)
	if err != nil || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))) {
		return errors.New("scanner launcher is inside assessed repository")
	}
	return nil
}

func canonicalRoot(root string) (string, error) {
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return "", errors.New("root is not a directory")
	}
	return resolved, nil
}

func (r *Runner) selectInputs(root string, selected []domain.File, resultID string) ([]sourceInput, []domain.Coverage, map[string]sourceInput, int, bool) {
	accepted := make([]sourceInput, 0, len(selected))
	coverage := make([]domain.Coverage, 0, len(selected))
	byPath := make(map[string]sourceInput)
	var total int64
	eligibleCount, inputErr := 0, false
	for _, file := range selected {
		if err := file.Validate(); err != nil {
			inputErr = true
			continue
		}
		key := domain.CoverageKey{FilePath: file.Path, Family: r.spec.Family, Category: r.spec.Category}
		row := domain.Coverage{Key: key, AdapterResultIDs: []string{resultID}}
		if !r.spec.eligible(file) {
			row.Status, row.Reason = domain.CoverageNotApplicable, "adapter language or extension is ineligible"
			coverage = append(coverage, row)
			continue
		}
		eligibleCount++
		if _, duplicate := byPath[file.Path]; duplicate {
			row.Status, row.Reason = domain.CoverageIncomplete, "selected input is duplicated"
			coverage = append(coverage, row)
			inputErr = true
			continue
		}
		input, err := fingerprint(root, file.Path, file.Language, r.spec.Inputs.MaxFileBytes)
		if err != nil {
			row.Status, row.Reason = domain.CoverageIncomplete, "eligible input is unreadable or exceeds the declared bound"
			coverage = append(coverage, row)
			inputErr = true
			continue
		}
		if len(accepted) == r.spec.Inputs.MaxFiles || total > r.spec.Inputs.MaxTotalBytes-input.manifest.Bytes {
			row.Status, row.Reason = domain.CoverageIncomplete, "eligible input exceeds the declared scan bound"
			coverage = append(coverage, row)
			inputErr = true
			continue
		}
		total += input.manifest.Bytes
		accepted, byPath[file.Path] = append(accepted, input), input
		coverage = append(coverage, row)
	}
	return accepted, coverage, byPath, eligibleCount, inputErr
}

func fingerprint(root, path, language string, maximum int64) (sourceInput, error) {
	if err := validRelativePath(path); err != nil {
		return sourceInput{}, err
	}
	joined := filepath.Join(root, filepath.FromSlash(path))
	resolved, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return sourceInput{}, err
	}
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return sourceInput{}, errors.New("input escapes root")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maximum {
		return sourceInput{}, errors.New("input is not a bounded regular file")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return sourceInput{}, err
	}
	defer file.Close()
	hash := sha256.New()
	lines, size := 1, int64(0)
	buffer := make([]byte, 32*1024)
	for {
		count, readErr := file.Read(buffer)
		if count > 0 {
			size += int64(count)
			if size > maximum {
				return sourceInput{}, errors.New("input exceeds bound")
			}
			hash.Write(buffer[:count])
			lines += bytes.Count(buffer[:count], []byte{'\n'})
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return sourceInput{}, readErr
		}
	}
	return sourceInput{manifest: ManifestInput{Path: path, Language: language, Digest: hex.EncodeToString(hash.Sum(nil)), Bytes: size, Lines: lines}}, nil
}

func (r *Runner) footprint(root string) ([]string, error) {
	paths := append(append(append([]string{}, r.spec.Inputs.ProjectFootprint...), r.spec.Rules.Files...), r.spec.Rules.ConfigFiles...)
	for _, path := range paths {
		if _, err := fingerprint(root, path, "", r.spec.Inputs.MaxFileBytes); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

func manifests(inputs []sourceInput) []ManifestInput {
	result := make([]ManifestInput, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, input.manifest)
	}
	return result
}

func newScannerState() (string, error) {
	state, err := os.MkdirTemp("", "scout-analyzer-")
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(state)
	if err != nil {
		os.RemoveAll(state)
		return "", err
	}
	state = resolved
	if err := os.Chmod(state, 0700); err != nil {
		os.RemoveAll(state)
		return "", err
	}
	return state, nil
}

type materializedRules struct {
	root    string
	paths   []string
	digests []RuleDigest
}

func (r *Runner) materializeBundledRules(state string) (materializedRules, error) {
	result := materializedRules{
		paths:   make([]string, 0, len(r.spec.Rules.BundledFiles)),
		digests: make([]RuleDigest, 0, len(r.spec.Rules.BundledFiles)),
	}
	for _, rulePath := range r.spec.Rules.BundledFiles {
		if err := validRelativePath(rulePath); err != nil {
			return materializedRules{}, err
		}
		content, err := bundledRule(rulePath)
		if err != nil || len(content) == 0 {
			return materializedRules{}, errors.New("bundled rule is unavailable")
		}
		if result.root == "" {
			result.root = filepath.Join(state, "rules")
		}
		path := filepath.Join(result.root, filepath.FromSlash(rulePath))
		relative, err := filepath.Rel(state, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return materializedRules{}, errors.New("bundled rule escapes private state")
		}
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return materializedRules{}, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return materializedRules{}, err
		}
		if _, err := file.Write(content); err != nil {
			file.Close()
			return materializedRules{}, err
		}
		if err := file.Close(); err != nil {
			return materializedRules{}, err
		}
		if err := os.Chmod(path, 0400); err != nil {
			return materializedRules{}, err
		}
		sum := sha256.Sum256(content)
		result.paths = append(result.paths, path)
		result.digests = append(result.digests, RuleDigest{Path: rulePath, Digest: hex.EncodeToString(sum[:])})
	}
	return result, nil
}

func (r *Runner) writeManifest(state string, manifest InputManifest) (string, error) {
	encoded, err := json.Marshal(manifest)
	if err != nil || len(encoded) > r.spec.Inputs.MaxManifestBytes {
		return "", errors.New("manifest exceeds bound")
	}
	file, err := os.CreateTemp(state, "adapter-manifest-*.json")
	if err != nil {
		return "", err
	}
	path := file.Name()
	if err := file.Chmod(0600); err != nil {
		file.Close()
		os.Remove(path)
		return "", errors.New("manifest permission change failed")
	}
	if _, err := file.Write(encoded); err != nil {
		file.Close()
		os.Remove(path)
		return "", errors.New("manifest write failed")
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", errors.New("manifest write failed")
	}
	return path, nil
}

func expandArgs(args []string, root, manifest string, inputs []ManifestInput, bundledRules []string, stageDir string) ([]string, error) {
	result := make([]string, 0, len(args)+len(inputs)+len(bundledRules))
	for _, arg := range args {
		switch arg {
		case "{input_manifest}":
			result = append(result, manifest)
		case "{input_paths}":
			for _, input := range inputs {
				result = append(result, input.Path)
			}
		case "{absolute_input_paths}":
			for _, input := range inputs {
				result = append(result, filepath.Join(root, input.Path))
			}
		case "{staged_input_dir}":
			result = append(result, stageDir)
		case "{bundled_rule_files}":
			result = append(result, bundledRules...)
		default:
			if strings.Contains(arg, "{") || strings.Contains(arg, "}") {
				return nil, errors.New("unknown invocation placeholder")
			}
			result = append(result, arg)
		}
	}
	return result, nil
}

func (r *Runner) verifyVersion(ctx context.Context, root, state, rulesRoot, binary string) error {
	r.verifyMu.Lock()
	defer r.verifyMu.Unlock()
	if r.verifiedBinary == binary {
		return nil
	}
	if err := r.checkVersion(ctx, root, state, rulesRoot, binary); err != nil {
		return err
	}
	r.verifiedBinary = binary
	return nil
}

func (r *Runner) checkVersion(ctx context.Context, root, state, rulesRoot, binary string) error {
	check, cancel := context.WithTimeout(ctx, minDuration(r.spec.Timeout, 10*time.Second))
	defer cancel()
	stdout, stderr, err := r.runCommand(check, root, state, rulesRoot, binary, r.spec.Binary.VersionArgs)
	if err != nil || check.Err() != nil || stdout.exceeded || stderr.exceeded {
		return errors.New("version command failed")
	}
	if !containsVersion(stdout.Bytes(), r.spec.Binary.Version) {
		return errors.New("version mismatch")
	}
	return nil
}

func containsVersion(output []byte, version string) bool {
	for _, token := range strings.FieldsFunc(string(output), func(r rune) bool {
		return !(r == '.' || r == '-' || r == '_' || r == '+' || ('0' <= r && r <= '9') || ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z'))
	}) {
		if token == version || strings.TrimPrefix(token, "v") == version {
			return true
		}
	}
	return false
}

func minDuration(left, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}

type outputBuffers struct {
	mu             sync.Mutex
	limit, used    int64
	exceeded       bool
	stdout, stderr bytes.Buffer
}
type outputWriter struct {
	owner       *outputBuffers
	destination *bytes.Buffer
}

func (w outputWriter) Write(value []byte) (int, error) {
	w.owner.mu.Lock()
	defer w.owner.mu.Unlock()
	available := w.owner.limit - w.owner.used
	if available <= 0 {
		w.owner.exceeded = true
		return len(value), nil
	}
	count := int64(len(value))
	if count > available {
		count, w.owner.exceeded = available, true
	}
	w.destination.Write(value[:count])
	w.owner.used += count
	return len(value), nil
}
func (b *outputBuffers) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte(nil), b.stdout.Bytes()...)
}

func (r *Runner) runCommand(ctx context.Context, root, state, rulesRoot, binary string, args []string) (*outputBuffers, *outputBuffers, error) {
	workDir := root
	if r.spec.Binary.PrivateWorkDir {
		workDir = state
	}
	command, err := sandboxedCommand(ctx, workDir, state, rulesRoot, binary, args)
	if err != nil {
		return &outputBuffers{}, &outputBuffers{}, err
	}
	stdout, stderr := &outputBuffers{limit: r.spec.MaxOutput}, &outputBuffers{limit: r.spec.MaxOutput}
	command.Stdout, command.Stderr = outputWriter{stdout, &stdout.stdout}, outputWriter{stderr, &stderr.stderr}
	command.Dir = workDir
	command.Env = scannerEnvironment(state)
	return stdout, stderr, command.Run()
}

// scannerEnvironment deliberately does not inherit the caller environment:
// analyzer subprocesses must not receive provider credentials or user config.
func scannerEnvironment(state string) []string {
	return []string{
		"HOME=" + state,
		"TMPDIR=" + state,
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"SEMGREP_ENABLE_VERSION_CHECK=0",
		"SEMGREP_SEND_METRICS=off",
		"LC_ALL=C",
	}
}

func sandboxBinary() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.LookPath("sandbox-exec")
	case "linux":
		return exec.LookPath("bwrap")
	default:
		return "", errors.New("read-only scanner isolation is unavailable")
	}
}

func sandboxedCommand(ctx context.Context, root, state, rulesRoot, binary string, args []string) (*exec.Cmd, error) {
	sandbox, err := sandboxBinary()
	if err != nil {
		return nil, errors.New("read-only scanner isolation is unavailable")
	}
	if runtime.GOOS == "darwin" {
		commandArgs := append([]string{"-p", darwinSandboxProfile(state, rulesRoot), binary}, args...)
		return exec.CommandContext(ctx, sandbox, commandArgs...), nil
	}
	commandArgs := []string{
		"--die-with-parent",
		"--unshare-net",
		"--ro-bind", "/", "/",
		"--bind", state, state,
	}
	if rulesRoot != "" {
		commandArgs = append(commandArgs, "--ro-bind", rulesRoot, rulesRoot)
	}
	commandArgs = append(commandArgs,
		"--unshare-pid",
		"--proc", "/proc",
		"--dev", "/dev",
		"--chdir", root,
		"--",
		binary,
	)
	commandArgs = append(commandArgs, args...)
	return exec.CommandContext(ctx, sandbox, commandArgs...), nil
}

func darwinSandboxProfile(state, rulesRoot string) string {
	profile := "(version 1)\n" +
		"(allow default)\n" +
		"(deny network*)\n" +
		"(deny file-write* (require-not (subpath " + strconv.Quote(state) + ")))\n"
	if rulesRoot != "" {
		profile += "(deny file-write* (subpath " + strconv.Quote(rulesRoot) + "))\n"
	}
	return profile
}

func (r *Runner) observations(root string, inputs map[string]sourceInput, findings []parsedFinding, observed time.Time) ([]domain.Observation, map[string]struct{}) {
	rejected := make(map[string]struct{})
	observations := make([]domain.Observation, 0, len(findings))
	for _, finding := range findings {
		input, found := inputs[finding.path]
		if !found || !validFindingRange(root, input, finding) {
			rejected[finding.path] = struct{}{}
			continue
		}
		mechanism, uncertainty, quote := "", "", ""
		severity, confidence := finding.severity, finding.confidence
		var err error
		if r.spec.Output.SecretDiagnostics {
			mechanism, severity, confidence = domain.SecretMechanism, domain.SeverityHigh, domain.ConfidenceMedium
			quote, err = sanitizeText(finding.quote, maxEvidenceText)
		} else {
			mechanism, err = sanitizeText(finding.mechanism, maxEvidenceText)
			if err == nil {
				uncertainty, err = sanitizeText(finding.uncertainty, maxEvidenceText)
			}
			if err == nil {
				quote, err = sanitizeText(finding.quote, maxEvidenceText)
			}
		}
		if err != nil || mechanism == "" {
			rejected[finding.path] = struct{}{}
			continue
		}
		location := domain.SourceRange{Path: finding.path, StartLine: finding.startLine, EndLine: finding.endLine, StartByte: finding.startByte, EndByte: finding.endByte}
		sourceFindingID := ""
		if r.spec.Output.SecretDiagnostics {
			// Identity of a secret is a digest of its raw value, computed in memory
			// here and never stored or logged raw, so a moved secret stays one issue.
			sourceFindingID = secretDigest(root, finding)
		} else {
			sourceFindingID = finding.id
		}
		observation := domain.Observation{ObservationID: digestID(r.runID, r.spec.ID, finding.id, finding.ruleID, finding.path, fmt.Sprint(finding.startLine)), RunID: r.runID, Source: r.spec.ID, SourceFindingID: sourceFindingID, Family: r.spec.Family, OriginalCategory: r.spec.Category, RuleID: finding.ruleID, RuleVersion: r.spec.Rules.Version, SourceSeverity: severity, EffectiveSeverity: severity, Confidence: confidence, Mechanism: mechanism, Uncertainty: uncertainty, Evidence: []domain.Evidence{{Location: location, RedactedQuote: quote}}, Validated: true, ObservedAt: observed}
		if err := observation.Validate(); err != nil {
			rejected[finding.path] = struct{}{}
			continue
		}
		observations = append(observations, observation)
	}
	return observations, rejected
}

func validFindingRange(root string, input sourceInput, finding parsedFinding) bool {
	current, err := fingerprint(root, finding.path, input.manifest.Language, input.manifest.Bytes)
	if err != nil || current.manifest.Digest != input.manifest.Digest || finding.startLine <= 0 || finding.endLine < finding.startLine || finding.endLine > current.manifest.Lines {
		return false
	}
	if (finding.startByte == 0) != (finding.endByte == 0) || finding.startByte < 0 || finding.endByte < finding.startByte || int64(finding.endByte) > current.manifest.Bytes {
		return false
	}
	return true
}

func (r *Runner) resultStatus(runErr error, parsed parsedOutput, inputs map[string]sourceInput, rejected map[string]struct{}, inputErr bool) (domain.AdapterStatus, string, string) {
	code := exitCode(runErr)
	meaning, known := r.spec.ExitCodes[code]
	if runErr != nil && code < 0 {
		return domain.AdapterFailed, "execution", "scanner process failed"
	}
	if !known || meaning == ExitFailure {
		return domain.AdapterFailed, "exit_code", "scanner returned an unsuccessful exit code"
	}
	if meaning == ExitPartial || !parsed.success || inputErr || len(rejected) != 0 {
		return domain.AdapterPartial, "partial", "scanner did not complete all declared inputs"
	}
	for path := range inputs {
		if _, errorPath := parsed.processingErrors[path]; errorPath {
			return domain.AdapterPartial, "processing_error", "scanner reported an input processing error"
		}
		if _, done := parsed.processed[path]; !done {
			return domain.AdapterPartial, "input_omission", "scanner omitted a declared input"
		}
	}
	for path := range parsed.processed {
		if _, known := inputs[path]; !known {
			return domain.AdapterPartial, "unknown_input", "scanner reported an undeclared input"
		}
	}
	return domain.AdapterCompleted, "", ""
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}
	return -1
}

func completeCoverage(coverage []domain.Coverage, inputs map[string]sourceInput, parsed parsedOutput, status domain.AdapterStatus) []domain.Coverage {
	for index := range coverage {
		row := &coverage[index]
		input, accepted := inputs[row.Key.FilePath]
		if !accepted || row.Status != "" {
			continue
		}
		if status == domain.AdapterCompleted && parsed.success {
			row.Status = domain.CoverageInspected
			row.InspectedRanges = []domain.SourceRange{{Path: input.manifest.Path, StartLine: 1, EndLine: input.manifest.Lines}}
			continue
		}
		row.Status, row.Reason = domain.CoverageIncomplete, "declared input was not completely inspected"
	}
	return coverage
}

func markAcceptedIncomplete(coverage []domain.Coverage, inputs map[string]sourceInput, reason string) []domain.Coverage {
	for index := range coverage {
		if _, ok := inputs[coverage[index].Key.FilePath]; ok && coverage[index].Status == "" {
			coverage[index].Status, coverage[index].Reason = domain.CoverageIncomplete, reason
		}
	}
	return coverage
}
func markPathIncomplete(coverage []domain.Coverage, path, reason string) []domain.Coverage {
	for index := range coverage {
		if coverage[index].Key.FilePath == path {
			coverage[index].Status, coverage[index].Reason, coverage[index].InspectedRanges = domain.CoverageIncomplete, reason, nil
		}
	}
	return coverage
}
func observationIDs(observations []domain.Observation) []string {
	result := make([]string, 0, len(observations))
	for _, observation := range observations {
		result = append(result, observation.ObservationID)
	}
	return result
}
