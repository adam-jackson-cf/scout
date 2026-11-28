package policy

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(key, "GIT_") || key == "GIT_CEILING_DIRECTORIES" {
			continue
		}
		if err := os.Unsetenv(key); err != nil {
			fmt.Fprintln(os.Stderr, "unset", key+":", err)
			os.Exit(2)
		}
	}
	os.Exit(m.Run())
}
