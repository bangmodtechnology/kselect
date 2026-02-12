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

# Download archive
ARCHIVE="kselect-${OS}-${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
echo "Downloading kselect $VERSION for $OS/$ARCH..."

TMP_DIR=$(mktemp -d)
TMP_ARCHIVE="$TMP_DIR/$ARCHIVE"
HTTP_CODE=$(curl -sSL -w "%{http_code}" -o "$TMP_ARCHIVE" "$URL")
if [ "$HTTP_CODE" != "200" ]; then
    rm -rf "$TMP_DIR"
    echo "Error: Download failed (HTTP $HTTP_CODE)"
    echo "URL: $URL"
    exit 1
fi

# Extract binary
echo "Extracting..."
tar -xzf "$TMP_ARCHIVE" -C "$TMP_DIR"
chmod +x "$TMP_DIR/kselect"

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/kselect" "$INSTALL_DIR/kselect"
else
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$TMP_DIR/kselect" "$INSTALL_DIR/kselect"
fi

# Cleanup
rm -rf "$TMP_DIR"

echo ""
echo "kselect $VERSION installed successfully!"
kselect --version
