package pathscope

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"scout/tools/gate/check"
)

type Hit struct {
	File string
	Line int
	Text string
}

func (s *Scope) Grep(words []string) ([]Hit, error) {
	args := []string{"grep", "-n", "-I", "-w", "-F", "--full-name", "-z"}
	if s.Run.Mode == check.Commit {
		args = append(args, "--cached")
	}
	for _, w := range words {
		args = append(args, "-e", w)
	}
	if s.Run.Mode != check.Commit {
		args = append(args, "HEAD")
	}
	args = append(args, "--", ".")
	for _, pat := range s.exclude {
		args = append(args, ":(exclude,glob)"+pat)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = s.Run.Root
	cmd.Env = s.Run.GitEnv()
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 && errb.Len() == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("git grep: %w: %s", err, strings.TrimSpace(errb.String()))
	}
	return parseGrep(out.String()), nil
}

func parseGrep(raw string) []Hit {
	var hits []Hit
	for _, line := range strings.Split(raw, "\n") {
		parts := strings.SplitN(line, "\x00", 3)
		if len(parts) < 3 {
			continue
		}
		file := strings.TrimPrefix(parts[0], "HEAD:")
		n, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		hits = append(hits, Hit{File: file, Line: n, Text: parts[2]})
	}
	return hits
}
