#!/bin/bash
set -e

VERSION=${1:-0.1.0}
ARCH=${2:-amd64}
BUILD_DIR="/tmp/juvia-build-$$"

echo "Building Juvia v${VERSION} for ${ARCH}"

mkdir -p "${BUILD_DIR}/DEBIAN"
mkdir -p "${BUILD_DIR}/usr/bin"
mkdir -p "${BUILD_DIR}/etc/systemd/system"
mkdir -p "${BUILD_DIR}/etc/juvia"

echo "Building binaries..."
cd "$(dirname "$0")/.."
CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -ldflags="-s -w" -o "${BUILD_DIR}/usr/bin/juvia" ./cmd/panel
CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} go build -ldflags="-s -w" -o "${BUILD_DIR}/usr/bin/juvia-agent" ./cmd/agent

echo "Installing systemd services..."
cp scripts/juvia.service "${BUILD_DIR}/etc/systemd/system/"
cp scripts/juvia-agent.service "${BUILD_DIR}/etc/systemd/system/"

echo "Creating DEB control file..."
cat > "${BUILD_DIR}/DEBIAN/control" <<EOF
Package: juvia
Version: ${VERSION}
Section: admin
Priority: optional
Architecture: ${ARCH}
Depends: sqlite3, sudo, ufw, curl, wget, rsync, unzip, git, net-tools
Recommends: nginx | apache2, php-fpm, php-cli, php-mysql, php-pgsql, php-curl, php-gd, php-mbstring, php-xml, php-zip, php-intl, php-bcmath, php-soap, php-opcache, php-apcu, mysql-server | mariadb-server, postgresql, postfix, dovecot-imapd, dovecot-pop3d, rspamd, bind9, certbot, python3-certbot-nginx | python3-certbot-apache, htop, tree
Suggests: python3-certbot-apache
Maintainer: Juvia Team <team@juvia.io>
Description: Server Control Panel
 Juvia is a modern server control panel for managing websites,
 email, databases, DNS, firewall, backups, and more.
EOF

cat > "${BUILD_DIR}/DEBIAN/postinst" <<'EOF'
#!/bin/bash
set -e

if [ "$1" = "configure" ]; then
    echo "Creating juvia user..."
    if ! id -u juvia >/dev/null 2>&1; then
        useradd -r -s /bin/bash -m juvia
    fi

    echo "Creating directories..."
    mkdir -p /var/run/juvia
    mkdir -p /var/log/juvia/terminal
    mkdir -p /var/lib/juvia
    mkdir -p /etc/juvia/nginx/sites
    mkdir -p /etc/juvia/php-fpm/pools

    echo "Setting permissions..."
    chown -R juvia:juvia /var/run/juvia
    chown -R juvia:juvia /var/log/juvia
    chown -R juvia:juvia /var/lib/juvia
    chown root:juvia /etc/juvia
    chmod 770 /etc/juvia
    if [ -S /var/run/juvia/agent.sock ]; then
        chmod 660 /var/run/juvia/agent.sock
    fi

    echo "Enabling services..."
    systemctl daemon-reload
    systemctl enable juvia
    systemctl enable juvia-agent

    echo "Starting services..."
    systemctl start juvia
    systemctl start juvia-agent
fi
EOF

cat > "${BUILD_DIR}/DEBIAN/prerm" <<'EOF'
#!/bin/bash
set -e

if [ "$1" = "remove" ] || [ "$1" = "purge" ]; then
    echo "Stopping services..."
    systemctl stop juvia 2>/dev/null || true
    systemctl stop juvia-agent 2>/dev/null || true
    systemctl disable juvia 2>/dev/null || true
    systemctl disable juvia-agent 2>/dev/null || true
fi
EOF

cat > "${BUILD_DIR}/DEBIAN/postrm" <<'EOF'
#!/bin/bash
set -e

if [ "$1" = "remove" ] || [ "$1" = "purge" ]; then
    echo "Removing systemd service files..."
    rm -f /etc/systemd/system/juvia.service
    rm -f /etc/systemd/system/juvia-agent.service
    systemctl daemon-reload

    if [ "$1" = "purge" ]; then
        echo "Removing Juvia user..."
        userdel juvia 2>/dev/null || true
        rm -rf /var/lib/juvia 2>/dev/null || true
        rm -rf /var/log/juvia 2>/dev/null || true
        rm -rf /var/run/juvia 2>/dev/null || true
        rm -rf /etc/juvia 2>/dev/null || true
    fi
fi
EOF

cat > "${BUILD_DIR}/DEBIAN/preinst" <<'EOF'
#!/bin/bash
set -e

if [ "$1" = "install" ] || [ "$1" = "upgrade" ]; then
    if id -u juvia >/dev/null 2>&1; then
        echo "User 'juvia' already exists."
    fi

    if systemctl is-active --quiet juvia 2>/dev/null; then
        echo "Warning: Juvia panel is running. It will be restarted after upgrade."
    fi
fi
EOF

chmod 755 "${BUILD_DIR}/DEBIAN/postinst"
chmod 755 "${BUILD_DIR}/DEBIAN/prerm"
chmod 755 "${BUILD_DIR}/DEBIAN/postrm"
chmod 755 "${BUILD_DIR}/DEBIAN/preinst"

echo "Building DEB package..."
dpkg-deb --build "${BUILD_DIR}" "juvia_${VERSION}_${ARCH}.deb"

echo "Cleaning up..."
rm -rf "${BUILD_DIR}"

echo "Done! Package created: juvia_${VERSION}_${ARCH}.deb"