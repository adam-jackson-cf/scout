//go:build boundary

package sandbox

import (
	"context"
	"os"
	"runtime"
	"testing"
)

func TestRunConfinesWritesToTheStateDirectory(t *testing.T) {
	if err := Available(); err != nil {
		t.Fatalf("scanner sandbox is not available: %v", err)
	}
	state, err := NewState("scout-sandbox-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(state); err != nil {
			t.Errorf("remove sandbox state: %v", err)
		}
	})
	outside := t.TempDir()
	request := Request{WorkDir: state, State: state, Binary: "/usr/bin/touch", StdoutLimit: 1 << 10, StderrLimit: 1 << 10}

	request.Args = []string{state + "/inside"}
	if _, _, err := Run(context.Background(), request); err != nil {
		t.Fatalf("write inside the state directory failed: %v", err)
	}
	request.Args = []string{outside + "/escaped"}
	if _, _, err := Run(context.Background(), request); err == nil {
		t.Fatal("write outside the state directory succeeded")
	}
	if _, err := os.Stat(outside + "/escaped"); err == nil {
		t.Fatal("sandbox wrote outside the state directory")
	}
	if runtime.GOOS == "darwin" {
		request.Binary = "/bin/sh"
		openNullSinkReadWriteLikePythonDevnull := []string{"-c", "exec 3<> /dev/null"}
		request.Args = openNullSinkReadWriteLikePythonDevnull
		if _, _, err := Run(context.Background(), request); err != nil {
			t.Fatalf("opening the null sink read/write failed: %v", err)
		}
		request.Args = []string{"-c", "exec 3<> /dev/zero"}
		if _, _, err := Run(context.Background(), request); err == nil {
			t.Fatal("opening another device for writing succeeded")
		}
	}
}
