package redact

import (
	"strings"
	"testing"
)

func TestTextMasksEachCredentialClass(t *testing.T) {
	cases := map[string]string{
		"private key":    "x -----BEGIN RSA PRIVATE KEY-----\nabc\n-----END RSA PRIVATE KEY----- y",
		"assignment":     "password = hunter2hunter2",
		"quoted key":     `{"api_key": "abcd1234"}`,
		"prefixed key":   "my_token=abcdef12",
		"url credential": "postgres://user:pass@host/db",
		"jwt":            "eyJhbGciOiJIUzI1NiJ9." + "eyJzdWIiOiIxMjM0NTY3ODkwIn0.abcdefghijklmnop",
		"aws key":        "AKIA" + "ABCDEFGHIJKLMNOP",
		"github token":   "ghp_abcdefghijklmnopqrstuvwxyz",
		"long token":     strings.Repeat("a1", 20),
	}
	secrets := map[string]string{
		"private key": "abc\n", "assignment": "hunter2hunter2", "quoted key": "abcd1234", "prefixed key": "abcdef12",
		"url credential": "user:pass", "jwt": "eyJhbGci", "aws key": "AKIA" + "ABCDEFGHIJKLMNOP",
		"github token": "ghp_abcdefghij", "long token": strings.Repeat("a1", 20),
	}
	for name, input := range cases {
		got, err := Text(input)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if strings.Contains(got, secrets[name]) || !strings.Contains(got, "REDACTED") {
			t.Errorf("%s: secret survived: %q", name, got)
		}
	}
}

func TestTextRejectsUnsafeInputAndKeepsOrdinaryProse(t *testing.T) {
	for _, bad := range []string{"a\x00b", "a\xffb"} {
		if _, err := Text(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	got, err := Text("tokenizer reads words\x07 here")
	if err != nil || got != "tokenizer reads words here" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestContainsFlagsCredentialsAndNotProse(t *testing.T) {
	if !Contains(`{"password":"x1"}`) || !Contains("a\x00") {
		t.Fatal("credential-bearing content was not flagged")
	}
	if Contains("the tokenizer is fast") {
		t.Fatal("ordinary prose flagged")
	}
}

func TestTruncateKeepsUTF8Boundary(t *testing.T) {
	got := Truncate("ééé", 3)
	if got != "é…" {
		t.Fatalf("got %q", got)
	}
	if Truncate("abc", 3) != "abc" {
		t.Fatal("short text changed")
	}
}
