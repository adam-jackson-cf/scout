package issues

import "fmt"

var mintID = func(n int) string { return fmt.Sprintf("id-%d", n) }

func Open(n int) string { return mintID(n) }
