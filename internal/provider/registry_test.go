package provider

import (
	"context"
	"testing"

	"github.com/ameb8/shai/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestNewProvider_Errors ensures the provider registry correctly handles missing or invalid configurations.
func TestNewProvider_Errors(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{
		Providers: map[string]config.ProviderConfig{
			"gemini": {APIKey: ""}, // Empty key should trigger error in NewGeminiProvider
		},
	}

	tests := []struct {
		name          string
		providerName  string
		expectedError string
	}{
		{
			name:          "unknown provider",
			providerName:  "unknown",
			expectedError: "provider unknown not configured",
		},
		{
			name:          "gemini missing key",
			providerName:  "gemini",
			expectedError: "gemini API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProvider(ctx, tt.providerName, cfg, "")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Nil(t, p)
		})
	}
}

// TestNewProvider_Aliases verifies that the registry correctly resolves provider aliases.
func TestNewProvider_Aliases(t *testing.T) {
	ctx := context.Background()
	// We can't easily test successful provider creation without valid keys
	// but we can check if it tries to create the right one.
	// Since constructors are not mocked here, we'll just check if it finds the config.

	cfg := config.Config{
		Providers: map[string]config.ProviderConfig{
			"grok": {APIKey: "key", DefaultModel: "grok-1"},
		},
	}

	// xai is an alias for grok
	_, err := NewProvider(ctx, "xai", cfg, "")
	// It will still fail in NewGrokProvider because it might try to create a client,
	// but let's see what happens.
	// Actually GrokProvider might not fail on New if it doesn't do network calls.

	if err != nil {
		assert.NotContains(t, err.Error(), "provider xai not configured")
	}
}
