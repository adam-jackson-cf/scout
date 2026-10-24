// Package clawpatch integrates Scout with the pinned Clawpatch CLI.
package clawpatch

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"scout/internal/domain"
	"scout/internal/modelcli"
)

const (
	Revision = "734f81bc1514b52a0627edb0c24ebe1a18b36105"
	Version  = "0.7.3"

	defaultMapTimeoutMS = "900000"
	stderrTailBytes     = 512

	// The agent map always runs on codex with this model and effort.
	mapModel           = "gpt-6-luna"
	mapReasoningEffort = "xhigh"
)

var (
	// ErrExternalStateRequiresAttachment prevents Scout from initializing over
	// developer-owned Clawpatch records.
	ErrExternalStateRequiresAttachment = errors.New("external Clawpatch state must already be initialized")
)

// Config identifies the reviewed executable and state location. BinaryPath is
// intentionally an absolute JavaScript entrypoint, never a command name.
type Config struct {
	BinaryPath         string
	NodePath           string
	StateDir           string
	AllowExternalState bool
	// EgressAuthorized explicitly permits Clawpatch map/review provider requests.
	EgressAuthorized bool
	// Provider is the Clawpatch provider for model-using calls: "codex" or
	// "claude". Empty falls back to CLAWPATCH_PROVIDER, then to whichever CLI
	// is installed. Any other value is rejected when a model call is made.
	Provider string
}

// Client serializes all state-mutating Clawpatch calls for one assessment root.
type Client struct {
	root             string
	stateDir         string
	binaryPath       string
	nodePath         string
	externalState    bool
	egressAuthorized bool
	provider         string

	mu          sync.Mutex
	initialized bool
	identityOK  bool
}

// New creates a client without invoking Clawpatch. Init performs the first
// connection and initialization when Scout owns the state directory.
func New(root string, config Config) (*Client, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve assessment root: %w", err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("stat assessment root: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("assessment root is not a directory: %s", absoluteRoot)
	}
	if config.BinaryPath == "" || !filepath.IsAbs(config.BinaryPath) {
		return nil, errors.New("Clawpatch binary path must be absolute")
	}
	binaryPath, err := filepath.EvalSymlinks(config.BinaryPath)
	if err != nil {
		return nil, fmt.Errorf("resolve Clawpatch binary: %w", err)
	}
	if err := validatePinnedBinaryPath(binaryPath); err != nil {
		return nil, err
	}

	stateDir := config.StateDir
	ownedState := StateDir(absoluteRoot)
	if stateDir == "" {
		stateDir = ownedState
	}
	if !filepath.IsAbs(stateDir) {
		return nil, errors.New("Clawpatch state directory must be absolute")
	}
	stateDir = filepath.Clean(stateDir)
	externalState := stateDir != ownedState
	if externalState && !config.AllowExternalState {
		return nil, errors.New("external Clawpatch state requires explicit opt-in")
	}

	nodePath := config.NodePath
	if nodePath == "" {
		nodePath, err = exec.LookPath("node")
		if err != nil {
			return nil, fmt.Errorf("locate Node runtime: %w", err)
		}
	}
	if !filepath.IsAbs(nodePath) {
		return nil, errors.New("Node runtime path must be absolute")
	}
	nodeInfo, err := os.Stat(nodePath)
	if err != nil || !nodeInfo.Mode().IsRegular() || nodeInfo.Mode().Perm()&0o111 == 0 {
		return nil, errors.New("Node runtime is not an executable file")
	}

	return &Client{
		root:             absoluteRoot,
		stateDir:         stateDir,
		binaryPath:       binaryPath,
		nodePath:         nodePath,
		externalState:    externalState,
		egressAuthorized: config.EgressAuthorized,
		provider:         config.Provider,
	}, nil
}

// Root returns the immutable assessment root.
func (c *Client) Root() string { return c.root }

// StateDir is where Scout keeps its Clawpatch state under an absolute root.
func StateDir(root string) string { return filepath.Join(root, ".scout", "clawpatch") }

// StateDir returns the configured Clawpatch state directory.
func (c *Client) StateDir() string { return c.stateDir }

// Init verifies the exact reviewed executable and initializes Scout-owned state
// exactly once. Existing external state is attached without reinitialization.
func (c *Client) Init(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.initialized {
		return nil
	}
	if err := c.verifyIdentityLocked(ctx); err != nil {
		return err
	}
	projectPath := filepath.Join(c.stateDir, "project.json")
	_, err := os.Stat(projectPath)
	switch {
	case err == nil:
		if err := c.ensureScoutExcluded(); err != nil {
			return err
		}
		c.initialized = true
		return nil
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("stat Clawpatch project state: %w", err)
	case c.externalState:
		return ErrExternalStateRequiresAttachment
	}
	if err := os.MkdirAll(c.stateDir, 0o700); err != nil {
		return fmt.Errorf("create Clawpatch state directory: %w", err)
	}
	if _, err := c.invokeLocked(ctx, "init"); err != nil {
		return fmt.Errorf("initialize Clawpatch state: %w", err)
	}
	if err := c.ensureScoutExcluded(); err != nil {
		return err
	}
	c.initialized = true
	return nil
}

// MapSource selects Clawpatch's mapper. Heuristic mapping is canonical and
// never contacts a provider; agent mapping needs provider egress.
type MapSource string

const (
	MapSourceHeuristic MapSource = "heuristic"
	MapSourceAgent     MapSource = "agent"
)

// Map maps only when its inputs changed. On an unchanged map it reads the prior
// authoritative records without a model or mapper invocation. A heuristic map
// runs with egress off. An agent map requires egress; if it fails, the
// heuristic map is kept and the result reports exploration as incomplete.
func (c *Client) Map(ctx context.Context, inputsChanged bool, source MapSource) (MapResult, error) {
	if source != MapSourceHeuristic && source != MapSourceAgent {
		err := fmt.Errorf("unknown Clawpatch map source %q", source)
		return MapResult{Status: MapBlocked, Reason: err.Error()}, err
	}
	if !inputsChanged {
		features, err := LoadFeatures(c.stateDir)
		if err != nil {
			return MapResult{Status: MapBlocked, Source: source, Reason: "canonical Clawpatch records are unavailable"}, err
		}
		return MapResult{Status: MapReused, Source: source, Features: features}, nil
	}
	var agentArgs []string
	var agentEnv []string
	if source == MapSourceAgent {
		provider, err := c.resolveMapProvider()
		if errors.Is(err, modelcli.ErrNoCLI) {
			return MapResult{Status: MapSkipped, Source: source, Reason: SkippedNoModelCLI}, nil
		}
		if err != nil {
			return MapResult{Status: MapBlocked, Source: source, Reason: err.Error()}, err
		}
		if !c.egressAuthorized {
			err := errors.New("Clawpatch mapping provider egress is not authorized")
			return MapResult{Status: MapBlocked, Source: source, Reason: err.Error()}, err
		}
		if err := c.verifyReviewSafetyPatch(); err != nil {
			return MapResult{Status: MapBlocked, Source: source, Reason: err.Error()}, err
		}
		var providerArgs []string
		providerArgs, agentEnv = modelInvocation(provider, mapEnvironment()...)
		agentArgs = append([]string{"map", "--source", "agent"}, providerArgs...)
		agentArgs = append(agentArgs, "--model", mapModel, "--reasoning-effort", mapReasoningEffort)
	}
	if err := c.Init(ctx); err != nil {
		return MapResult{Status: MapBlocked, Source: source, Reason: err.Error()}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if source == MapSourceHeuristic {
		return c.mapHeuristicLocked(ctx)
	}
	if _, err := c.invokeEnvLocked(ctx, agentEnv, agentArgs...); err != nil {
		fallback, fallbackErr := c.mapHeuristicLocked(ctx)
		if fallbackErr != nil {
			return fallback, errors.Join(err, fallbackErr)
		}
		fallback.ExplorationIncomplete = true
		fallback.Reason = "agent mapping did not complete; heuristic map kept and exploration is incomplete"
		return fallback, nil
	}
	features, err := LoadFeatures(c.stateDir)
	if err != nil {
		return MapResult{Status: MapBlocked, Source: source, Reason: "canonical Clawpatch records are unavailable"}, err
	}
	if len(features) == 0 {
		return MapResult{Status: MapMapped, Source: source, ExplorationIncomplete: true, Reason: "agent mapping produced zero features"}, nil
	}
	return MapResult{Status: MapMapped, Source: source, Features: features}, nil
}

func (c *Client) mapHeuristicLocked(ctx context.Context) (MapResult, error) {
	if _, err := c.invokeLocked(ctx, "map", "--source", "heuristic"); err != nil {
		features, loadErr := LoadFeatures(c.stateDir)
		if loadErr != nil {
			return MapResult{Status: MapBlocked, Source: MapSourceHeuristic, Reason: "pinned Clawpatch map did not complete"}, errors.Join(err, loadErr)
		}
		return MapResult{
			Status:   MapBlocked,
			Source:   MapSourceHeuristic,
			Features: features,
			Reason:   "pinned Clawpatch map did not complete; prior canonical records were retained",
		}, err
	}
	features, err := LoadFeatures(c.stateDir)
	if err != nil {
		return MapResult{Status: MapBlocked, Source: MapSourceHeuristic, Reason: "canonical Clawpatch records are unavailable"}, err
	}
	return MapResult{Status: MapMapped, Source: MapSourceHeuristic, Features: features}, nil
}

// CheckPrerequisites verifies the pinned executable without creating target
// state or contacting a provider.
func (c *Client) CheckPrerequisites(ctx context.Context) error {
	if err := c.verifyReviewSafetyPatch(); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.verifyIdentityLocked(ctx)
}

func (c *Client) verifyIdentityLocked(ctx context.Context) error {
	if c.identityOK {
		return nil
	}
	output, err := c.invokeLocked(ctx, "--version")
	if err != nil {
		return fmt.Errorf("verify pinned Clawpatch identity: %w", err)
	}
	if strings.TrimSpace(string(output)) != Version {
		return fmt.Errorf("Clawpatch reports a non-pinned version; expected %s", Version)
	}
	c.identityOK = true
	return nil
}

func (c *Client) invokeLocked(ctx context.Context, args ...string) ([]byte, error) {
	return c.invokeEnvLocked(ctx, nil, args...)
}

// mapEnvironment returns the Clawpatch provider timeouts for the map
// invocation. Clawpatch's own defaults (codex 300 s, claude 180 s) are too
// short for an agent map (2–7 min observed), so both providers get 900000 ms
// unless Scout's environment sets CLAWPATCH_CODEX_TIMEOUT_MS /
// CLAWPATCH_CLAUDE_TIMEOUT_MS, which win.
func mapEnvironment() []string {
	env := make([]string, 0, 2)
	for _, name := range []string{"CLAWPATCH_CODEX_TIMEOUT_MS", "CLAWPATCH_CLAUDE_TIMEOUT_MS"} {
		timeout := strings.TrimSpace(os.Getenv(name))
		if timeout == "" {
			timeout = defaultMapTimeoutMS
		}
		env = append(env, name+"="+timeout)
	}
	return env
}

func (c *Client) invokeEnvLocked(ctx context.Context, extraEnv []string, args ...string) ([]byte, error) {
	commandArgs := make([]string, 0, len(args)+6)
	commandArgs = append(commandArgs, c.binaryPath, "--root", c.root, "--state-dir", c.stateDir, "--json")
	commandArgs = append(commandArgs, args...)
	command := exec.CommandContext(ctx, c.nodePath, commandArgs...)
	// No Clawpatch call ever sees an API key or provider re-route (REQ-31).
	command.Env = modelcli.ScrubEnv(os.Environ(), extraEnv...)
	output, err := command.Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, context.DeadlineExceeded
		}
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			if tail := stderrTail(exit.Stderr); tail != "" {
				return nil, fmt.Errorf("pinned Clawpatch command failed: %w: %s", err, tail)
			}
		}
		return nil, fmt.Errorf("pinned Clawpatch command failed: %w", err)
	}
	return output, nil
}

// stderrTail is the redacted last stderrTailBytes of a failed command's stderr.
func stderrTail(stderr []byte) string {
	if len(stderr) > stderrTailBytes {
		stderr = stderr[len(stderr)-stderrTailBytes:]
	}
	return sanitizeFindingText(compactWhitespace(string(stderr)))
}

func validatePinnedBinaryPath(path string) error {
	if filepath.Base(path) != "cli.js" || filepath.Base(filepath.Dir(path)) != "dist" {
		return errors.New("Clawpatch binary must be the pinned dist/cli.js entrypoint")
	}
	if filepath.Base(filepath.Dir(filepath.Dir(path))) != Revision {
		return fmt.Errorf("Clawpatch binary must be built from pinned revision %s", Revision)
	}
	return nil
}

// MapStatus distinguishes an executed canonical map from record reuse.
type MapStatus string

const (
	MapMapped  MapStatus = "mapped"
	MapReused  MapStatus = "reused"
	MapBlocked MapStatus = "blocked"
	// MapSkipped means no model CLI exists for an agent map; heuristic records
	// are untouched and the lane reports `skipped (no model CLI)`.
	MapSkipped MapStatus = "skipped"
)

// MapResult never substitutes a partial map for a successful map.
// ExplorationIncomplete marks a requested agent map that failed or found
// nothing; Features then hold the heuristic map.
type MapResult struct {
	Status                MapStatus
	Source                MapSource
	Features              []domain.Feature
	Reason                string
	ExplorationIncomplete bool
}
