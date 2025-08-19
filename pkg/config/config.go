package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type Config struct {
	DefaultProfile string             `yaml:"default_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
}

var globalConfigFile string

func SetGlobalConfigFile(path string) {
	globalConfigFile = path
}

func GetConfigDir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "mbx"), nil
}

func GetConfigPath() (string, error) {
	// 1. CLI flag has highest priority
	if globalConfigFile != "" {
		return globalConfigFile, nil
	}

	// 2. Default location (XDG compliant)
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			DefaultProfile: "",
			Profiles:       make(map[string]Profile),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create directory for config file if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func ResolveConfiguration(flagURL, flagToken, flagProfile string) (string, string, error) {
	var metabaseURL, apiToken string

	// 1. Start with CLI flags (highest priority)
	if flagURL != "" {
		metabaseURL = flagURL
	}
	if flagToken != "" {
		apiToken = flagToken
	}

	// 2. Try config file
	if metabaseURL == "" || apiToken == "" {
		config, err := LoadConfig()
		if err != nil {
			return "", "", fmt.Errorf("failed to load config: %v", err)
		}

		profileName := flagProfile
		if profileName == "" {
			profileName = config.DefaultProfile
		}

		if profileName != "" {
			if profile, exists := config.Profiles[profileName]; exists {
				if metabaseURL == "" && profile.URL != "" {
					metabaseURL = profile.URL
				}
				if apiToken == "" && profile.Token != "" {
					apiToken = profile.Token
				}
			}
		}
	}

	// 3. Check if we have everything we need
	if metabaseURL == "" || apiToken == "" {
		return "", "", fmt.Errorf("missing configuration: URL=%s, Token=%s",
			map[bool]string{true: "✓", false: "✗"}[metabaseURL != ""],
			map[bool]string{true: "✓", false: "✗"}[apiToken != ""])
	}

	return metabaseURL, apiToken, nil
}
