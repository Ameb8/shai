package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ameb8/shai/cmd"
	"github.com/ameb8/shai/internal/testutil/mockprovider"
)

// TestCLI_OutputContract verifies that the CLI correctly displays the command
// and explanation in the expected format.
func TestCLI_OutputContract(t *testing.T) {
	finalContent := "find . -name '*.go'\n# find all Go source files"
	// Setup a mock provider to simulate a standard LLM response.
	script := []mockprovider.Turn{
		{
			FinalContent: finalContent,
		},
	}
	cmd.ProviderOverride = mockprovider.New(t, script)
	t.Cleanup(func() { cmd.ProviderOverride = nil })

	// Execute the query command and capture output.
	stdout, stderr, err := RunCLI(t, "query", "--shell", "/bin/bash", "find all go files")

	if err != nil {
		t.Fatalf("RunCLI failed: %v\nStderr: %s", err, stderr)
	}
	// Validate that the output contains the command and its explanation.
	if !strings.Contains(stdout, "> find . -name '*.go'") {
		t.Errorf("expected stdout to contain command, got %q", stdout)
	}
	if !strings.Contains(stdout, "# find all Go source files") {
		t.Errorf("expected stdout to contain explanation, got %q", stdout)
	}
}

// TestCLI_NoExplain ensures that the --no-explain flag suppresses the explanation
// part of the agent's response.
func TestCLI_NoExplain(t *testing.T) {
	finalContent := "find . -name '*.go'\n# find all Go source files"
	// Mock a provider that returns both a command and an explanation.
	script := []mockprovider.Turn{
		{
			FinalContent: finalContent,
		},
	}
	cmd.ProviderOverride = mockprovider.New(t, script)
	t.Cleanup(func() { cmd.ProviderOverride = nil })

	// Run with the suppression flag enabled.
	stdout, _, err := RunCLI(t, "query", "--shell", "/bin/bash", "--no-explain", "find all go files")

	if err != nil {
		t.Fatalf("RunCLI failed: %v", err)
	}
	// Ensure only the command is present in the output.
	if !strings.Contains(stdout, "> find . -name '*.go'") {
		t.Errorf("expected stdout to contain command, got %q", stdout)
	}
	if strings.Contains(stdout, "# find all Go source files") {
		t.Errorf("expected stdout NOT to contain explanation, got %q", stdout)
	}
}

// TestCLI_DryRun verifies that the --dry-run flag prevents any side effects like
// writing the command to a file.
func TestCLI_DryRun(t *testing.T) {
	finalContent := "echo hello\n# print hello"
	script := []mockprovider.Turn{
		{
			FinalContent: finalContent,
		},
	}
	cmd.ProviderOverride = mockprovider.New(t, script)
	t.Cleanup(func() { cmd.ProviderOverride = nil })

	// Prepare a temporary path for the command file.
	tmpDir := t.TempDir()
	cmdFilePath := filepath.Join(tmpDir, "shai-output")

	// Execute with dry-run and command file flags.
	stdout, _, err := RunCLI(t, "query", "--shell", "/bin/bash", "--dry-run", "--cmd-file", cmdFilePath, "say hello")

	if err != nil {
		t.Fatalf("RunCLI failed: %v", err)
	}
	// Validate the command was printed but the file was not created.
	if !strings.Contains(stdout, "> echo hello") {
		t.Errorf("expected stdout to contain command, got %q", stdout)
	}
	if _, err := os.Stat(cmdFilePath); !os.IsNotExist(err) {
		t.Errorf("expected cmd-file NOT to exist in dry-run mode, but it does")
	}
}

// TestCLI_CmdFile validates that the --cmd-file flag correctly redirects the
// generated command to the specified file.
func TestCLI_CmdFile(t *testing.T) {
	finalContent := "echo hello\n# print hello"
	script := []mockprovider.Turn{
		{
			FinalContent: finalContent,
		},
	}
	cmd.ProviderOverride = mockprovider.New(t, script)
	t.Cleanup(func() { cmd.ProviderOverride = nil })

	// Define the output file for the command.
	tmpDir := t.TempDir()
	cmdFilePath := filepath.Join(tmpDir, "cmd.txt")

	// Run the query with file redirection.
	stdout, _, err := RunCLI(t, "query", "--shell", "/bin/bash", "--cmd-file", cmdFilePath, "say hello")

	if err != nil {
		t.Fatalf("RunCLI failed: %v", err)
	}
	// Verify the command was NOT printed to stdout.
	if strings.Contains(stdout, "> echo hello") {
		t.Errorf("expected stdout NOT to contain command when cmd-file is used, got %q", stdout)
	}

	// Read and verify the file content.
	content, err := os.ReadFile(cmdFilePath)
	if err != nil {
		t.Fatalf("failed to read cmd-file: %v", err)
	}
	if string(content) != "echo hello" {
		t.Errorf("expected cmd-file content %q, got %q", "echo hello", string(content))
	}
}
