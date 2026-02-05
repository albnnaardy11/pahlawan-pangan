# Pahlawan Pangan Staging Deployment Script (v2.0-Alpha)
# This script ensures the system is production-ready before pushing.

$ErrorActionPreference = "Stop"

Write-Host "--- 🛡️ Starting Pahlawan Pangan Deployment (v2.0-Alpha) ---" -ForegroundColor Cyan

# 1. Quality Control
Write-Host "[1/4] Running Linting & Formatting..." -ForegroundColor Yellow
go fmt ./...
go vet ./...
# golangci-lint run  # Uncomment in real CI environment

# 2. Testing Logic
Write-Host "[2/4] Executing Mission-Critical Unit Tests..." -ForegroundColor Yellow
go test -v ./internal/matching/... ./internal/api/... -race

# 3. Documentation Sync
Write-Host "[3/4] Syncing Documentation Portal..." -ForegroundColor Yellow
Set-Location portal
npm run gen-api-docs all
npm run build
Set-Location ..

# 4. Binary Build
Write-Host "[4/4] Building Production Binaries..." -ForegroundColor Yellow
if (!(Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" }
go build -o bin/pahlawan-server cmd/server/main.go

Write-Host "`n--- ✅ DEPLOYMENT READY: Pahlawan Pangan v2.0-Alpha ---" -ForegroundColor Green
Write-Host "Platform is stable. Documentation is built. Ready for Staging." -ForegroundColor Green
