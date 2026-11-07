package association

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"scout/internal/areas"
	"scout/internal/modelcli"
	"scout/internal/policy"
)

// TerraPromptDigest pins the measured, data-not-instructions prompt.
const TerraPromptDigest = "ca24fd764b745dbcb261a1aa6a0d740710809e82a13a8a103e55bbc4d8555e76"

//go:embed terra_prompt.txt
var terraPrompt string

// TerraPrompt returns the exact pilot prompt bytes.
func TerraPrompt() string { return terraPrompt }

// TerraAnswer is the validated placement of a residual cluster.
type TerraAnswer struct {
	Decision   string `json:"decision"`
	AreaID     string `json:"area_id"`
	NewTitle   string `json:"new_title"`
	NewSummary string `json:"new_summary"`
}

// TerraDecider owns the pinned, isolated codex boundary.
type TerraDecider struct {
	config                      policy.AssociationFallback
	attempts                    int
	latency                     time.Duration
	raw, model, effort, version string
}

// NewTerraDecider configures the per-cluster boundary; Decide retries once.
func NewTerraDecider(config policy.AssociationFallback) *TerraDecider {
	return &TerraDecider{config: config}
}

var terraModel = regexp.MustCompile(`(?m)^model:\s*(\S+)`)
var terraEffort = regexp.MustCompile(`(?m)^reasoning effort:\s*(\S+)`)
var terraCommand = regexp.MustCompile(`(?m)^exec\b|"type"\s*:\s*"(?:command_execution|exec_command|function_call)"`)
var terraVersion = regexp.MustCompile(`(?m)^OpenAI Codex\s+(\S+)`)

// Decide sends only the supplied bounded state, never repository files or rules.
func (d *TerraDecider) Decide(ctx context.Context, input json.RawMessage, catalogue []areas.Area) (TerraAnswer, error) {
	sum := sha256.Sum256([]byte(terraPrompt))
	if hex.EncodeToString(sum[:]) != TerraPromptDigest {
		return TerraAnswer{}, fmt.Errorf("Terra prompt digest mismatch")
	}
	_, executable, err := modelcli.Detect("codex")
	if err != nil {
		return TerraAnswer{}, err
	}
	d.attempts, d.latency = 0, 0
	var answer TerraAnswer
	for range 2 {
		d.attempts++
		started := time.Now()
		answer, err = d.invoke(ctx, executable, input, catalogue)
		d.latency += time.Since(started)
		if err == nil {
			return answer, nil
		}
	}
	return TerraAnswer{}, err
}

func (d *TerraDecider) invoke(ctx context.Context, executable string, input json.RawMessage, catalogue []areas.Area) (TerraAnswer, error) {
	dir, err := os.MkdirTemp("", "scout-terra-")
	if err != nil {
		return TerraAnswer{}, err
	}
	defer os.RemoveAll(dir)
	// Keep the answer outside the working directory so codex starts in an empty directory.
	answerPath := filepath.Join(dir, "answer")
	work := filepath.Join(dir, "work")
	if err := os.Mkdir(work, 0700); err != nil {
		return TerraAnswer{}, err
	}
	d.raw = ""
	callCtx, cancel := context.WithTimeout(ctx, d.config.Timeout)
	defer cancel()
	cmd := exec.CommandContext(callCtx, executable, "exec", "-m", d.config.Model, "-c", "model_reasoning_effort="+d.config.Effort, "-s", "read-only", "--ephemeral", "--ignore-user-config", "--ignore-rules", "--skip-git-repo-check", "-C", work, "-o", answerPath, "-")
	cmd.Dir, cmd.Env = work, modelcli.Env()
	cmd.Stdin = strings.NewReader(terraPrompt + "\n=== STATE (JSON) ===\n" + string(input) + "\n")
	var diagnostic bytes.Buffer
	cmd.Stdout, cmd.Stderr = &diagnostic, &diagnostic
	err = cmd.Run()
	both := diagnostic.String()
	d.model, d.effort, d.version = "", "", ""
	if m := terraModel.FindStringSubmatch(both); m != nil {
		d.model = m[1]
	}
	if m := terraEffort.FindStringSubmatch(both); m != nil {
		d.effort = m[1]
	}
	if m := terraVersion.FindStringSubmatch(both); m != nil {
		d.version = m[1]
	}
	if err != nil {
		return TerraAnswer{}, fmt.Errorf("Terra process failed: %w", err)
	}
	if terraCommand.MatchString(both) {
		return TerraAnswer{}, fmt.Errorf("Terra executed a command")
	}
	if d.model != d.config.Model || d.effort != d.config.Effort {
		return TerraAnswer{}, fmt.Errorf("Terra model/effort mismatch")
	}
	raw, err := os.ReadFile(answerPath)
	if err != nil {
		return TerraAnswer{}, err
	}
	d.raw = string(raw)
	var answer TerraAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		return TerraAnswer{}, fmt.Errorf("Terra answer: %w", err)
	}
	switch answer.Decision {
	case "existing":
		for _, area := range catalogue {
			if area.Active() && area.ScoutAreaID != "" && area.ScoutAreaID == answer.AreaID {
				return answer, nil
			}
		}
		return TerraAnswer{}, fmt.Errorf("Terra returned unknown area")
	case "new":
		if strings.TrimSpace(answer.NewTitle) != "" && strings.TrimSpace(answer.NewSummary) != "" {
			return answer, nil
		}
	}
	return TerraAnswer{}, fmt.Errorf("invalid Terra answer")
}
