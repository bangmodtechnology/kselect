#!/bin/sh
# kselect installer
# Usage: curl -sSL https://raw.githubusercontent.com/bangmodtechnology/kselect/master/install.sh | sh

set -e

REPO="bangmodtechnology/kselect"
INSTALL_DIR="/usr/local/bin"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    linux|darwin) ;;
    *) echo "Error: Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Error: Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release tag
echo "Fetching latest version..."
VERSION=$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//')
if [ -z "$VERSION" ]; then
    echo "Error: Failed to get latest version"
    exit 1
fi

# Download binary
BINARY="kselect-${OS}-${ARCH}"
URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY"
echo "Downloading kselect $VERSION for $OS/$ARCH..."

TMP=$(mktemp)
HTTP_CODE=$(curl -sSL -w "%{http_code}" -o "$TMP" "$URL")
if [ "$HTTP_CODE" != "200" ]; then
    rm -f "$TMP"
    echo "Error: Download failed (HTTP $HTTP_CODE)"
    echo "URL: $URL"
    exit 1
fi

chmod +x "$TMP"

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP" "$INSTALL_DIR/kselect"
else
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$TMP" "$INSTALL_DIR/kselect"
fi

echo ""
echo "kselect $VERSION installed successfully!"
kselect --version
