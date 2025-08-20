package cli

import (
	"fmt"
	"os"

	"github.com/amureki/metabase-explorer/pkg/config"
	"github.com/amureki/metabase-explorer/pkg/tui"
	"github.com/amureki/metabase-explorer/pkg/util"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func printHelp() {
	fmt.Printf(`mbx - Metabase Explorer %s

A Terminal User Interface for exploring Metabase database metadata.

USAGE:
    mbx [OPTIONS]
    mbx init
    mbx config <command> [arguments]
    mbx update

OPTIONS:
    -h, --help                Show this help message
    -v, --version             Show version information
    -u, --url <url>           Metabase URL (overrides config)
    -t, --token <token>       API token (overrides config)
    -p, --profile <name>      Configuration profile to use
    -c, --config <path>       Custom config file location

COMMANDS:
    init                               Interactive setup wizard
    config <subcommand>                Configuration management
    update                             Update to the latest version

CONFIGURATION:
    mbx init                           # Interactive setup wizard
    mbx config set url "https://your-metabase-instance.com/"
    mbx config set token "your-api-token-here"
    mbx config list                    # Show all profiles
    mbx config switch <profile>        # Change default profile

    Default config location: ~/.config/mbx/config.yaml
    Custom location: --config <path>
    See https://www.metabase.com/docs/latest/people-and-groups/api-keys for API token setup

For more information, visit: https://github.com/amureki/metabase-explorer
`, version)
}

func Execute(args []string, ver string) {
	version = ver
	var showVersion, showHelp bool
	var metabaseURL, apiToken, profile, configFile string

	// Basic flag parsing
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-v", "--version":
			showVersion = true
		case "-h", "--help":
			showHelp = true
		case "-u", "--url":
			if i+1 < len(args) {
				metabaseURL = args[i+1]
				i++
			}
		case "-t", "--token":
			if i+1 < len(args) {
				apiToken = args[i+1]
				i++
			}
		case "-p", "--profile":
			if i+1 < len(args) {
				profile = args[i+1]
				i++
			}
		case "-c", "--config":
			if i+1 < len(args) {
				configFile = args[i+1]
				i++
			}
		}
	}

	if configFile != "" {
		config.SetGlobalConfigFile(configFile)
	}

	if len(args) > 0 {
		switch args[0] {
		case "init":
			handleConfigInit()
			return
		case "config":
			handleConfigCommand(args[1:])
			return
		case "update":
			util.HandleUpdateCommand(version)
			return
		}
	}

	if showVersion {
		fmt.Printf("mbx version %s\n", version)
		return
	}

	if showHelp {
		printHelp()
		return
	}

	p := tea.NewProgram(tui.InitialModel(metabaseURL, apiToken, profile, version), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
