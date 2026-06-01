# SIPKEU — Sistem Informasi Penatausahaan Keuangan

Aplikasi penatausahaan keuangan daerah berbasis web dengan Go backend dan tampilan Bootstrap, lengkap dengan dashboard grafik, manajemen transaksi, dan cetak kwitansi resmi.

## Run & Operate

- `PORT=3000 go run go-app/main.go` — jalankan aplikasi keuangan (port 3000)
- Workflow: **Aplikasi Keuangan** — jalankan via Replit workflow

## Stack

- **Backend:** Go (net/http standard library, embed FS)
- **Frontend:** HTML + Bootstrap 5 + Chart.js 4
- **Storage:** In-memory (mutex-safe, data hilang saat restart)

## Where things live

- `go-app/main.go` — Go server, REST API, in-memory storage, sample data
- `go-app/index.html` — Frontend SPA: dashboard, form, tabel, kwitansi
- `go-app/go.mod` — Go module file

## Architecture decisions

- In-memory storage dengan sync.Mutex untuk thread safety
- `//go:embed index.html` untuk embed HTML ke binary tanpa file server terpisah
- REST API `/api/transactions`, `/api/transactions/{id}`, `/api/dashboard`
- CSS `@media print` untuk menyembunyikan form saat cetak kwitansi
- Terbilang (angka ke kata) diimplementasi di JavaScript client-side

## Product

- **Dashboard:** Statistik total transaksi, nilai, pajak, nilai bersih + 3 grafik (bar kegiatan, doughnut komposisi, line bulanan)
- **Data Transaksi:** Form input 12 kolom + tabel + search + edit + hapus
- **Cetak Kwitansi:** Format kwitansi resmi dengan ruang tanda tangan 3 pejabat
- **Rekapitulasi:** Ringkasan per kegiatan dengan progress bar persentase

## User preferences

- Bahasa Indonesia untuk semua label dan teks UI
- Format rupiah: `Rp X.XXX.XXX`
- Data sample sudah tersedia saat startup

## Gotchas

- Data hilang saat aplikasi di-restart (in-memory). Untuk persistensi gunakan SQLite atau PostgreSQL.
- `go run` butuh beberapa detik untuk compile sebelum port tersedia.
- Port default: 3000 (bisa diubah via env `PORT=xxxx`)
