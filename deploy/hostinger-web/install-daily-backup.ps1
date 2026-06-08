# Daftarkan backup harian SIPKeu ke Windows Task Scheduler
param(
    [string]$Time = "06:00"
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$DailyScript = Join-Path $ScriptDir "backup-daily.ps1"
$TaskName = "SIPKeu Daily Backup"

if (-not (Test-Path $DailyScript)) {
    Write-Error "Tidak ditemukan: $DailyScript"
}

$parts = $Time -split ":"
$hour = [int]$parts[0]
$minute = if ($parts.Length -gt 1) { [int]$parts[1] } else { 0 }

$existing = Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
if ($existing) {
    Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
    Write-Host "Task lama dihapus, mendaftar ulang..."
}

$action = New-ScheduledTaskAction `
    -Execute "powershell.exe" `
    -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$DailyScript`""

$trigger = New-ScheduledTaskTrigger -Daily -At ([datetime]::Today.AddHours($hour).AddMinutes($minute))

$settings = New-ScheduledTaskSettingsSet `
    -AllowStartIfOnBatteries `
    -DontStopIfGoingOnBatteries `
    -StartWhenAvailable `
    -ExecutionTimeLimit (New-TimeSpan -Hours 1)

Register-ScheduledTask `
    -TaskName $TaskName `
    -Action $action `
    -Trigger $trigger `
    -Settings $settings `
    -Description "Backup data SIPKeu dari server Hostinger ke OneDrive, sekali sehari." `
    | Out-Null

Write-Host ""
Write-Host "Task terdaftar: $TaskName"
Write-Host "Jadwal       : setiap hari pukul $Time"
Write-Host "Script       : $DailyScript"
Write-Host "Log          : $ScriptDir\logs\"
Write-Host ""
Write-Host "Pastikan deploy\.env berisi SSH_PASSWORD sebelum jadwal jalan."
Write-Host "Tes manual   : powershell -File `"$DailyScript`""
