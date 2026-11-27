//go:build boundary

package runner

import (
	"os/exec"
	"testing"
)

func TestGitReportsItsVersion(t *testing.T) {
	if err := exec.Command("git", "--version").Run(); err != nil {
		t.Fatal(err)
	}
}
