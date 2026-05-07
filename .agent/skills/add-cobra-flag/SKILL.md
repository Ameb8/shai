---
name: add-cobra-flag
description: Guides the addition of new CLI flags and positional arguments using the Cobra library in Go. Use when the user wants to extend the CLI interface with new options, parameters, or argument validation rules.
---

# Add Cobra Flag

This skill provides a structured workflow for adding new flags and arguments to a Cobra-based CLI application.

## Workflow

1.  **Identify the Target Command**: Determine which command the flag or argument should belong to (e.g., `rootCmd` for global flags, or a specific subcommand).
2.  **Define the Flag/Argument**:
    *   For **Flags**: Decide on the type (String, Bool, Int, etc.), name, shorthand, default value, and description.
    *   For **Arguments**: Decide on the validation logic (e.g., `ExactArgs(1)`, `MinimumNArgs(2)`).
3.  **Register in `init()`**:
    *   Use `Flags().TypeP(...)` for local flags.
    *   Use `PersistentFlags().TypeP(...)` for global flags.
    *   Update the `Args` field in the command definition if adding positional arguments.
4.  **Implement Usage**: Retrieve and use the flag/argument value within the command's `Run` or `RunE` function.
5.  **Optional: Bind to Viper**: If the project uses Viper, bind the flag to a configuration key using `viper.BindPFlag`.

## Reference Patterns

See [cobra-patterns.md](references/cobra-patterns.md) for concrete code examples of:
- Persistent and Local flags
- Required flags
- Positional argument validation
- Retrieving values in `Run`
- Viper binding

## Examples

### Adding a boolean flag to root
**User Request**: "Add a --debug flag to the root command"
**Action**:
1. Open `cmd/root.go`.
2. In `init()`, add `rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")`.
3. In `RunE` or wherever needed, use `debug, _ := cmd.Flags().GetBool("debug")`.

### Adding a required string argument to a subcommand
**User Request**: "The 'shai config set-provider' command should require exactly one argument: the provider name."
**Action**:
1. Open `cmd/config.go`.
2. Locate `setProviderCmd`.
3. Set `Args: cobra.ExactArgs(1)`.
4. Update `RunE` to use `args[0]`.
