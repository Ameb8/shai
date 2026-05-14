package provider

import (
	"context"
	"testing"

	"github.com/ameb8/shai/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestNewGeminiProvider_BaseURL verifies that the Gemini provider correctly uses the BaseURL from configuration.
func TestNewGeminiProvider_BaseURL(t *testing.T) {
	ctx := context.Background()

	t.Run("default url", func(t *testing.T) {
		p, err := NewGeminiProvider(ctx, config.ProviderConfig{APIKey: "test-key"}, "")
		if err != nil {
			t.Fatalf("failed to create provider: %v", err)
		}
		assert.Equal(t, "", p.url)
	})

	t.Run("custom url", func(t *testing.T) {
		customURL := "https://custom.gemini.ai/v1"
		p, err := NewGeminiProvider(ctx, config.ProviderConfig{APIKey: "test-key", BaseURL: customURL}, "")
		if err != nil {
			t.Fatalf("failed to create provider: %v", err)
		}
		assert.Equal(t, customURL, p.url)
	})
}
// TestGeminiProvider_Name verifies that the Gemini provider returns the correct identifier.
func TestGeminiProvider_Name(t *testing.T) {
	p := &GeminiProvider{}
	assert.Equal(t, "gemini", p.Name())
}

// TestNewGeminiProvider_Errors ensures that the Gemini provider constructor handles invalid configurations correctly.
func TestNewGeminiProvider_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("missing api key", func(t *testing.T) {
		p, err := NewGeminiProvider(ctx, config.ProviderConfig{APIKey: ""}, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "gemini API key is required")
		assert.Nil(t, p)
	})
}
