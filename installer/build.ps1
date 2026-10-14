$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$payload = Join-Path $PSScriptRoot "payload"
$dist = Join-Path $PSScriptRoot "dist"

New-Item -ItemType Directory -Force -Path $payload | Out-Null
New-Item -ItemType Directory -Force -Path $dist | Out-Null

Push-Location $root
try {
    go build -o (Join-Path $payload "luminalang.exe") .
    go build -o (Join-Path $payload "pho.exe") ./cmd/pho
}
finally {
    Pop-Location
}

dotnet publish (Join-Path $PSScriptRoot "LuminaInstaller.csproj") -c Release -o $dist

Get-ChildItem $dist -Filter "*.exe" | ForEach-Object {
    Write-Host ("built {0} ({1:N1} MB)" -f $_.Name, ($_.Length / 1MB))
}
