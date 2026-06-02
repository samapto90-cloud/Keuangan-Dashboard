---
name: SIPKEU Stack & Routing
description: Go backend + Bootstrap 5 SPA keuangan daerah — konvensi penting untuk pengembangan lanjutan
---

## Rules

- Server Go di `go-app/main.go`, HTML di `go-app/index.html` di-embed via `//go:embed index.html`
- API prefix: `/data/transactions`, `/data/dashboard` — bukan `/api/`
- Route `/data/transactions/import` harus didaftarkan SEBELUM `/data/transactions/` (lebih spesifik dulu)
- In-memory store dengan `sync.Mutex`, data hilang saat restart
- Jalankan: `PORT=3000 go run go-app/main.go`

**Why:** Pernah ada bug jika /import didaftarkan setelah /transactions/ maka Go mengarahkan ke handler by-ID.

## Struct JSON keys (snake_case dari Go json tag)
- `jenis_pajak`, `ntpn`, `kode_billing`, `ntb` — field pajak baru
- `no_bpk`, `no_bast`, `no_kontrak` — 3 field nomor dokumen terpisah
- `pekerjaan` (dropdown cascading) vs `uraian` (free-text textarea)

## Frontend Conventions
- TARIF_PAJAK map di JS: PPh 21 5%, 2.5%; PPh 23 2%; PPh 22 1.5%; PPh Ps.4(2) 10%
- EXPORT_COLS array 19 kolom — dipakai untuk ekspor xlsx DAN mapping impor
- SheetJS CDN: `xlsx@0.18.5` di `cdn.jsdelivr.net/npm/xlsx@0.18.5/dist/xlsx.full.min.js`
- Kondisional NTPN/KodeBilling/NTB: tampil hanya jika jenis_pajak dipilih (`row-ntpn`, `row-kbilling`, `row-ntb`)

**How to apply:** Setiap tambah field baru ke Transaction struct Go, update EXPORT_COLS di JS agar impor/ekspor otomatis ikut.
