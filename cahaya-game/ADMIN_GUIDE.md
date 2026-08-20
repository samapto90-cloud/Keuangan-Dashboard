# Panduan Admin — Ular Tangga Nusantara

## Login Admin

1. Buka `/admin` atau `/cahaya/admin`
2. Login dengan akun yang memiliki role **MODERATOR**, **ADMIN**, atau **SUPER_ADMIN**
3. Alternatif: header `X-Ular-Admin` / `Authorization: Bearer` dengan `ULAR_ADMIN_TOKEN`

## Peran

| Role | Akses |
|---|---|
| MODERATOR | Player view, match view, reports, audit |
| ADMIN | + questions, config view, sanctions, achievements |
| SUPER_ADMIN | + config edit, admin manage, seasons, rank edit |

## Player Management

- Daftar pemain, filter status/rank
- Detail: profil, history, rank history, sanctions
- **Sanction**: warning, mute, temp/permanent ban

## Question Management

- CRUD soal per mata pelajaran
- Import CSV
- Validasi bank soal

## Report Moderation

- Laporan player & feedback pemain
- Resolve / dismiss dengan catatan

## Config & Seasons

- Reward XP/coins, RR, daily cycle
- Feature flags (ranked, chat, daily reward)
- Season management

## Audit Log

Semua aksi admin tercatat. Feedback pemain masuk sebagai `PLAYER_FEEDBACK:*`.

## Backup

- Env `ULAR_BACKUP_ENABLED=1` untuk backup harian otomatis
- Admin endpoint: `POST /cahaya/api/admin/backup/restore-test` (SUPER_ADMIN / CONFIG_EDIT)
