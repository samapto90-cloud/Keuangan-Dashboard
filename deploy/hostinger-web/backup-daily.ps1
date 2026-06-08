# Jalankan backup harian SIPKeu -> OneDrive
$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Root = Resolve-Path (Join-Path $ScriptDir "..\..")
$LogDir = Join-Path $Root "deploy\hostinger-web\logs"
$EnvFile = Join-Path $Root "deploy\.env"
$NodeScript = Join-Path $ScriptDir "backup-remote.mjs"

New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
$LogFile = Join-Path $LogDir ("backup-{0:yyyyMMdd-HHmmss}.log" -f (Get-Date))

function Write-Log([string]$Message) {
    $line = "[{0:yyyy-MM-dd HH:mm:ss}] {1}" -f (Get-Date), $Message
    Add-Content -Path $LogFile -Value $line
    Write-Output $line
}

try {
    Write-Log "=== SIPKeu daily backup start ==="

    if (-not (Test-Path $EnvFile)) {
        throw "File $EnvFile belum ada. Salin deploy\.env.example ke deploy\.env dan isi SSH_PASSWORD."
    }

    foreach ($line in Get-Content $EnvFile) {
        $t = $line.Trim()
        if (-not $t -or $t.StartsWith("#")) { continue }
        $eq = $t.IndexOf("=")
        if ($eq -le 0) { continue }
        $key = $t.Substring(0, $eq).Trim()
        $val = $t.Substring($eq + 1).Trim().Trim('"').Trim("'")
        if ($key -eq "SSH_PASSWORD") { $env:SSH_PASSWORD = $val }
    }

    if (-not $env:SSH_PASSWORD) {
        throw "SSH_PASSWORD kosong di deploy\.env"
    }

    $node = (Get-Command node -ErrorAction Stop).Source
    Write-Log "Node: $node"
    Write-Log "Script: $NodeScript"

    Push-Location $ScriptDir
    & $node $NodeScript 2>&1 | ForEach-Object { Write-Log $_ }
    if ($LASTEXITCODE -and $LASTEXITCODE -ne 0) {
        throw "backup-remote.mjs exit code $LASTEXITCODE"
    }
    Pop-Location

    # Hapus log lebih dari 14 hari
    Get-ChildItem $LogDir -Filter "backup-*.log" -ErrorAction SilentlyContinue |
        Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-14) } |
        Remove-Item -Force -ErrorAction SilentlyContinue

    Write-Log "=== SIPKeu daily backup selesai ==="
    exit 0
}
catch {
    Write-Log "ERROR: $($_.Exception.Message)"
    if ($_.ScriptStackTrace) { Write-Log $_.ScriptStackTrace }
    exit 1
}
finally {
    if ((Get-Location).Path -ne $PWD.Path) { Pop-Location -ErrorAction SilentlyContinue }
}
