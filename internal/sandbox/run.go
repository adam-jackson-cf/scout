package sandbox

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Request describes one sandboxed scanner process.
type Request struct {
	WorkDir   string
	State     string // private writable directory; also HOME and TMPDIR
	RulesRoot string // optional read-only rule directory
	Binary    string // resolved scanner executable
	Args      []string
	ExtraPath []string // directories searched before the system PATH
	// StdoutLimit and StderrLimit bound each stream in bytes; output beyond a
	// limit is dropped and reported by Capture.Exceeded.
	StdoutLimit, StderrLimit int64
}

// Run executes the request. The scanner inherits nothing from the caller's
// environment: provider credentials and user configuration never reach it.
func Run(ctx context.Context, request Request) (stdout, stderr *Capture, err error) {
	stdout, stderr = &Capture{limit: request.StdoutLimit}, &Capture{limit: request.StderrLimit}
	command, err := command(ctx, request.WorkDir, request.State, request.RulesRoot, request.Binary, request.Args)
	if err != nil {
		return stdout, stderr, err
	}
	command.Stdout, command.Stderr = stdout, stderr
	command.Env = environment(request.State, request.ExtraPath)
	return stdout, stderr, command.Run()
}

func environment(state string, extraPath []string) []string {
	path := append(append([]string{}, extraPath...), "/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin")
	return []string{
		"HOME=" + state,
		"TMPDIR=" + state,
		"PATH=" + strings.Join(path, ":"),
		"SEMGREP_ENABLE_VERSION_CHECK=0",
		"SEMGREP_SEND_METRICS=off",
		"LC_ALL=C",
	}
}

// Capture is a concurrency-safe output buffer that keeps at most its limit.
type Capture struct {
	mu       sync.Mutex
	limit    int64
	buffer   bytes.Buffer
	exceeded bool
}

// Write keeps what fits and records any overflow; it never fails the process.
func (c *Capture) Write(value []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	available := c.limit - int64(c.buffer.Len())
	if available <= 0 {
		c.exceeded = true
		return len(value), nil
	}
	count := int64(len(value))
	if count > available {
		count, c.exceeded = available, true
	}
	c.buffer.Write(value[:count])
	return len(value), nil
}

// Bytes returns a copy of the captured output.
func (c *Capture) Bytes() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]byte(nil), c.buffer.Bytes()...)
}

// Exceeded reports whether output was dropped at the limit.
func (c *Capture) Exceeded() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.exceeded
}

// NewState creates a private (0700) scanner state directory with symlinks
// resolved, so sandbox write rules and path checks see the same path.
func NewState(prefix string) (string, error) {
	state, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(state)
	if err != nil {
		os.RemoveAll(state)
		return "", err
	}
	if err := os.Chmod(resolved, 0700); err != nil {
		os.RemoveAll(resolved)
		return "", err
	}
	return resolved, nil
}

// VerifyVersion runs request (a version invocation) and requires the output to
// name exactly the pinned version as a token. Errors never carry scanner output.
func VerifyVersion(ctx context.Context, request Request, want string) error {
	stdout, stderr, err := Run(ctx, request)
	if err != nil || ctx.Err() != nil || stdout.Exceeded() || stderr.Exceeded() {
		return errors.New("version command failed")
	}
	if !ContainsVersion(stdout.Bytes(), want) {
		return errors.New("version mismatch")
	}
	return nil
}

// ContainsVersion reports whether output has a whitespace- or punctuation-
// delimited token equal to version (optionally prefixed by "v").
func ContainsVersion(output []byte, version string) bool {
	for _, token := range strings.FieldsFunc(string(output), func(r rune) bool {
		return !(r == '.' || r == '-' || r == '_' || r == '+' || ('0' <= r && r <= '9') || ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z'))
	}) {
		if token == version || strings.TrimPrefix(token, "v") == version {
			return true
		}
	}
	return false
}
