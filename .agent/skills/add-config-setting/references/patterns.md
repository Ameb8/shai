# Configuration Patterns for Shai

This document provides patterns for adding new configuration settings to the `shai` application.

## 1. Adding a Simple String or Boolean Setting

### internal/config/config.go
Add the field to the `Config` struct.

```go
type Config struct {
    // ... existing fields
    MaxTokens int    `mapstructure:"max_tokens" toml:"max_tokens"`
    Debug     bool   `mapstructure:"debug" toml:"debug"`
    DefaultShell string `mapstructure:"default_shell" toml:"default_shell"`
}
```

### cmd/config.go
Add a new command to set the value.

```go
var setMaxTokensCmd = &cobra.Command{
    Use:   "set-max-tokens [value]",
    Short: "Set the maximum tokens for LLM response",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        val, err := strconv.Atoi(args[0])
        if err != nil {
            return fmt.Errorf("invalid integer: %w", err)
        }
        cfg.MaxTokens = val
        if err := cfg.Save(); err != nil {
            return fmt.Errorf("failed to save config: %w", err)
        }
        fmt.Printf("Max tokens set to %d\n", val)
        return nil
    },
}

func init() {
    // ...
    configCmd.AddCommand(setMaxTokensCmd)
}
```

## 2. Adding a Setting to a Sub-struct (e.g., ActiveConfig)

### internal/config/config.go
```go
type ActiveConfig struct {
    Provider string `mapstructure:"provider" toml:"provider"`
    Theme    string `mapstructure:"theme" toml:"theme"`
}
```

### cmd/config.go
```go
var setThemeCmd = &cobra.Command{
    Use:   "set-theme [theme]",
    Short: "Set the UI theme",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg.Active.Theme = args[0]
        return cfg.Save()
    },
}
```

## 3. Adding a Map Setting (e.g., Aliases)

### internal/config/config.go
```go
type Config struct {
    // ...
    CustomHeaders map[string]string `mapstructure:"custom_headers" toml:"custom_headers"`
}

// In Load() ensure map is initialized
if cfg.CustomHeaders == nil {
    cfg.CustomHeaders = make(map[string]string)
}
```

### cmd/config.go
```go
var setHeaderCmd = &cobra.Command{
    Use:   "set-header [key] [value]",
    Short: "Set a custom header",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg.CustomHeaders[args[0]] = args[1]
        return cfg.Save()
    },
}
```

## Implementation Checklist

1. [ ] Update `internal/config/config.go` with new struct fields and tags.
2. [ ] Initialize any new maps in `internal/config/config.go:Load()`.
3. [ ] Update `internal/config/config.go:Save()` if needed (Viper usually handles this via `viper.Set`).
4. [ ] Create a new `cobra.Command` in `cmd/config.go`.
5. [ ] Register the command in `cmd/config.go:init()`.
6. [ ] Update `getCmd` in `cmd/config.go` to display the new setting.
