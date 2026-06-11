#!/bin/bash
set -e

UNINSTALL_VERSION="1.0.0"

echo "Juvia Uninstaller v${UNINSTALL_VERSION}"
echo "========================================"

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root."
    exit 1
fi

echo ""
echo "WARNING: This will remove Juvia and all its data."
echo "         Websites, email accounts, databases, and configs will be DELETED."
echo ""
read -p "Type 'yes' to confirm: " -r
if [ "$REPLY" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo "Stopping services..."
systemctl stop juvia 2>/dev/null || true
systemctl stop juvia-agent 2>/dev/null || true
systemctl disable juvia 2>/dev/null || true
systemctl disable juvia-agent 2>/dev/null || true

echo "Removing package..."
dpkg -P juvia 2>/dev/null || apt-get remove -y juvia 2>/dev/null || true

echo "Removing systemd service files..."
rm -f /etc/systemd/system/juvia.service
rm -f /etc/systemd/system/juvia-agent.service
systemctl daemon-reload

echo "Removing binaries..."
rm -f /usr/bin/juvia
rm -f /usr/bin/juvia-agent

echo ""
echo "========================================"
echo "Juvia has been uninstalled."
echo ""
echo "NOTE: User 'juvia', /var/lib/juvia, /var/log/juvia,"
echo "      /var/run/juvia, and /etc/juvia have been removed."
echo ""
echo "If you want to also remove these, run:"
echo "  userdel juvia"
echo "  rm -rf /var/lib/juvia"
echo "  rm -rf /var/log/juvia"
echo "  rm -rf /var/run/juvia"
echo "  rm -rf /etc/juvia"
echo "========================================"