# Hostinger MCP di Cursor

Integrasi MCP memungkinkan Cursor mengelola hosting Hostinger (domain, DNS, deploy) langsung dari chat.

## 1. Buat API Token Hostinger

1. Login [hPanel](https://hpanel.hostinger.com/)
2. **Profil** → **Account Information** → **API**
3. Klik **Generate token** / **New token**
4. Nama: `Cursor-MCP-SIPKEU`
5. **Copy token** — hanya muncul sekali!

Link langsung: https://hpanel.hostinger.com/profile/api

## 2. Set environment variable (Windows)

**PowerShell (sebagai User):**

```powershell
[System.Environment]::SetEnvironmentVariable('HOSTINGER_API_TOKEN', 'PASTE_TOKEN_ANDA_DI_SINI', 'User')
```

Ganti `PASTE_TOKEN_ANDA_DI_SINI` dengan token dari hPanel.

Tutup dan buka ulang Cursor agar variabel terbaca.

## 3. Konfigurasi MCP

Sudah dibuat di:

- Global: `%USERPROFILE%\.cursor\mcp.json`
- Proyek: `.cursor/mcp.json`

Token **tidak** disimpan di file — memakai `${env:HOSTINGER_API_TOKEN}`.

## 4. Restart Cursor

Tutup Cursor sepenuhnya (`File → Exit`), buka lagi.

Verifikasi: **Settings** → **Tools & MCP** → pastikan `hostinger-mcp` status **Connected**.

## 5. Contoh perintah di Cursor

Setelah MCP aktif, Anda bisa minta:

- "Cek DNS domain sakubijak.com"
- "Deploy SIPKEU ke Hostinger"
- "List website di akun Hostinger saya"

## Troubleshooting

| Masalah | Solusi |
|---------|--------|
| MCP disconnected | Pastikan Node.js terinstall (`node -v`) |
| Invalid token | Buat token baru di hPanel, update env variable |
| npx error | Jalankan: `npm install -g hostinger-api-mcp` |
| Token tidak terbaca | Restart Cursor setelah set env variable |

## Keamanan

- Jangan commit token ke Git
- Jangan ganti `${env:HOSTINGER_API_TOKEN}` dengan token langsung di `mcp.json`
- Revoke token lama jika bocor
