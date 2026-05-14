package cmd

import (
	"fmt"
	"os"

	"github.com/ameb8/shai/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var versionInfo = formatVersion("dev", "none", "unknown")

// rootCmd represents the entry point for the shai CLI application.
var rootCmd = &cobra.Command{
	Use:     "shai [query]",
	Short:   "Natural language shell assistant",
	Version: versionInfo,
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

// SetVersion configures build metadata shown by the root command's --version flag.
func SetVersion(version, commit, date string) {
	versionInfo = formatVersion(version, commit, date)
	rootCmd.Version = versionInfo
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main() to initiate the command execution lifecycle.
func Execute() error {
	return rootCmd.Execute()
}

// NewRootCmd creates and configures a new instance of the root command.
// If cfgOverride is provided, it is used as the global configuration and
// automatic configuration initialization is skipped.
func NewRootCmd(cfgOverride *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shai [query]",
		Short:   "Natural language shell assistant",
		Version: versionInfo,
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

	if cfgOverride != nil {
		cfg = cfgOverride
	} else {
		cobra.OnInitialize(initConfig)
	}

	setupRootFlags(cmd)
	addConfigurationCommands(cmd)

	return cmd
}

func init() {
	// Register the configuration initializer to run before command execution.
	cobra.OnInitialize(initConfig)

	setupRootFlags(rootCmd)
}

// setupRootFlags defines the persistent CLI flags that are available globally across all subcommands.
func setupRootFlags(cmd *cobra.Command) {
	// Define persistent flags that are available globally across all subcommands.
	cmd.PersistentFlags().StringP("provider", "p", "", "LLM provider override")
	cmd.PersistentFlags().StringP("model", "m", "", "Model or alias override")
	cmd.PersistentFlags().Bool("think", false, "Use smart model")
	cmd.PersistentFlags().Bool("dry-run", false, "Print command only, do not inject")
	cmd.PersistentFlags().BoolP("copy", "c", false, "Copy the generated command to clipboard")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Print token usage and latency")
	cmd.PersistentFlags().Bool("no-explain", false, "Suppress explanation lines")
	cmd.PersistentFlags().String("shell", "", "Override shell (bash|zsh|fish)")
	cmd.PersistentFlags().String("cmd-file", "", "Write the final command to this file instead of stdout")

	// Map CLI flags to Viper configuration keys for unified access.
	viper.BindPFlag("active.provider", cmd.PersistentFlags().Lookup("provider"))
}

func formatVersion(version, commit, date string) string {
	return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}

// cfg maintains the application state and configuration throughout the execution.
var cfg *config.Config

// initConfig loads the application configuration from files or environment variables.
// It terminates the process if the configuration fails to load.
func initConfig() {
	if cfg != nil {
		return
	}
	var err error
	// Load the unified configuration from default paths and environment.
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}
