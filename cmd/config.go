package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// configCmd represents the base command for shai configuration management.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage shai configuration",
}

// setKeyCmd defines the command to securely store an API key for a specific provider.
var setKeyCmd = &cobra.Command{
	Use:   "set-key",
	Short: "Set API key for a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Resolve provider name from flags or interactive prompt.
		providerName, _ := cmd.Flags().GetString("provider")
		if providerName == "" {
			fmt.Print("Enter provider (gemini/grok/mistral): ")
			fmt.Scanln(&providerName)
		}
		providerName = strings.ToLower(strings.TrimSpace(providerName))

		// Read API key securely without echoing to the terminal.
		fmt.Printf("Enter API key for %s: ", providerName)
		byteKey, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println() // New line after password entry
		key := strings.TrimSpace(string(byteKey))

		// Update the configuration in-memory.
		pCfg := cfg.Providers[providerName]
		pCfg.APIKey = key
		cfg.Providers[providerName] = pCfg

		// Persist changes to the configuration file.
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("API key for %s set successfully.\n", providerName)
		return nil
	},
}

// getCmd provides a human-readable overview of the current application configuration.
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Print current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Active Provider: %s\n", cfg.Active.Provider)

		// Display model aliases if any are configured.
		fmt.Println("\nModels (Aliases):")
		if len(cfg.Models) == 0 {
			fmt.Println("  None configured.")
		}
		for alias, model := range cfg.Models {
			fmt.Printf("  %s: %s\n", alias, model)
		}

		// List details for each configured provider, including redacted API keys.
		fmt.Println("\nProviders:")
		if len(cfg.Providers) == 0 {
			fmt.Println("  None configured.")
		}
		for name, pCfg := range cfg.Providers {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    Default Model: %s\n", pCfg.DefaultModel)

			// Redact API keys to show status without leaking secrets.
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

// setProviderCmd updates the global active provider used for inference.
var setProviderCmd = &cobra.Command{
	Use:   "set-provider [provider]",
	Short: "Set the active provider",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine provider name from arguments or interactive input.
		var providerName string
		if len(args) > 0 {
			providerName = args[0]
		} else {
			fmt.Print("Enter active provider: ")
			fmt.Scanln(&providerName)
		}
		providerName = strings.ToLower(strings.TrimSpace(providerName))

		// Update and save the active provider state.
		cfg.Active.Provider = providerName
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Active provider set to %s.\n", providerName)
		return nil
	},
}

// setModelCmd assigns a default model to a provider or configures a global alias.
var setModelCmd = &cobra.Command{
	Use:   "set-model [model]",
	Short: "Set model for a provider or alias",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		providerName, _ := cmd.Flags().GetString("provider")
		alias, _ := cmd.Flags().GetString("alias")

		// Resolve model name from arguments or stdin.
		var modelName string
		if len(args) > 0 {
			modelName = args[0]
		} else {
			fmt.Print("Enter model name: ")
			reader := bufio.NewReader(os.Stdin)
			modelName, _ = reader.ReadString('\n')
			modelName = strings.TrimSpace(modelName)
		}

		// Apply the model change based on the provided flags.
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

		// Persist the updated model configuration.
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(setKeyCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setProviderCmd)
	configCmd.AddCommand(setModelCmd)

	setKeyCmd.Flags().StringP("provider", "p", "", "Provider to set key for")
	setModelCmd.Flags().StringP("provider", "p", "", "Provider to set model for")
	setModelCmd.Flags().StringP("alias", "a", "", "Alias to set (e.g., smart, fast)")
}
