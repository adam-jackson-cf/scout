package analyzers

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
)

const semgrepPythonUnsafeYAMLLoadRule = "security/semgrep-python-unsafe-yaml-load.yaml"
const semgrepActionsMutableRefRule = "security/semgrep-actions-mutable-ref.yaml"
const semgrepPythonMultipartRule = "security/semgrep-python-multipart-cve-2024-24762.yaml"
const semgrepJavaScriptCurrentScriptRule = "security/semgrep-javascript-current-script-dom-clobbering.yaml"
const semgrepCSSHeightTransitionRule = "performance/semgrep-css-height-transition.yaml"
const gitleaksDefaultRule = "security/gitleaks-default.toml"

//go:embed rules/security/semgrep-python-unsafe-yaml-load.yaml
var semgrepPythonUnsafeYAMLLoadRuleContent []byte

//go:embed rules/security/semgrep-actions-mutable-ref.yaml
var semgrepActionsMutableRefRuleContent []byte

//go:embed rules/security/semgrep-python-multipart-cve-2024-24762.yaml
var semgrepPythonMultipartRuleContent []byte

//go:embed rules/security/semgrep-javascript-current-script-dom-clobbering.yaml
var semgrepJavaScriptCurrentScriptRuleContent []byte

//go:embed rules/performance/semgrep-css-height-transition.yaml
var semgrepCSSHeightTransitionRuleContent []byte

//go:embed rules/security/gitleaks-default.toml
var gitleaksDefaultRuleContent []byte

func bundledRule(path string) ([]byte, error) {
	var content []byte
	switch path {
	case semgrepPythonUnsafeYAMLLoadRule:
		content = semgrepPythonUnsafeYAMLLoadRuleContent
	case semgrepActionsMutableRefRule:
		content = semgrepActionsMutableRefRuleContent
	case semgrepPythonMultipartRule:
		content = semgrepPythonMultipartRuleContent
	case semgrepJavaScriptCurrentScriptRule:
		content = semgrepJavaScriptCurrentScriptRuleContent
	case semgrepCSSHeightTransitionRule:
		content = semgrepCSSHeightTransitionRuleContent
	case gitleaksDefaultRule:
		content = gitleaksDefaultRuleContent
	default:
		return nil, errors.New("unknown bundled rule: " + path)
	}
	return append([]byte(nil), content...), nil
}

func bundledRuleDigest(path string) (string, error) {
	content, err := bundledRule(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}
