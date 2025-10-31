#!/bin/bash

# sesn Installation Script
# A tmux session manager with a beautiful TUI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/thetanav/sesn.git"
BINARY_NAME="sesn"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TEMP_DIR=$(mktemp -d)

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

trap cleanup EXIT

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if running as root or sudo
check_permissions() {
    if [ "$INSTALL_DIR" = "/usr/local/bin" ] && [ "$EUID" -ne 0 ]; then
        log_warning "Installing to $INSTALL_DIR requires root privileges."
        log_info "Run with sudo or set INSTALL_DIR to a user directory (e.g., export INSTALL_DIR=\$HOME/bin)"
        exit 1
    fi
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."

    # Check Go
    if ! command_exists go; then
        log_error "Go is not installed. Please install Go first:"
        log_info "  - Ubuntu/Debian: sudo apt install golang-go"
        log_info "  - macOS: brew install go"
        log_info "  - Or download from: https://golang.org/dl/"
        exit 1
    fi

    # Check tmux
    if ! command_exists tmux; then
        log_error "tmux is not installed. Please install tmux first:"
        log_info "  - Ubuntu/Debian: sudo apt install tmux"
        log_info "  - macOS: brew install tmux"
        log_info "  - CentOS/RHEL: sudo yum install tmux"
        exit 1
    fi

    # Check git
    if ! command_exists git; then
        log_error "git is not installed. Please install git first."
        exit 1
    fi

    log_success "All dependencies are installed"
}

# Clone repository
clone_repo() {
    log_info "Cloning sesn repository..."
    if ! git clone "$REPO_URL" "$TEMP_DIR"; then
        log_error "Failed to clone repository"
        exit 1
    fi
    log_success "Repository cloned successfully"
}

# Build binary
build_binary() {
    log_info "Building sesn binary..."
    cd "$TEMP_DIR"
    if ! go build -o "$BINARY_NAME" .; then
        log_error "Failed to build sesn"
        exit 1
    fi
    log_success "Binary built successfully"
}

# Install binary
install_binary() {
    log_info "Installing sesn to $INSTALL_DIR..."

    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR"
    fi

    # Move binary to install location
    if ! mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"; then
        log_error "Failed to install binary to $INSTALL_DIR"
        exit 1
    fi

    # Make sure it's executable
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    log_success "sesn installed successfully to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."
    if command_exists sesn; then
        log_success "sesn is now available in your PATH"
        sesn --help >/dev/null 2>&1 || log_warning "sesn installed but may not be fully functional"
    else
        log_warning "sesn installed but not found in PATH. You may need to restart your shell or add $INSTALL_DIR to your PATH"
        log_info "You can run it directly: $INSTALL_DIR/sesn"
    fi
}

# Print usage information
print_usage() {
    echo
    log_success "Installation complete!"
    echo
    echo "Usage:"
    echo "  sesn              # Launch the TUI"
    echo "  sesn -f           # Use fuzzy finder mode"
    echo
    echo "Keybindings in TUI:"
    echo "  c                 # Create new session"
    echo "  d                 # Delete selected session"
    echo "  r                 # Rename selected session"
    echo "  k                 # Kill selected session"
    echo "  enter             # Attach to selected session"
    echo "  /                 # Fuzzy find mode"
    echo "  ctrl+c            # Quit"
    echo
    echo "For more information, visit: https://github.com/thetanav/sesn"
}

# Main installation process
main() {
    echo
    echo "  ___  ___  ___ _ __  "
    echo " / __|/ _ \/ __| '_ \ "
    echo " \__ \  __/\__ \ | | |"
    echo " |___/\___||___/_| |_|"
    echo
    echo "Installing sesn - A tmux session manager with a beautiful TUI"
    echo

    check_permissions
    check_dependencies
    clone_repo
    build_binary
    install_binary
    verify_installation
    print_usage
}

# Run main function
main "$@"
<filePath>/home/thetanav/Code/go/sesn/install.sh