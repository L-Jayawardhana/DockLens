#!/bin/sh
# scripts/install.sh
#
# Universal installer for DockLens
# Usage: curl -fsSL https://raw.githubusercontent.com/L-Jayawardhana/DockLens/main/scripts/install.sh | sh
#
# What this does:
#   1. Detects your OS and CPU architecture
#   2. Downloads the correct binary from GitHub Releases
#   3. Places it in /usr/local/bin so you can run `docklens` anywhere

set -e  # Exit immediately if any command fails

# ─── Configuration ─────────────────────────────────────────────────────────────
REPO="L-Jayawardhana/DockLens"
BINARY_NAME="docklens"
INSTALL_DIR="/usr/local/bin"

# ─── Colors for output ────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RESET='\033[0m'

info()    { printf "${CYAN}[docklens]${RESET} %s\n" "$1"; }
success() { printf "${GREEN}[docklens]${RESET} %s\n" "$1"; }
warn()    { printf "${YELLOW}[docklens]${RESET} %s\n" "$1"; }
error()   { printf "${RED}[docklens] ERROR:${RESET} %s\n" "$1" >&2; exit 1; }

# ─── Check required tools ─────────────────────────────────────────────────────
need_cmd() {
  if ! command -v "$1" > /dev/null 2>&1; then
    error "Required command not found: '$1'. Please install it and try again."
  fi
}

need_cmd curl
need_cmd tar

# ─── Detect OS ────────────────────────────────────────────────────────────────
detect_os() {
  OS="$(uname -s)"
  case "$OS" in
    Linux*)   echo "linux" ;;
    Darwin*)  echo "darwin" ;;
    *)        error "Unsupported OS: $OS. Install manually from https://github.com/$REPO/releases" ;;
  esac
}

# ─── Detect CPU architecture ──────────────────────────────────────────────────
detect_arch() {
  ARCH="$(uname -m)"
  case "$ARCH" in
    x86_64 | amd64)  echo "amd64" ;;
    aarch64 | arm64) echo "arm64" ;;
    *)               error "Unsupported architecture: $ARCH. Install manually from https://github.com/$REPO/releases" ;;
  esac
}

# ─── Get latest version from GitHub API ───────────────────────────────────────
get_latest_version() {
  LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"([^"]+)".*/\1/')

  if [ -z "$LATEST" ]; then
    error "Could not fetch latest version. Check your internet connection or visit https://github.com/$REPO/releases"
  fi

  echo "$LATEST"
}

# ─── Main install logic ───────────────────────────────────────────────────────
main() {
  info "Starting DockLens installation..."

  OS=$(detect_os)
  ARCH=$(detect_arch)
  VERSION=$(get_latest_version)

  # Strip the "v" prefix for the filename (GoReleaser uses the version without "v" in filenames)
  VERSION_NUM="${VERSION#v}"

  # Build the filename exactly as GoReleaser names it
  ARCHIVE_NAME="${BINARY_NAME}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE_NAME"
  CHECKSUM_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"

  info "Detected OS:   $OS"
  info "Detected Arch: $ARCH"
  info "Version:       $VERSION"
  info "Downloading:   $DOWNLOAD_URL"

  # Create a temp directory that gets cleaned up automatically
  TMP_DIR=$(mktemp -d)
  trap 'rm -rf "$TMP_DIR"' EXIT  # Always cleanup on exit

  # Download the archive
  curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE_NAME" || \
    error "Download failed. Check your internet connection."

  # Download and verify checksum
  info "Verifying checksum..."
  curl -fsSL "$CHECKSUM_URL" -o "$TMP_DIR/checksums.txt" || \
    warn "Could not download checksums. Skipping verification."

  if [ -f "$TMP_DIR/checksums.txt" ]; then
    cd "$TMP_DIR"
    if command -v sha256sum > /dev/null 2>&1; then
      grep "$ARCHIVE_NAME" checksums.txt | sha256sum -c - || error "Checksum verification failed!"
    elif command -v shasum > /dev/null 2>&1; then
      grep "$ARCHIVE_NAME" checksums.txt | shasum -a 256 -c - || error "Checksum verification failed!"
    fi
    info "Checksum verified ✓"
    cd - > /dev/null
  fi

  # Extract the binary
  tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR"

  # Install the binary
  if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
  else
    info "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
  fi

  # Verify it works
  if command -v "$BINARY_NAME" > /dev/null 2>&1; then
    success "DockLens $VERSION installed successfully!"
    success "Run it with: docklens"
  else
    warn "Binary installed but not found in PATH."
    warn "Add $INSTALL_DIR to your PATH, or run: $INSTALL_DIR/$BINARY_NAME"
  fi
}

main "$@"
