package sandbox

import (
	"strings"
	"testing"
)

func TestContainsVersionMatchesWholeTokensOnly(t *testing.T) {
	for output, want := range map[string]bool{
		"1.21.7\n":             true,
		"lizard 1.21.7":        true,
		"ruff v0.15.8 (abc)":   true,
		"version: 1.21.70":     false,
		"1.21.7-rc1":           false,
		"11.21.7":              false,
		"":                     false,
		"semgrep\n1.182.0\nok": true,
		"pre-1.182.0":          false,
	} {
		version := "1.21.7"
		if strings.Contains(output, "0.15.8") {
			version = "0.15.8"
		}
		if strings.Contains(output, "1.182.0") {
			version = "1.182.0"
		}
		if got := ContainsVersion([]byte(output), version); got != want {
			t.Errorf("ContainsVersion(%q, %q) = %t, want %t", output, version, got, want)
		}
	}
}

func TestCaptureKeepsAtMostItsLimitAndReportsOverflow(t *testing.T) {
	capture := &Capture{limit: 5}
	if n, err := capture.Write([]byte("abc")); n != 3 || err != nil || capture.Exceeded() {
		t.Fatalf("first write = %d, %v, exceeded=%t", n, err, capture.Exceeded())
	}
	if n, err := capture.Write([]byte("defgh")); n != 5 || err != nil || !capture.Exceeded() {
		t.Fatalf("overflowing write = %d, %v, exceeded=%t", n, err, capture.Exceeded())
	}
	if got := string(capture.Bytes()); got != "abcde" {
		t.Fatalf("kept %q", got)
	}
	capture.Bytes()[0] = 'z'
	if string(capture.Bytes()) != "abcde" {
		t.Fatal("Bytes aliases the internal buffer")
	}
}

func TestEnvironmentInheritsNothingFromTheCaller(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "should-not-leak")
	t.Setenv("HOME", "/home/caller")
	env := environment("/state", []string{"/opt/tool"})
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "should-not-leak") || strings.Contains(joined, "/home/caller") {
		t.Fatalf("caller environment leaked: %v", env)
	}
	for _, want := range []string{"HOME=/state", "TMPDIR=/state", "PATH=/opt/tool:/usr/local/sbin:"} {
		if !strings.Contains(joined, want) {
			t.Errorf("environment lacks %q: %v", want, env)
		}
	}
}
