package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProfile(t *testing.T) {
	profile := Profile{
		URL:   "https://example.metabase.com",
		Token: "test-token",
	}

	if profile.URL != "https://example.metabase.com" {
		t.Errorf("Profile.URL = %s, want https://example.metabase.com", profile.URL)
	}
	if profile.Token != "test-token" {
		t.Errorf("Profile.Token = %s, want test-token", profile.Token)
	}
}

func TestGetConfigDir(t *testing.T) {
	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

		testDir := "/tmp/test-config"
		os.Setenv("XDG_CONFIG_HOME", testDir)

		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("GetConfigDir() error = %v", err)
		}

		expected := filepath.Join(testDir, "mbx")
		if configDir != expected {
			t.Errorf("GetConfigDir() = %s, want %s", configDir, expected)
		}
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

		os.Unsetenv("XDG_CONFIG_HOME")

		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("GetConfigDir() error = %v", err)
		}

		homeDir, _ := os.UserHomeDir()
		expected := filepath.Join(homeDir, ".config", "mbx")
		if configDir != expected {
			t.Errorf("GetConfigDir() = %s, want %s", configDir, expected)
		}
	})
}

func TestGetConfigPath(t *testing.T) {
	t.Run("with global config file", func(t *testing.T) {
		originalGlobal := globalConfigFile
		defer func() { globalConfigFile = originalGlobal }()

		testPath := "/tmp/test-config.yaml"
		SetGlobalConfigFile(testPath)

		configPath, err := GetConfigPath()
		if err != nil {
			t.Fatalf("GetConfigPath() error = %v", err)
		}

		if configPath != testPath {
			t.Errorf("GetConfigPath() = %s, want %s", configPath, testPath)
		}
	})

	t.Run("without global config file", func(t *testing.T) {
		originalGlobal := globalConfigFile
		defer func() { globalConfigFile = originalGlobal }()

		SetGlobalConfigFile("")

		configPath, err := GetConfigPath()
		if err != nil {
			t.Fatalf("GetConfigPath() error = %v", err)
		}

		configDir, _ := GetConfigDir()
		expected := filepath.Join(configDir, "config.yaml")
		if configPath != expected {
			t.Errorf("GetConfigPath() = %s, want %s", configPath, expected)
		}
	})
}

func TestLoadConfig_NonexistentFile(t *testing.T) {
	originalGlobal := globalConfigFile
	defer func() { globalConfigFile = originalGlobal }()

	SetGlobalConfigFile("/tmp/nonexistent-config.yaml")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() should not error on nonexistent file, got: %v", err)
	}

	if config.DefaultProfile != "" {
		t.Errorf("LoadConfig() DefaultProfile = %s, want empty string", config.DefaultProfile)
	}

	if config.Profiles == nil {
		t.Error("LoadConfig() Profiles should be initialized")
	}

	if len(config.Profiles) != 0 {
		t.Errorf("LoadConfig() Profiles length = %d, want 0", len(config.Profiles))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mbx-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalGlobal := globalConfigFile
	defer func() { globalConfigFile = originalGlobal }()

	configPath := filepath.Join(tempDir, "config.yaml")
	SetGlobalConfigFile(configPath)

	config := &Config{
		DefaultProfile: "work",
		Profiles: map[string]Profile{
			"work": {
				URL:   "https://work.metabase.com",
				Token: "work-token",
			},
			"dev": {
				URL:   "https://dev.metabase.com",
				Token: "dev-token",
			},
		},
	}

	err = SaveConfig(config)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loadedConfig.DefaultProfile != "work" {
		t.Errorf("LoadConfig() DefaultProfile = %s, want work", loadedConfig.DefaultProfile)
	}

	if len(loadedConfig.Profiles) != 2 {
		t.Errorf("LoadConfig() Profiles length = %d, want 2", len(loadedConfig.Profiles))
	}

	workProfile, exists := loadedConfig.Profiles["work"]
	if !exists {
		t.Error("LoadConfig() work profile not found")
	} else {
		if workProfile.URL != "https://work.metabase.com" {
			t.Errorf("work profile URL = %s, want https://work.metabase.com", workProfile.URL)
		}
		if workProfile.Token != "work-token" {
			t.Errorf("work profile Token = %s, want work-token", workProfile.Token)
		}
	}
}

func TestResolveConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		flagURL     string
		flagToken   string
		flagProfile string
		config      *Config
		wantURL     string
		wantToken   string
		wantError   bool
	}{
		{
			name:      "flags only",
			flagURL:   "https://flag.metabase.com",
			flagToken: "flag-token",
			wantURL:   "https://flag.metabase.com",
			wantToken: "flag-token",
			wantError: false,
		},
		{
			name:        "profile from config",
			flagProfile: "test",
			config: &Config{
				DefaultProfile: "test",
				Profiles: map[string]Profile{
					"test": {
						URL:   "https://test.metabase.com",
						Token: "test-token",
					},
				},
			},
			wantURL:   "https://test.metabase.com",
			wantToken: "test-token",
			wantError: false,
		},
		{
			name:        "flag URL overrides config",
			flagURL:     "https://override.metabase.com",
			flagProfile: "test",
			config: &Config{
				Profiles: map[string]Profile{
					"test": {
						URL:   "https://test.metabase.com",
						Token: "test-token",
					},
				},
			},
			wantURL:   "https://override.metabase.com",
			wantToken: "test-token",
			wantError: false,
		},
		{
			name:      "missing configuration",
			config:    &Config{Profiles: make(map[string]Profile)},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config != nil {
				tempDir, err := os.MkdirTemp("", "mbx-resolve-test")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				defer os.RemoveAll(tempDir)

				originalGlobal := globalConfigFile
				defer func() { globalConfigFile = originalGlobal }()

				configPath := filepath.Join(tempDir, "config.yaml")
				SetGlobalConfigFile(configPath)

				err = SaveConfig(tt.config)
				if err != nil {
					t.Fatalf("Failed to save test config: %v", err)
				}
			}

			gotURL, gotToken, err := ResolveConfiguration(tt.flagURL, tt.flagToken, tt.flagProfile)

			if tt.wantError {
				if err == nil {
					t.Errorf("ResolveConfiguration() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveConfiguration() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if gotURL != tt.wantURL {
				t.Errorf("ResolveConfiguration() URL = %s, want %s", gotURL, tt.wantURL)
			}

			if gotToken != tt.wantToken {
				t.Errorf("ResolveConfiguration() Token = %s, want %s", gotToken, tt.wantToken)
			}
		})
	}
}
