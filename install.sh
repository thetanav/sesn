#!/bin/bash

set -e

# sesn installer

echo "Installing sesn..."

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
    linux) OS="linux" ;;
    darwin) OS="darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get the latest release URL
REPO="thetanav/tmuxly"
LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"
DOWNLOAD_URL=$(curl -s $LATEST_URL | grep "browser_download_url.*sesn-$OS-$ARCH" | head -1 | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Could not find binary for $OS-$ARCH"
    exit 1
fi

# Download the binary
echo "Downloading sesn from $DOWNLOAD_URL"
curl -L -o sesn $DOWNLOAD_URL

# Make executable
chmod +x sesn

# Install to /usr/local/bin (requires sudo)
echo "Installing to /usr/local/bin (requires sudo)"
sudo mv sesn /usr/local/bin/

echo "sesn installed successfully! Run 'sesn' to start."