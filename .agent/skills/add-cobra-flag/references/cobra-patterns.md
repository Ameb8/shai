# Cobra Flag and Argument Patterns

## Adding a Persistent Flag
Persistent flags are available to the command they are added to and all subcommands.

```go
func init() {
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
```

## Adding a Local Flag
Local flags are only available to the specific command.

```go
func init() {
    myCmd.Flags().StringP("name", "n", "", "name to display")
}
```

## Required Flags
You can mark a flag as required.

```go
func init() {
    myCmd.Flags().StringP("region", "r", "", "AWS region")
    myCmd.MarkFlagRequired("region")
}
```

## Positional Arguments
Cobra provides several built-in validators for positional arguments:

- `cobra.NoArgs`: Report an error if there are any positional args.
- `cobra.ArbitraryArgs`: Accept any number of positional args.
- `cobra.MinimumNArgs(int)`: Report an error if less than N positional args.
- `cobra.MaximumNArgs(int)`: Report an error if more than N positional args.
- `cobra.ExactArgs(int)`: Report an error if not exactly N positional args.

```go
var myCmd = &cobra.Command{
    Use:   "mycmd [arg]",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Arg:", args[0])
    },
}
```

## Retrieving Flag Values
In the `Run` or `RunE` function:

```go
Run: func(cmd *cobra.Command, args []string) {
    verbose, _ := cmd.Flags().GetBool("verbose")
    name, _ := cmd.Flags().GetString("name")
}
```

## Binding Flags to Viper
If using Viper for configuration:

```go
func init() {
    rootCmd.PersistentFlags().String("config", "", "config file")
    viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
}
```
