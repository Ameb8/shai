# Adding Base URL Support to `shai` Provider Infrastructure

This document covers the plumbing changes needed to thread a configurable `BaseURL`
through the config, registry, and provider layer — without adding any specific
provider implementation. Once these changes are in place, adding an OpenAI-compatible
provider is a matter of writing one new file.

---

## 1. Update `config.ProviderConfig`

Add a `BaseURL` field to the shared provider config struct. It is optional — providers
that don't need it simply ignore it.

```go
// internal/config/config.go (or wherever ProviderConfig lives)

type ProviderConfig struct {
    APIKey       string `yaml:"api_key"`
    DefaultModel string `yaml:"default_model"`
    BaseURL      string `yaml:"base_url,omitempty"` // optional; overrides the provider's default endpoint
}
```

**Why `omitempty`:** Keeps existing config files valid — users who don't set `base_url`
get zero-value (`""`), which providers interpret as "use the built-in default."

---

## 2. No Changes Needed to `provider.Provider`

The `BaseURL` is a construction-time concern, not a per-call one. The `Provider`
interface stays exactly as-is:

```go
type Provider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
    ValidateKey(ctx context.Context) error
    Name() string
}
```

Providers receive `cfg config.ProviderConfig` in their constructor and read
`cfg.BaseURL` there. No interface changes required.

---

## 3. `registry.go` — No Structural Changes Required

`NewProvider` already passes the full `pCfg` to each constructor, so `BaseURL` flows
through automatically once it exists on `ProviderConfig`. The only future change is
adding a new `case` for each new provider:

```go
switch name {
case "gemini":
    return NewGeminiProvider(ctx, pCfg, modelName)
case "mistral":
    return NewMistralProvider(ctx, pCfg, modelName)
case "grok", "xai":
    return NewGrokProvider(ctx, pCfg, modelName)
// Add new providers here — pCfg.BaseURL is already available to them.
default:
    return nil, fmt.Errorf("unknown provider: %s", name)
}
```

---

## 4. Convention for New Providers That Use `BaseURL`

Any provider that supports a configurable endpoint should follow this pattern in
its constructor:

```go
const myProviderDefaultURL = "https://api.example.com/v1/chat/completions"

func NewMyProvider(_ context.Context, cfg config.ProviderConfig, modelName string) (*MyProvider, error) {
    // ...

    url := cfg.BaseURL
    if url == "" {
        url = myProviderDefaultURL
    }

    return &MyProvider{
        apiKey: cfg.APIKey,
        model:  modelName,
        client: &http.Client{Timeout: 120 * time.Second},
        url:    url,
    }, nil
}
```

**Rules:**
- Store the resolved URL on the struct (not `cfg`) so it is immutable after construction.
- Fall back to the provider's own default constant when `cfg.BaseURL` is empty.
- Never read `cfg` after the constructor returns.

---

## 5. User-Facing Config Examples

### Default endpoint (no `base_url` needed)

```yaml
providers:
  myprovider:
    api_key: sk-...
    default_model: some-model
```

### Custom / compatible endpoint (LiteLLM, local inference, proxy, etc.)

```yaml
providers:
  myprovider:
    api_key: placeholder   # some servers still require a non-empty value
    base_url: http://localhost:4000/v1/chat/completions
    default_model: some-model
```

---

## 6. Checklist

- [ ] Add `BaseURL string` to `config.ProviderConfig` with `yaml:"base_url,omitempty"`
- [ ] Confirm existing provider constructors compile cleanly (they ignore `BaseURL`)
- [ ] Verify existing YAML configs still parse correctly (field is omitempty)
- [ ] Document `base_url` in your config schema / README for users
- [ ] New providers: read `cfg.BaseURL` in constructor, fall back to built-in default
- [ ] New providers: add a `case` in `registry.go`