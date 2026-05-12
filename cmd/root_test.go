package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ameb8/shai/internal/config"
)

// TestRootCmdVersion ensures the root command outputs the correct version, commit, and build date.
func TestRootCmdVersion(t *testing.T) {
	// Set temporary version info for the test and cleanup after.
	SetVersion("1.2.3", "abc123", "2026-05-12")
	t.Cleanup(func() {
		SetVersion("dev", "none", "unknown")
	})

	// Initialize the root command with an empty configuration.
	root := NewRootCmd(&config.Config{
		Providers: make(map[string]config.ProviderConfig),
		Models:    make(map[string]string),
	})

	// Execute the command with the version flag and capture stdout.
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Verify the version string matches the expected format.
	got := out.String()
	want := "shai version 1.2.3 (abc123, 2026-05-12)"
	if !strings.Contains(got, want) {
		t.Fatalf("version output = %q, want it to contain %q", got, want)
	}
}
