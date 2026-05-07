# LLM Provider Implementation Guide

To add a new LLM provider to `shai`, follow these steps:

## 1. Implement the Provider Interface

Create a new file in `internal/provider/` (e.g., `openai.go`). Your provider must implement the `Provider` interface defined in `internal/provider/provider.go`:

```go
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	ValidateKey(ctx context.Context) error
	Name() string
}
```

### Key Considerations:
- **Struct Definition**: Include API key, model name, and an HTTP client.
- **Constructor**: Create a `New<ProviderName>Provider` function.
- **Mapping**: Implement helper functions to map `shai`'s internal `Message` and `ToolDefinition` types to the provider's specific API format.
- **Error Handling**: Ensure HTTP errors and empty responses are handled gracefully.

## 2. Register the Provider

Update `internal/provider/registry.go` to include your new provider in the `NewProvider` factory function.

```go
func NewProvider(ctx context.Context, name string, cfg config.Config, modelOverride string) (Provider, error) {
    // ...
    switch name {
    case "gemini":
        return NewGeminiProvider(ctx, pCfg, modelName)
    case "<your-provider-name>":
        return New<YourProviderName>Provider(ctx, pCfg, modelName)
    // ...
    }
}
```

## 3. Add Unit Tests

Create a test file (e.g., `openai_test.go`) and implement tests using a mock HTTP server to verify the `Complete` and `ValidateKey` methods. Refer to `mistral_test.go` or `grok_test.go` for examples.
