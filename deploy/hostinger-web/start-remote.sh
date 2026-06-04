#!/bin/bash
set -euo pipefail

APP_DIR="$HOME/sipkeu"
DATA_DIR="$HOME/sipkeu-data"
LOG="$HOME/sipkeu.log"
PORT=8888

mkdir -p "$APP_DIR" "$DATA_DIR"
chmod +x "$APP_DIR/keuangan"

export PORT="$PORT"
export DATA_DIR="$DATA_DIR"
export ANGGARAN_FILE="$APP_DIR/Anggaran.xlsx"
export ALLOWED_ORIGIN="https://sakubijak.com"
export TZ="Asia/Jakarta"

if [ -f "$HOME/hostinger-web/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  source "$HOME/hostinger-web/.env"
  set +a
fi

pkill -f "$APP_DIR/keuangan" 2>/dev/null || true
pkill -x keuangan 2>/dev/null || true
sleep 2

cd "$APP_DIR"
nohup ./keuangan >> "$LOG" 2>&1 &
sleep 2

if curl -sf "http://127.0.0.1:${PORT}/health" >/dev/null; then
  echo "OK: SIPKEU running on port ${PORT}"
  echo "Data: ${DATA_DIR}"
else
  echo "FAIL: health check"
  tail -30 "$LOG"
  exit 1
fi
