package cmd

import (
	"fmt"
	"os"

	"github.com/ameb8/shai/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the entry point for the shai CLI application.
var rootCmd = &cobra.Command{
	Use:   "shai [query]",
	Short: "Natural language shell assistant",
	Long: `shai is a terminal-native AI agent that converts natural language into a single shell command, 
staged in the user’s prompt buffer for manual execution.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runQuery(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main() to initiate the command execution lifecycle.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Register the configuration initializer to run before command execution.
	cobra.OnInitialize(initConfig)

	// Define persistent flags that are available globally across all subcommands.
	rootCmd.PersistentFlags().StringP("provider", "p", "", "LLM provider override")
	rootCmd.PersistentFlags().StringP("model", "m", "", "Model or alias override")
	rootCmd.PersistentFlags().Bool("think", false, "Use smart model")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Print command only, do not inject")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Print token usage and latency")
	rootCmd.PersistentFlags().Bool("no-explain", false, "Suppress explanation lines")
	rootCmd.PersistentFlags().String("shell", "", "Override shell (bash|zsh|fish)")
	rootCmd.PersistentFlags().String("cmd-file", "", "Write the final command to this file instead of stdout")

	// Map CLI flags to Viper configuration keys for unified access.
	viper.BindPFlag("active.provider", rootCmd.PersistentFlags().Lookup("provider"))
}

// cfg maintains the application state and configuration throughout the execution.
var cfg *config.Config

// initConfig loads the application configuration from files or environment variables.
// It terminates the process if the configuration fails to load.
func initConfig() {
	var err error
	// Load the unified configuration from default paths and environment.
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}
