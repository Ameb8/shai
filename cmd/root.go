package cmd

import (
	"fmt"
	"os"

	"shai/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags available to all subcommands
	rootCmd.PersistentFlags().StringP("provider", "p", "", "LLM provider override")
	rootCmd.PersistentFlags().StringP("model", "m", "", "Model or alias override")
	rootCmd.PersistentFlags().Bool("think", false, "Use smart model")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Print command only, do not inject")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Print token usage and latency")
	rootCmd.PersistentFlags().Bool("no-explain", false, "Suppress explanation lines")
	rootCmd.PersistentFlags().String("shell", "", "Override shell (bash|zsh|fish)")

	// Bind flags to viper
	viper.BindPFlag("active.provider", rootCmd.PersistentFlags().Lookup("provider"))
}

var cfg *config.Config

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}
