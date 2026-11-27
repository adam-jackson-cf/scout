//go:build linux

package fixture

import _ "embed"

//go:embed payload.txt
var payload string

var endpoint = "https://example.com/path"

var pattern = `/* raw string, not a comment */`

func run() error { return nil } //nolint:unused // kept for the canary
