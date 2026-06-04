# Deploy SIPKEU ke sakubijak.com

Panduan menjalankan aplikasi **SIPKEU** (Keuangan Dashboard) di domain **https://sakubijak.com**.

## Prasyarat

| Item | Keterangan |
|------|------------|
| VPS / cloud server | Ubuntu 22.04+ atau Debian 12+, min. 1 GB RAM |
| Domain | `sakubijak.com` sudah dimiliki |
| DNS | Record **A** `@` → IP server, **A** `www` → IP server (atau CNAME `www` → `@`) |
| Port terbuka | **80** dan **443** (firewall) |

Frontend memakai API relatif (`/data`), jadi **tidak perlu** mengubah kode untuk domain baru.

---

## Opsi A — Docker + Caddy (disarankan)

Caddy otomatis mengurus sertifikat HTTPS (Let's Encrypt).

### 1. Clone repo di server

```bash
git clone https://github.com/samapto90-cloud/Keuangan-Dashboard.git
cd Keuangan-Dashboard
```

### 2. Siapkan environment

```bash
cp deploy/.env.example .env
nano .env
```

Wajib ganti:

- `SIPKEU_ADMIN_PASSWORD`
- `SIPKEU_OPERATOR_PASSWORD`
- `ALLOWED_ORIGIN=https://sakubijak.com`

> **Hostinger VPS:** lihat panduan lengkap di [`deploy/HOSTINGER.md`](deploy/HOSTINGER.md) — termasuk deploy otomatis via GitHub Actions.

### 3. Pastikan DNS sudah mengarah ke server

```bash
dig +short sakubijak.com
# harus menampilkan IP server Anda
```

### 4. Jalankan

```bash
docker compose up -d --build
```

### 5. Verifikasi

```bash
curl -s https://sakubijak.com/health
# {"status":"ok"}
```

Buka **https://sakubijak.com** di browser → portal SIPKEU.

### Perintah berguna

```bash
docker compose logs -f sipkeu    # log aplikasi
docker compose restart sipkeu  # restart setelah update
docker compose pull && docker compose up -d --build  # update versi
```

---

## Opsi B — Binary + systemd + Nginx

Tanpa Docker; cocok jika server sudah punya Nginx.

### 1. Build binary di server

```bash
sudo apt update && sudo apt install -y golang nginx certbot python3-certbot-nginx
git clone https://github.com/samapto90-cloud/Keuangan-Dashboard.git
cd Keuangan-Dashboard/go-app
go build -ldflags="-s -w" -o keuangan .
```

### 2. Install ke `/opt/sipkeu`

```bash
sudo useradd -r -s /bin/false sipkeu || true
sudo mkdir -p /opt/sipkeu
sudo cp keuangan Anggaran.xlsx /opt/sipkeu/
sudo cp ../deploy/.env.example /opt/sipkeu/.env
sudo nano /opt/sipkeu/.env   # ganti password & ALLOWED_ORIGIN
sudo chown -R sipkeu:sipkeu /opt/sipkeu
```

### 3. Aktifkan systemd

```bash
sudo cp deploy/systemd/sipkeu.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now sipkeu
curl http://127.0.0.1:3000/health
```

### 4. Nginx + SSL

```bash
sudo cp deploy/nginx/sakubijak.com.conf /etc/nginx/sites-available/sakubijak.com
sudo ln -s /etc/nginx/sites-available/sakubijak.com /etc/nginx/sites-enabled/
sudo certbot --nginx -d sakubijak.com -d www.sakubijak.com
sudo nginx -t && sudo systemctl reload nginx
```

---

## Variabel environment

| Variabel | Default | Keterangan |
|----------|---------|------------|
| `PORT` | `3000` | Port HTTP internal |
| `ANGGARAN_FILE` | `Anggaran.xlsx` | Path file pagu RAK |
| `ALLOWED_ORIGIN` | `*` (kosong) | Set `https://sakubijak.com` di production |
| `SIPKEU_ADMIN_USER` | `admin` | Username admin |
| `SIPKEU_ADMIN_PASSWORD` | `admin2026` | **Ganti di production** |
| `SIPKEU_OPERATOR_USER` | `operator` | Username operator |
| `SIPKEU_OPERATOR_PASSWORD` | `operator2026` | **Ganti di production** |
| `TZ` | — | Contoh: `Asia/Jakarta` |

---

## Update file anggaran

Ganti `go-app/Anggaran.xlsx`, lalu:

- **Docker:** rebuild image (`docker compose up -d --build`) atau mount volume ke `/app/Anggaran.xlsx`
- **systemd:** copy ke `/opt/sipkeu/Anggaran.xlsx` dan restart `sipkeu`

Admin juga bisa impor anggaran baru lewat menu **Impor Anggaran** di aplikasi.

---

## Catatan penting

1. **Data in-memory** — transaksi hilang saat restart. Untuk production jangka panjang, pertimbangkan persistensi (SQLite/PostgreSQL).
2. **Firewall** — buka 80/443 ke publik; **jangan** expose port 3000 langsung ke internet.
3. **Backup** — export transaksi secara berkala lewat fitur export di aplikasi.
4. **CDN** — aplikasi memuat Bootstrap, Chart.js, dll. dari CDN; server perlu akses internet.

---

## Troubleshooting

| Gejala | Solusi |
|--------|--------|
| Sertifikat SSL gagal | Pastikan DNS sudah propagate; port 80 tidak diblokir |
| 502 Bad Gateway | Cek `docker compose logs sipkeu` atau `journalctl -u sipkeu` |
| Login gagal | Pastikan password di `.env` benar; restart service |
| Pagu kosong | Pastikan `Anggaran.xlsx` ada dan `ANGGARAN_FILE` benar |

---

## Struktur file deploy

```
deploy/
  Caddyfile              # Reverse proxy + HTTPS (Docker)
  .env.example           # Template environment
  nginx/sakubijak.com.conf
  systemd/sipkeu.service
docker-compose.yml
go-app/Dockerfile
```
