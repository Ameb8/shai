package provider

import (
	"context"
	"fmt"
	"shai/internal/config"
)

// NewProvider returns a concrete provider implementation based on the name.
func NewProvider(ctx context.Context, name string, cfg config.Config, modelOverride string) (Provider, error) {
	pCfg, ok := cfg.Providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not configured", name)
	}

	modelName := modelOverride
	if modelName == "" {
		// Model resolution logic
		// 1. check for aliases in models map
		// This is a bit complex here as we don't know if modelOverride was an alias or not.
		// For now, let's keep it simple.
		modelName = pCfg.DefaultModel
	}

	switch name {
	case "gemini":
		return NewGeminiProvider(ctx, pCfg, modelName)
	// case "openai": ...
	// case "anthropic": ...
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
