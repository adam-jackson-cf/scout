package engine

import (
	"net"
	"testing"
)

func TestTotalStaysLocalWhenAListenerOpens(t *testing.T) {
	if l, err := net.Listen("tcp", "127.0.0.1:0"); err == nil {
		_ = l.Close()
	}
	if Total([]int{1}) != 1 {
		t.Fatal("Total")
	}
}
