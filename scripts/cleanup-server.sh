#!/bin/bash
# Server cleanup script - Option 2: Fresh Slate
# WARNING: This will remove aaPanel, nginx, Caddy, and other services!

set -e

echo "=== SERVER CLEANUP - FRESH SLATE ==="
echo "This will stop and remove conflicting services..."
echo ""

echo "[1/9] Stopping Juvia services..."
sudo systemctl stop juvia juvia-agent 2>/dev/null || true

echo "[2/9] Removing Juvia package..."
sudo dpkg -r juvia 2>/dev/null || true

echo "[3/9] Removing Juvia user and group..."
sudo userdel juvia 2>/dev/null || true
sudo groupdel juvia 2>/dev/null || true

echo "[4/9] Cleaning Juvia directories..."
sudo rm -rf /var/lib/juvia /var/log/juvia /etc/juvia /var/run/juvia

echo "[5/9] Stopping conflicting services..."
sudo systemctl stop nginx php8.1-fpm php8.2-fpm php8.3-fpm mysql mariadb postfix dovecot named caddy 2>/dev/null || true

echo "[6/9] Disabling conflicting services..."
sudo systemctl disable nginx php8.1-fpm php8.2-fpm php8.3-fpm mysql mariadb postfix dovecot named caddy 2>/dev/null || true

echo "[7/9] Removing conflicting packages..."
sudo apt-get remove -y nginx nginx-common nginx-full apache2 apache2-* php8.1-fpm php8.2-fpm php8.3-fpm mysql-server mariadb-server postfix dovecot bind9 caddy 2>/dev/null || true

echo "[8/9] Cleaning up remaining files..."
sudo rm -rf /etc/nginx /etc/php /var/www/html /var/lib/mysql /var/lib/bind /etc/bind /etc/juvia /etc/caddy* /usr/share/caddy 2>/dev/null || true

echo "[9/9] Purging configuration files..."
sudo apt-get purge -y nginx nginx-common apache2 php8.*-fpm mysql-server mariadb-server postfix dovecot bind9 caddy 2>/dev/null || true

echo ""
echo "=== Cleanup complete ==="
echo "Server is now clean. Ready for fresh Juvia install."