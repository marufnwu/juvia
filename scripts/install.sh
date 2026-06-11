#!/bin/bash
set -e

INSTALL_SCRIPT_VERSION="1.1.0"

echo "Juvia Installer v${INSTALL_SCRIPT_VERSION}"
echo "========================================"
echo ""

detect_arch() {
    case $(uname -m) in
        x86_64|amd64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) echo "amd64" ;;
    esac
}

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        case $ID in
            ubuntu|debian) echo "debian" ;;
            *) echo "unknown" ;;
        esac
    else
        echo "unknown"
    fi
}

ARCH=$(detect_arch)
OS=$(detect_os)

if [ "$OS" = "unknown" ]; then
    echo "Error: This installer only supports Debian/Ubuntu-based systems."
    exit 1
fi

echo "Detected: $OS ($ARCH)"

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root."
    exit 1
fi

LATEST_VERSION=$(curl -s https://api.github.com/repos/marufnwu/juvia/releases/latest 2>/dev/null | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4 || echo "0.1.0")
VERSION=${VERSION:-$LATEST_VERSION}

echo ""
echo "This will install the following packages on your server:"
echo ""
echo "  Web Server:     nginx, apache2"
echo "  PHP:            php-fpm (8.1, 8.2, 8.3) with all extensions"
echo "  Databases:      mysql-server, postgresql"
echo "  Email:          postfix, dovecot, rspamd"
echo "  DNS:            bind9"
echo "  SSL:            certbot, python3-certbot-nginx"
echo "  Firewall:       ufw"
echo "  Utilities:      git, curl, wget, rsync, unzip, sqlite3, htop, tree, net-tools"
echo ""
if [ "$AUTO_INSTALL" != "1" ]; then
  read -p "Continue with installation? [y/N] " -n 1 -r
  echo ""
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
  fi
fi

echo ""
echo "Installing system dependencies..."

apt-get update

echo ""
echo "Installing web servers..."
apt-get install -y nginx apache2

echo ""
echo "Installing PHP and extensions..."
apt-get install -y php-fpm php-cli
apt-get install -y php-mysql php-pgsql php-curl php-gd php-mbstring php-xml php-zip php-intl php-bcmath php-soap php-opcache php-apcu

echo ""
echo "Installing databases..."
apt-get install -y mysql-server postgresql

echo ""
echo "Installing email services..."
apt-get install -y postfix dovecot-imapd dovecot-pop3d rspamd

echo ""
echo "Installing DNS server..."
apt-get install -y bind9

echo ""
echo "Installing SSL tools..."
apt-get install -y certbot python3-certbot-nginx python3-certbot-apache

echo ""
echo "Installing firewall..."
apt-get install -y ufw

echo ""
echo "Installing utilities..."
apt-get install -y git curl wget rsync unzip sqlite3 htop tree net-tools sudo

echo ""
echo "System dependencies installed successfully."
echo ""

echo "Installing Juvia v${VERSION}..."

TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "Downloading package..."
PACKAGE_URL="https://github.com/marufnwu/juvia/releases/download/${VERSION}/juvia_${VERSION}_${ARCH}.deb"

if ! curl -fsSL "$PACKAGE_URL" -o juvia.deb 2>/dev/null; then
    echo "Error: Failed to download package from $PACKAGE_URL"
    echo "Please check if version ${VERSION} exists for architecture ${ARCH}"
    rm -rf "$TEMP_DIR"
    exit 1
fi

echo "Installing Juvia package..."
dpkg -i juvia.deb || {
    echo "Installing missing dependencies..."
    apt-get update
    apt-get install -f -y
    dpkg -i juvia.deb
}

echo ""
echo "========================================"
echo "Juvia has been installed successfully!"
echo ""
echo "Access the control panel at:"
echo "  http://localhost:18473"
echo ""
echo "Note: If you configure a hostname (domain) for the panel,"
echo "      the port will change to 8080."
echo ""
echo "Useful commands:"
echo "  systemctl status juvia          - Check panel status"
echo "  systemctl status juvia-agent    - Check agent status"
echo "  journalctl -u juvia -f          - View panel logs"
echo "  journalctl -u juvia-agent -f    - View agent logs"
echo ""
echo "========================================"

rm -rf "$TEMP_DIR"