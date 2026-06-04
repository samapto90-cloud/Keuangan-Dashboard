# Deploy SIPKEU ke Hostinger Web Hosting
# Jalankan di PowerShell: .\deploy\hostinger-web\deploy.ps1
# Akan meminta password SSH Hostinger (dari hPanel → SSH Access)

$ErrorActionPreference = "Stop"

$HostingerHost = "145.79.14.155"
$HostingerPort = "65002"
$HostingerUser = "u657726332"
$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$GoApp = Join-Path $Root "go-app"
$Binary = Join-Path $GoApp "keuangan-linux-amd64"
$Anggaran = Join-Path $GoApp "Anggaran.xlsx"

Write-Host "==> Build binary Linux..." -ForegroundColor Cyan
Push-Location $GoApp
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -ldflags="-s -w" -o keuangan-linux-amd64 .
Pop-Location

if (-not (Test-Path $Binary)) { throw "Build gagal: $Binary tidak ditemukan" }
if (-not (Test-Path $Anggaran)) { throw "Anggaran.xlsx tidak ditemukan" }

Write-Host "==> Upload ke Hostinger (masukkan password SSH saat diminta)..." -ForegroundColor Cyan
ssh -p $HostingerPort "${HostingerUser}@${HostingerHost}" "mkdir -p ~/sipkeu ~/sipkeu-data ~/hostinger-web"
scp -P $HostingerPort $Binary "${HostingerUser}@${HostingerHost}:~/sipkeu/keuangan"
scp -P $HostingerPort $Anggaran "${HostingerUser}@${HostingerHost}:~/sipkeu/Anggaran.xlsx"
scp -P $HostingerPort (Join-Path $PSScriptRoot "start-remote.sh") "${HostingerUser}@${HostingerHost}:~/hostinger-web/start-remote.sh"
scp -P $HostingerPort (Join-Path $PSScriptRoot ".env.production") "${HostingerUser}@${HostingerHost}:~/hostinger-web/.env"

Write-Host "==> Start aplikasi di server..." -ForegroundColor Cyan
ssh -p $HostingerPort "${HostingerUser}@${HostingerHost}" "chmod +x ~/hostinger-web/start-remote.sh && bash ~/hostinger-web/start-remote.sh"

Write-Host ""
Write-Host "Deploy selesai! Buka: https://sakubijak.com:8888" -ForegroundColor Green
Write-Host "Login admin: admin / (password di deploy/hostinger-web/.env.production)" -ForegroundColor Yellow
Write-Host "Data tersimpan di: ~/sipkeu-data/ di server Hostinger" -ForegroundColor Gray
