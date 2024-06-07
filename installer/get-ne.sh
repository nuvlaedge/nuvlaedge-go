#!/bin/sh

BASE_URL="https://github.com/nuvlaedge/nuvlaedge-go/releases/download"

# Get the version from the command line arguments, or use the latest version
if [ -z "$1" ]; then
    VERSION=$(curl --silent "https://api.github.com/repos/nuvlaedge/nuvlaedge-go/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
else
    VERSION="$1"
fi

# Detect the operating system
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Detect the architecture
ARCH_RAW=$(uname -m)
if [ "$ARCH_RAW" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH_RAW" = "aarch64" ]; then
    ARCH="arm64"
elif [ "$ARCH_RAW" = "arm64" ]; then
    ARCH="arm64"
elif [ "$ARCH_RAW" = "armv7l" ]; then
    ARCH="arm"
else
    echo "Unsupported architecture: $ARCH_RAW"
    exit 1
fi

# Construct the download URL
URL="${BASE_URL}/${VERSION}/nuvlaedge-${OS}-${ARCH}-${VERSION}"

echo "Downloading NuvlaEdge CLI from $URL"
# Download the binary
curl -L -o nuvlaedge "$URL"

# Make the binary executable
chmod +x nuvlaedge-cli
