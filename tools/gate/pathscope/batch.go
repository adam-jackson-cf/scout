package pathscope

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func (s *Scope) ReadHeadMany(paths []string) (map[string][]byte, error) {
	if s.Fixture != "" || s.Full || s.Run.Explicit {
		return s.readDiskMany(paths)
	}
	raw, err := s.catFileBatch(paths)
	if err != nil {
		return nil, err
	}
	return parseBatch(raw, paths)
}

func (s *Scope) readDiskMany(paths []string) (map[string][]byte, error) {
	out := make(map[string][]byte, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(s.Disk(p))
		if err != nil {
			return nil, err
		}
		out[p] = data
	}
	return out, nil
}

func (s *Scope) catFileBatch(paths []string) ([]byte, error) {
	var in bytes.Buffer
	for _, p := range paths {
		fmt.Fprintf(&in, "%s:%s\n", s.HeadRev(), p)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "cat-file", "--batch")
	cmd.Dir, cmd.Stdin, cmd.Env = s.Run.Root, &in, s.Run.GitEnv()
	var errb bytes.Buffer
	cmd.Stderr = &errb
	raw, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git cat-file --batch: %w: %s", err, strings.TrimSpace(errb.String()))
	}
	return raw, nil
}

func parseBatch(raw []byte, paths []string) (map[string][]byte, error) {
	out := make(map[string][]byte, len(paths))
	rd := bufio.NewReader(bytes.NewReader(raw))
	for _, p := range paths {
		header, err := rd.ReadString('\n')
		if err != nil {
			return nil, err
		}
		fields := strings.Fields(header)
		if len(fields) != 3 {
			return nil, fmt.Errorf("git cat-file: %s: %s", p, strings.TrimSpace(header))
		}
		size, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		data := make([]byte, size+1)
		if _, err := io.ReadFull(rd, data); err != nil {
			return nil, err
		}
		out[p] = data[:size]
	}
	return out, nil
}
