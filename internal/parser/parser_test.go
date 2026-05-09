package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParse verifies that the Parse function correctly extracts commands,
// explanations, and warning flags from various LLM response formats.
func TestParse(t *testing.T) {
	// Define test cases covering simple commands, markdown fences, shell prefixes, and edge cases.
	tests := []struct {
		name     string
		raw      string
		expected ParsedResponse
	}{
		{
			name: "simple command and explanation",
			raw: `ls -la
# List all files including hidden ones`,
			expected: ParsedResponse{
				Command:     "ls -la",
				Explanation: []string{"List all files including hidden ones"},
				HasWarning:  false,
			},
		},
		{
			name: "command with warning",
			raw: `rm -rf /
# Delete everything
# WARNING: This is destructive`,
			expected: ParsedResponse{
				Command:     "rm -rf /",
				Explanation: []string{"Delete everything", "WARNING: This is destructive"},
				HasWarning:  true,
			},
		},
		{
			name: "markdown code fences",
			raw:  "```bash\ndu -sh .\n# Show disk usage\n```",
			expected: ParsedResponse{
				Command:     "du -sh .",
				Explanation: []string{"Show disk usage"},
				HasWarning:  false,
			},
		},
		{
			name: "multiple command lines",
			raw: `echo "hello"
echo "world"
# Multiple echoes`,
			expected: ParsedResponse{
				Command:     "echo \"hello\"\necho \"world\"",
				Explanation: []string{"Multiple echoes"},
				HasWarning:  false,
			},
		},
		{
			name: "command with $ or > prefix",
			raw: `$ ls
> pwd
# Shell prefixes should be stripped`,
			expected: ParsedResponse{
				Command:     "ls\npwd",
				Explanation: []string{"Shell prefixes should be stripped"},
				HasWarning:  false,
			},
		},
		{
			name: "empty input",
			raw:  "",
			expected: ParsedResponse{
				Command:     "",
				Explanation: nil,
				HasWarning:  false,
			},
		},
	}

	// Run subtests for each scenario to ensure parser robustness.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
