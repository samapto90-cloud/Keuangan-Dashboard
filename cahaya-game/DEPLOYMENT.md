# DEPLOYMENT

Version 1.0.0-beta. Staging dulu. Jangan production deploy otomatis.

## Requirements

- Go 1.22+
- Node 20+
- HTTPS reverse proxy (nginx / Hostinger)

## Environment

Salin `deploy/.env.example`. Wajib: `PORT`, `ALLOWED_ORIGIN`, password admin, `CAHAYA_SOCIAL_STORE`.

`SIPKEU_ENV=staging` atau `production`.

## Build

```bash
cd cahaya-game && npm ci && npm run build
cd ../go-app && go test ./mmo/ -count=1
go build -ldflags "-X main.buildSHA=$(git rev-parse --short HEAD)" -o keuangan .
```

Frontend masuk `go-app/cahaya-dist/`.

## Database / persist

- SIPKEU: file JSON modul keuangan (lihat `storageInfo()`).
- MMO runtime: `data/cahaya-social.json`.
- Backup harian: `deploy/hostinger-web/backup-daily.ps1`.
- Restore: hentikan proses, salin backup ke path persist, start ulang.

## Server

Binary Go menyajikan `/` (keuangan), `/cahaya/` (game), `/cahaya/ws`, `/health`, `/ready`.

## Reverse proxy / HTTPS

Proxy WebSocket `/cahaya/ws` dengan upgrade. Set `ALLOWED_ORIGIN` ke domain staging.

## Monitoring

- `/health` — liveness
- `/ready` — persist path writable
- Log process CPU/RAM, error rate, jumlah koneksi WS

## Rollback

Simpan artifact binary + `cahaya-dist` + file persist. Rollback = ganti binary lama dan restore JSON backup.

## Smoke (staging)

Jangan production deploy otomatis. Staging lokal:

```bash
cd go-app
SIPKEU_ENV=staging CAHAYA_ENV=staging PORT=38991 \
  CAHAYA_SOCIAL_STORE=./data/cahaya-staging-smoke.json \
  ./keuangan
```

Di terminal lain:

```bash
SMOKE_URL=http://127.0.0.1:38991 go run ./cmd/cahaya-smoke
```

Checklist manual: login, play, combat, quest, inventory, logout. Dua pemain untuk party.

Restore persist (setelah backup): `RestoreRuntime(path.bak-YYYYMMDD-HHMMSS)` — file backup tidak dihapus; live di-snapshot dulu.
