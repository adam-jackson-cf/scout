package violating

import "os/exec"

func Version() ([]byte, error) {
	return exec.Command("git", "--version").Output()
}
