# Ular Tangga Nusantara Online

**Version:** 1.0.0 — Season One  
**Status:** Production release ready

Game ular tangga multiplayer edukasi dengan tantangan PAI, Matematika, Bahasa Inggris, dan Bahasa Jawa. Pemain dapat bermain online 2–4 orang, naik rank, mengumpulkan XP/coins, dan bersaing di leaderboard.

## Tech Stack

| Layer | Technology |
|---|---|
| Client | Vite, TypeScript, CSS design tokens |
| Server | Go (`go-app/mmo/`) |
| Realtime | WebSocket `/cahaya/ws` |
| Storage | JSON file stores (`ular-*.json`, `cahaya-accounts.json`) |
| Admin | SPA at `/admin` |

## Installation

```bash
cd cahaya-game && npm install
cd ../go-app && go mod download
```

## Development

```bash
# Terminal 1 — backend
cd go-app && go run .

# Terminal 2 — frontend
cd cahaya-game && npm run dev
```

Buka http://localhost:5174/cahaya/

## Environment Variables

| Variable | Description |
|---|---|
| `PORT` | HTTP port (default 3000) |
| `ULAR_OPS_STORE` | Path to ops JSON |
| `ULAR_PROGRESS_STORE` | Path to progress JSON |
| `ULAR_BACKUP_ENABLED` | `1` enables daily backup |
| `ULAR_BACKUP_ROOT` | Backup directory (default `data/ular-backups`) |
| `ULAR_ADMIN_TOKEN` | Optional admin bearer token |
| `ULAR_SUPER_ADMIN` | Bootstrap super-admin username |

## Testing

```bash
cd go-app && go test ./...
cd cahaya-game && npm run test && npm run build
```

## Build

```bash
cd cahaya-game && npm run build
cd ../go-app && go build -o keuangan .
```

Output frontend: `go-app/cahaya-dist/` served at `/cahaya/`.

## Deployment

See [DEPLOYMENT.md](./DEPLOYMENT.md).

## Architecture

```
cahaya-game/          Client SPA (board, online, profile, admin UI)
  src/ui/             Screens, onboarding, landing, settings
  src/game/           Board engine, dice, snakes, ladders
  src/audio/          Procedural audio manager
go-app/mmo/           Authoritative server (WS + REST)
  ular_match.go       Match lifecycle
  ular_question_flow  Question + answer validation
  ular_progress.go    XP, coins, rank settlement
  ular_admin_http.go  Admin APIs + RBAC
  ular_backup.go      Backup/restore automation
```

## Documentation

- [USER_GUIDE.md](./USER_GUIDE.md) — Panduan pemain
- [ADMIN_GUIDE.md](./ADMIN_GUIDE.md) — Panduan admin
- [CHANGELOG.md](./CHANGELOG.md) — Riwayat versi
- [DEPLOYMENT.md](./DEPLOYMENT.md) — Deploy & recovery

## Security

Semua hasil dadu, posisi, jawaban, XP, coins, dan rank dihitung server-side. Lihat Phase 9 security hardening di backend.
