# Jalankan SIPKEU lokal di Windows 10/11 (tanpa error Origin)
# Usage: .\run-local.ps1

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$env:PORT = "3000"
$env:ALLOWED_ORIGIN = ""
$env:SIPKEU_ALLOW_LOCALHOST = "1"
$env:DATA_DIR = Join-Path $PSScriptRoot "sipkeu-data-local"

Write-Host "SIPKEU lokal: http://localhost:3000" -ForegroundColor Green
Write-Host "Login admin: admin / admin2026 (default)" -ForegroundColor Yellow
Write-Host "Data: $env:DATA_DIR" -ForegroundColor Gray

go run .
