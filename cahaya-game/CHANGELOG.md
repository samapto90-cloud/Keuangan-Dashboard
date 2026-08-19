# Changelog

## 1.0.0-beta — 2026-08-18

Phase 30/30 final integration overlay.

- Version string 1.0.0-beta on HUD, title, `/health`, endgame
- `/ready` probe + persist backup copy
- Title screen: Mulai / Lanjutkan / Pengaturan / Kredit
- Graphics AUTO + LOW/MEDIUM/HIGH/ULTRA
- Accessibility: subtitle, text size, camera shake, VFX, master/music/SFX volume
- Loading tips and release kicker
- Docs: README, SECURITY, DEPLOYMENT, GAME_DESIGN, API, DATABASE
- No large gameplay features; no production deploy; no real-money systems
- Staging smoke: `/health` `/ready` `/cahaya/`, backup restore roundtrip, 2-player party
- `go run ./cmd/cahaya-smoke` against a local staging process

Phases 1–29 remain integrated (character, combat, quest, economy, guild, raid, 33 siluman, endgame).
