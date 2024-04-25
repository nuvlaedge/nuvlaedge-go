#!/bin/bash

# Initialize variables
VERSION=""
DEST_DIR=""
UUID=""

# Parse command-line arguments
while (( "$#" )); do
  case "$1" in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --dir)
      DEST_DIR="$2"
      shift 2
      ;;
    --uuid)
      UUID="$2"
      shift 2
      ;;
    --) # end argument parsing
      shift
      break
      ;;
    -*|--*=) # unsupported flags
      echo "Error: Unsupported flag $1" >&2
      exit 1
      ;;
    *) # preserve positional arguments
      PARAMS="$PARAMS $1"
      shift
      ;;
  esac
done
# set positional arguments in their proper place
eval set -- "$PARAMS"

echo "Installing NuvlaEdge with parameters: version=$VERSION, dir=$DEST_DIR, uuid=$UUID"

# Determine the OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map the architecture names to the ones used by NuvlaEdge
case $ARCH in
    x86_64) ARCH=amd64 ;;
    aarch64) ARCH=arm64 ;;
    arm64) ARCH=arm64 ;;
    armv*) ARCH=arm ;;
    *) echo "Unsupported architecture: $ARCH" ; exit 1 ;;
esac

# Map the OS names to the ones used by NuvlaEdge
case $OS in
    darwin) OS=darwin ;;
    linux) OS=linux ;;
    *) echo "Unsupported OS: $OS" ; exit 1 ;;
esac

# If no version is provided, get the latest version from the GitHub API
if [ -z "$VERSION" ]
then
    VERSION=$(curl --silent "https://api.github.com/repos/nuvlaedge/nuvlaedge-go/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

echo "Downloading NuvlaEdge version $VERSION for $OS/$ARCH"

# Construct the URL for the binary
URL="https://github.com/nuvlaedge/nuvlaedge-go/releases/download/$VERSION/nuvlaedge-$OS-$ARCH-$VERSION"

echo "Downloading NuvlaEdge from $URL"

# Download the binary
curl -L -O "$URL"

# Determine the destination directory
if [ -z "$DEST_DIR" ]
then
    if [ "$(id -u)" -eq 0 ]
    then
        DEST_DIR="/usr/local/bin"
        CONF_DIR="/etc/nuvlaedge"
        mkdir -p "$CONF_DIR"
    else
        DEST_DIR="$HOME/.nuvlaedge/bin"
        mkdir -p "$DEST_DIR"
    fi
fi

# Move the binary to the destination directory
mv "nuvlaedge-$OS-$ARCH-$VERSION" "$DEST_DIR/nuvlaedge"
chmod +x "$DEST_DIR/nuvlaedge"

# Download the template.toml file from the main branch
curl -L -o "$CONF_DIR/template.toml" "https://raw.githubusercontent.com/nuvlaedge/nuvlaedge-go/main/config/template.toml"

export NUVLAEDGE_SETTINGS="$CONF_DIR/template.toml"


if [ -n "$UUID" ]
then
    echo "Starting NuvlaEdge with UUID $UUID..."

    export NUVLAEDGE_UUID="$UUID"
    export NUVLAEDGE_SETTINGS="$DEST_DIR/template.toml"
    echo $NUVLAEDGE_UUID
    echo $NUVLAEDGE_SETTINGS
    cd $DEST_DIR
    ./nuvlaedge
fi

echo "NuvlaEdge has been installed to $DEST_DIR"