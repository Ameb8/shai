package integration

import (
	"bytes"
	"testing"

	"github.com/ameb8/shai/cmd"
	"github.com/ameb8/shai/internal/config"
	"github.com/ameb8/shai/internal/prompt"
	"github.com/ameb8/shai/internal/provider"
)

// buildMessages constructs a standard message slice for Path 1 tests.
// Always passes "/bin/bash" as the shell override to make the system prompt deterministic.
func buildMessages(t *testing.T, query string) []provider.Message {
	sysPrompt, err := prompt.BuildSystemPrompt("/bin/bash")
	if err != nil {
		t.Fatalf("failed to build system prompt: %v", err)
	}

	return []provider.Message{
		{Role: "system", Content: sysPrompt},
		{Role: "user", Content: query},
	}
}

// minimalConfig returns a *config.Config with empty-but-initialized maps,
// suitable for Path 2 tests where ProviderOverride is used.
func minimalConfig() *config.Config {
	return &config.Config{
		Active:    config.ActiveConfig{Provider: "mock"},
		Providers: make(map[string]config.ProviderConfig),
		Models:    make(map[string]string),
	}
}

// RunCLI executes the shai command tree in-process and captures stdout/stderr.
// It uses NewRootCmd with a minimal configuration to avoid reading from disk.
func RunCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	outBuf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	root := cmd.NewRootCmd(minimalConfig())
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	root.SetArgs(args)

	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}
