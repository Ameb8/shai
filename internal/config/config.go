package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the shai configuration schema.
type Config struct {
	Active    ActiveConfig              `mapstructure:"active" toml:"active"`
	Models    map[string]string         `mapstructure:"models" toml:"models"` // alias → model string
	Providers map[string]ProviderConfig `mapstructure:"providers" toml:"providers"`
}

type ActiveConfig struct {
	Provider string `mapstructure:"provider" toml:"provider"`
}

type ProviderConfig struct {
	APIKey       string `mapstructure:"api_key" toml:"api_key"`
	DefaultModel string `mapstructure:"default_model" toml:"default_model"`
}

// GetConfigPath returns the default path to the config file.
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "shai", "config.toml"), nil
}

// Load reads the configuration from disk.
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("toml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// If file doesn't exist, we'll return an empty config instead of erroring
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Initialize maps if they are nil
	if cfg.Models == nil {
		cfg.Models = make(map[string]string)
	}
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

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
		// If WriteConfigAs fails because the file doesn't exist, we try SafeWriteConfigAs
		// but actually WriteConfigAs should work if we provide the path.
		// Let's just use WriteConfigAs.
		return err
	}

	// Enforce 600 permissions
	return os.Chmod(path, 0600)
}
