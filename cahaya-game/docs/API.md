# API

Transport: WebSocket `/cahaya/ws` (pesan bertipe string di `go-app/mmo/protocol.go`). HTTP game static: `/cahaya/`. Ops: `GET /health`, `GET /ready`.

## Authentication

HTTP:

- `POST /cahaya/api/register` username, email, password, confirmPassword
- `POST /cahaya/api/login` username atau email + password
- `POST /cahaya/api/logout` Bearer token
- `POST /cahaya/api/reset-password` username + email + password baru
- `GET /cahaya/api/profile` Bearer token — data milik sesi itu saja

WebSocket `/cahaya/ws`: pesan pertama `AUTH` dengan `token` dari login. Tanpa token valid, `AUTH_FAIL`.

Klien tidak mengirim damage/gold. Progress tidak memakai localStorage sebagai sumber kebenaran.

## Character / combat / inventory / quest

`GET_PROGRESSION`, `REQUEST_TRANSFORMATION` — server menolak `SET_LEVEL` / `SET_DAMAGE`.  
`ATTACK` / skills — cooldown dan damage server.  
`GET_INVENTORY`, `EQUIP_ITEM`, `PICKUP_ITEM` — ownership + jarak.  
`QUEST_ACCEPT`, `QUEST_CLAIM`, `EDUCATION_ANSWER` — state machine server.

## Crafting / economy / guild / party

`GATHER`, `CRAFT`, `NPC_SHOP_BUY/SELL`, `GUILD_*`, `PARTY_*`, `TRADE_*` — transaksi atomik, rate limit.

## Dungeon / raid / event

`DUNGEON_ENTER`, `CLAIM_LOOT`, `JOIN_WORLD_EVENT`, `CLAIM_EVENT_REWARD`, `CLAIM_WORLD_BOSS` — syarat party/level/partisipasi server.

Cheats (`GIVE_GOLD`, `TELEPORT`, `COMPLETE_DUNGEON`, …) ditolak `server_authoritative`.
