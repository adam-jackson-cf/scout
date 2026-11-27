package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"scout/internal/artifacts"
	"scout/internal/clawpatch"
	"scout/internal/contextual"
	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/issues"
	"scout/internal/policy"
	"scout/internal/prmeta"
	"scout/internal/state"
	"scout/internal/web"
)

const (
	exitComplete    = 0
	exitPartial     = 2
	exitUnavailable = 3
)

type stringsFlag []string

func (s *stringsFlag) String() string { return strings.Join(*s, ",") }
func (s *stringsFlag) Set(value string) error {
	if value == "" {
		return errors.New("empty path or family")
	}
	*s = append(*s, value)
	return nil
}

type commonFlags struct {
	root   string
	binary string
	json   bool
}

func common(set *flag.FlagSet) *commonFlags {
	flags := &commonFlags{}
	set.StringVar(&flags.root, "root", ".", "repository root")
	set.StringVar(&flags.binary, "clawpatch", os.Getenv("SCOUT_CLAWPATCH_BINARY"), "absolute pinned Clawpatch dist/cli.js")
	set.BoolVar(&flags.json, "json", false, "emit machine-readable JSON")
	return flags
}

type outcome struct {
	Status   string `json:"status"`
	ExitCode int    `json:"exitCode"`
	Result   any    `json:"result,omitempty"`
	Error    string `json:"error,omitempty"`
}

// withheldText replaces error or reason text the sanitiser rejects (NUL bytes,
// invalid UTF-8); the failure is reported, never an empty string.
const withheldText = "text withheld: contains unsafe characters"

// sanitized returns text safe to print, or withheldText when it cannot be.
func sanitized(text string) string {
	safe, err := issues.SanitizeText(text)
	if err != nil {
		return withheldText
	}
	return safe
}

// capabilityReason preserves the catalog's public rule-pack fingerprint while
// sanitizing policy-controlled values and other free text in the explanation.
func capabilityReason(check engine.Capability) string {
	if check.PublicRuleVersion == "" {
		return sanitized(check.Reason)
	}
	parts := strings.Split(check.Reason, check.PublicRuleVersion)
	for index := range parts {
		parts[index] = sanitized(parts[index])
	}
	return strings.Join(parts, check.PublicRuleVersion)
}

func emit(flags *commonFlags, status string, code int, result any, err error) int {
	reply := outcome{Status: status, ExitCode: code, Result: result}
	if err != nil {
		reply.Error = sanitized(err.Error())
	}
	if flags != nil && flags.json {
		if encodeErr := json.NewEncoder(os.Stdout).Encode(reply); encodeErr != nil {
			return exitUnavailable
		}
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "scout:", reply.Error)
	} else if result != nil {
		encoded, marshalErr := json.MarshalIndent(reply, "", "  ")
		if marshalErr != nil {
			fmt.Fprintln(os.Stderr, "scout:", marshalErr)
			return exitUnavailable
		}
		if _, writeErr := fmt.Fprintln(os.Stdout, string(encoded)); writeErr != nil {
			return exitUnavailable
		}
	}
	return code
}

func printStdout(text string, code int) int {
	if _, err := fmt.Fprint(os.Stdout, text); err != nil {
		fmt.Fprintln(os.Stderr, "scout:", sanitized(err.Error()))
		return exitUnavailable
	}
	return code
}

func closeStore(store *state.Store) {
	if err := store.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "scout: close state store:", sanitized(err.Error()))
	}
}

// commands is the dispatch table; the usage line is generated from it.
var commands = []struct {
	name string
	run  func(context.Context, []string) int
}{
	{"init", initCmd},
	{"readiness", readiness},
	{"assess", assess},
	{"diff", diff},
	{"issues", listIssues},
	{"issue", showIssue},
	{"mapping", mapping},
	{"map", mapCmd},
	{"areas", areasCmd},
	{"coverage", coverage},
	{"triage", triage},
	{"decisions", decisionsCmd},
	{"signals", signalsCmd},
	{"export", exportCmd},
	{"import", importCmd},
	{"state", stateCmd},
	{"serve", serve},
}

func main() { os.Exit(run(context.Background(), os.Args[1:])) }
func run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		names := make([]string, len(commands))
		for index, command := range commands {
			names[index] = command.name
		}
		fmt.Fprintf(os.Stderr, "usage: scout <%s> [flags]\n", strings.Join(names, "|"))
		return exitUnavailable
	}
	for _, command := range commands {
		if command.name == args[0] {
			return command.run(ctx, args[1:])
		}
	}
	fmt.Fprintf(os.Stderr, "unknown Scout command %q\n", args[0])
	return exitUnavailable
}
func parse(set *flag.FlagSet, args []string, flags *commonFlags) bool {
	set.SetOutput(os.Stderr)
	if err := set.Parse(args); err != nil {
		emit(flags, "unavailable", exitUnavailable, nil, err)
		return false
	}
	if len(set.Args()) != 0 {
		emit(flags, "unavailable", exitUnavailable, nil, errors.New("unexpected positional arguments"))
		return false
	}
	return true
}

// providerConfig is the model-CLI adapter's environment layer. SCOUT_MODEL_CLI
// (claude|codex) is a local override honoured only when the committed policy
// leaves model_cli.provider unset (see engine.ApplyModelCLIPolicy); with
// neither, the first CLI on PATH runs. SCOUT_MODEL is optional and the CLI uses
// its own default model without it. No credential is read or forwarded.
func providerConfig(authorized bool) contextual.Config {
	return contextual.Config{CLI: os.Getenv("SCOUT_MODEL_CLI"), Model: os.Getenv("SCOUT_MODEL"), EgressAuthorized: authorized}
}

// engineConfig builds the engine config with model-CLI consent from the policy
// committed at HEAD: a lane runs only when opted in, and the policy's provider
// outranks SCOUT_MODEL_CLI. A repository without a policy has every lane off.
func engineConfig(flags *commonFlags, authorized bool) (engine.Config, error) {
	pol, err := headPolicy(flags.root)
	if err != nil {
		return engine.Config{}, err
	}
	cfg := engine.Config{Root: flags.root, ClawpatchBinary: flags.binary, ClawpatchEgressAuthorized: authorized, Provider: providerConfig(authorized)}
	return engine.ApplyModelCLIPolicy(cfg, pol), nil
}
func selectedFamilies(values []string) ([]domain.Family, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make([]domain.Family, 0, len(values))
	seen := map[domain.Family]bool{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			family := domain.Family(strings.TrimSpace(part))
			if !family.Valid() || family == domain.FamilyClawpatch {
				return nil, fmt.Errorf("invalid family %q", part)
			}
			if !seen[family] {
				result = append(result, family)
				seen[family] = true
			}
		}
	}
	return result, nil
}
func scopeOptions(kind string, paths []string, from, to string, families []string) (domain.Scope, error) {
	selected, err := selectedFamilies(families)
	if err != nil {
		return domain.Scope{}, err
	}
	scope := domain.Scope{Kind: domain.ScopeKind(kind), Paths: paths, FromRevision: from, ToRevision: to, SelectedFamilies: selected}
	return scope, scope.Validate()
}

type assessFlags struct {
	full, authorized, fork, noEgress, fetchPRs                             bool
	kind, from, to, prMetadata, prRepo, artifactRevision, artifactProducer string
	paths, excludes, families, artifactPaths                               stringsFlag
}

func registerAssessFlags(set *flag.FlagSet) *assessFlags {
	opts := &assessFlags{}
	set.BoolVar(&opts.full, "full", false, "force full rescan")
	set.BoolVar(&opts.authorized, "allow-egress", false, "authorize source egress for providers")
	set.BoolVar(&opts.fork, "fork", false, "fork run: association never calls Jev")
	set.BoolVar(&opts.noEgress, "no-egress", false, "forbid association Jev calls even when consented")
	set.StringVar(&opts.kind, "scope", "repository", "repository|subpath|staged|commit_range|paths")
	set.StringVar(&opts.from, "from", "", "commit-range start")
	set.StringVar(&opts.to, "to", "", "commit-range end")
	set.Var(&opts.paths, "path", "selected relative path (repeatable)")
	set.Var(&opts.excludes, "exclude", "excluded relative path (repeatable)")
	set.Var(&opts.families, "family", "selected family (repeatable)")
	set.StringVar(&opts.prMetadata, "pr-metadata", "", "merged-PR metadata JSON (scout.pr-metadata@1) for pr_effort, batch_size, dispersion")
	set.BoolVar(&opts.fetchPRs, "fetch-pr-metadata", false, "fetch merged-PR metadata from GitHub (needs GITHUB_TOKEN with pull-requests: read and --pr-repo)")
	set.StringVar(&opts.prRepo, "pr-repo", "", "owner/name for --fetch-pr-metadata")
	set.Var(&opts.artifactPaths, "artifact", "coverage (LCOV, Cobertura XML) or JUnit XML file or glob, relative to the repository (repeatable); CRAP needs a revision-matched coverage artifact")
	set.StringVar(&opts.artifactRevision, "artifact-revision", "", "commit the --artifact files were produced at, for artifacts without a "+artifacts.SidecarSuffix+" provenance file")
	set.StringVar(&opts.artifactProducer, "artifact-producer", "", "producer name recorded for --artifact files without a sidecar")
	return opts
}

func (opts *assessFlags) measures(scope domain.Scope) (*engine.MeasuresOptions, error) {
	measureOpts := &engine.MeasuresOptions{History: scope.Kind == domain.ScopeRepository}
	if opts.prMetadata != "" && opts.fetchPRs {
		return nil, errors.New("--pr-metadata and --fetch-pr-metadata are mutually exclusive")
	}
	if opts.prMetadata != "" || opts.fetchPRs {
		measureOpts.PRMetadata = &engine.PRMetadataOptions{File: opts.prMetadata, Fetch: prmeta.FetchOptions{Enabled: opts.fetchPRs, Repo: opts.prRepo, Token: os.Getenv("GITHUB_TOKEN")}}
	}
	for _, path := range opts.artifactPaths {
		measureOpts.Artifacts = append(measureOpts.Artifacts, artifacts.Spec{Path: path, Revision: opts.artifactRevision, Producer: opts.artifactProducer})
	}
	return measureOpts, nil
}

func assessInputs(flags *commonFlags, opts *assessFlags) (engine.Config, policy.Policy, engine.Options, error) {
	scope, err := scopeOptions(opts.kind, opts.paths, opts.from, opts.to, opts.families)
	if err != nil {
		return engine.Config{}, policy.Policy{}, engine.Options{}, err
	}
	cfg, err := engineConfig(flags, opts.authorized)
	if err != nil {
		return engine.Config{}, policy.Policy{}, engine.Options{}, err
	}
	pol, err := headPolicy(flags.root)
	if err != nil {
		return engine.Config{}, policy.Policy{}, engine.Options{}, err
	}
	associationOpts, trigger, err := associationOptions(pol, opts.from, opts.to, opts.fork, opts.noEgress)
	if err != nil {
		return engine.Config{}, policy.Policy{}, engine.Options{}, err
	}
	measureOpts, err := opts.measures(scope)
	if err != nil {
		return engine.Config{}, policy.Policy{}, engine.Options{}, err
	}
	return cfg, pol, engine.Options{Scope: scope, ForceFull: opts.full, Exclusions: opts.excludes, Trigger: trigger, Measures: measureOpts, Association: &associationOpts}, nil
}

func assess(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("assess", flag.ContinueOnError)
	flags := common(set)
	opts := registerAssessFlags(set)
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	cfg, pol, engineOpts, err := assessInputs(flags, opts)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	result, err := engine.Assess(ctx, cfg, engineOpts)
	if err != nil {
		unavailable := engine.UnavailableReport(err)
		return emit(flags, unavailable.Gate.Status, unavailable.Gate.ExitCode, unavailable, err)
	}
	return emitAssessment(ctx, flags, pol, result)
}

func emitAssessment(ctx context.Context, flags *commonFlags, pol policy.Policy, result engine.Result) int {
	verdict := assessGate(pol, result)
	ranking, health, err := rankAssessment(ctx, flags, pol, result, verdict)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	pending, err := pendingAssessmentAreas(ctx, flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	report := assessmentReport{result.Run, result.Plan, result.Diff, result.Map.Status, result.MappingGaps, result.Issues, result.Adapters, verdict.Gate, health, []domain.WouldBlock{}, ranking, pending, result.Associations, associationGuardReport{result.AssociationGuard}}
	return emit(flags, verdict.Gate.Status, verdict.Gate.ExitCode, report, nil)
}

type capability struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Ready    bool   `json:"ready"`
	Reason   string `json:"reason,omitempty"`
}

func readiness(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("readiness", flag.ContinueOnError)
	flags := common(set)
	shadowOpts := registerShadowFlags(set)
	var kind, from, to string
	var paths, families, excludes stringsFlag
	var full, authorized, matrix bool
	set.StringVar(&kind, "scope", "repository", "scope kind")
	set.StringVar(&from, "from", "", "revision start")
	set.StringVar(&to, "to", "", "revision end")
	set.Var(&paths, "path", "relative path")
	set.Var(&families, "family", "family")
	set.Var(&excludes, "exclude", "excluded path")
	set.BoolVar(&full, "full", false, "evaluate a forced full rescan")
	set.BoolVar(&matrix, "matrix", false, "print the capability matrix (REQ-18) for this repository and exit")
	set.BoolVar(&authorized, "allow-egress", false, "authorize selected provider calls")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if shadowOpts.enabled {
		return readinessShadow(ctx, flags, shadowOpts, set)
	}
	if shadowOnlyFlagGiven(set) {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("--check, --version, --as-of and --encoding require --shadow"))
	}
	scope, err := scopeOptions(kind, paths, from, to, families)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	cfg, err := engineConfig(flags, authorized)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if matrix {
		return readinessMatrix(ctx, flags, cfg)
	}
	assessmentReadiness, err := engine.AssessReadiness(ctx, cfg, engine.Options{Scope: scope, ForceFull: full, Exclusions: excludes})
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	checks := readinessCapabilities(assessmentReadiness.Capabilities)
	report := struct {
		Ready        bool         `json:"ready"`
		Capabilities []capability `json:"capabilities"`
		associationGuardReport
	}{assessmentReadiness.Ready, checks, associationGuardReport{assessmentReadiness.AssociationGuard}}
	if !assessmentReadiness.Ready {
		return emit(flags, "unavailable", exitUnavailable, report, nil)
	}
	return emit(flags, "complete", exitComplete, report, nil)
}

func shadowOnlyFlagGiven(set *flag.FlagSet) bool {
	var shadowOnly bool
	set.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "check", "version", "as-of", "encoding":
			shadowOnly = true
		}
	})
	return shadowOnly
}

func readinessCapabilities(capabilities []engine.Capability) []capability {
	checks := make([]capability, len(capabilities))
	for index, check := range capabilities {
		reason := capabilityReason(check)
		checks[index] = capability{Name: check.Name, Required: check.Required, Ready: check.Ready, Reason: reason}
	}
	return checks
}

func openStore(flags *commonFlags) (*state.Store, error) { return state.Open(flags.root) }

type issueFilter struct {
	status, minSeverity, family, file, feature, runID string
}

func (f issueFilter) validate() error {
	if f.status != "" && !domain.IssueState(f.status).Valid() {
		return errors.New("invalid issue state")
	}
	if f.minSeverity != "" && !domain.Severity(f.minSeverity).Valid() {
		return errors.New("invalid severity")
	}
	if f.family != "" && !domain.Family(f.family).Valid() {
		return errors.New("invalid family")
	}
	return nil
}

func (f issueFilter) matchesIssue(issue domain.Issue) bool {
	if f.status != "" && string(issue.State) != f.status || f.minSeverity != "" && issue.EffectiveSeverity.Rank() < domain.Severity(f.minSeverity).Rank() {
		return false
	}
	if f.file != "" && !slices.Contains(issue.Files, f.file) || f.feature != "" && !slices.Contains(issue.FeatureIDs, f.feature) {
		return false
	}
	return true
}

func filterIssues(ctx context.Context, store *state.Store, all []domain.Issue, f issueFilter) ([]domain.Issue, error) {
	filtered := make([]domain.Issue, 0, len(all))
	for _, issue := range all {
		if !f.matchesIssue(issue) {
			continue
		}
		if f.runID != "" || f.family != "" {
			observations, loadErr := store.ListIssueObservations(ctx, issue.IssueID)
			if loadErr != nil {
				return nil, loadErr
			}
			if !hasMatchingObservation(observations, f.runID, domain.Family(f.family)) {
				continue
			}
		}
		filtered = append(filtered, issue)
	}
	return filtered, nil
}

func listIssues(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("issues", flag.ContinueOnError)
	flags := common(set)
	var filter issueFilter
	set.StringVar(&filter.status, "status", "", "issue state")
	set.StringVar(&filter.minSeverity, "min-severity", "", "minimum severity")
	set.StringVar(&filter.family, "family", "", "observation family")
	set.StringVar(&filter.file, "file", "", "relative file")
	set.StringVar(&filter.feature, "feature", "", "feature ID")
	set.StringVar(&filter.runID, "run", "", "observation run ID")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if err := filter.validate(); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer closeStore(store)
	all, err := store.ListIssues(ctx)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	filtered, err := filterIssues(ctx, store, all, filter)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, filtered, nil)
}
func hasMatchingObservation(observations []domain.Observation, runID string, family domain.Family) bool {
	for _, observation := range observations {
		if runID != "" && observation.RunID != runID {
			continue
		}
		if family != "" && observation.Family != family {
			continue
		}
		return true
	}
	return false
}
func showIssue(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("issue", flag.ContinueOnError)
	flags := common(set)
	var id, runID string
	set.StringVar(&id, "id", "", "issue ID")
	set.StringVar(&runID, "run", "", "include run evidence")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if id == "" {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("issue ID required"))
	}
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer closeStore(store)
	issue, err := store.LoadIssue(ctx, id)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if runID == "" {
		return emit(flags, "complete", exitComplete, issue, nil)
	}
	run, err := store.LoadRun(ctx, runID)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	observations, err := store.LoadObservations(ctx, runID)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	associated := []domain.Observation{}
	for _, observation := range observations {
		if slices.Contains(issue.ObservationIDs, observation.ObservationID) {
			associated = append(associated, observation)
		}
	}
	return emit(flags, "complete", exitComplete, struct {
		Issue        domain.Issue         `json:"issue"`
		Run          domain.Run           `json:"run"`
		Observations []domain.Observation `json:"observations"`
	}{issue, run, associated}, nil)
}
func mapping(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("mapping", flag.ContinueOnError)
	flags := common(set)
	var listAreas, listUnowned bool
	set.BoolVar(&listAreas, "areas", false, "list area labels")
	set.BoolVar(&listUnowned, "unowned", false, "list directory scopes of files no heuristic feature owns")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if listAreas && listUnowned {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("--areas and --unowned are mutually exclusive"))
	}
	if listAreas {
		return mappingAreas(ctx, flags)
	}
	if listUnowned {
		return mappingUnowned(ctx, flags)
	}
	root, err := filepath.Abs(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	features, err := clawpatch.LoadFeatures(clawpatch.StateDir(root))
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	type mapped struct {
		ID      string                  `json:"featureId"`
		Kind    string                  `json:"kind"`
		Source  string                  `json:"source"`
		Owned   []domain.FeatureFileRef `json:"ownedFiles"`
		Context []domain.FeatureFileRef `json:"contextFiles"`
		Status  string                  `json:"status"`
	}
	output := make([]mapped, 0, len(features))
	for _, feature := range features {
		output = append(output, mapped{feature.FeatureID, feature.Kind, feature.Source, feature.OwnedFiles, feature.ContextFiles, feature.Status})
	}
	return emit(flags, "complete", exitComplete, output, nil)
}
func coverage(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("coverage", flag.ContinueOnError)
	flags := common(set)
	var runID string
	set.StringVar(&runID, "run", "", "run ID (default latest)")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer closeStore(store)
	var run domain.Run
	if runID == "" {
		run, err = store.LatestRecordedRun(ctx)
	} else {
		run, err = store.LoadRun(ctx, runID)
	}
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	plan, err := store.LoadPlan(ctx, run.PlanID)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, struct {
		Run  domain.Run  `json:"run"`
		Plan domain.Plan `json:"plan"`
	}{run, plan}, nil)
}
func requireAuthorization(authorized bool, actor, reason string) (string, string, error) {
	if !authorized {
		return "", "", errors.New("explicit --authorize required")
	}
	if actor == "" || reason == "" {
		return "", "", errors.New("actor and reason required")
	}
	safeActor, err := issues.SanitizeText(actor)
	if err != nil {
		return "", "", err
	}
	safeReason, err := issues.SanitizeText(reason)
	if err != nil {
		return "", "", err
	}
	if safeActor == "" || safeReason == "" {
		return "", "", errors.New("actor and reason must not be empty after sanitation")
	}
	return safeActor, safeReason, nil
}

type triageFlags struct {
	id, status, actor, reason, owner, disposition, runID, artifactID string
	authorized                                                       bool
}

type triageRequest struct {
	id                                                   string
	desired                                              domain.IssueState
	owner, disposition, actor, reason, runID, artifactID string
	needsEvidence                                        bool
}

func triageNeedsEvidence(desired domain.IssueState, disposition, runID, artifactID string) (bool, error) {
	needsEvidence := desired == domain.IssueFalsePositive || desired == domain.IssueWontFix
	if needsEvidence && (disposition == "" || runID == "" || artifactID == "") {
		return true, errors.New("false-positive and won't-fix triage require disposition, run, and sanitized artifact evidence")
	}
	if !needsEvidence && (runID != "" || artifactID != "") {
		return false, errors.New("evidence run and artifact apply only to false-positive or won't-fix triage")
	}
	return needsEvidence, nil
}

func (f triageFlags) request() (triageRequest, error) {
	safeActor, safeReason, err := requireAuthorization(f.authorized, f.actor, f.reason)
	if err != nil {
		return triageRequest{}, err
	}
	desired := domain.IssueState(f.status)
	if f.id == "" || !desired.Valid() || desired == domain.IssueResolved {
		return triageRequest{}, errors.New("issue ID and non-resolved valid state required")
	}
	safeOwner, err := issues.SanitizeText(f.owner)
	if err != nil {
		return triageRequest{}, err
	}
	safeDisposition, err := issues.SanitizeText(f.disposition)
	if err != nil {
		return triageRequest{}, err
	}
	needsEvidence, err := triageNeedsEvidence(desired, safeDisposition, f.runID, f.artifactID)
	if err != nil {
		return triageRequest{}, err
	}
	return triageRequest{id: f.id, desired: desired, owner: safeOwner, disposition: safeDisposition, actor: safeActor, reason: safeReason, runID: f.runID, artifactID: f.artifactID, needsEvidence: needsEvidence}, nil
}

func applyTriage(ctx context.Context, store *state.Store, req triageRequest) (domain.Issue, error) {
	if !req.needsEvidence {
		return store.TriageIssue(ctx, state.Triage{IssueID: req.id, State: req.desired, Owner: req.owner, DispositionReason: req.disposition, Actor: req.actor, Reason: req.reason})
	}
	run, err := store.LoadRun(ctx, req.runID)
	if err != nil {
		return domain.Issue{}, err
	}
	artifact, err := artifactInRun(run, req.artifactID)
	if err != nil {
		return domain.Issue{}, err
	}
	return store.TriageIssueWithEvidence(ctx, state.Triage{IssueID: req.id, State: req.desired, Owner: req.owner, DispositionReason: req.disposition, Actor: req.actor, Reason: req.reason}, []domain.SafeArtifactRef{artifact})
}

func triage(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("triage", flag.ContinueOnError)
	flags := common(set)
	var opts triageFlags
	set.StringVar(&opts.id, "id", "", "issue ID")
	set.StringVar(&opts.status, "state", "", "investigating|accepted|false_positive|wont_fix")
	set.StringVar(&opts.actor, "actor", "", "human actor")
	set.StringVar(&opts.reason, "reason", "", "audit reason")
	set.StringVar(&opts.owner, "owner", "", "owner")
	set.StringVar(&opts.disposition, "disposition", "", "contradictory or scope evidence for false positive/won't fix")
	set.StringVar(&opts.runID, "run", "", "run containing cited disposition evidence")
	set.StringVar(&opts.artifactID, "artifact", "", "sanitized evidence artifact ID in run")
	set.BoolVar(&opts.authorized, "authorize", false, "explicitly authorize triage mutation")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	req, err := opts.request()
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer closeStore(store)
	issue, err := applyTriage(ctx, store, req)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return emit(flags, "complete", exitComplete, issue, nil)
}
func artifactInRun(run domain.Run, artifactID string) (domain.SafeArtifactRef, error) {
	if artifactID == "" {
		return domain.SafeArtifactRef{}, errors.New("artifact ID required")
	}
	for _, artifact := range run.ArtifactRefs {
		if artifact.ArtifactID == artifactID {
			return artifact, nil
		}
	}
	return domain.SafeArtifactRef{}, errors.New("artifact is not in run")
}

var errLoopbackOnly = errors.New("web server accepts only loopback listeners")

func requireLoopback(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	if host != "localhost" {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return errLoopbackOnly
		}
	}
	return nil
}

func listenLoopback(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	if tcp, ok := listener.Addr().(*net.TCPAddr); !ok || !tcp.IP.IsLoopback() {
		_ = listener.Close()
		return nil, errLoopbackOnly
	}
	return listener, nil
}

func announceServing(flags *commonFlags, address string) error {
	if flags.json {
		return json.NewEncoder(os.Stdout).Encode(outcome{Status: "serving", ExitCode: exitComplete, Result: map[string]string{"address": address}})
	}
	_, err := fmt.Fprintln(os.Stdout, "Scout read-only web:", address)
	return err
}

func serve(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags := common(set)
	var address string
	set.StringVar(&address, "listen", "127.0.0.1:8347", "loopback address")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if err := requireLoopback(address); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	store, err := state.OpenExistingReadOnly(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer closeStore(store)
	listener, err := listenLoopback(address)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	server := &http.Server{Handler: web.New(store), ReadHeaderTimeout: 5 * time.Second}
	stopped, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-stopped.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	if err := announceServing(flags, listener.Addr().String()); err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	err = server.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return exitComplete
}
