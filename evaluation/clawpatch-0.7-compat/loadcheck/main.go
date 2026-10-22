// Throwaway: decodes every features/*.json in the given Clawpatch state dirs with
// Scout's own loader (clawpatch.DecodeFeature / LoadFeatures) and reports failures.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"scout/internal/clawpatch"
)

func main() {
	failed := false
	for _, state := range os.Args[1:] {
		files, _ := filepath.Glob(filepath.Join(state, "features", "*.json"))
		sort.Strings(files)
		ok, bad := 0, 0
		firstErrs := map[string]int{}
		for _, f := range files {
			raw, err := os.ReadFile(f)
			if err != nil {
				fmt.Println("read error", f, err)
				bad++
				continue
			}
			if _, err := clawpatch.DecodeFeature(raw); err != nil {
				bad++
				firstErrs[err.Error()]++
				continue
			}
			ok++
		}
		_, loadErr := clawpatch.LoadFeatures(state)
		fmt.Printf("%s: files=%d decoded=%d failed=%d LoadFeatures_err=%v\n", state, len(files), ok, bad, loadErr)
		for msg, n := range firstErrs {
			fmt.Printf("  %dx %s\n", n, msg)
		}
		if bad > 0 || loadErr != nil {
			failed = true
		}
	}
	if failed {
		os.Exit(1)
	}
}
