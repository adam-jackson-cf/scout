// Package codex invokes the authenticated Codex CLI without granting it direct
// access to the assessed repository. Callers must redact prompts before Run.
package codex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const MaxPromptBytes = 4 << 20

// Check confirms a usable local subscription and a source-denying OS sandbox.
// It does not read repository source or disclose authentication output.
func Check(ctx context.Context, sourceRoot string) (string, error) {
	if _, err := deniedRoot(sourceRoot); err != nil {
		return "", err
	}
	if runtime.GOOS != "darwin" {
		return "", errors.New("Codex source isolation requires the Darwin sandbox")
	}
	if _, err := exec.LookPath("sandbox-exec"); err != nil {
		return "", errors.New("Codex source isolation sandbox is unavailable")
	}
	binary, err := exec.LookPath("codex")
	if err != nil {
		return "", errors.New("authenticated Codex CLI is unavailable")
	}
	check := exec.CommandContext(ctx, binary, "login", "status")
	check.Stdout, check.Stderr = io.Discard, io.Discard
	if err := check.Run(); err != nil {
		return "", errors.New("Codex CLI subscription is not authenticated")
	}
	return binary, nil
}

// Run executes one structured prompt in an empty directory. The OS sandbox
// denies reads beneath sourceRoot even if untrusted prompt text asks for tools.
func Run(parent context.Context, sourceRoot, executable string, prompt, schema []byte, model string, timeout time.Duration) ([]byte, error) {
	root, err := deniedRoot(sourceRoot)
	if err != nil {
		return nil, err
	}
	if runtime.GOOS != "darwin" {
		return nil, errors.New("Codex source isolation requires the Darwin sandbox")
	}
	if len(prompt) == 0 || len(prompt) > MaxPromptBytes {
		return nil, errors.New("Codex prompt is empty or exceeds its byte limit")
	}
	if executable == "" || !filepath.IsAbs(executable) {
		return nil, errors.New("Codex executable must be an absolute path")
	}
	if len(schema) != 0 && !json.Valid(schema) {
		return nil, errors.New("Codex output schema is invalid JSON")
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	work, err := os.MkdirTemp("", "scout-codex-")
	if err != nil {
		return nil, fmt.Errorf("create isolated Codex workspace: %w", err)
	}
	defer os.RemoveAll(work)
	args := []string{"exec", "--ephemeral", "--ignore-user-config", "--skip-git-repo-check", "--sandbox", "read-only", "--cd", work}
	if model != "" && model != "default" {
		args = append(args, "--model", model)
	}
	if len(schema) != 0 {
		schemaPath := filepath.Join(work, "schema.json")
		if err := os.WriteFile(schemaPath, schema, 0o600); err != nil {
			return nil, fmt.Errorf("write Codex output schema: %w", err)
		}
		args = append(args, "--output-schema", schemaPath)
	}
	outputPath := filepath.Join(work, "output.json")
	args = append(args, "--output-last-message", outputPath, "-")
	profile := "(version 1)(allow default)(deny file-read* (subpath " + strconv.Quote(root) + "))"
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	command := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", append([]string{"-p", profile, executable}, args...)...)
	command.Dir = work
	command.Stdin = bytes.NewReader(prompt)
	// Codex output may echo prompt/source on errors; never publish its streams.
	command.Stdout, command.Stderr = io.Discard, io.Discard
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("Codex provider timed out or was cancelled: %w", ctx.Err())
		}
		return nil, errors.New("Codex provider failed without a structured response")
	}
	output, err := os.ReadFile(outputPath)
	if err != nil || len(output) == 0 || len(output) > MaxPromptBytes || !json.Valid(output) {
		return nil, errors.New("Codex provider returned no valid bounded JSON response")
	}
	return output, nil
}

func deniedRoot(sourceRoot string) (string, error) {
	if sourceRoot == "" {
		return "", errors.New("Codex source isolation requires an assessment root")
	}
	root, err := filepath.EvalSymlinks(sourceRoot)
	if err != nil {
		return "", fmt.Errorf("resolve assessment root for Codex isolation: %w", err)
	}
	root, err = filepath.Abs(root)
	if err != nil || root == string(filepath.Separator) {
		return "", errors.New("Codex source isolation requires a bounded assessment root")
	}
	return root, nil
}
