package version

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"os"
	"runtime/debug"
	"testing"
)

func TestCommand(t *testing.T) {
	cmd := Command()
	if cmd == nil {
		t.Fatal("Expected command to be non-nil")
	}

	if cmd.Describe() != "print the version of sws" {
		t.Fatalf("Unexpected describe: %v", cmd.Describe())
	}

	if fs := cmd.Flagset(); fs == nil {
		t.Fatal("Expected flagset to be non-nil")
	}

	if help := cmd.Help(); help != "Print the version of sws" {
		t.Fatalf("Unexpected help output: %v", help)
	}
}

func TestRun(t *testing.T) {
	cmd := Command()
	ctx := context.Background()
	t.Run("it should print version info correctly", func(t *testing.T) {
		cmd.getVersionCmd = func() (*buildinfo.BuildInfo, bool) {
			return &buildinfo.BuildInfo{
				Main: debug.Module{
					Version: "v1.2.3",
					Sum:     "h1:checksum",
				},
				GoVersion: "go1.23.5",
			}, true
		}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := cmd.Run(ctx)
		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		// Читаем вывод
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		got := buf.String()

		expected := "version: v1.2.3, go version: go1.23.5, checksum: h1:checksum\n"
		if got != expected {
			t.Fatalf("Expected output %q, got %q", expected, got)
		}
	})
}
