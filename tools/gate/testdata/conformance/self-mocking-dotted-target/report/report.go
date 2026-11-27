package report

import "canary/clock"

func Stamp() string { return clock.Now().Format("2006") }
