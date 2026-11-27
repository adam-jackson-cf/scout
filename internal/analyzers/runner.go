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
	"regexp"
	"strings"
	"sync"
	"time"

	"scout/internal/domain"
	"scout/internal/sandbox"
)

// Coverage reasons the runner writes and reads back. A batch that left an
// input incomplete for one of the proof reasons is rescanned one input per
// process; the retry decision keys on these constants, never on prose elsewhere.
const (
	reasonNotInspected    = "declared input was not completely inspected"
	reasonOutputInvalid   = "scanner output failed structured validation"
	reasonOutputBound     = "scanner output exceeded the declared bound"
	reasonEvidenceInvalid = "scanner evidence failed current-source validation"
)

var proofReasons = map[string]struct{}{
	reasonNotInspected:    {},
	reasonOutputInvalid:   {},
	reasonOutputBound:     {},
	reasonEvidenceInvalid: {},
}

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
	return r.retryUnproven(ctx, root, selected, batchOutcome{result: result, found: found, err: err})
}

func (r *Runner) stagesInputs() bool { return r.spec.gitleaks() }

// retryUnproven rescans, one input per process, each input the batch left
// incomplete for a proof reason, and replaces that input's coverage row and
// observations with the single-file outcome. Timeouts, missing tools and
// setup failures are not retried. The batch is completed only when every row
// ends inspected or not applicable.
type batchOutcome struct {
	result domain.AdapterResult
	found  []domain.Observation
	err    error
}

func (r *Runner) retryUnproven(ctx context.Context, root string, selected []domain.File, outcome batchOutcome) (domain.AdapterResult, []domain.Observation, error) {
	result, found, batchErr := outcome.result, outcome.found, outcome.err
	rows := unprovenRows(result.Coverage)
	if len(rows) == 0 {
		return result, found, batchErr
	}
	retried, replaced := r.rescanSingly(ctx, root, selected, rows, &result)
	kept := withoutRetried(found, retried)
	for _, file := range selected {
		kept = append(kept, replaced[file.Path]...)
		delete(replaced, file.Path)
	}
	complete := coverageSettled(result.Coverage)
	linkObservations(result.Coverage, kept)
	result.Observations = observationIDs(kept)
	finished := time.Now().UTC()
	result.FinishedAt = &finished
	if complete {
		result.Status, result.Failures = domain.AdapterCompleted, nil
		return result, kept, nil
	}
	return result, kept, batchErr
}

func unprovenRows(coverage []domain.Coverage) map[string]int {
	rows := make(map[string]int, len(coverage))
	for index, row := range coverage {
		if _, ok := proofReasons[row.Reason]; ok && row.Status == domain.CoverageIncomplete {
			rows[row.Key.FilePath] = index
		}
	}
	return rows
}

func (r *Runner) rescanSingly(ctx context.Context, root string, selected []domain.File, rows map[string]int, result *domain.AdapterResult) (map[string]struct{}, map[string][]domain.Observation) {
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
	return retried, replaced
}

func withoutRetried(found []domain.Observation, retried map[string]struct{}) []domain.Observation {
	kept := make([]domain.Observation, 0, len(found))
	for _, observation := range found {
		if len(observation.Evidence) > 0 {
			if _, ok := retried[observation.Evidence[0].Location.Path]; ok {
				continue
			}
		}
		kept = append(kept, observation)
	}
	return kept
}

func coverageSettled(coverage []domain.Coverage) bool {
	for _, row := range coverage {
		if row.Status != domain.CoverageInspected && row.Status != domain.CoverageNotApplicable {
			return false
		}
	}
	return true
}

func linkObservations(coverage []domain.Coverage, observations []domain.Observation) {
	byPath := make(map[string][]string, len(observations))
	for _, observation := range observations {
		path := observation.Evidence[0].Location.Path
		byPath[path] = append(byPath[path], observation.ObservationID)
	}
	for index := range coverage {
		coverage[index].ObservationIDs = byPath[coverage[index].Key.FilePath]
	}
}

type inputSelection struct {
	accepted      []sourceInput
	coverage      []domain.Coverage
	byPath        map[string]sourceInput
	eligibleCount int
	inputErr      bool
	totalBytes    int64
}

type scanPlan struct {
	state        string
	manifestPath string
	binary       string
	stageDir     string
	scanRoot     string
	rules        materializedRules
	attested     map[string]sourceInput
	args         []string
}

type scanOutput struct {
	parsed parsedOutput
	stderr []byte
	runErr error
}

type batchFailure struct {
	status  domain.AdapterStatus
	code    string
	reason  string
	message string
	err     string
}

func failure(status domain.AdapterStatus, code, reason, message, err string) *batchFailure {
	return &batchFailure{status: status, code: code, reason: reason, message: message, err: err}
}

type batch struct {
	runner  *Runner
	root    string
	started time.Time
	result  domain.AdapterResult
	inputs  inputSelection
	plan    scanPlan
}

// runBatch returns all validated, sanitized observations even when the adapter is
// partial or failed. It never returns raw scanner output in its error value.
func (r *Runner) runBatch(ctx context.Context, root string, selected []domain.File) (domain.AdapterResult, []domain.Observation, error) {
	started := time.Now().UTC()
	b := &batch{runner: r, started: started, result: r.newResult(digestID(r.runID, r.spec.ID, started.Format(time.RFC3339Nano)), started)}
	canonical, err := canonicalRoot(root)
	if err != nil {
		return r.finish(&b.result, domain.AdapterFailed, nil, "invalid_root", "assessment root is unavailable"), nil, errors.New("adapter root is unavailable")
	}
	b.root = canonical
	b.inputs = r.selectInputs(canonical, selected, b.result.ResultID)
	if len(b.inputs.accepted) == 0 {
		return b.finishWithoutInputs(len(selected))
	}
	state, err := sandbox.NewState("scout-analyzer-")
	if err != nil {
		return b.fail(failure(domain.AdapterFailed, "state", "private scanner state could not be created", "private scanner state could not be created", "adapter state setup failed"))
	}
	b.plan.state = state
	defer b.removeState()
	if failed := b.prepareInputs(); failed != nil {
		return b.fail(failed)
	}
	if failed := b.prepareScanner(ctx); failed != nil {
		return b.fail(failed)
	}
	output, failed := b.execute(ctx)
	if failed != nil {
		return b.fail(failed)
	}
	return b.complete(ctx, output)
}

func (b *batch) removeState() {
	if b.plan.manifestPath != "" {
		_ = os.Remove(b.plan.manifestPath)
	}
	_ = os.RemoveAll(b.plan.state)
}

func (b *batch) finishWithoutInputs(selectedCount int) (domain.AdapterResult, []domain.Observation, error) {
	status, code, message := domain.AdapterNoApplicableInput, "no_applicable_input", "no applicable selected inputs"
	if b.inputs.inputErr {
		status, code, message = domain.AdapterPartial, "input_unavailable", "eligible input could not be safely read"
	} else if selectedCount != 0 && b.inputs.eligibleCount == 0 {
		status, code, message = domain.AdapterIneligible, "ineligible", "selected inputs are ineligible for this adapter"
	}
	if status == domain.AdapterNoApplicableInput || status == domain.AdapterIneligible {
		return b.runner.finish(&b.result, status, b.inputs.coverage, "", ""), nil, nil
	}
	return b.runner.finish(&b.result, status, b.inputs.coverage, code, message), nil, errors.New("adapter input is incomplete")
}

func (b *batch) fail(failed *batchFailure) (domain.AdapterResult, []domain.Observation, error) {
	coverage := markAcceptedIncomplete(b.inputs.coverage, b.inputs.byPath, failed.reason)
	var err error
	if failed.err != "" {
		err = errors.New(failed.err)
	}
	return b.runner.finish(&b.result, failed.status, coverage, failed.code, failed.message), nil, err
}

func (b *batch) prepareInputs() *batchFailure {
	r := b.runner
	rules, err := r.materializeBundledRules(b.plan.state)
	if err != nil {
		return failure(domain.AdapterFailed, "bundled_rules", "bundled scanner rules could not be materialized", "bundled scanner rules could not be materialized", "bundled scanner rules unavailable")
	}
	b.plan.rules = rules
	footprint, err := r.footprint(b.root)
	if err != nil {
		return failure(domain.AdapterIneligible, "project_footprint", "declared project footprint is unavailable", "declared project footprint is unavailable", "")
	}
	manifest := InputManifest{Version: manifestVersion, Inputs: manifests(b.inputs.accepted), ProjectFootprint: footprint, RuleFiles: r.spec.Rules.Files, ConfigFiles: r.spec.Rules.ConfigFiles, BundledRules: rules.digests}
	manifestPath, err := r.writeManifest(b.plan.state, manifest)
	if err != nil {
		return failure(domain.AdapterFailed, "manifest", "input manifest could not be safely created", "input manifest could not be safely created", "adapter input manifest failed")
	}
	b.plan.manifestPath = manifestPath
	return nil
}

func (b *batch) prepareScanner(ctx context.Context) *batchFailure {
	r := b.runner
	if err := sandbox.Available(); err != nil {
		return failure(domain.AdapterMissingTool, "missing_sandbox", "required scanner sandbox is not installed", "required scanner sandbox is not installed", "required scanner sandbox is not installed")
	}
	binary, err := exec.LookPath(r.spec.Binary.Path)
	if err != nil {
		return failure(domain.AdapterMissingTool, "missing_tool", "required scanner binary is not installed", "required scanner binary is not installed", "required scanner binary is not installed")
	}
	if err := externalScannerBinary(b.root, binary); err != nil {
		return failure(domain.AdapterFailed, "project_script", "scanner launcher resolves inside the assessed repository", "scanner launcher must not come from the assessed repository", "project scanner launcher is prohibited")
	}
	if err := r.verifyVersion(ctx, b.root, b.plan.state, b.plan.rules.root, binary); err != nil {
		return failure(domain.AdapterFailed, "version", "required scanner version is unavailable", "required scanner version is unavailable", "required scanner version is unavailable")
	}
	b.plan.binary = binary
	return b.prepareInvocation()
}

func (b *batch) prepareInvocation() *batchFailure {
	r := b.runner
	b.plan.attested, b.plan.scanRoot = b.inputs.byPath, b.root
	if r.stagesInputs() {
		stageDir, attested, err := stageGitleaksInputs(b.root, b.plan.state, b.inputs.byPath)
		if err != nil {
			return failure(domain.AdapterFailed, "staging", "private input staging failed", "private input staging failed", "adapter input staging failed")
		}
		b.plan.stageDir, b.plan.attested, b.plan.scanRoot = stageDir, attested, stageDir
	}
	args, err := expandArgs(r.spec.Binary.Args, invocationValues{root: b.root, manifest: b.plan.manifestPath, inputs: manifests(b.inputs.accepted), bundledRules: b.plan.rules.paths, stageDir: b.plan.stageDir})
	if err != nil {
		return failure(domain.AdapterFailed, "invocation", "declared scanner invocation is invalid", "declared scanner invocation is invalid", "adapter invocation is invalid")
	}
	b.plan.args = args
	return nil
}

func (b *batch) execute(ctx context.Context) (scanOutput, *batchFailure) {
	r := b.runner
	runContext, cancel := context.WithTimeout(ctx, r.spec.Timeout)
	defer cancel()
	stdout, stderr, runErr := r.runCommand(runContext, b.root, &b.plan, b.plan.args)
	if runContext.Err() == context.DeadlineExceeded {
		return scanOutput{}, failure(domain.AdapterTimedOut, "timed_out", "scanner timed out", "scanner timed out", "scanner timed out")
	}
	if stdout.Exceeded() || stderr.Exceeded() {
		return scanOutput{}, failure(domain.AdapterFailed, "output_limit", reasonOutputBound, reasonOutputBound, reasonOutputBound)
	}
	parsed, parseErr := parseOutput(r.spec, stdout.Bytes(), b.plan.scanRoot)
	if parseErr != nil {
		return scanOutput{}, failure(domain.AdapterFailed, "output_schema", reasonOutputInvalid, reasonOutputInvalid, reasonOutputInvalid)
	}
	return scanOutput{parsed: parsed, stderr: stderr.Bytes(), runErr: runErr}, nil
}

func (b *batch) complete(ctx context.Context, output scanOutput) (domain.AdapterResult, []domain.Observation, error) {
	r := b.runner
	switch r.spec.Attestation {
	case AttestGitleaksScanLog:
		attestGitleaksInput(&output.parsed, b.plan.attested, output.stderr, b.plan.stageDir)
	case AttestRuffFileListing:
		b.attestRuffInputs(ctx, &output.parsed)
	}
	observations, rejected := r.observations(b.root, b.inputs.byPath, output.parsed.findings, b.started)
	coverage := b.inputs.coverage
	for path := range rejected {
		coverage = markPathIncomplete(coverage, path, reasonEvidenceInvalid)
	}
	status, code, message := r.resultStatus(output.runErr, output.parsed, b.inputs.byPath, rejected, b.inputs.inputErr)
	coverage = completeCoverage(coverage, b.inputs.byPath, output.parsed, status)
	linkObservations(coverage, observations)
	result := r.finish(&b.result, status, coverage, code, message)
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

func (r *Runner) selectInputs(root string, selected []domain.File, resultID string) inputSelection {
	selection := inputSelection{
		accepted: make([]sourceInput, 0, len(selected)),
		coverage: make([]domain.Coverage, 0, len(selected)),
		byPath:   make(map[string]sourceInput),
	}
	allowlist, allowlistErr := r.globalPathAllowlist()
	selection.inputErr = allowlistErr
	for _, file := range selected {
		if err := file.Validate(); err != nil {
			selection.inputErr = true
			continue
		}
		key := domain.CoverageKey{FilePath: file.Path, Family: r.spec.Family, Category: r.spec.Category}
		row := domain.Coverage{Key: key, AdapterResultIDs: []string{resultID}}
		row.Status, row.Reason = r.classifyInput(root, file, allowlist, &selection)
		selection.coverage = append(selection.coverage, row)
	}
	return selection
}

func (r *Runner) globalPathAllowlist() ([]*regexp.Regexp, bool) {
	if !r.spec.gitleaks() {
		return nil, false
	}
	pack, err := bundledRule(gitleaksDefaultRule)
	if err != nil {
		return nil, true
	}
	allowlist, err := gitleaksGlobalPathAllowlist(pack)
	return allowlist, err != nil
}

func (r *Runner) classifyInput(root string, file domain.File, allowlist []*regexp.Regexp, selection *inputSelection) (domain.CoverageStatus, string) {
	if file.Eligible && !file.Excluded && pathAllowlisted(allowlist, file.Path) {
		return domain.CoverageNotApplicable, gitleaksAllowlistReason
	}
	if !r.spec.Applies(file) {
		return domain.CoverageNotApplicable, "adapter language or extension is ineligible"
	}
	selection.eligibleCount++
	if _, duplicate := selection.byPath[file.Path]; duplicate {
		selection.inputErr = true
		return domain.CoverageIncomplete, "selected input is duplicated"
	}
	input, err := fingerprint(root, file.Path, file.Language, r.spec.Inputs.MaxFileBytes)
	if err != nil {
		selection.inputErr = true
		return domain.CoverageIncomplete, "eligible input is unreadable or exceeds the declared bound"
	}
	if len(selection.accepted) == r.spec.Inputs.MaxFiles || selection.totalBytes > r.spec.Inputs.MaxTotalBytes-input.manifest.Bytes {
		selection.inputErr = true
		return domain.CoverageIncomplete, "eligible input exceeds the declared scan bound"
	}
	selection.totalBytes += input.manifest.Bytes
	selection.accepted, selection.byPath[file.Path] = append(selection.accepted, input), input
	return "", ""
}

func fingerprint(root, path, language string, maximum int64) (sourceInput, error) {
	resolved, err := boundedRegularFile(root, path, maximum)
	if err != nil {
		return sourceInput{}, err
	}
	digest, size, lines, err := digestFile(resolved, maximum)
	if err != nil {
		return sourceInput{}, err
	}
	return sourceInput{manifest: ManifestInput{Path: path, Language: language, Digest: digest, Bytes: size, Lines: lines}}, nil
}

func boundedRegularFile(root, path string, maximum int64) (string, error) {
	if err := validRelativePath(path); err != nil {
		return "", err
	}
	joined := filepath.Join(root, filepath.FromSlash(path))
	resolved, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("input escapes root")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maximum {
		return "", errors.New("input is not a bounded regular file")
	}
	return resolved, nil
}

func digestFile(path string, maximum int64) (digest string, size int64, lines int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, 0, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	hash := sha256.New()
	lines = 1
	buffer := make([]byte, 32*1024)
	for {
		count, readErr := file.Read(buffer)
		if count > 0 {
			size += int64(count)
			if size > maximum {
				return "", 0, 0, errors.New("input exceeds bound")
			}
			hash.Write(buffer[:count])
			lines += bytes.Count(buffer[:count], []byte{'\n'})
		}
		if readErr == io.EOF {
			return hex.EncodeToString(hash.Sum(nil)), size, lines, nil
		}
		if readErr != nil {
			return "", 0, 0, readErr
		}
	}
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
		if err := writePrivateRule(state, path, content); err != nil {
			return materializedRules{}, err
		}
		sum := sha256.Sum256(content)
		result.paths = append(result.paths, path)
		result.digests = append(result.digests, RuleDigest{Path: rulePath, Digest: hex.EncodeToString(sum[:])})
	}
	return result, nil
}

func writePrivateRule(state, path string, content []byte) error {
	relative, err := filepath.Rel(state, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("bundled rule escapes private state")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err := file.Write(content); err != nil {
		return errors.Join(err, file.Close())
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Chmod(path, 0400)
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
		return "", errors.Join(errors.New("manifest permission change failed"), file.Close(), os.Remove(path))
	}
	if _, err := file.Write(encoded); err != nil {
		return "", errors.Join(errors.New("manifest write failed"), file.Close(), os.Remove(path))
	}
	if err := file.Close(); err != nil {
		return "", errors.Join(errors.New("manifest write failed"), os.Remove(path))
	}
	return path, nil
}

type invocationValues struct {
	root         string
	manifest     string
	inputs       []ManifestInput
	bundledRules []string
	stageDir     string
}

func expandArgs(args []string, values invocationValues) ([]string, error) {
	root, manifest, inputs, bundledRules, stageDir := values.root, values.manifest, values.inputs, values.bundledRules, values.stageDir
	result := make([]string, 0, len(args)+len(inputs)+len(bundledRules))
	for _, arg := range args {
		switch arg {
		case "{input_manifest}":
			result = append(result, manifest)
		case "{input_paths}":
			result = append(result, relativeInputPaths(inputs)...)
		case "{absolute_input_paths}":
			result = append(result, absoluteInputPaths(root, inputs)...)
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

func relativeInputPaths(inputs []ManifestInput) []string {
	paths := make([]string, 0, len(inputs))
	for _, input := range inputs {
		paths = append(paths, input.Path)
	}
	return paths
}

func absoluteInputPaths(root string, inputs []ManifestInput) []string {
	paths := make([]string, 0, len(inputs))
	for _, input := range inputs {
		paths = append(paths, filepath.Join(root, input.Path))
	}
	return paths
}

func (r *Runner) verifyVersion(ctx context.Context, root, state, rulesRoot, binary string) error {
	r.verifyMu.Lock()
	defer r.verifyMu.Unlock()
	if r.verifiedBinary == binary {
		return nil
	}
	// Startup is scanner work too: use its declared budget rather than a
	// shorter probe-only deadline that rejects cold or contended executables.
	check, cancel := context.WithTimeout(ctx, r.spec.Timeout)
	defer cancel()
	if err := sandbox.VerifyVersion(check, r.request(root, state, rulesRoot, binary, r.spec.Binary.VersionArgs), r.spec.Binary.Version); err != nil {
		return err
	}
	r.verifiedBinary = binary
	return nil
}

func (r *Runner) request(root, state, rulesRoot, binary string, args []string) sandbox.Request {
	workDir := root
	if r.spec.Binary.PrivateWorkDir {
		workDir = state
	}
	return sandbox.Request{WorkDir: workDir, State: state, RulesRoot: rulesRoot, Binary: binary, Args: args, StdoutLimit: r.spec.MaxOutput, StderrLimit: r.spec.MaxOutput}
}

func (r *Runner) runCommand(ctx context.Context, root string, plan *scanPlan, args []string) (*sandbox.Capture, *sandbox.Capture, error) {
	return sandbox.Run(ctx, r.request(root, plan.state, plan.rules.root, plan.binary, args))
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
		observation, err := r.observation(root, finding, observed)
		if err != nil {
			rejected[finding.path] = struct{}{}
			continue
		}
		observations = append(observations, observation)
	}
	return observations, rejected
}

type findingText struct {
	mechanism   string
	uncertainty string
	quote       string
	severity    domain.Severity
	confidence  domain.Confidence
}

func (r *Runner) findingText(finding parsedFinding) (findingText, error) {
	text := findingText{severity: finding.severity, confidence: finding.confidence}
	var err error
	if r.spec.Output.SecretDiagnostics {
		text.mechanism, text.severity, text.confidence = domain.SecretMechanism, domain.SeverityHigh, domain.ConfidenceMedium
		text.quote, err = sanitizeText(finding.quote, maxEvidenceText)
		return text, err
	}
	if text.mechanism, err = sanitizeText(finding.mechanism, maxEvidenceText); err != nil {
		return text, err
	}
	if text.uncertainty, err = sanitizeText(finding.uncertainty, maxEvidenceText); err != nil {
		return text, err
	}
	text.quote, err = sanitizeText(finding.quote, maxEvidenceText)
	return text, err
}

func (r *Runner) observation(root string, finding parsedFinding, observed time.Time) (domain.Observation, error) {
	text, err := r.findingText(finding)
	if err != nil || text.mechanism == "" {
		return domain.Observation{}, errors.New("finding text failed sanitization")
	}
	location := domain.SourceRange{Path: finding.path, StartLine: finding.startLine, EndLine: finding.endLine, StartByte: finding.startByte, EndByte: finding.endByte}
	sourceFindingID := finding.id
	if r.spec.Output.SecretDiagnostics {
		// Identity of a secret is a digest of its raw value, computed in memory
		// here and never stored or logged raw, so a moved secret stays one issue.
		var digestErr error
		if sourceFindingID, digestErr = secretDigest(root, finding); digestErr != nil {
			return domain.Observation{}, digestErr
		}
	}
	observation := domain.Observation{ObservationID: digestID(r.runID, r.spec.ID, finding.id, finding.ruleID, finding.path, fmt.Sprint(finding.startLine)), RunID: r.runID, Source: r.spec.ID, SourceFindingID: sourceFindingID, Family: r.spec.Family, OriginalCategory: r.spec.Category, RuleID: finding.ruleID, RuleVersion: r.spec.Rules.Version, SourceSeverity: text.severity, EffectiveSeverity: text.severity, Confidence: text.confidence, Mechanism: text.mechanism, Uncertainty: text.uncertainty, Evidence: []domain.Evidence{{Location: location, RedactedQuote: text.quote}}, Validated: true, ObservedAt: observed}
	if err := observation.Validate(); err != nil {
		return domain.Observation{}, err
	}
	return observation, nil
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
	return inputAccountingStatus(parsed, inputs)
}

func inputAccountingStatus(parsed parsedOutput, inputs map[string]sourceInput) (domain.AdapterStatus, string, string) {
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
		row.Status, row.Reason = domain.CoverageIncomplete, reasonNotInspected
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
