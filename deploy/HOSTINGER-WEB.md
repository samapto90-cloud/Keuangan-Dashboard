# Deploy SIPKEU di Hostinger Web Hosting (sakubijak.com)

Panduan untuk **Hostinger Web Hosting** dengan SSH — sesuai panel Anda:

| Item | Nilai |
|------|-------|
| IP | `145.79.14.155` |
| Port SSH | `65002` |
| Username | `u657726332` |
| Perintah SSH | `ssh -p 65002 u657726332@145.79.14.155` |
| URL aplikasi | `https://sakubijak.com:8888` |

---

## Fitur penyimpanan data

Data sekarang **disimpan permanen** di folder server (`DATA_DIR`):

| File | Isi |
|------|-----|
| `sekretariat.json` | Transaksi + pengaturan portal Sekretariat |
| `paud.json` | Transaksi + pengaturan portal PAUD |
| `kas-belanja.json` | Data realisasi anggaran kas |

Data **tidak hilang** saat aplikasi di-restart.

Default lokasi di Hostinger: `~/sipkeu-data/`

---

## Langkah 1 — Aktifkan SSH (sudah aktif ✓)

hPanel → **Website** → **sakubijak.com** → **Tingkat Lanjut** → **SSH Access**

---

## Langkah 2 — Build binary Linux (dari komputer Anda)

Di folder proyek:

```powershell
cd d:\Keuangan-Dashboard\Keuangan-Dashboard\go-app
$env:GOOS="linux"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -ldflags="-s -w" -o keuangan .
```

---

## Langkah 3 — Upload ke Hostinger

```powershell
scp -P 65002 keuangan u657726332@145.79.14.155:~/keuangan
scp -P 65002 Anggaran.xlsx u657726332@145.79.14.155:~/Anggaran.xlsx
scp -P 65002 -r deploy/hostinger-web u657726332@145.79.14.155:~/hostinger-web
```

---

## Langkah 4 — Jalankan di server (SSH)

```bash
ssh -p 65002 u657726332@145.79.14.155
```

```bash
mkdir -p ~/sipkeu-data ~/sipkeu
mv ~/keuangan ~/sipkeu/
mv ~/Anggaran.xlsx ~/sipkeu/
chmod +x ~/sipkeu/keuangan

# Buat file environment
cp ~/hostinger-web/.env.example ~/hostinger-web/.env
nano ~/hostinger-web/.env
# Ganti SIPKEU_ADMIN_PASSWORD dan SIPKEU_OPERATOR_PASSWORD

export PORT=8888
export DATA_DIR=$HOME/sipkeu-data
export ANGGARAN_FILE=$HOME/sipkeu/Anggaran.xlsx
export ALLOWED_ORIGIN=https://sakubijak.com
export TZ=Asia/Jakarta
# muat password dari .env:
set -a && source ~/hostinger-web/.env && set +a

# Stop proses lama
pkill -f sipkeu/keuangan 2>/dev/null || true

# Jalankan
cd ~/sipkeu
nohup ./keuangan >> ~/sipkeu.log 2>&1 &
sleep 2
curl http://127.0.0.1:8888/health
```

Harus muncul: `{"status":"ok"}`

Buka browser: **https://sakubijak.com:8888**

---

## Langkah 5 — Sinkronisasi dengan GitHub (update otomatis)

### Opsi A — GitHub Actions ke port 8888 (disarankan)

Setiap push ke `main`, workflow **Deploy to Hostinger Web (8888)** meng-upload binary ke server SSH.

**Sekali saja:** GitHub → Settings → Secrets → Actions → tambahkan:

| Secret | Nilai |
|--------|-------|
| `HOSTINGER_SSH_PASSWORD` | Password SSH dari hPanel → SSH Access |

Workflow memakai `deploy/hostinger-web/deploy-node.mjs` (sama seperti deploy manual).

> **Penting:** Workflow **Deploy to Hostinger** (Docker VPS) **tidak** meng-update `https://sakubijak.com:8888`. Production SIPKEU di port 8888 memakai binary di `~/sipkeu/keuangan` pada web hosting SSH.

Verifikasi setelah deploy: buka `https://sakubijak.com:8888/health` — harus ada `"build":"abc1234"`.

### Opsi B — Git pull + rebuild (jika Go tersedia di server)

```bash
cd ~/sipkeu-src   # clone repo sekali
git clone https://github.com/samapto90-cloud/Keuangan-Dashboard.git ~/sipkeu-src
bash ~/sipkeu-src/deploy/hostinger-web/update.sh
```

### Opsi C — Upload binary baru (disarankan jika Actions belum ada secret)

**Cara termudah (Windows):** double-click file:

`deploy\hostinger-web\deploy.bat`

Atau di **Command Prompt** (bukan PowerShell):

```cmd
cd d:\Keuangan-Dashboard\Keuangan-Dashboard\deploy\hostinger-web
deploy.bat
```

Script akan build binary, minta **password SSH Hostinger**, upload, dan restart otomatis.

> Jika `deploy.ps1` error *running scripts is disabled*, jangan pakai PowerShell langsung — gunakan **deploy.bat** saja.

Setiap ada update di GitHub (manual):

1. Build ulang binary Linux di komputer (Langkah 2)
2. Upload: `scp -P 65002 keuangan u657726332@145.79.14.155:~/sipkeu/`
3. Restart:
   ```bash
   ssh -p 65002 u657726332@145.79.14.155 "pkill -f sipkeu/keuangan; cd ~/sipkeu && nohup ./keuangan >> ~/sipkeu.log 2>&1 &"
   ```

### Opsi C — Hostinger Git (hPanel)

hPanel → **Website** → **GIT** → tambah repository:

```
https://github.com/samapto90-cloud/Keuangan-Dashboard.git
```

Branch: `main`

> Git di hPanel hanya menarik source code. Anda tetap perlu build/upload binary `keuangan` secara manual karena shared hosting tidak menjalankan Go secara otomatis.

---

## Backup data

```bash
ssh -p 65002 u657726332@145.79.14.155
tar -czf sipkeu-backup-$(date +%Y%m%d).tar.gz ~/sipkeu-data/
```

Download backup ke komputer:

```powershell
scp -P 65002 u657726332@145.79.14.155:~/sipkeu-backup-*.tar.gz .
```

---

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| Tidak bisa SSH | Pastikan SSH status ACTIVE di hPanel |
| Port 8888 tidak bisa diakses | Cek firewall Hostinger; hubungi support buka port 8888 |
| Data hilang setelah restart | Pastikan `DATA_DIR=~/sipkeu-data` diset sebelum jalankan |
| Login gagal | Cek password di `~/hostinger-web/.env` |
| CORS error | Set `ALLOWED_ORIGIN=https://sakubijak.com` |

---

## Login default (ganti segera!)

| Role | Username | Password (.env) |
|------|----------|-----------------|
| Admin | admin | `SIPKEU_ADMIN_PASSWORD` |
| Operator | operator | `SIPKEU_OPERATOR_PASSWORD` |
