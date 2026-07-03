param(
    [string]$Output = "bin/openusage.exe"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$outputPath = Join-Path $repoRoot $Output
$outputDir = Split-Path -Parent $outputPath

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is not installed or not on PATH."
}

if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) {
    throw "gcc is not installed or not on PATH. Install MinGW-w64 or scoop gcc for CGO builds."
}

New-Item -ItemType Directory -Path $outputDir -Force | Out-Null

$env:CGO_ENABLED = "1"

Push-Location $repoRoot
try {
    go build -o $outputPath ./cmd/openusage
    Write-Host "Built $outputPath"
}
finally {
    Pop-Location
}
