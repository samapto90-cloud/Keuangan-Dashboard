#!/bin/bash
# Keepalive SIPKEU: restart Go backend hanya jika health check gagal.
# Dipasang via cron (lihat deploy-node.mjs) supaya proses tidak mati permanen
# saat shared hosting membersihkan proses background.
PORT=8888
APP_DIR="$HOME/sipkeu"
DATA_DIR="$HOME/sipkeu-data"
LOG="$HOME/sipkeu.log"
KLOG="$HOME/sipkeu-keepalive.log"

if curl -sf "http://127.0.0.1:${PORT}/health" >/dev/null 2>&1; then
  exit 0
fi

mkdir -p "$APP_DIR" "$DATA_DIR"
chmod +x "$APP_DIR/keuangan" 2>/dev/null || true

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
export SIPKEU_API_RATE_LIMIT="${SIPKEU_API_RATE_LIMIT:-1200}"
export SIPKEU_LOGIN_RATE_LIMIT="${SIPKEU_LOGIN_RATE_LIMIT:-60}"
export SIPKEU_MAX_SESSIONS="${SIPKEU_MAX_SESSIONS:-10000}"

pkill -x keuangan 2>/dev/null || true
sleep 1
cd "$APP_DIR"
nohup ./keuangan >> "$LOG" 2>&1 &
echo "$(date '+%Y-%m-%d %H:%M:%S') keepalive: SIPKEU di-restart" >> "$KLOG"
