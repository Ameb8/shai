package prompt

import (
	"os"
	"runtime"
	"testing"

	"github.com/ameb8/shai/internal/sysenv"
	"github.com/stretchr/testify/assert"
)

// TestBuildSystemPrompt verifies that the system prompt is correctly constructed
// with the appropriate OS and shell information, respecting environment overrides.
func TestBuildSystemPrompt(t *testing.T) {
	// Preserve the original SHELL environment variable to avoid side effects on the host system.
	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)

	// Define scenarios for shell resolution, including environment defaults and manual overrides.
	tests := []struct {
		name          string
		shellOverride string
		envShell      string
		expectedOS    string
		expectedShell string
	}{
		{
			name:          "default shell from env",
			shellOverride: "",
			envShell:      "/bin/zsh",
			expectedOS:    runtime.GOOS,
			expectedShell: "/bin/zsh",
		},
		{
			name:          "shell override",
			shellOverride: "/bin/bash",
			envShell:      "/bin/zsh",
			expectedOS:    runtime.GOOS,
			expectedShell: "/bin/bash",
		},
		{
			name:          "no shell anywhere",
			shellOverride: "",
			envShell:      "",
			expectedOS:    runtime.GOOS,
			expectedShell: "unknown",
		},
	}

	// Execute each test case, mocking the shell environment as needed.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sysenv.ResetRuntimeForTest()
			if tt.envShell != "" {
				os.Setenv("SHELL", tt.envShell)
			} else {
				os.Unsetenv("SHELL")
			}

			prompt, err := BuildSystemPrompt(tt.shellOverride)
			assert.NoError(t, err)

			// Validate that the generated prompt contains critical identity and environment markers.
			assert.Contains(t, prompt, "You are shai")
			assert.Contains(t, prompt, tt.expectedOS)
			assert.Contains(t, prompt, tt.expectedShell)
		})
	}
}
