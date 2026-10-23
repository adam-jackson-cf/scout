// Package measures computes Scout's principal (snapshot) quality measures
// (design §8.2): complexity from lizard, duplication from jscpd, over the
// eligible production-source population of each feature or directory scope.
// Every signal is descriptive and carries numerator, denominator, population
// digest and tool/config digests; a measure that could not run is
// `unmeasured`, one without population is `insufficient`, never zero.
package measures

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Eligibility constants. The extension set and exclusion globs are the ones
// E1 preregistered (`evaluation/inertia-v1/collect.py` EXTS and
// `exclusions.json` "base"), so scope populations stay comparable with it.
const (
	// MinFileNLOC is the smallest file (lizard NLOC) in the eligible population.
	MinFileNLOC = 10
	// HighCC is the threshold above which a function counts as high complexity
	// (CC > 10).
	HighCC = 10
)

var sourceExtensions = map[string]bool{
	".py": true, ".js": true, ".mjs": true, ".cjs": true, ".ts": true, ".tsx": true,
	".go": true, ".rs": true, ".java": true, ".cs": true,
}

// nonProductionGlobs name test, docs, example, benchmark, vendored, generated
// and tooling paths by role. `**/` matches zero or more directories.
var nonProductionGlobs = []string{
	"**/test/**", "**/tests/**", "**/__tests__/**", "**/testdata/**", "**/fixtures/**",
	"**/examples/**", "**/docs/**", "**/docs_src/**", "**/benchmark/**", "**/benchmarks/**",
	"**/vendor/**", "**/node_modules/**", "**/third_party/**", "**/generated/**",
	"**/*.d.ts", "**/*.test.*", "**/*.spec.*", "**/*_test.go", "**/*.pb.go",
	"**/*.Designer.cs", ".github/**",
}

var nonProductionPatterns = compileGlobs(nonProductionGlobs)

// compileGlobs mirrors E1's glob_re: `**/` crosses directories (and may match
// none), `**` matches anything, `*` and `?` stay within one path segment.
func compileGlobs(globs []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(globs))
	for _, glob := range globs {
		var out strings.Builder
		for i := 0; i < len(glob); {
			switch {
			case strings.HasPrefix(glob[i:], "**/"):
				out.WriteString("(?:.*/)?")
				i += 3
			case strings.HasPrefix(glob[i:], "**"):
				out.WriteString(".*")
				i += 2
			case glob[i] == '*':
				out.WriteString("[^/]*")
				i++
			case glob[i] == '?':
				out.WriteString("[^/]")
				i++
			default:
				out.WriteString(regexp.QuoteMeta(glob[i : i+1]))
				i++
			}
		}
		compiled = append(compiled, regexp.MustCompile("^"+out.String()+"$"))
	}
	return compiled
}

// IsProductionSource reports whether path is a source file of a supported
// language outside the test, docs, vendored and generated path roles.
func IsProductionSource(path string) bool {
	if !sourceExtensions[strings.ToLower(filepath.Ext(path))] {
		return false
	}
	for _, pattern := range nonProductionPatterns {
		if pattern.MatchString(path) {
			return false
		}
	}
	return true
}
