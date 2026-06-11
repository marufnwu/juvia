#!/bin/bash
# Complete server cleanup - Remove EVERYTHING
# WARNING: This will completely wipe aaPanel, nginx, php, mysql, and all other services!

set -e

echo "========================================"
echo "COMPLETE SERVER CLEANUP"
echo "WARNING: This removes EVERYTHING"
echo "========================================"

echo "[1/12] Stopping all services..."
sudo systemctl stop juvia juvia-agent nginx php8.1-fpm php8.2-fpm php8.3-fpm mysql mariadb postfix dovecot named caddy 2>/dev/null || true
sudo killall -9 nginx php-fpm mysqld mariadbd named caddy 2>/dev/null || true

echo "[2/12] Stopping aaPanel/s6 supervision..."
sudo pkill -9 -u 9999 2>/dev/null || true
sudo rm -f /etc/systemd/system/juvia* /etc/systemd/system/nginx* /etc/systemd/system/php* 2>/dev/null || true

echo "[3/12] Removing Juvia..."
sudo dpkg -r juvia 2>/dev/null || true
sudo userdel juvia 2>/dev/null || true
sudo groupdel juvia 2>/dev/null || true
sudo rm -rf /var/lib/juvia /var/log/juvia /etc/juvia /var/run/juvia /run/juvia 2>/dev/null || true

echo "[4/12] Removing nginx (all variants)..."
sudo apt-get purge -y nginx nginx-common nginx-full apache2 apache2-* 2>/dev/null || true
sudo rm -rf /etc/nginx /var/log/nginx /var/www/html 2>/dev/null || true

echo "[5/12] Removing PHP (all versions)..."
sudo apt-get purge -y php* php8.* php-fpm 2>/dev/null || true
sudo rm -rf /etc/php /var/log/php* /var/lib/php 2>/dev/null || true

echo "[6/12] Removing MySQL/MariaDB..."
sudo apt-get purge -y mysql-server mysql-client mariadb-server mariadb-client 2>/dev/null || true
sudo rm -rf /etc/mysql /var/lib/mysql /var/log/mysql /run/mysql 2>/dev/null || true

echo "[7/12] Removing PostgreSQL..."
sudo apt-get purge -y postgresql* 2>/dev/null || true
sudo rm -rf /etc/postgresql /var/lib/postgresql /var/log/postgresql 2>/dev/null || true

echo "[8/12] Removing Postfix/Dovecot..."
sudo apt-get purge -y postfix postfix-* dovecot dovecot-* 2>/dev/null || true
sudo rm -rf /etc/postfix /var/spool/postfix /var/log/mail* 2>/dev/null || true

echo "[9/12] Removing BIND9..."
sudo apt-get purge -y bind9 bind9-* 2>/dev/null || true
sudo rm -rf /etc/bind /var/lib/bind /var/log/named /run/named 2>/dev/null || true

echo "[10/12] Removing Caddy..."
sudo apt-get purge -y caddy 2>/dev/null || true
sudo rm -rf /etc/caddy* /usr/share/caddy 2>/dev/null || true

echo "[11/12] Removing aaPanel..."
sudo rm -rf /www/server /www/www 2>/dev/null || true
sudo rm -rf /panel 2>/dev/null || true
sudo userdel www 2>/dev/null || true
sudo userdel panel 2>/dev/null || true

echo "[12/12] Cleaning up leftover files..."
sudo rm -rf /tmp/juvia* /tmp/juvia 2>/dev/null || true
sudo rm -rf /var/cache/apt/archives/* 2>/dev/null || true
sudo apt autoremove -y 2>/dev/null || true

echo ""
echo "========================================"
echo "CLEANUP COMPLETE"
echo "Server is now completely clean."
echo "Run install.sh to install Juvia."
echo "========================================"