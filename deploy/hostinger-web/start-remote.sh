#!/bin/bash
set -eu

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

export SIPKEU_TRUST_PROXY="${SIPKEU_TRUST_PROXY:-1}"
export SIPKEU_API_RATE_LIMIT="${SIPKEU_API_RATE_LIMIT:-3000}"
export SIPKEU_LOGIN_RATE_LIMIT="${SIPKEU_LOGIN_RATE_LIMIT:-180}"
export SIPKEU_IP_RATE_LIMIT="${SIPKEU_IP_RATE_LIMIT:-1500}"
export SIPKEU_ASSET_RATE_LIMIT="${SIPKEU_ASSET_RATE_LIMIT:-4000}"
export SIPKEU_PORTAL_STATUS_RATE="${SIPKEU_PORTAL_STATUS_RATE:-360}"
export SIPKEU_MAX_CONN_PER_IP="${SIPKEU_MAX_CONN_PER_IP:-72}"
export SIPKEU_GLOBAL_MAX_CONN="${SIPKEU_GLOBAL_MAX_CONN:-22000}"
export SIPKEU_MAX_SESSIONS="${SIPKEU_MAX_SESSIONS:-25000}"

pkill -f "$APP_DIR/keuangan" 2>/dev/null || true
pkill -x keuangan 2>/dev/null || true
sleep 2

cd "$APP_DIR"
nohup ./keuangan >> "$LOG" 2>&1 &

for i in 1 2 3 4 5; do
  sleep 2
  if curl -sf "http://127.0.0.1:${PORT}/health" >/dev/null; then
    echo "OK: SIPKEU running on port ${PORT}"
    echo "Data: ${DATA_DIR}"
    exit 0
  fi
done

echo "FAIL: health check"
tail -40 "$LOG"
exit 1
