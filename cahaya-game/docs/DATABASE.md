# DATABASE

MMO prototype tidak memakai skema SQL terpisah. Persist logis:

## Player log (QuestRepo)

`PlayerLog` di `quest.go`: quests, flags, gold, inventory refs, titles, achievements, season, raid lockout, story choices, professions. Index logis: playerId, questId, recipeId.

Progress pemain disimpan atomik per `PlayerID` di `data/cahaya-players/<id>.json` (quest + equipment + checkpoint). `MemoryQuestRepo` hanya untuk unit test. Restart proses memuat ulang file tersebut.

Akun: `data/cahaya-accounts.json` — username, email, **password bcrypt**, sesi token. Tidak ada password plain text.

## Runtime JSON

`data/cahaya-social.json` atau `CAHAYA_SOCIAL_STORE`:

guilds, social, audit, reports, transactions seen, pvp, community event, horizon leaderboard, market, housing, economy log, chat log, dungeon board.

Atomic write: file `.tmp` lalu rename.

## Backup / restore

`BackupRuntime()` menyalin ke `*.bak-YYYYMMDD-HHMMSS`.  
Restore: stop server, ganti file live dengan backup, start. Jangan hapus migration/backup.

## Relasi logis

player → guild, party, raid instance, event participation, boss contribution. Unique: playerId, instanceId, transactionId. Foreign key logis divalidasi di kode, bukan RDBMS.
