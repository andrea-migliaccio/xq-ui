#!/bin/bash
# XQ-UI Installation Script

set -e

# Configuration
REPO="andrea-migliaccio/xq-ui"
BINARY_NAME="xq-ui"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            error "Unsupported operating system: $os"
            ;;
    esac
    
    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
    
    info "Detected platform: ${OS}_${ARCH}"
}

# Get latest release version
get_latest_version() {
    info "Fetching latest release information..."
    local release_url="https://api.github.com/repos/${REPO}/releases/latest"
    
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "${release_url}" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "${release_url}" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
    
    if [ -z "$VERSION" ]; then
        error "Could not fetch latest version"
    fi
    
    info "Latest version: $VERSION"
}

# Download and extract binary
download_binary() {
    local filename="${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${filename}"
    local temp_dir=$(mktemp -d)
    
    info "Downloading $filename..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "$download_url" -o "${temp_dir}/${filename}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "${temp_dir}/${filename}"
    fi
    
    if [ ! -f "${temp_dir}/${filename}" ]; then
        error "Download failed"
    fi
    
    info "Extracting binary..."
    tar -xzf "${temp_dir}/${filename}" -C "$temp_dir"
    
    if [ ! -f "${temp_dir}/${BINARY_NAME}" ]; then
        error "Binary not found in archive"
    fi
    
    BINARY_PATH="${temp_dir}/${BINARY_NAME}"
}

# Install binary
install_binary() {
    info "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
    
    # Check if we need sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        if command -v sudo >/dev/null 2>&1; then
            sudo cp "$BINARY_PATH" "$INSTALL_DIR/"
            sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
        else
            error "Cannot write to $INSTALL_DIR and sudo is not available"
        fi
    else
        cp "$BINARY_PATH" "$INSTALL_DIR/"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Cleanup
    rm -rf "$(dirname "$BINARY_PATH")"
    
    success "${BINARY_NAME} installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local installed_version=$("$BINARY_NAME" --help | head -n2 | tail -n1 | awk '{print $2}')
        success "Installation verified: $BINARY_NAME $installed_version"
        
        info "You can now run: $BINARY_NAME --help"
        
        # Check dependencies
        info "Checking dependencies..."
        if ! command -v jq >/dev/null 2>&1; then
            warn "jq is not installed. Install with: sudo apt install jq (Ubuntu/Debian) or brew install jq (macOS)"
        fi
        
        if ! command -v yq >/dev/null 2>&1; then
            warn "yq is not installed. Install from: https://github.com/mikefarah/yq#install"
        fi
        
        info "For AI features, set your OpenAI API key:"
        info "export OPENAI_API_KEY='your-api-key-here'"
    else
        error "Installation verification failed"
    fi
}

# Main installation flow
main() {
    echo -e "${BLUE}XQ-UI Installation Script${NC}"
    echo "==============================="
    
    detect_platform
    get_latest_version
    download_binary
    install_binary
    verify_installation
    
    echo
    success "Installation complete! 🎉"
}

# Run main function
main "$@"