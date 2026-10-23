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
	"strings"
	"syscall"
	"time"

	"scout/internal/artifacts"
	"scout/internal/clawpatch"
	"scout/internal/contextual"
	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/inventory"
	"scout/internal/issues"
	"scout/internal/prmeta"
	"scout/internal/state"
	"scout/internal/web"
)

const (
	exitComplete    = 0
	exitGate        = 1
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

func emit(flags *commonFlags, status string, code int, result any, err error) int {
	reply := outcome{Status: status, ExitCode: code, Result: result}
	if err != nil {
		reply.Error, _ = issues.SanitizeText(err.Error())
	}
	if flags != nil && flags.json {
		if encodeErr := json.NewEncoder(os.Stdout).Encode(reply); encodeErr != nil {
			return exitUnavailable
		}
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "scout:", reply.Error)
	} else if result != nil {
		encoded, _ := json.MarshalIndent(reply, "", "  ")
		fmt.Fprintln(os.Stdout, string(encoded))
	}
	return code
}
func main() { os.Exit(run(context.Background(), os.Args[1:])) }
func run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: scout <init|readiness|assess|diff|issues|issue|mapping|coverage|triage|export|import|state|signals|serve> [flags]")
		return exitUnavailable
	}
	switch args[0] {
	case "init":
		return initCmd(ctx, args[1:])
	case "readiness":
		return readiness(ctx, args[1:])
	case "assess":
		return assess(ctx, args[1:])
	case "diff":
		return diff(ctx, args[1:])
	case "issues":
		return listIssues(ctx, args[1:])
	case "issue":
		return showIssue(ctx, args[1:])
	case "mapping":
		return mapping(ctx, args[1:])
	case "map":
		return mapCmd(ctx, args[1:])
	case "coverage":
		return coverage(ctx, args[1:])
	case "triage":
		return triage(ctx, args[1:])
	case "signals":
		return signalsCmd(ctx, args[1:])
	case "export":
		return exportCmd(ctx, args[1:])
	case "import":
		return importCmd(ctx, args[1:])
	case "state":
		return stateCmd(ctx, args[1:])
	case "serve":
		return serve(ctx, args[1:])
	default:
		fmt.Fprintln(os.Stderr, "unknown Scout command")
		return exitUnavailable
	}
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

func assess(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("assess", flag.ContinueOnError)
	flags := common(set)
	var full, authorized bool
	var kind, from, to string
	var paths, excludes, families stringsFlag
	set.BoolVar(&full, "full", false, "force full rescan")
	set.BoolVar(&authorized, "allow-egress", false, "authorize source egress for providers")
	set.StringVar(&kind, "scope", "repository", "repository|subpath|staged|commit_range|paths")
	set.StringVar(&from, "from", "", "commit-range start")
	set.StringVar(&to, "to", "", "commit-range end")
	set.Var(&paths, "path", "selected relative path (repeatable)")
	set.Var(&excludes, "exclude", "excluded relative path (repeatable)")
	set.Var(&families, "family", "selected family (repeatable)")
	var prMetadata, prRepo string
	var fetchPRs bool
	set.StringVar(&prMetadata, "pr-metadata", "", "merged-PR metadata JSON (scout.pr-metadata@1) for pr_effort, batch_size, dispersion")
	set.BoolVar(&fetchPRs, "fetch-pr-metadata", false, "fetch merged-PR metadata from GitHub (needs GITHUB_TOKEN with pull-requests: read and --pr-repo)")
	set.StringVar(&prRepo, "pr-repo", "", "owner/name for --fetch-pr-metadata")
	var artifactPaths stringsFlag
	var artifactRevision, artifactProducer string
	set.Var(&artifactPaths, "artifact", "coverage (LCOV, Cobertura XML) or JUnit XML file or glob, relative to the repository (repeatable); CRAP needs a revision-matched coverage artifact")
	set.StringVar(&artifactRevision, "artifact-revision", "", "commit the --artifact files were produced at, for artifacts without a "+artifacts.SidecarSuffix+" provenance file")
	set.StringVar(&artifactProducer, "artifact-producer", "", "producer name recorded for --artifact files without a sidecar")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	scope, err := scopeOptions(kind, paths, from, to, families)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	cfg, err := engineConfig(flags, authorized)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	measureOpts := &engine.MeasuresOptions{History: scope.Kind == domain.ScopeRepository}
	if prMetadata != "" || fetchPRs {
		measureOpts.PRMetadata = &engine.PRMetadataOptions{File: prMetadata, Fetch: prmeta.FetchOptions{Enabled: fetchPRs, Repo: prRepo, Token: os.Getenv("GITHUB_TOKEN")}}
	}
	for _, path := range artifactPaths {
		measureOpts.Artifacts = append(measureOpts.Artifacts, artifacts.Spec{Path: path, Revision: artifactRevision, Producer: artifactProducer})
	}
	result, err := engine.Assess(ctx, cfg, engine.Options{Scope: scope, ForceFull: full, Exclusions: excludes, Trigger: domain.TriggerPush, Measures: measureOpts})
	if err != nil {
		unavailable := engine.UnavailableReport(err)
		return emit(flags, unavailable.Gate.Status, unavailable.Gate.ExitCode, unavailable, err)
	}
	verdict := assessGate(result)
	ranking, health, err := rankAssessment(ctx, flags, result, verdict)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	report := struct {
		Run           domain.Run             `json:"run"`
		Plan          domain.Plan            `json:"plan"`
		Diff          inventory.DiffResult   `json:"diff"`
		MappingStatus clawpatch.MapStatus    `json:"mappingStatus"`
		MappingGaps   []inventory.MappingGap `json:"mappingGaps"`
		Issues        []domain.Issue         `json:"issues"`
		Adapters      []domain.AdapterResult `json:"adapters"`
		Gate          domain.Gate            `json:"gate"`
		Health        []domain.LaneHealth    `json:"health"`
		WouldBlock    []domain.WouldBlock    `json:"would_block"`
		Ranking       engine.Ranking         `json:"ranking"`
	}{result.Run, result.Plan, result.Diff, result.Map.Status, result.MappingGaps, result.Issues, result.Adapters, verdict.Gate, health, []domain.WouldBlock{}, ranking}
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
	set.BoolVar(&matrix, "matrix", false, "print the capability matrix (design §5.3) for this repository and exit")
	set.BoolVar(&authorized, "allow-egress", false, "authorize selected provider calls")
	if !parse(set, args, flags) {
		return exitUnavailable
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
	checks := make([]capability, len(assessmentReadiness.Capabilities))
	for index, check := range assessmentReadiness.Capabilities {
		reason, _ := issues.SanitizeText(check.Reason)
		checks[index] = capability{Name: check.Name, Required: check.Required, Ready: check.Ready, Reason: reason}
	}
	report := struct {
		Ready        bool         `json:"ready"`
		Capabilities []capability `json:"capabilities"`
	}{assessmentReadiness.Ready, checks}
	if !assessmentReadiness.Ready {
		return emit(flags, "unavailable", exitUnavailable, report, nil)
	}
	return emit(flags, "complete", exitComplete, report, nil)
}

func openStore(flags *commonFlags) (*state.Store, error) { return state.Open(flags.root) }
func listIssues(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("issues", flag.ContinueOnError)
	flags := common(set)
	var status, minSeverity, family, file, feature, runID string
	set.StringVar(&status, "status", "", "issue state")
	set.StringVar(&minSeverity, "min-severity", "", "minimum severity")
	set.StringVar(&family, "family", "", "observation family")
	set.StringVar(&file, "file", "", "relative file")
	set.StringVar(&feature, "feature", "", "feature ID")
	set.StringVar(&runID, "run", "", "observation run ID")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	if status != "" && !domain.IssueState(status).Valid() {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("invalid issue state"))
	}
	if minSeverity != "" && !domain.Severity(minSeverity).Valid() {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("invalid severity"))
	}
	if family != "" && !domain.Family(family).Valid() {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("invalid family"))
	}
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer store.Close()
	all, err := store.ListIssues(ctx)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	filtered := make([]domain.Issue, 0, len(all))
	for _, issue := range all {
		if status != "" && string(issue.State) != status || minSeverity != "" && issue.EffectiveSeverity.Rank() < domain.Severity(minSeverity).Rank() {
			continue
		}
		if file != "" && !contains(issue.Files, file) || feature != "" && !contains(issue.FeatureIDs, feature) {
			continue
		}
		if runID != "" || family != "" {
			observations, loadErr := store.ListIssueObservations(ctx, issue.IssueID)
			if loadErr != nil {
				return emit(flags, "unavailable", exitUnavailable, nil, loadErr)
			}
			if !hasMatchingObservation(observations, runID, domain.Family(family)) {
				continue
			}
		}
		filtered = append(filtered, issue)
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
func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
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
	defer store.Close()
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
		if contains(issue.ObservationIDs, observation.ObservationID) {
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
	features, err := clawpatch.LoadFeatures(filepath.Join(root, ".scout", "clawpatch"))
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
	defer store.Close()
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
func triage(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("triage", flag.ContinueOnError)
	flags := common(set)
	var id, status, actor, reason, owner, disposition, runID, artifactID string
	var authorized bool
	set.StringVar(&id, "id", "", "issue ID")
	set.StringVar(&status, "state", "", "investigating|accepted|false_positive|wont_fix")
	set.StringVar(&actor, "actor", "", "human actor")
	set.StringVar(&reason, "reason", "", "audit reason")
	set.StringVar(&owner, "owner", "", "owner")
	set.StringVar(&disposition, "disposition", "", "contradictory or scope evidence for false positive/won't fix")
	set.StringVar(&runID, "run", "", "run containing cited disposition evidence")
	set.StringVar(&artifactID, "artifact", "", "sanitized evidence artifact ID in run")
	set.BoolVar(&authorized, "authorize", false, "explicitly authorize triage mutation")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	safeActor, safeReason, err := requireAuthorization(authorized, actor, reason)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	desired := domain.IssueState(status)
	if id == "" || !desired.Valid() || desired == domain.IssueResolved {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("issue ID and non-resolved valid state required"))
	}
	safeOwner, err := issues.SanitizeText(owner)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	safeDisposition, err := issues.SanitizeText(disposition)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	needsEvidence := desired == domain.IssueFalsePositive || desired == domain.IssueWontFix
	if needsEvidence && (safeDisposition == "" || runID == "" || artifactID == "") {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("false-positive and won't-fix triage require disposition, run, and sanitized artifact evidence"))
	}
	if !needsEvidence && (runID != "" || artifactID != "") {
		return emit(flags, "unavailable", exitUnavailable, nil, errors.New("evidence run and artifact apply only to false-positive or won't-fix triage"))
	}
	store, err := openStore(flags)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer store.Close()
	if needsEvidence {
		run, loadErr := store.LoadRun(ctx, runID)
		if loadErr != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, loadErr)
		}
		artifact, artifactErr := artifactInRun(run, artifactID)
		if artifactErr != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, artifactErr)
		}
		issue, triageErr := store.TriageIssueWithEvidence(ctx, id, desired, safeOwner, safeDisposition, safeActor, safeReason, []domain.SafeArtifactRef{artifact})
		if triageErr != nil {
			return emit(flags, "unavailable", exitUnavailable, nil, triageErr)
		}
		return emit(flags, "complete", exitComplete, issue, nil)
	}
	issue, err := store.TriageIssue(ctx, id, desired, safeOwner, safeDisposition, safeActor, safeReason)
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
func serve(ctx context.Context, args []string) int {
	set := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags := common(set)
	var address string
	set.StringVar(&address, "listen", "127.0.0.1:8347", "loopback address")
	if !parse(set, args, flags) {
		return exitUnavailable
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	if host != "localhost" {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			return emit(flags, "unavailable", exitUnavailable, nil, errors.New("web server accepts only loopback listeners"))
		}
	}
	store, err := state.OpenExistingReadOnly(flags.root)
	if err != nil {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	defer store.Close()
	listener, err := net.Listen("tcp", address)
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
	if flags.json {
		_ = json.NewEncoder(os.Stdout).Encode(outcome{Status: "serving", ExitCode: exitComplete, Result: map[string]string{"address": listener.Addr().String()}})
	} else {
		fmt.Fprintln(os.Stdout, "Scout read-only web:", listener.Addr().String())
	}
	err = server.Serve(listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return emit(flags, "unavailable", exitUnavailable, nil, err)
	}
	return exitComplete
}
