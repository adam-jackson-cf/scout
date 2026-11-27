package testrun

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type result struct {
	out     []byte
	exit    int
	elapsed time.Duration
	started time.Time
}

func goCommand(dir string, env []string, limit time.Duration, args ...string) (result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = env
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	start := time.Now()
	err := cmd.Run()
	res := result{out: out.Bytes(), elapsed: time.Since(start), started: start}
	if ctx.Err() != nil {
		return res, fmt.Errorf("go %s: exceeded %s", strings.Join(args, " "), limit)
	}
	if exit, ok := err.(*exec.ExitError); ok {
		res.exit = exit.ExitCode()
		return res, nil
	}
	return res, err
}

func goEnv(dir string, keys ...string) (map[string]string, error) {
	res, err := goCommand(dir, nil, time.Minute, append([]string{"env"}, keys...)...)
	if err != nil {
		return nil, err
	}
	if res.exit != 0 {
		return nil, fmt.Errorf("go env: %s", strings.TrimSpace(string(res.out)))
	}
	lines := strings.Split(strings.TrimRight(string(res.out), "\n"), "\n")
	out := map[string]string{}
	for i, k := range keys {
		if i < len(lines) {
			out[k] = lines[i]
		}
	}
	return out, nil
}
