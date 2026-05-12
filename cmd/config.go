package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newConfigCmd creates and returns the "config" command for managing shai settings.
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage shai configuration",
	}

	setKeyCmd := &cobra.Command{
		Use:   "set-key",
		Short: "Set API key for a provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the target provider from flags or interactive input.
			providerName, _ := cmd.Flags().GetString("provider")
			if providerName == "" {
				fmt.Print("Enter provider (gemini/grok/mistral): ")
				fmt.Scanln(&providerName)
			}
			providerName = strings.ToLower(strings.TrimSpace(providerName))

			// Prompt for and read the API key securely.
			fmt.Printf("Enter API key for %s: ", providerName)
			byteKey, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return err
			}
			fmt.Println() // Ensure terminal cursor moves to a new line.
			key := strings.TrimSpace(string(byteKey))

			// Update the in-memory configuration with the new key.
			pCfg := cfg.Providers[providerName]
			pCfg.APIKey = key
			cfg.Providers[providerName] = pCfg

			// Save the updated configuration to disk.
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("API key for %s set successfully.\n", providerName)
			return nil
		},
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Print current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Active Provider: %s\n", cfg.Active.Provider)

			// Display configured model aliases for quick reference.
			fmt.Println("\nModels (Aliases):")
			if len(cfg.Models) == 0 {
				fmt.Println("  None configured.")
			}
			for alias, model := range cfg.Models {
				fmt.Printf("  %s: %s\n", alias, model)
			}

			// Show detailed settings for each provider while protecting sensitive keys.
			fmt.Println("\nProviders:")
			if len(cfg.Providers) == 0 {
				fmt.Println("  None configured.")
			}
			for name, pCfg := range cfg.Providers {
				fmt.Printf("  %s:\n", name)
				fmt.Printf("    Default Model: %s\n", pCfg.DefaultModel)

				// Mask API keys to confirm existence without exposing them in cleartext.
				keyStatus := "Not set"
				if pCfg.APIKey != "" {
					keyStatus = "Set (redacted)"
					if len(pCfg.APIKey) > 8 {
						keyStatus = fmt.Sprintf("%s...%s (redacted)", pCfg.APIKey[:4], pCfg.APIKey[len(pCfg.APIKey)-4:])
					}
				}
				fmt.Printf("    API Key: %s\n", keyStatus)
			}
		},
	}

	setProviderCmd := &cobra.Command{
		Use:   "set-provider [provider]",
		Short: "Set the active provider",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the provider name from positional arguments or interactive input.
			var providerName string
			if len(args) > 0 {
				providerName = args[0]
			} else {
				fmt.Print("Enter active provider: ")
				fmt.Scanln(&providerName)
			}
			providerName = strings.ToLower(strings.TrimSpace(providerName))

			// Update the active provider and persist the change.
			cfg.Active.Provider = providerName
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Active provider set to %s.\n", providerName)
			return nil
		},
	}

	setModelCmd := &cobra.Command{
		Use:   "set-model [model]",
		Short: "Set model for a provider or alias",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerName, _ := cmd.Flags().GetString("provider")
			alias, _ := cmd.Flags().GetString("alias")

			// Resolve the model name from arguments or standard input.
			var modelName string
			if len(args) > 0 {
				modelName = args[0]
			} else {
				fmt.Print("Enter model name: ")
				reader := bufio.NewReader(os.Stdin)
				modelName, _ = reader.ReadString('\n')
				modelName = strings.TrimSpace(modelName)
			}

			// Assign the model to either an alias or a provider-specific default.
			if alias != "" {
				cfg.Models[alias] = modelName
				fmt.Printf("Alias '%s' set to model '%s'.\n", alias, modelName)
			} else if providerName != "" {
				pCfg := cfg.Providers[providerName]
				pCfg.DefaultModel = modelName
				cfg.Providers[providerName] = pCfg
				fmt.Printf("Default model for provider '%s' set to '%s'.\n", providerName, modelName)
			} else {
				return fmt.Errorf("must specify either --provider or --alias")
			}

			// Persist the model configuration update.
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			return nil
		},
	}

	cmd.AddCommand(setKeyCmd)
	cmd.AddCommand(getCmd)
	cmd.AddCommand(setProviderCmd)
	cmd.AddCommand(setModelCmd)

	setKeyCmd.Flags().StringP("provider", "p", "", "Provider to set key for")
	setModelCmd.Flags().StringP("provider", "p", "", "Provider to set model for")
	setModelCmd.Flags().StringP("alias", "a", "", "Alias to set (e.g., smart, fast)")

	return cmd
}

// addConfigurationCommands registers configuration-related subcommands to the provided root command.
func addConfigurationCommands(root *cobra.Command) {
	root.AddCommand(newConfigCmd())
}

func init() {
	addConfigurationCommands(rootCmd)
}
