#!/bin/bash
set -e

UPDATE_VERSION="1.0.0"

echo "Juvia Updater v${UPDATE_VERSION}"
echo "========================================"

detect_arch() {
    case $(uname -m) in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) echo "amd64" ;;
    esac
}

ARCH=$(detect_arch)

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root."
    exit 1
fi

CURRENT_VERSION=$(dpkg -s juvia 2>/dev/null | grep '^Version:' | awk '{print $2}' || echo "unknown")
echo "Current version: ${CURRENT_VERSION}"

LATEST_VERSION=$(curl -s https://api.github.com/repos/marufnwu/juvia/releases/latest 2>/dev/null | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4 || echo "")

if [ -z "$LATEST_VERSION" ]; then
    echo "Error: Could not fetch latest version from GitHub."
    exit 1
fi

echo "Latest version: ${LATEST_VERSION}"

if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
    echo "You are already on the latest version."
    exit 0
fi

echo ""
echo "Updating from ${CURRENT_VERSION} to ${LATEST_VERSION}..."
echo ""

TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Downloading package..."
PACKAGE_URL="https://github.com/marufnwu/juvia/releases/download/${LATEST_VERSION}/juvia_${LATEST_VERSION}_${ARCH}.deb"

if ! curl -fsSL "$PACKAGE_URL" -o juvia.deb 2>/dev/null; then
    echo "Error: Failed to download package from $PACKAGE_URL"
    echo "Please check if version ${LATEST_VERSION} exists for architecture ${ARCH}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

echo "Installing package..."
dpkg -i juvia.deb || {
    echo "Installing dependencies..."
    apt-get update
    apt-get install -f -y
    dpkg -i juvia.deb
}

echo ""
echo "========================================"
echo "Juvia has been updated to v${LATEST_VERSION}!"
echo ""
echo "Access the control panel at:"
echo "  http://localhost:8080"
echo ""
echo "Useful commands:"
echo "  systemctl restart juvia    - Restart panel"
echo "  systemctl restart juvia-agent - Restart agent"
echo "  systemctl status juvia     - Check panel status"
echo "========================================"

rm -rf "$TEMP_DIR"