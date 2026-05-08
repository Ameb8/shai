package {{.Package}}

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test{{.FunctionName}}(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
		wantErr  bool
	}{
		{
			name:     "initial case",
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// implementation here
		})
	}
}
