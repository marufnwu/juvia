#!/bin/bash
set -e

echo "Building Juvia..."

# Build frontend
cd web
npm install
npm run build
cd ..

# Copy frontend to embed directory
mkdir -p internal/web
rm -rf internal/web/dist
cp -r web/dist internal/web/dist

# Build Go binaries
go build -o juvia ./cmd/panel
go build -o juvia-agent ./cmd/agent

echo "Build complete:"
ls -la juvia juvia-agent
