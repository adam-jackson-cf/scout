package analyzers

import (
	"strings"
	"testing"
)

func TestSanitizeTextRemovesEveryCredentialClassBeforeEvidenceLeaves(t *testing.T) {
	secrets := map[string]string{
		"private key":    "-----BEGIN RSA " + "PRIVATE KEY-----\nMIIBOgIBAAJBAKj34GkxFhD90vcNLYLInFEX6Ppy1tPf9Cnzj4p4WGeKLs1Pt8Qu\n-----END RSA PRIVATE KEY-----",
		"assignment":     "password = hunter2hunter2",
		"url credential": "https://deploy:s3cretpw@example.com/repo.git",
		"jwt":            "eyJhbGciOiJIUzI1NiJ9." + "eyJzdWIiOiIxMjM0NTY3ODkwIn0.abcdefghijklmnop",
		"aws key":        "AKIA" + "ABCDEFGHIJKLMNOP",
		"github token":   "ghp_abcdefghijklmnopqrstuvwxyz0123",
		"long token":     strings.Repeat("Ab1", 15),
	}
	for name, secret := range secrets {
		got, err := sanitizeText("found "+secret+" in config", maxEvidenceText)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, fragment := range []string{"hunter2hunter2", "s3cretpw", "MIIBOgIBAAJB", "AKIA" + "ABCDEFGHIJKLMNOP", "ghp_abcdef", "eyJzdWIi", strings.Repeat("Ab1", 15)} {
			if strings.Contains(got, fragment) {
				t.Errorf("%s: %q survived sanitization: %q", name, fragment, got)
			}
		}
	}
}

func TestSanitizeTextRejectsUnsafeTextAndTruncatesOnRuneBoundary(t *testing.T) {
	for _, value := range []string{"nul\x00byte", "bad\xffutf8"} {
		if _, err := sanitizeText(value, maxEvidenceText); err == nil {
			t.Errorf("%q was accepted", value)
		}
	}
	if _, err := sanitizeText("text", 0); err == nil {
		t.Error("a zero bound was accepted")
	}
	got, err := sanitizeText(strings.Repeat("é", 10), 5)
	if err != nil || got != "éé…" {
		t.Fatalf("truncation = %q, %v", got, err)
	}
	got, err = sanitizeText("a\x07b\tc\n", maxEvidenceText)
	if err != nil || got != "ab\tc\n" {
		t.Fatalf("control characters = %q, %v", got, err)
	}
}
