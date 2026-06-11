# Juvia Build Script for Windows
# Build frontend and Go binaries on Windows

$ErrorActionPreference = "Stop"

Write-Host "Building Juvia..." -ForegroundColor Cyan

$ProjectRoot = Split-Path -Parent $PSScriptRoot

# Build frontend
Write-Host "Building frontend..." -ForegroundColor Yellow
Push-Location "$ProjectRoot\web"
try {
    npm install
    npm run build
} finally {
    Pop-Location
}

# Copy frontend to embed directory
Write-Host "Copying frontend to embed directory..." -ForegroundColor Yellow
$webDistDir = "$ProjectRoot\web\dist"
$embedDir = "$ProjectRoot\internal\web\dist"

if (Test-Path $embedDir) {
    Remove-Item -Recurse -Force $embedDir
}
New-Item -ItemType Directory -Path $embedDir -Force | Out-Null
Copy-Item -Recurse -Path "$webDistDir\*" -Destination $embedDir

# Build Go binaries
Write-Host "Building Go binaries..." -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    go build -o juvia.exe ./cmd/panel
    go build -o juvia-agent.exe ./cmd/agent
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "Build complete:" -ForegroundColor Green
Get-ChildItem -Path $ProjectRoot -Filter "*.exe" | Format-Table Name, Length -AutoSize
Write-Host ""
Write-Host "Binaries: juvia.exe, juvia-agent.exe" -ForegroundColor Cyan