# Metabase Explorer

A Terminal User Interface (TUI) for exploring Metabase database metadata. Browse databases, schemas, tables, and fields through an interactive interface that connects to Metabase via REST API.

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

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` with your Metabase details:
```bash
METABASE_URL="https://your-metabase-instance.url/"
METABASE_API_TOKEN="your-api-token-here"
```

### Getting an API Token
1. Go to your Metabase Admin Settings
2. Navigate to "API Keys"
3. Create a new API key
4. Copy the token to your `.env` file

## Usage

Run the application:
```bash
./mbx
```

The application provides keyboard shortcuts and help information directly in the interface.

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
