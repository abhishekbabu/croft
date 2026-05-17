#!/bin/sh
# Install croft — https://github.com/abhishekbabu/croft
#
#   curl -fsSL https://raw.githubusercontent.com/abhishekbabu/croft/main/install.sh | sh
#
# Environment:
#   CROFT_VERSION  version tag to install (default: latest release)
#   CROFT_BIN_DIR  install directory (default: /usr/local/bin)
set -eu

REPO="abhishekbabu/croft"
BIN_DIR="${CROFT_BIN_DIR:-/usr/local/bin}"

die() { echo "install: $*" >&2; exit 1; }

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux | darwin) ;;
  *) die "unsupported OS: $os" ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) die "unsupported architecture: $arch" ;;
esac

version="${CROFT_VERSION:-}"
if [ -z "$version" ]; then
  version=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4)
  [ -n "$version" ] || die "could not determine the latest version (set CROFT_VERSION)"
fi
num=${version#v}

tarball="croft_${num}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$version"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading croft $version ($os/$arch)..."
curl -fsSL "$base/$tarball" -o "$tmp/$tarball" || die "download failed: $base/$tarball"

# Best-effort checksum verification.
if curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt" 2>/dev/null; then
  if command -v sha256sum >/dev/null 2>&1; then
    sumtool="sha256sum"
  elif command -v shasum >/dev/null 2>&1; then
    sumtool="shasum -a 256"
  else
    sumtool=""
  fi
  if [ -n "$sumtool" ]; then
    expected=$(grep " $tarball\$" "$tmp/checksums.txt" | cut -d' ' -f1)
    actual=$($sumtool "$tmp/$tarball" | cut -d' ' -f1)
    [ "$expected" = "$actual" ] || die "checksum mismatch for $tarball"
    echo "Checksum verified."
  fi
fi

tar -xzf "$tmp/$tarball" -C "$tmp" croft

if [ -w "$BIN_DIR" ]; then
  mv "$tmp/croft" "$BIN_DIR/croft"
  chmod +x "$BIN_DIR/croft"
else
  echo "Installing to $BIN_DIR (requires sudo)..."
  sudo mv "$tmp/croft" "$BIN_DIR/croft"
  sudo chmod +x "$BIN_DIR/croft"
fi

echo "croft installed to $BIN_DIR/croft"
"$BIN_DIR/croft" --version
