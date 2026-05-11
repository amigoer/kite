#!/usr/bin/env sh
# Install the latest kite binary into ~/.local/bin (or $PREFIX/bin).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/amigoer/kite/main/install.sh | sh
#   PREFIX=/usr/local sh install.sh

set -eu

REPO="amigoer/kite"
PREFIX="${PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"

uname_s=$(uname -s | tr '[:upper:]' '[:lower:]')
uname_m=$(uname -m)

case "$uname_s" in
  linux)  os="linux" ;;
  darwin) os="darwin" ;;
  *) echo "kite: unsupported OS: $uname_s" >&2; exit 1 ;;
esac

case "$uname_m" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "kite: unsupported arch: $uname_m" >&2; exit 1 ;;
esac

# Resolve the latest tag from GitHub.
tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
  grep '"tag_name":' | head -1 | sed -E 's/.*"v?([^"]+)".*/\1/')
if [ -z "$tag" ]; then
  echo "kite: could not determine latest release" >&2
  exit 1
fi

archive="kite_${tag}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/v${tag}/${archive}"

echo "kite: downloading $url"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL "$url" -o "$tmp/$archive"
tar -xzf "$tmp/$archive" -C "$tmp"

mkdir -p "$BIN_DIR"
mv "$tmp/kite" "$BIN_DIR/kite"
chmod +x "$BIN_DIR/kite"

echo
echo "kite installed to $BIN_DIR/kite"
echo
case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) echo "Add this to your shell rc:"; echo "  export PATH=\"$BIN_DIR:\$PATH\"" ;;
esac
echo
echo "Try it: kite version"
