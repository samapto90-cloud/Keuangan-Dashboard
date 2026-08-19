# Petualangan Menuju Cahaya

Game 3D MMORPG web original (Vite + TypeScript + Three.js) dengan backend Go yang sama dengan SIPKEU.

**Version:** 1.0.0-beta  
**Phase:** 30/30  
**Status:** Ready for external QA — bukan rilis produksi final.

## Overview

Pemain menjelajahi delapan wilayah, menghadapi 33 siluman original, menyelesaikan quest pendidikan, crafting, party, guild, dungeon, raid, world event, dan endgame. NPC utama berbicara Bahasa Jawa; UI dan soal pendidikan memakai Bahasa Indonesia.

Transformasi memakai desain original: Wujud Asal, Wujud Cahaya, Wujud Kilat, Wujud Agung, Wujud Fajar.

## Architecture

- Client: `cahaya-game/` (Three.js, HUD, interpolation, prediction)
- Server: `go-app/mmo/` (authoritative combat, economy, quest, persist JSON)
- Serve path: `/cahaya/` dari binary Go; WebSocket `/cahaya/ws`

## Installation & development

```bash
cd cahaya-game
npm install
npm run dev
```

Buka http://localhost:5174/cahaya/  
Dev server mem-proxy `/cahaya/ws` ke backend di `127.0.0.1:8888`.

Backend:

```bash
cd go-app
go test ./mmo/ -count=1
go build -o keuangan.exe .
```

## Environment

| Variable | Fungsi |
|---|---|
| `PORT` | HTTP port (default 3000) |
| `CAHAYA_SOCIAL_STORE` | Path JSON runtime (guild/social/market) |
| `SIPKEU_ENV` / `CAHAYA_ENV` | `production` menandai rilis |
| `SIPKEU_ADMIN_PASSWORD` | Wajib diubah dari default |
| `ALLOWED_ORIGIN` | CORS |

Rahasia hanya lewat environment. Jangan hardcode.

## Database

MMO memakai `MemoryQuestRepo` untuk progress pemain dalam proses, plus file JSON runtime `data/cahaya-social.json` (atau `CAHAYA_SOCIAL_STORE`). Backup: salin file itu. Lihat `docs/DATABASE.md`.

## Testing

```bash
cd go-app && go test ./mmo/ -count=1
cd cahaya-game && npm run build
```

## Build

```bash
cd cahaya-game && npm run build
```

Output: `go-app/cahaya-dist/` dilayani di `/cahaya/`.

## Deployment

Lihat `DEPLOYMENT.md`. Staging dulu. Jangan production deploy otomatis sebelum QA eksternal.

## Security

Lihat `SECURITY.md`. Semua damage, gold, item, quest claim, dan leaderboard dihitung server.

## Troubleshooting

- WS gagal: pastikan backend hidup dan proxy `/cahaya/ws`.
- Progress hilang setelah restart: set `CAHAYA_SOCIAL_STORE` dan pastikan proses tidak di-kill sebelum `SaveRuntime`.
- Graphics berat: pilih AUTO atau LOW di Pengaturan.

## Kontrol

Desktop: WASD, Space dodge, Shift block, E interact, Q, 1–4 skill, mouse kamera/serang.  
Mobile: virtual joystick + tombol serang/skill/lompat/dodge/target.
