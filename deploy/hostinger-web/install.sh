#!/bin/bash
# Instalasi SIPKEU di Hostinger Web Hosting (SSH)
# Jalankan: bash install.sh

set -euo pipefail

APP_DIR="${HOME}/sipkeu"
DATA_DIR="${HOME}/sipkeu-data"
REPO="https://github.com/samapto90-cloud/Keuangan-Dashboard.git"
PORT="${PORT:-8888}"

echo "==> SIPKEU Hostinger installer"
echo "    App dir : $APP_DIR"
echo "    Data dir: $DATA_DIR"
echo "    Port    : $PORT"

mkdir -p "$DATA_DIR"

if [ ! -d "$APP_DIR/.git" ]; then
  echo "==> Clone repository..."
  git clone "$REPO" "$APP_DIR"
else
  echo "==> Update repository..."
  cd "$APP_DIR"
  git pull origin main
fi

cd "$APP_DIR/go-app"

if ! command -v go >/dev/null 2>&1; then
  echo "ERROR: Go tidak ditemukan di server."
  echo "       Hostinger web hosting tidak mendukung build Go secara native."
  echo "       Upload binary Linux hasil build: keuangan-linux-amd64"
  echo "       Lalu jalankan: chmod +x keuangan && ./keuangan"
  exit 1
fi

echo "==> Build aplikasi..."
go build -ldflags="-s -w" -o keuangan .

if [ ! -f "$APP_DIR/deploy/hostinger-web/.env" ]; then
  cp "$APP_DIR/deploy/hostinger-web/.env.example" "$APP_DIR/deploy/hostinger-web/.env"
  echo "==> Buat file .env — edit password sebelum production:"
  echo "    nano $APP_DIR/deploy/hostinger-web/.env"
fi

set -a
# shellcheck disable=SC1091
source "$APP_DIR/deploy/hostinger-web/.env"
set +a

export DATA_DIR="$DATA_DIR"
export PORT="$PORT"
export ALLOWED_ORIGIN="${ALLOWED_ORIGIN:-https://sakubijak.com}"

echo "==> Stop proses lama (jika ada)..."
pkill -f "${APP_DIR}/go-app/keuangan" 2>/dev/null || true
sleep 1

echo "==> Jalankan SIPKEU..."
cd "$APP_DIR/go-app"
nohup ./keuangan >> "${HOME}/sipkeu.log" 2>&1 &
sleep 2

if curl -sf "http://127.0.0.1:${PORT}/health" >/dev/null; then
  echo "==> SIPKEU berjalan! Health check OK."
  echo "    URL: https://sakubijak.com:${PORT}"
  echo "    Log: tail -f ${HOME}/sipkeu.log"
  echo "    Data tersimpan di: ${DATA_DIR}"
else
  echo "ERROR: Health check gagal. Cek log:"
  tail -20 "${HOME}/sipkeu.log"
  exit 1
fi
