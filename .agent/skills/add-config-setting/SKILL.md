---
name: add-config-setting
description: Guide for adding new configuration settings using Viper and Cobra. Use when adding new user-configurable options, API settings, or application behaviors that need to be persisted in the config file.
---

# Add Config Setting

## Overview

This skill guides you through the process of adding new configuration settings to the `shai` application. It involves updating the core configuration struct and adding CLI commands to manage these settings.

## Workflow

To add a new configuration setting, follow these steps:

### 1. Update the Configuration Struct
Modify `internal/config/config.go` to add the new field to the `Config` struct (or one of its sub-structs). Ensure you use both `mapstructure` (for Viper) and `toml` (for persistence) tags.

### 2. Update Configuration Logic
If your new setting is a map or a complex type, ensure it is properly initialized in the `Load()` function in `internal/config/config.go`. Also, ensure `Save()` persists the new field by adding a corresponding `viper.Set()` call if necessary (though Viper often handles this if the struct is passed correctly).

### 3. Implement the Management Command
Add a new `cobra.Command` in `cmd/config.go`. This command should:
- Parse arguments or flags.
- Update the global `cfg` object.
- Call `cfg.Save()` to persist the changes.

### 4. Register and Verify
- Register the new command in `cmd/config.go:init()`.
- Update `getCmd` in `cmd/config.go` to display the new setting when the user runs `shai config get`.

## Patterns and Examples

For concrete code snippets and common patterns (strings, booleans, maps), refer to [references/patterns.md](references/patterns.md).

## Verification Checklist

1. [ ] Field added to `internal/config/config.go` with correct tags.
2. [ ] Map/complex type initialization added to `Load()`.
3. [ ] Command implemented and registered in `cmd/config.go`.
4. [ ] `cfg.Save()` called in the command.
5. [ ] `shai config get` updated to show the new setting.
