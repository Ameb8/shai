package provider

import (
	"context"
	"fmt"

	"github.com/ameb8/shai/internal/config"
)

// NewProvider creates and initializes a concrete Provider implementation based on the requested name and configuration.
// It handles provider aliasing and resolves the appropriate model to use for completion requests.
func NewProvider(ctx context.Context, name string, cfg config.Config, modelOverride string) (Provider, error) {
	// Normalize provider names and handle common aliases (e.g., xai maps to grok).
	configName := name
	if name == "xai" {
		name = "grok"
	}

	// Retrieve provider-specific configuration from the global config object.
	pCfg, ok := cfg.Providers[name]
	if !ok && configName != name {
		pCfg, ok = cfg.Providers[configName]
	}
	if !ok {
		return nil, fmt.Errorf("provider %s not configured", configName)
	}

	// Determine which model to use, prioritizing manual overrides over default settings.
	modelName := modelOverride
	if modelName == "" {
		modelName = pCfg.DefaultModel
	}

	// Instantiate the specific provider implementation based on the resolved name.
	switch name {
	case "gemini":
		return NewGeminiProvider(ctx, pCfg, modelName)
	case "mistral":
		return NewMistralProvider(ctx, pCfg, modelName)
	case "grok", "xai":
		return NewGrokProvider(ctx, pCfg, modelName)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
