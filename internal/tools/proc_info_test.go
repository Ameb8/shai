package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseProcInfoArgs verifies that raw input maps are correctly transformed
// into ProcInfoArgs structs, handling type conversions for ports and names.
func TestParseProcInfoArgs(t *testing.T) {
	// Define test cases covering various input formats and edge cases.
	tests := []struct {
		name     string
		raw      map[string]any
		expected ProcInfoArgs
		wantErr  bool
	}{
		{
			name: "full arguments",
			raw: map[string]any{
				"port": 3000.0,
				"name": "node",
				"user": "dev",
			},
			expected: ProcInfoArgs{
				Port: 3000,
				Name: "node",
				User: "dev",
			},
			wantErr: false,
		},
		{
			name: "integer port",
			raw: map[string]any{
				"port": 3000,
			},
			expected: ProcInfoArgs{
				Port: 3000,
			},
			wantErr: false,
		},
		{
			name: "string port",
			raw: map[string]any{
				"port": "8080",
			},
			expected: ProcInfoArgs{
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "invalid port type",
			raw: map[string]any{
				"port": true,
			},
			wantErr: true,
		},
		{
			name: "only name",
			raw: map[string]any{
				"name": "python",
			},
			expected: ProcInfoArgs{
				Name: "python",
			},
			wantErr: false,
		},
		{
			name:     "empty args",
			raw:      map[string]any{},
			expected: ProcInfoArgs{},
			wantErr:  false,
		},
	}

	// Execute subtests for each scenario to ensure robustness.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := parseProcInfoArgs(tt.raw)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, args)
			}
		})
	}
}
