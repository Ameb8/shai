package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateRunQueryArgs ensures that command validation enforces safety by allowing
// only read-only or whitelisted inspection commands while rejecting destructive actions.
func TestValidateRunQueryArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        RunQueryArgs
		shouldError bool
		errMsg      string
	}{
		{
			name: "ls allowed",
			args: RunQueryArgs{
				Executable: "ls",
				Args:       []string{"-la"},
			},
			shouldError: false,
		},
		{
			name: "path executable rejected",
			args: RunQueryArgs{
				Executable: "/bin/ls",
				Args:       []string{"-la"},
			},
			shouldError: true,
			errMsg:      "executable must be a bare command name",
		},
		{
			name: "docker-compose ps allowed",
			args: RunQueryArgs{
				Executable: "docker-compose",
				Args:       []string{"ps"},
			},
			shouldError: false,
		},
		{
			name: "docker-compose down rejected",
			args: RunQueryArgs{
				Executable: "docker-compose",
				Args:       []string{"down"},
			},
			shouldError: true,
			errMsg:      "subcommand \"down\" is not allowed",
		},
		{
			name: "git remote -v allowed",
			args: RunQueryArgs{
				Executable: "git",
				Args:       []string{"remote", "-v"},
			},
			shouldError: false,
		},
		{
			name: "git remote add rejected",
			args: RunQueryArgs{
				Executable: "git",
				Args:       []string{"remote", "add", "origin", "url"},
			},
			shouldError: true,
			errMsg:      "git remote command \"add\" is not allowed",
		},
		{
			name: "git branch -a allowed",
			args: RunQueryArgs{
				Executable: "git",
				Args:       []string{"branch", "-a"},
			},
			shouldError: false,
		},
		{
			name: "git branch -D rejected",
			args: RunQueryArgs{
				Executable: "git",
				Args:       []string{"branch", "-D", "main"},
			},
			shouldError: true,
			errMsg:      "git branch flag \"-D\" is not allowed",
		},
		{
			name: "docker compose ps allowed",
			args: RunQueryArgs{
				Executable: "docker",
				Args:       []string{"compose", "ps"},
			},
			shouldError: false,
		},
		{
			name: "curl GET allowed",
			args: RunQueryArgs{
				Executable: "curl",
				Args:       []string{"-sS", "-X", "GET", "https://example.com"},
			},
			shouldError: false,
		},
		{
			name: "curl POST rejected",
			args: RunQueryArgs{
				Executable: "curl",
				Args:       []string{"-X", "POST", "https://example.com"},
			},
			shouldError: true,
			errMsg:      "curl may only use GET requests",
		},
		{
			name: "curl file URL rejected",
			args: RunQueryArgs{
				Executable: "curl",
				Args:       []string{"file:///etc/passwd"},
			},
			shouldError: true,
			errMsg:      "curl URL scheme \"file\" is not allowed",
		},
		{
			name: "find -delete rejected",
			args: RunQueryArgs{
				Executable: "find",
				Args:       []string{".", "-delete"},
			},
			shouldError: true,
			errMsg:      "argument \"-delete\" is not allowed",
		},
	}

	// Execute table-driven tests to verify the whitelist and blacklist policy enforcement.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRunQueryArgs(tt.args)
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
