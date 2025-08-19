package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/amureki/metabase-explorer/pkg/config"
)

func handleConfigCommand(args []string) {
	if len(args) == 0 {
		fmt.Print(`mbx config - Configuration management

USAGE:
    mbx config <command> [arguments]

COMMANDS:
    list                    Show all profiles
    get [profile]           Show profile details (default profile if none specified)  
    set <key> <value>       Set configuration value in default profile
    set --profile <name> <key> <value>  Set configuration value in specific profile
    delete <profile>        Delete a profile
    switch <profile>        Set default profile

EXAMPLES:
    mbx config list
    mbx config set url "https://metabase.company.com/"
    mbx config set --profile work token "abc123"
    mbx config get work
    mbx config switch work
`)
		return
	}

	cmd := args[0]
	switch cmd {
	case "list":
		handleConfigList()
	case "get":
		profile := ""
		if len(args) > 1 {
			profile = args[1]
		}
		handleConfigShow(profile)
	case "set":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: 'set' requires key and value\nUsage: mbx config set <key> <value>\n")
			os.Exit(1)
		}
		profileName := ""
		key, value := args[1], args[2]

		// Check for --profile flag
		if len(args) >= 5 && args[1] == "--profile" {
			profileName = args[2]
			key, value = args[3], args[4]
		}
		handleConfigSet(profileName, key, value)
	case "delete":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Error: 'delete' requires profile name\nUsage: mbx config delete <profile>\n")
			os.Exit(1)
		}
		handleConfigDelete(args[1])
	case "switch":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Error: 'switch' requires profile name\nUsage: mbx config switch <profile>\n")
			os.Exit(1)
		}
		handleConfigSwitch(args[1])
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown config command '%s'\n", cmd)
		os.Exit(1)
	}
}

func handleConfigInit() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Metabase Explorer Configuration Setup")
	fmt.Println("====================================")

	// Show existing configuration if any
	if len(cfg.Profiles) > 0 {
		fmt.Println("\nExisting configuration:")
		for name := range cfg.Profiles {
			marker := "  "
			if name == cfg.DefaultProfile {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, name)
		}
		fmt.Println()
	}

	var url, token, profileName string

	fmt.Print("Profile name [default]: ")
	fmt.Scanln(&profileName)
	if profileName == "" {
		profileName = "default"
	}

	// Check if profile already exists
	if existingProfile, exists := cfg.Profiles[profileName]; exists {
		fmt.Printf("\nProfile '%s' already exists:\n", profileName)
		fmt.Printf("  URL: %s\n", existingProfile.URL)
		if len(existingProfile.Token) > 8 {
			fmt.Printf("  Token: %s...%s\n", existingProfile.Token[:4], existingProfile.Token[len(existingProfile.Token)-4:])
		} else {
			fmt.Printf("  Token: %s\n", existingProfile.Token)
		}

		var overwrite string
		fmt.Print("\nOverwrite existing profile? [y/N]: ")
		fmt.Scanln(&overwrite)
		if strings.ToLower(overwrite) != "y" && strings.ToLower(overwrite) != "yes" {
			fmt.Println("Configuration unchanged.")
			return
		}

		// Pre-fill with existing values
		fmt.Printf("\nMetabase URL [%s]: ", existingProfile.URL)
		fmt.Scanln(&url)
		if url == "" {
			url = existingProfile.URL
		}

		fmt.Printf("API Token [keep existing]: ")
		fmt.Scanln(&token)
		if token == "" {
			token = existingProfile.Token
		}
	} else {
		fmt.Print("\nMetabase URL: ")
		fmt.Scanln(&url)

		fmt.Print("API Token: ")
		fmt.Scanln(&token)
	}

	if url == "" || token == "" {
		fmt.Fprintf(os.Stderr, "Error: URL and token are required\n")
		os.Exit(1)
	}

	cfg.Profiles[profileName] = config.Profile{URL: url, Token: token}
	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = profileName
	}

	err = config.SaveConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Configuration saved for profile '%s'\n", profileName)
	if cfg.DefaultProfile == profileName {
		fmt.Println("✓ Set as default profile")
	}
}

func handleConfigList() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured. Run 'mbx init' to get started.")
		return
	}

	fmt.Println("Configured profiles:")
	for name := range cfg.Profiles {
		marker := "  "
		if name == cfg.DefaultProfile {
			marker = "* "
		}
		fmt.Printf("%s%s\n", marker, name)
	}
}

func handleConfigShow(profileName string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if profileName == "" {
		profileName = cfg.DefaultProfile
	}

	if profileName == "" {
		fmt.Println("No default profile set. Run 'mbx init' or specify a profile name.")
		return
	}

	profile, exists := cfg.Profiles[profileName]
	if !exists {
		fmt.Fprintf(os.Stderr, "Profile '%s' not found\n", profileName)
		os.Exit(1)
	}

	fmt.Printf("Profile: %s\n", profileName)
	if profileName == cfg.DefaultProfile {
		fmt.Println("(default)")
	}
	fmt.Printf("URL: %s\n", profile.URL)
	if len(profile.Token) > 8 {
		fmt.Printf("Token: %s...%s\n", profile.Token[:4], profile.Token[len(profile.Token)-4:])
	} else {
		fmt.Printf("Token: %s\n", profile.Token)
	}
}

func handleConfigSet(profileName, key, value string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if profileName == "" {
		if cfg.DefaultProfile == "" {
			profileName = "default"
		} else {
			profileName = cfg.DefaultProfile
		}
	}

	profile := cfg.Profiles[profileName]
	switch strings.ToLower(key) {
	case "url":
		profile.URL = value
	case "token":
		profile.Token = value
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown key '%s'. Valid keys: url, token\n", key)
		os.Exit(1)
	}

	cfg.Profiles[profileName] = profile
	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = profileName
	}

	err = config.SaveConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Set %s for profile '%s'\n", key, profileName)
}

func handleConfigDelete(profileName string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if _, exists := cfg.Profiles[profileName]; !exists {
		fmt.Fprintf(os.Stderr, "Profile '%s' not found\n", profileName)
		os.Exit(1)
	}

	delete(cfg.Profiles, profileName)

	if cfg.DefaultProfile == profileName {
		cfg.DefaultProfile = ""
		if len(cfg.Profiles) > 0 {
			for name := range cfg.Profiles {
				cfg.DefaultProfile = name
				break
			}
		}
	}

	err = config.SaveConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Deleted profile '%s'\n", profileName)
	if cfg.DefaultProfile != "" {
		fmt.Printf("✓ Default profile is now '%s'\n", cfg.DefaultProfile)
	}
}

func handleConfigSwitch(profileName string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if _, exists := cfg.Profiles[profileName]; !exists {
		fmt.Fprintf(os.Stderr, "Profile '%s' not found\n", profileName)
		os.Exit(1)
	}

	cfg.DefaultProfile = profileName

	err = config.SaveConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Switched to profile '%s'\n", profileName)
}
