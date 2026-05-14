package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config defines the application-wide configuration schema for shai.
type Config struct {
	Active    ActiveConfig              `mapstructure:"active" toml:"active"`
	Models    map[string]string         `mapstructure:"models" toml:"models"` // Models maps model aliases to their full provider identifiers.
	Providers map[string]ProviderConfig `mapstructure:"providers" toml:"providers"`
}

// ActiveConfig identifies the currently selected provider.
type ActiveConfig struct {
	Provider string `mapstructure:"provider" toml:"provider"`
}

// ProviderConfig stores credentials and default settings for a specific LLM provider.
type ProviderConfig struct {
	APIKey       string `mapstructure:"api_key" toml:"api_key"`
	DefaultModel string `mapstructure:"default_model" toml:"default_model"`
	BaseURL      string `mapstructure:"base_url,omitempty" toml:"base_url,omitempty"`
}

// GetConfigPath returns the absolute path to the shai configuration file.
// It defaults to ~/.config/shai/config.toml but respects standard user home discovery.
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "shai", "config.toml"), nil
}

// Load reads the shai configuration from the default disk location.
// If the configuration directory does not exist, it will be created.
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Ensure the configuration directory exists before attempting to read.
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("toml")

	// Read the configuration file, ignoring missing file errors to allow for defaults.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Initialize maps to prevent nil pointer exceptions during runtime updates.
	if cfg.Models == nil {
		cfg.Models = make(map[string]string)
	}
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}

	return &cfg, nil
}

// Save persists the current configuration state to disk.
// It ensures restrictive file permissions (0600) to protect sensitive API keys.
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create parent directories if they were removed since Load was called.
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	viper.Set("active", c.Active)
	viper.Set("models", c.Models)
	viper.Set("providers", c.Providers)

	if err := viper.WriteConfigAs(path); err != nil {
		return err
	}

	// Restrict access to the configuration file for security.
	return os.Chmod(path, 0600)
}
