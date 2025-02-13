package version

import "testing"

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
