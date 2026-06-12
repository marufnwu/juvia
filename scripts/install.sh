#!/bin/bash
set -e

INSTALL_SCRIPT_VERSION="1.3.0"

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

wait_for_apt_lock() {
    local max_attempts=30
    local attempt=0
    echo "Waiting for apt lock..."
    while fuser /var/lib/apt/lists/lock /var/cache/apt/archives/lock /var/lib/dpkg/lock-frontend /var/lib/dpkg/lock >/dev/null 2>&1; do
        attempt=$((attempt + 1))
        if [ $attempt -ge $max_attempts ]; then
            echo "Error: apt is locked by another process for too long. Please wait and retry."
            exit 1
        fi
        echo "  Attempt $attempt/$max_attempts - waiting 10s..."
        sleep 10
    done
}

detect_existing_services() {
    HAS_MYSQL=0
    HAS_POSTGRES=0
    HAS_NGINX=0
    HAS_APACHE=0
    HAS_CADDY=0
    HAS_PHP=0

    [ -x "$(command -v mysqld)" ] || [ -x "$(command -v mariadbd)" ] && HAS_MYSQL=1
    [ -x "$(command -v postgres)" ] && HAS_POSTGRES=1
    [ -x "$(command -v nginx)" ] && HAS_NGINX=1
    [ -x "$(command -v apache2)" ] && HAS_APACHE=1
    [ -x "$(command -v caddy)" ] && HAS_CADDY=1
    [ -x "$(command -v php)" ] || dpkg -l php-fpm >/dev/null 2>&1 && HAS_PHP=1
}

verify_install() {
    local errors=0
    local port="${PANEL_PORT:-18473}"

    echo ""
    echo "Verifying installation..."

    if ! systemctl is-active --quiet juvia 2>/dev/null; then
        echo "  ERROR: juvia service is not running"
        errors=$((errors + 1))
    else
        echo "  OK: juvia service is running"
    fi

    if ! systemctl is-active --quiet juvia-agent 2>/dev/null; then
        echo "  ERROR: juvia-agent service is not running"
        errors=$((errors + 1))
    else
        echo "  OK: juvia-agent service is running"
    fi

    if [ -S /var/run/juvia/agent.sock ]; then
        echo "  OK: agent socket exists"
    else
        echo "  ERROR: agent socket not found"
        errors=$((errors + 1))
    fi

    if curl -sf "http://127.0.0.1:${port}/api/v1/health" >/dev/null 2>&1; then
        echo "  OK: panel responding on port ${port}"
    else
        echo "  WARNING: panel not responding on port ${port} (may still be starting)"
    fi

    if command -v certbot >/dev/null 2>&1; then
        echo "  OK: certbot installed"
    else
        echo "  WARNING: certbot not found in PATH"
    fi

    if command -v dovecot >/dev/null 2>&1; then
        echo "  OK: dovecot installed"
    else
        echo "  WARNING: dovecot not found"
    fi

    if systemctl is-active --quiet nginx 2>/dev/null; then
        echo "  OK: nginx running"
    else
        echo "  WARNING: nginx not running"
    fi

    if [ $errors -gt 0 ]; then
        echo ""
        echo "Installation completed with $errors error(s). Check logs above."
        return 1
    fi

    echo ""
    echo "Installation verified successfully!"
    return 0
}

repair_services() {
    echo ""
    echo "Repairing services..."

    echo "  Checking nginx..."
    if command -v nginx >/dev/null 2>&1; then
        nginx -t 2>/dev/null && echo "    OK: nginx config valid" || echo "    ERROR: nginx config invalid"
    fi

    echo "  Checking certbot..."
    if ! command -v certbot >/dev/null 2>&1 && command -v python3 >/dev/null 2>&1; then
        python3 -m pip install certbot --quiet 2>/dev/null || true
        if [ -f /usr/local/bin/certbot ]; then
            ln -sf /usr/local/bin/certbot /usr/bin/certbot 2>/dev/null || true
        fi
    fi

    echo "  Checking dovecot..."
    if command -v dovecot >/dev/null 2>&1; then
        if ! systemctl is-active --quiet dovecot 2>/dev/null; then
            systemctl start dovecot 2>/dev/null || true
        fi
        if ! command -v doveadm >/dev/null 2>&1; then
            apt-get install -y dovecot-core 2>/dev/null || true
        fi
    fi

    echo "  Fixing nginx dead symlinks..."
    if [ -d /etc/nginx/sites-enabled ]; then
        for link in /etc/nginx/sites-enabled/*.conf; do
            if [ -L "$link" ] && [ ! -e "$link" ]; then
                echo "    Removing dead symlink: $link"
                rm -f "$link"
            fi
        done
    fi

    echo "  Restarting services..."
    systemctl restart nginx 2>/dev/null || true
    systemctl restart dovecot 2>/dev/null || true

    echo "  Repair complete."
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

REPAIR_MODE=0
if [ "$1" = "--repair" ]; then
    REPAIR_MODE=1
fi

wait_for_apt_lock

detect_existing_services

LATEST_VERSION=$(curl -s https://api.github.com/repos/marufnwu/juvia/releases/latest 2>/dev/null | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4 || echo "0.1.0")
VERSION=${VERSION:-$LATEST_VERSION}

echo ""
echo "This will install the following packages on your server:"
echo ""
echo "  Web Server:     nginx or apache2 (if not already installed)"
echo "  PHP:            php-fpm with extensions (if not already installed)"
echo "  Databases:      mysql or postgresql (if not already installed)"
echo "  Email:          postfix, dovecot, rspamd (if not already installed)"
echo "  DNS:            bind9 (if not already installed)"
echo "  SSL:            certbot (if not already installed)"
echo "  Firewall:       ufw"
echo "  Utilities:      git, curl, wget, rsync, unzip, sqlite3, htop, tree, net-tools, sudo"
echo ""

if [ "$HAS_NGINX" = "1" ]; then echo "  Note: nginx already installed, skipping"; fi
if [ "$HAS_APACHE" = "1" ]; then echo "  Note: apache2 already installed, skipping"; fi
if [ "$HAS_MYSQL" = "1" ]; then echo "  Note: MySQL/MariaDB already installed, skipping"; fi
if [ "$HAS_POSTGRES" = "1" ]; then echo "  Note: PostgreSQL already installed, skipping"; fi
if [ "$HAS_CADDY" = "1" ]; then echo "  Note: Caddy already installed (unsupported, install nginx/apache separately)"; fi
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
echo "Updating package lists..."
apt-get update -qq

echo ""
echo "Installing base utilities..."
apt-get install -y git curl wget rsync unzip sqlite3 htop tree net-tools sudo ufw

echo ""
echo "Installing web server..."
if [ "$HAS_NGINX" != "1" ] && [ "$HAS_APACHE" != "1" ]; then
    apt-get install -y nginx
else
    echo "  Skipping (already installed)"
fi

echo ""
echo "Installing PHP and extensions..."
if [ "$HAS_PHP" != "1" ]; then
    apt-get install -y php-fpm php-cli
    apt-get install -y php-mysql php-pgsql php-curl php-gd php-mbstring php-xml php-zip php-intl php-bcmath php-soap php-opcache php-apcu
else
    echo "  Skipping (already installed)"
fi

echo ""
echo "Installing databases..."
if [ "$HAS_MYSQL" != "1" ]; then
    apt-get install -y mysql-server || apt-get install -y mariadb-server
else
    echo "  Skipping MySQL/MariaDB (already installed)"
fi

if [ "$HAS_POSTGRES" != "1" ]; then
    apt-get install -y postgresql
else
    echo "  Skipping PostgreSQL (already installed)"
fi

echo ""
echo "Starting database services..."
if command -v mysqld >/dev/null 2>&1; then
    systemctl enable mysql 2>/dev/null || true
    systemctl start mysql 2>/dev/null || true
fi
if command -v mariadbd >/dev/null 2>&1; then
    systemctl enable mariadb 2>/dev/null || true
    systemctl start mariadb 2>/dev/null || true
fi
if command -v pg_ctl >/dev/null 2>&1; then
    systemctl enable postgresql 2>/dev/null || true
    systemctl start postgresql 2>/dev/null || true
fi
if command -v named >/dev/null 2>&1; then
    systemctl enable bind9 2>/dev/null || true
    systemctl start bind9 2>/dev/null || true
fi

echo ""
echo "Starting web server..."
if command -v nginx >/dev/null 2>&1; then
    systemctl enable nginx 2>/dev/null || true
    systemctl start nginx 2>/dev/null || true
fi
if command -v apache2 >/dev/null 2>&1; then
    systemctl enable apache2 2>/dev/null || true
    systemctl start apache2 2>/dev/null || true
fi

echo ""
echo "Starting email services..."
if command -v dovecot >/dev/null 2>&1; then
    systemctl enable dovecot 2>/dev/null || true
    systemctl start dovecot 2>/dev/null || true
fi
if command -v postfix >/dev/null 2>&1; then
    systemctl enable postfix 2>/dev/null || true
    systemctl start postfix 2>/dev/null || true
fi

echo ""
echo "Setting up certbot..."
if command -v certbot >/dev/null 2>&1; then
    echo "  certbot already installed"
elif command -v snap >/dev/null 2>&1; then
    snap install certbot --classic 2>/dev/null || true
    ln -sf /snap/bin/certbot /usr/bin/certbot 2>/dev/null || true
elif command -v python3 >/dev/null 2>&1; then
    python3 -m pip install certbot --quiet 2>/dev/null || true
    ln -sf /usr/local/bin/certbot /usr/bin/certbot 2>/dev/null || true
fi
if command -v certbot >/dev/null 2>&1; then
    mkdir -p /etc/juvia/ssl
    touch /etc/juvia/ssl/renovate.json
fi

echo ""
echo "Installing email services..."
apt-get install -y postfix dovecot-imapd dovecot-pop3d rspamd

echo ""
echo "Installing DNS server..."
apt-get install -y bind9

echo ""
echo "Setting up Juvia runtime directories and permissions..."
id -u juvia >/dev/null 2>&1 || useradd -r -s /usr/sbin/nologin -d /var/lib/juvia -m juvia 2>/dev/null || true

mkdir -p /etc/juvia/nginx/sites
mkdir -p /etc/juvia/nginx/pool.d
mkdir -p /etc/juvia/php-fpm/pools
mkdir -p /etc/juvia/postfix
mkdir -p /etc/juvia/dovecot
mkdir -p /etc/juvia/bind/zones
mkdir -p /etc/juvia/ufw
mkdir -p /var/lib/juvia
mkdir -p /var/log/juvia
mkdir -p /var/log/juvia/terminal
mkdir -p /var/run/juvia
mkdir -p /var/backups/juvia

chown -R www-data:www-data /etc/juvia/nginx/sites /etc/juvia/nginx/pool.d
chown -R www-data:www-data /etc/juvia/php-fpm/pools
chmod -R 755 /etc/juvia/nginx /etc/juvia/php-fpm
chown -R juvia:juvia /var/lib/juvia /var/log/juvia /var/run/juvia
chmod 755 /var/lib/juvia /var/log/juvia /var/run/juvia
chmod 770 /var/log/juvia/terminal
chown -R juvia:juvia /var/backups/juvia
chmod 755 /var/backups/juvia

if [ -d /etc/nginx/sites-enabled ]; then
    chown root:www-data /etc/nginx/sites-enabled
    chmod 755 /etc/nginx/sites-enabled
fi

if [ -d /etc/php ]; then
    for fpm_ver in /etc/php/*/fpm; do
        if [ -d "$fpm_ver/pool.d" ]; then
            chmod 755 "$fpm_ver/pool.d"
        fi
    done
fi

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
    apt-get update -qq
    apt-get install -f -y --no-remove
    dpkg -i juvia.deb
}

PANEL_PORT=$(sqlite3 /var/lib/juvia/juvia.db "SELECT value FROM settings WHERE key='panel_port'" 2>/dev/null || echo "18473")
echo ""
echo "Configuring firewall..."
ufw --force enable 2>/dev/null || true
ufw allow "$PANEL_PORT/tcp" comment 'juvia-panel' 2>/dev/null || true
ufw allow 22/tcp comment 'ssh' 2>/dev/null || true
ufw reload 2>/dev/null || true

echo ""
echo "========================================"
echo "Juvia has been installed successfully!"
echo ""
echo "Access the control panel at:"
echo "  http://localhost:${PANEL_PORT}"
echo ""
echo "Useful commands:"
echo "  systemctl status juvia          - Check panel status"
echo "  systemctl status juvia-agent    - Check agent status"
echo "  journalctl -u juvia -f          - View panel logs"
echo "  journalctl -u juvia-agent -f    - View agent logs"
echo ""
echo "========================================"

verify_install

if [ "$REPAIR_MODE" = "1" ]; then
    repair_services
fi

rm -rf "$TEMP_DIR"