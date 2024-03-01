#!/bin/bash
set -euxo pipefail

# Function to check if kitty is installed
is_kitty_installed() {
    if command -v kitty &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to install kitty on macOS
install_kitty_mac() {
    echo ">>>>> Will install kitty"
    brew tap ImSingee/kitty
    brew install kitty
}

# Function to install kitty on Linux
install_kitty_linux() {
    echo ">>>>> Will install kitty"
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        *)
            echo "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    KITTY_URL="https://go.singee.site/kitty/latest-linux-$ARCH.tar.gz"
    curl -o /tmp/kitty.tar.gz $KITTY_URL
    tar -xzf /tmp/kitty.tar.gz -C /tmp
    mv /tmp/kitty /usr/local/bin/kitty
    chmod +x /usr/local/bin/kitty
}

# Function to run kitty install command
run_kitty_install() {
    kitty install --from-direnv
}

# Detect the operating system
OS=$(uname)

if is_kitty_installed; then
    echo "Kitty is already installed."
else
    if [ "$OS" = "Darwin" ]; then
        # macOS system detected
        install_kitty_mac
    elif [ "$OS" = "Linux" ]; then
        # Linux system detected
        install_kitty_linux
    else
        echo "Unsupported operating system: $OS"
        exit 1
    fi
fi

# Run the kitty install command
if [ -d ".git" ]; then
  run_kitty_install
fi
