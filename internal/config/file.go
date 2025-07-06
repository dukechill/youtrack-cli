package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config defines the structure of the YouTrack CLI configuration.
type Config struct {
	URL           string `yaml:"url"`
	Token         string `yaml:"token"`
	DefaultSprint string `yaml:"default_sprint,omitempty"`
	BoardName     string `yaml:"board_name,omitempty"`
}

// configFilePath returns the absolute path to the configuration file.
func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".youtrack-cli.yaml"), nil
}

// Load loads the configuration from the file.
func Load() (Config, error) {
	var cfg Config
	path, err := configFilePath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, fmt.Errorf("config file not found at %s, please run 'youtrack-cli configure' or 'youtrack-cli config set'", path)
		}
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return cfg, nil
}

// Save saves the configuration to the file.
func Save(cfg Config) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", path, err)
	}
	return nil
}

// SetValue updates a specific configuration key.
func SetValue(key, value string) error {
	cfg, err := Load()
	if err != nil {
		// If config doesn't exist, create a new one
		if os.IsNotExist(err) {
			cfg = Config{}
		} else {
			return err
		}
	}

	switch key {
	case "url":
		cfg.URL = value
	case "token":
		cfg.Token = value
	case "sprint":
		cfg.DefaultSprint = value
	case "board":
		cfg.BoardName = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return Save(cfg)
}

// PrintRaw prints the raw configuration (for view command).
func PrintRaw(cfg Config) {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		fmt.Printf("Error formatting config: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// PrintMasked prints the configuration with sensitive parts masked (for show command).
func PrintMasked(cfg Config) {
	fmt.Printf("YouTrack URL: %s\n", cfg.URL)
	// Mask the token for security
	if len(cfg.Token) > 8 {
		fmt.Printf("API Token: %s...%s\n", cfg.Token[:4], cfg.Token[len(cfg.Token)-4:])
	} else {
		fmt.Printf("API Token: %s\n", cfg.Token)
	}
	fmt.Printf("Default Board: %s\n", cfg.BoardName)
	fmt.Printf("Default Sprint: %s\n", cfg.DefaultSprint)
}