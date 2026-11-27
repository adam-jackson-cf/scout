package testrun

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func MachineID(goVersion string) string {
	parts := []string{runtime.GOOS, runtime.GOARCH, cpuModel(), goVersion}
	return strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(strings.Join(parts, "-")), "-"), "-")
}

func cpuModel() string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.CommandContext(ctx, "sysctl", "-n", "machdep.cpu.brand_string").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "linux":
		data, err := os.ReadFile("/proc/cpuinfo")
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if k, v, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(k) == "model name" {
					return strings.TrimSpace(v)
				}
			}
		}
	}
	return "unknown-cpu"
}
