#!/bin/bash
set -e

# Metabase Explorer installer script

REPO="amureki/metabase-explorer"
BINARY_NAME="mbx"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Failed to get latest release"
    exit 1
fi

echo "Installing $BINARY_NAME $LATEST_RELEASE for $OS-$ARCH..."

# Download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/${BINARY_NAME}_${LATEST_RELEASE#v}_${OS}_${ARCH}.tar.gz"

if [ "$OS" = "windows" ]; then
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/${BINARY_NAME}_${LATEST_RELEASE#v}_${OS}_${ARCH}.zip"
fi

# Create temp directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Download and extract
echo "Downloading from $DOWNLOAD_URL..."
if command -v curl &> /dev/null; then
    curl -sL "$DOWNLOAD_URL" -o archive
elif command -v wget &> /dev/null; then
    wget -q "$DOWNLOAD_URL" -O archive
else
    echo "Error: curl or wget is required"
    exit 1
fi

# Extract based on file type
if [ "$OS" = "windows" ]; then
    unzip -q archive
else
    tar -xzf archive
fi

# Install binary
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

if [ -f "$BINARY_NAME" ]; then
    chmod +x "$BINARY_NAME"
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    echo "$BINARY_NAME installed to $INSTALL_DIR"
else
    echo "Error: Binary not found in archive"
    exit 1
fi

# Check if configuration already exists
if [ -f "$HOME/.config/mbx/config.yaml" ]; then
    echo "Existing configuration found at ~/.config/mbx/config.yaml"
fi

# Cleanup
cd /
rm -rf "$TEMP_DIR"

echo ""
echo "Installation complete! 🎉"
echo ""
echo "Make sure $INSTALL_DIR is in your PATH:"
echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
echo ""
echo "Next steps:"
echo "1. Run: $BINARY_NAME init"
echo "   This will guide you through setting up your Metabase connection"
echo "2. Then run: $BINARY_NAME"
echo ""
echo "Configuration:"
echo "• Config file: ~/.config/mbx/config.yaml"
echo "• Multiple profiles: $BINARY_NAME config --help"
echo "• API token setup: https://www.metabase.com/docs/latest/people-and-groups/api-keys"
