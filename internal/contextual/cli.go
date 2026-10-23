package contextual

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"scout/internal/modelcli"
)

const maximumCLIOutputBytes = 2 * 1024 * 1024

// claudeArgs is the fixed, tool-less invocation: no tools, no MCP servers, no
// slash commands, no browser, no session persistence, instructions as the
// system prompt, source on stdin. The model flag is the only variable part.
func claudeArgs(model string) []string {
	args := []string{
		"-p", "--output-format", "json",
		"--tools", "",
		"--permission-mode", "dontAsk",
		"--no-session-persistence",
		"--disable-slash-commands",
		"--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`,
		"--no-chrome",
		"--system-prompt", reviewInstructions,
	}
	if model != "" {
		args = append(args, "--model", model)
	}
	return args
}

// codexArgs has no "no tools" switch, so it combines a read-only sandbox, an
// ephemeral session, and an empty working directory (the subprocess cwd) with
// the fixed prompt on stdin ("-"). The answer is read from lastMessage.
func codexArgs(model, lastMessage string) []string {
	args := []string{
		"exec",
		"--sandbox", "read-only",
		"--skip-git-repo-check",
		"--ephemeral",
		"--output-last-message", lastMessage,
	}
	if model != "" {
		args = append(args, "--model", model)
	}
	return append(args, "-")
}

// run executes the CLI once and returns the model's raw answer text, which the
// caller parses strictly, with the model the CLI reports it resolved to.
func (c *Client) run(parent context.Context, cli modelcli.Name, binary string, prompt []byte) ([]byte, ModelRecord, error) {
	ctx, cancel := context.WithTimeout(parent, c.config.Timeout)
	defer cancel()
	workdir, err := os.MkdirTemp("", "scout-model-cli-")
	if err != nil {
		return nil, ModelRecord{}, err
	}
	defer os.RemoveAll(workdir)

	var args []string
	var stdin []byte
	lastMessage := filepath.Join(workdir, "last-message.txt")
	switch cli {
	case modelcli.Claude:
		args, stdin = claudeArgs(c.config.Model), prompt
	case modelcli.Codex:
		args = codexArgs(c.config.Model, lastMessage)
		stdin = append(append([]byte(reviewInstructions), '\n', '\n'), prompt...)
	default:
		return nil, ModelRecord{}, fmt.Errorf("unsupported model CLI %q", cli)
	}
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = workdir
	command.Env = modelcli.Env()
	command.Stdin = bytes.NewReader(stdin)
	var stdout, stderr limitedBuffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, ModelRecord{}, ctx.Err()
		}
		return nil, ModelRecord{}, fmt.Errorf("%s exited unsuccessfully", cli)
	}
	if stdout.over || stderr.over {
		return nil, ModelRecord{}, errors.New("model CLI output exceeds safe size limit")
	}

	switch cli {
	case modelcli.Claude:
		return parseClaude(stdout.Bytes())
	default:
		answer, err := os.ReadFile(lastMessage)
		if err != nil {
			return nil, ModelRecord{}, errors.New("codex produced no final message")
		}
		if len(bytes.TrimSpace(answer)) == 0 {
			return nil, ModelRecord{}, errors.New("codex final message is empty")
		}
		return answer, ModelRecord{CLI: string(cli), ResolvedModel: codexModel(stdout.Bytes(), stderr.Bytes())}, nil
	}
}

// parseClaude reads `claude -p --output-format json`: one result object whose
// `result` is the model text and whose `modelUsage` keys name the resolved model.
func parseClaude(raw []byte) ([]byte, ModelRecord, error) {
	var envelope struct {
		Type       string                     `json:"type"`
		IsError    bool                       `json:"is_error"`
		Result     *string                    `json:"result"`
		ModelUsage map[string]json.RawMessage `json:"modelUsage"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, ModelRecord{}, errors.New("claude output is not a JSON result object")
	}
	if envelope.IsError || envelope.Result == nil || strings.TrimSpace(*envelope.Result) == "" {
		return nil, ModelRecord{}, errors.New("claude reported an error or an empty result")
	}
	models := make([]string, 0, len(envelope.ModelUsage))
	for name := range envelope.ModelUsage {
		models = append(models, name)
	}
	sort.Strings(models)
	resolved := strings.Join(models, ",")
	if resolved == "" {
		resolved = "claude:unreported"
	}
	return []byte(*envelope.Result), ModelRecord{CLI: string(modelcli.Claude), ResolvedModel: resolved}, nil
}

var codexModelLine = regexp.MustCompile(`(?m)^\s*model:\s*(\S+)\s*$`)

// codexModel reads the model line codex prints in its run header
// [INFERENCE: format taken from codex exec's banner; "codex:unreported" when absent].
func codexModel(stdout, stderr []byte) string {
	for _, stream := range [][]byte{stderr, stdout} {
		if match := codexModelLine.FindSubmatch(stream); match != nil {
			return string(match[1])
		}
	}
	return "codex:unreported"
}

// limitedBuffer keeps at most maximumCLIOutputBytes and notes any overflow.
type limitedBuffer struct {
	bytes.Buffer
	over bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if room := maximumCLIOutputBytes - b.Len(); len(p) > room {
		b.over = true
		if room > 0 {
			b.Buffer.Write(p[:room])
		}
		return len(p), nil
	}
	return b.Buffer.Write(p)
}
