#!/bin/bash
set -euo pipefail

REPO="Priyans-hu/argus"
BINARY="argus"

echo "Installing ${BINARY}..."

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
VERSION=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d '"' -f 4)
if [ -z "$VERSION" ]; then
  echo "Error: Could not determine latest version."
  exit 1
fi
echo "Found ${BINARY} ${VERSION}"

# Determine file extension
EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi

FILENAME="${BINARY}_${VERSION#v}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "${TMPDIR}"' EXIT

echo "Downloading ${FILENAME}..."
curl -sL "${URL}" -o "${TMPDIR}/${FILENAME}"

echo "Extracting..."
if [ "$EXT" = "tar.gz" ]; then
  tar -xzf "${TMPDIR}/${FILENAME}" -C "${TMPDIR}"
else
  unzip -q "${TMPDIR}/${FILENAME}" -d "${TMPDIR}"
fi

# Install binary
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi
chmod +x "${INSTALL_DIR}/${BINARY}"

echo ""
echo "${BINARY} ${VERSION} installed successfully!"
echo ""
echo "Run 'argus --help' to get started."
echo ""
echo "⭐ Star us on GitHub: https://github.com/${REPO}"
