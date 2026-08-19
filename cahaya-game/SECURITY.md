# SECURITY

Petualangan Menuju Cahaya — 1.0.0-beta

## Authentication

- Login SIPKEU: session cookie, password bcrypt, rate limit login.
- MMO: register/login HTTP, password bcrypt, sesi token. WebSocket `AUTH` wajib token. Data hanya untuk `PlayerID` sesi.

## Authorization

- Admin HTTP: `requireAuth` + role portal/settings.
- MMO: party/guild role dicek server (invite, kick, leader, storage).

## Anti-cheat

Server authoritative. Client tidak boleh:

- damage arbitrary
- teleport
- give item / gold
- ubah stats / level
- complete quest / claim reward tanpa syarat

Deteksi: input rate, kecepatan gerak, jarak pickup, cooldown, ownership.

## Rate limit

HTTP: `withAPIRateLimit`, IP shield.  
MMO: combat, gather, craft, buy, sell, chat, login.

## Validation

Posisi, distance, cooldown, resource, ownership, level, damage dihitung ulang di server.

## Secret management

- Password hashed (bcrypt).
- Set `SIPKEU_ADMIN_PASSWORD` dan jangan pakai default.
- Jangan commit `.env` berisi kredensial.

## Backup

Salin `CAHAYA_SOCIAL_STORE` / `data/cahaya-social.json`.  
Salin juga `data/cahaya-accounts.json` dan folder `data/cahaya-players/`.
`BackupRuntime()` menulis salinan `.bak-YYYYMMDD-HHMMSS` tanpa menimpa file hidup.  
`RestoreRuntime(backup)` men-snapshot live dulu, lalu menimpa store dari backup. File `.bak-*` tidak dihapus.

## Production

- Disable debug overlay (build production tidak mengaktifkan `DEV`).
- `/health` dan `/ready` tanpa stack trace.
- Jangan expose email, token, atau kredensial database di error klien.
