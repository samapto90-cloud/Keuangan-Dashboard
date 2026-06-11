@echo off
setlocal EnableExtensions
title Deploy SIPKEU ke Hostinger (port 8888)
cd /d "%~dp0"

echo.
echo ========================================
echo   Deploy SIPKEU - sakubijak.com:8888
echo ========================================
echo.

where go >nul 2>&1
if errorlevel 1 (
  echo ERROR: Go belum terinstall. Install dari https://go.dev/dl/
  pause
  exit /b 1
)

where node >nul 2>&1
if errorlevel 1 (
  echo ERROR: Node.js belum terinstall. Install dari https://nodejs.org/
  pause
  exit /b 1
)

echo [1/3] Build binary Linux...
cd /d "%~dp0..\..\go-app"
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags="-s -w -X main.buildSHA=manual" -o keuangan-linux-amd64 .
if errorlevel 1 (
  echo Build GAGAL.
  pause
  exit /b 1
)

echo [2/3] Siapkan modul deploy...
cd /d "%~dp0"
call npm install ssh2 --no-save --silent 2>nul

echo [3/3] Upload ke server Hostinger...
echo       Password SSH = dari hPanel ^> Advanced ^> SSH Access
echo.
node deploy-node.mjs
set ERR=%ERRORLEVEL%

echo.
if %ERR% neq 0 (
  echo Deploy GAGAL. Cek password SSH atau koneksi internet.
) else (
  echo Selesai. Buka https://sakubijak.com:8888 lalu Ctrl+F5
)
pause
exit /b %ERR%
