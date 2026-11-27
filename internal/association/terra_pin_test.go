package association

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestTerraPromptPin(t *testing.T) {
	sum := sha256.Sum256([]byte(terraPrompt))
	got := hex.EncodeToString(sum[:])
	const want = "ca24fd764b745dbcb261a1aa6a0d740710809e82a13a8a103e55bbc4d8555e76"
	if got != want || TerraPromptDigest != want {
		t.Fatalf("prompt digest %s constant %s", got, TerraPromptDigest)
	}
}

func TestTerraStalePrompt(t *testing.T) {
	staleProof(t, func(f *Fingerprint) { f.PromptDigest = "changed-terra-prompt" }, true)
}

func TestTerraStaleModel(t *testing.T) {
	staleProof(t, func(f *Fingerprint) { f.FallbackModel = "gpt-6.1-terra-other-pin" }, true)
}
