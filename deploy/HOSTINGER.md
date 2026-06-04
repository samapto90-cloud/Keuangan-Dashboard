# Deploy SIPKEU di Hostinger (sakubijak.com)

Panduan khusus untuk **Hostinger VPS** + domain **sakubijak.com**.

> **Catatan:** Aplikasi Go ini **tidak bisa** di-host di shared hosting Hostinger (hPanel biasa). Anda membutuhkan **VPS Hostinger** dengan template **Docker**.

---

## Langkah 1 — Beli & siapkan VPS Hostinger

1. Login [hPanel Hostinger](https://hpanel.hostinger.com/)
2. **VPS** → pilih paket (min. KVM 1 / 1 GB RAM)
3. Saat setup OS, pilih template **Docker** (Ubuntu + Docker pre-installed)
4. Catat **IP VPS** dan **VM ID** (angka di URL: `hpanel.hostinger.com/vps/123456/overview` → ID = `123456`)

---

## Langkah 2 — Hubungkan domain sakubijak.com

Di hPanel → **Domains** → **sakubijak.com** → **DNS / DNS Zone**:

| Type | Name | Points to | TTL |
|------|------|-----------|-----|
| A | `@` | IP VPS Anda | 14400 |
| A | `www` | IP VPS Anda | 14400 |

Tunggu propagasi DNS (5–30 menit). Cek:

```bash
dig +short sakubijak.com
```

---

## Langkah 3 — Buka firewall VPS

SSH ke VPS:

```bash
ssh root@IP_VPS_ANDA
```

```bash
ufw allow 22
ufw allow 80
ufw allow 80/tcp
ufw allow 443
ufw allow 443/tcp
ufw --force enable
```

---

## Opsi A — Deploy otomatis via GitHub Actions (disarankan)

### 3a. Buat API Key Hostinger

1. hPanel → **Profil** → [API](https://hpanel.hostinger.com/profile/api)
2. Generate API key → salin

### 3b. Atur secrets di GitHub

Repo → **Settings** → **Secrets and variables** → **Actions**

**Secrets** (tab Secrets):

| Nama | Nilai |
|------|-------|
| `HOSTINGER_API_KEY` | API key dari hPanel |
| `SIPKEU_ADMIN_PASSWORD` | Password admin production |
| `SIPKEU_OPERATOR_PASSWORD` | Password operator production |

**Variables** (tab Variables):

| Nama | Nilai |
|------|-------|
| `HOSTINGER_VM_ID` | ID VPS (contoh: `123456`) |

### 3c. Deploy

Setiap push ke branch `main`, GitHub Actions otomatis deploy ke VPS.

Deploy manual: **Actions** → **Deploy to Hostinger** → **Run workflow**

### 3d. Verifikasi

```bash
curl https://sakubijak.com/health
# {"status":"ok"}
```

---

## Opsi B — Deploy manual via Docker Manager (hPanel)

1. hPanel → **VPS** → **Manage** → **Docker Manager**
2. Klik **Compose from URL**
3. Paste URL repo GitHub:
   ```
   https://github.com/samapto90-cloud/Keuangan-Dashboard
   ```
4. Atau paste link raw `docker-compose.yml`:
   ```
   https://raw.githubusercontent.com/samapto90-cloud/Keuangan-Dashboard/main/docker-compose.yml
   ```
5. Project name: `sipkeu-sakubijak`
6. Tambahkan environment variables (ganti password):
   ```
   ALLOWED_ORIGIN=https://sakubijak.com
   SIPKEU_ADMIN_PASSWORD=password_kuat_admin
   SIPKEU_OPERATOR_PASSWORD=password_kuat_operator
   TZ=Asia/Jakarta
   ```
7. Klik **Deploy**

Caddy otomatis mengurus HTTPS Let's Encrypt untuk `sakubijak.com`.

---

## Opsi C — Deploy manual via SSH

```bash
ssh root@IP_VPS_ANDA

apt update && apt install -y git
git clone https://github.com/samapto90-cloud/Keuangan-Dashboard.git
cd Keuangan-Dashboard

cp deploy/.env.example .env
nano .env   # ganti password

docker compose up -d --build
docker compose ps
docker compose logs -f sipkeu
```

---

## Update aplikasi

**GitHub Actions:** push ke `main` → deploy otomatis.

**SSH manual:**

```bash
cd Keuangan-Dashboard
git pull
docker compose up -d --build
```

---

## Troubleshooting Hostinger

| Masalah | Solusi |
|---------|--------|
| SSL gagal / certificate error | Pastikan DNS A record sudah mengarah ke IP VPS; port 80 terbuka |
| 502 Bad Gateway | `docker compose logs sipkeu` — tunggu healthcheck hijau |
| GitHub Action gagal | Cek `HOSTINGER_API_KEY` dan `HOSTINGER_VM_ID` |
| Repo private | Tambah SSH deploy key VPS ke GitHub repo settings |
| Domain tidak load | Cek DNS di hPanel; flush DNS lokal |

---

## Login aplikasi

- URL: **https://sakubijak.com**
- Admin: username `admin` + password dari `SIPKEU_ADMIN_PASSWORD`
- Operator: username `operator` + password dari `SIPKEU_OPERATOR_PASSWORD`

Lihat juga: [`DEPLOY.md`](../DEPLOY.md) untuk detail teknis umum.
