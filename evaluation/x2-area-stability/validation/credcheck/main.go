// Command credcheck prints modelcli.IsCredential for each environment variable name read from stdin (one per line) as "name<TAB>0|1".
// validation/validate.py compares it with common.is_credential, so the collector's environment scrub is the Go one name for name.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"scout/internal/modelcli"
)

func main() {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		name := strings.TrimSpace(sc.Text())
		if name == "" {
			continue
		}
		v := 0
		if modelcli.IsCredential(name) {
			v = 1
		}
		fmt.Printf("%s\t%d\n", name, v)
	}
}
