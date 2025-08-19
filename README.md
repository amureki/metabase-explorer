# Metabase Explorer

A Terminal User Interface (TUI) for exploring Metabase collections and database metadata. Browse collections, cards, dashboards, databases, schemas, tables, and fields through an interactive interface that connects to Metabase via REST API.

<img src="https://github.com/user-attachments/assets/1e29f0f9-7ab3-48bf-a22a-436bc50fd285" alt="Metabase Explorer Demo" width="500" />

## Installation

### Quick Install (Recommended)
```bash
curl -sSL https://raw.githubusercontent.com/amureki/metabase-explorer/main/install.sh | bash
```

### Download Pre-built Binaries
Download the latest release for your platform from [GitHub Releases](https://github.com/amureki/metabase-explorer/releases/latest).

### Build from Source
If you have Go 1.23.0+ installed:
```bash
git clone https://github.com/amureki/metabase-explorer.git
cd metabase-explorer
go build -o mbx .
```

## Configuration

### Interactive Setup (Recommended)
```bash
mbx init
```

This will guide you through setting up your Metabase connection with an interactive wizard.

### Manual Configuration
```bash
# Set up your default profile
mbx config set url "https://your-metabase-instance.com/"
mbx config set token "your-api-token-here"

# Or create named profiles for different environments
mbx config set --profile work url "https://work.metabase.com/"
mbx config set --profile work token "work-token"
mbx config switch work
```

### Multiple Profiles
```bash
mbx config list                    # Show all profiles
mbx config get work                 # Show specific profile
mbx config switch staging          # Change default profile
mbx --profile work                  # Use specific profile once
```

### Getting an API Token
See the [Metabase API Keys documentation](https://www.metabase.com/docs/latest/people-and-groups/api-keys) for instructions on creating an API token.

## Usage

```bash
# Use default profile
mbx

# Use specific profile
mbx --profile work

# Override with flags
mbx --url https://demo.metabase.com --token your-token
```

The application provides keyboard shortcuts and help information directly in the interface.

## Configuration Files

Configuration is stored in `~/.config/mbx/config.yaml` by default, or you can specify a custom location:

```bash
mbx --config /path/to/custom/config.yaml config list
```

## Updating

To update to the latest version:

```bash
curl -sSL https://raw.githubusercontent.com/amureki/metabase-explorer/main/install.sh | bash
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Issues & Support

[Report bugs or request features](https://github.com/amureki/metabase-explorer/issues)

[Sponsor Development](https://github.com/sponsors/amureki)
