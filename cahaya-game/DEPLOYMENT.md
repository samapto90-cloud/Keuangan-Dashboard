# Deployment — Ular Tangga Nusantara Online v1.0.0

## Prerequisites

- Go 1.22+
- Node.js 20+ (build only)
- Linux/Windows server with persistent disk for JSON stores

## Environment Setup

```bash
export PORT=3000
export ULAR_BACKUP_ENABLED=1
export ULAR_BACKUP_ROOT=/var/data/ular-backups
# Optional:
# export ULAR_ADMIN_TOKEN=<strong-random-token>
# export ULAR_SUPER_ADMIN=adminusername
```

**Never commit secrets.** Use environment or secret manager.

## Build

```bash
cd cahaya-game && npm ci && npm run build
cd ../go-app && go test ./... && go build -o keuangan .
```

## Deploy

1. Copy binary + `go-app/cahaya-dist/` embedded (via `go build` with embed)
2. Ensure `data/` directory writable for JSON stores
3. Run behind reverse proxy (HTTPS required for PWA install)
4. WebSocket upgrade path: `/cahaya/ws`

## Database / Storage

JSON files (default under `data/`):

- `cahaya-accounts.json`
- `ular-progress.json`
- `ular-ops.json`
- `ular-social.json`
- `ular-matches.json`
- `ular-attempts.json`

## Backup

- Automatic: `ULAR_BACKUP_ENABLED=1` (daily, retention 7d/4w/3m)
- Manual: copy entire `data/` directory
- Restore test: `POST /cahaya/api/admin/backup/restore-test` with admin auth

## Rollback

1. Stop process
2. Restore `data/` from last known-good backup
3. Deploy previous binary version
4. Start process
5. Verify `/health`, `/ready`, `/cahaya/`

## Health Checks

- `GET /health` — liveness
- `GET /ready` — readiness
- `GET /cahaya/api/admin/status` — game subsystem status (auth required)

## Zero-Downtime Notes

- Graceful shutdown: finish in-flight WS before kill
- Config changes affect **future matches only** (frozen rewards at match start)

## Staging Checklist

- [ ] Register / login / onboarding
- [ ] 2-player match end-to-end
- [ ] Ranked settlement
- [ ] Admin login + question CRUD
- [ ] Backup + restore test
- [ ] PWA install on mobile browser
