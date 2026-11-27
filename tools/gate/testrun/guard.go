package testrun

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type sandbox struct {
	exec     string
	homeOnly string
	kills    bool
	summary  string
}

func realPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

func t1Guard(tmp, goTmp, realHome string) (sandbox, error) {
	switch runtime.GOOS {
	case "darwin":
		return darwinGuard(tmp, goTmp, realHome)
	case "linux":
		return linuxGuard(goTmp, realHome)
	}
	return sandbox{}, fmt.Errorf("no T1 guard for %s; run the gate on darwin or linux", runtime.GOOS)
}

func darwinGuard(tmp, goTmp, realHome string) (sandbox, error) {
	bin, err := exec.LookPath("sandbox-exec")
	if err != nil {
		return sandbox{}, errors.New("T1 guard needs /usr/bin/sandbox-exec")
	}
	profile := strings.Join([]string{
		"(version 1)",
		"(allow default)",
		"(deny process-fork (with send-signal SIGKILL))",
		"(deny process-exec (with send-signal SIGKILL))",
		fmt.Sprintf("(allow process-exec (subpath %q))", goTmp),
		"(deny network* (with send-signal SIGKILL))",
		fmt.Sprintf("(deny file-write* (with send-signal SIGKILL) (subpath %q))", realHome),
		"",
	}, "\n")
	path := filepath.Join(tmp, "t1-guard.sb")
	if err := os.WriteFile(path, []byte(profile), 0o600); err != nil {
		return sandbox{}, err
	}
	home := filepath.Join(tmp, "home-guard.sb")
	homeProfile := fmt.Sprintf("(version 1)\n(allow default)\n(deny file-write* (with send-signal SIGKILL) (subpath %q))\n", realHome)
	if err := os.WriteFile(home, []byte(homeProfile), 0o600); err != nil {
		return sandbox{}, err
	}
	return sandbox{exec: bin + " -f " + path, homeOnly: bin + " -f " + home, kills: true, summary: "sandbox-exec kills on process creation, network use or a write under " + realHome}, nil
}

func linuxGuard(goTmp, realHome string) (sandbox, error) {
	bin, err := exec.LookPath("bwrap")
	if err != nil {
		return sandbox{}, errors.New("T1 guard needs bubblewrap (bwrap); run scripts/ci-sandbox.sh")
	}
	args := []string{bin, "--unshare-net", "--die-with-parent", "--ro-bind", "/", "/", "--dev", "/dev", "--proc", "/proc", "--bind", os.TempDir(), os.TempDir(), "--bind", goTmp, goTmp}
	if !strings.HasPrefix(goTmp, realHome+"/") {
		args = append(args, "--ro-bind", realHome, realHome)
	}
	return sandbox{exec: strings.Join(args, " "), summary: "bubblewrap denies network and makes " + realHome + " read-only; process creation is not blocked on linux"}, nil
}
