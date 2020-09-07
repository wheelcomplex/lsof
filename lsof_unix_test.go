// +build linux plan9 freebsd solaris

package lsof

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkScan(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := Scan(func(string) bool { return false })
		if err != nil {
			b.Fatal("Failed to scan:", err)
		}
	}
}

func TestScan(t *testing.T) {
	var pidFile = newFile(t)

	p, err := Scan(func(path string) bool {
		return pidFile == path
	})

	if err != nil {
		t.Fatal("Failed to scan:", err)
	}

	if len(p) != 1 {
		t.Fatal("Mismatch results length != 1:", len(p))
	}

	pids, ok := p[pidFile]
	if !ok {
		t.Fatal("PIDs not found:", p)
	}

	if len(pids) != 1 {
		t.Fatal("Mismatch PIDs length != 1:", len(pids))
	}

	if selfPID := os.Getpid(); pids[0] != selfPID {
		t.Fatal("Incorrect PID (found/self):", pids[0], "!=", selfPID)
	}
}

func newFile(t *testing.T) string {
	var path = filepath.Join(
		os.TempDir(),
		fmt.Sprintf("lsof-test-%d", os.Getpid()),
	)

	f, err := os.Create(path)
	if err != nil {
		t.Fatal("Failed to create temp file:", err)
	}

	t.Cleanup(func() {
		f.Close()
		os.Remove(path)
	})

	return path
}
