$ErrorActionPreference = 'Stop'

function Fail([string]$Message) {
    Write-Host ""
    Write-Host "ERROR: $Message" -ForegroundColor Red
    exit 1
}

function Require-Command([string]$Name, [string]$Message) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Fail $Message
    }
}

$root = Split-Path -Parent $PSScriptRoot
$backend = Join-Path $root 'backend-go'
$desktop = Join-Path $root 'desktop-wails'
$frontend = Join-Path $desktop 'frontend'
$envFile = Join-Path $root '.env'

foreach ($directory in @($backend, $desktop, $frontend)) {
    if (-not (Test-Path -LiteralPath $directory -PathType Container)) {
        Fail "Required directory does not exist: $directory"
    }
}

Require-Command 'go' 'Go is not installed or is not available in PATH.'
Require-Command 'npm' 'npm is not installed or is not available in PATH.'
Require-Command 'wails' 'Wails CLI is not installed or is not available in PATH.'
$npmCommand = (Get-Command 'npm.cmd' -ErrorAction SilentlyContinue).Source
if (-not $npmCommand) {
    Fail 'npm.cmd is not installed or is not available in PATH.'
}

if (-not (Test-Path -LiteralPath $envFile -PathType Leaf)) {
    Fail "Missing environment file: $envFile"
}

if (-not (Test-Path -LiteralPath (Join-Path $frontend 'node_modules') -PathType Container)) {
    Write-Host "Installing frontend dependencies..." -ForegroundColor Cyan
    Push-Location $frontend
    try {
        & npm install
        if ($LASTEXITCODE -ne 0) { Fail 'npm install failed.' }
    }
    finally {
        Pop-Location
    }
}

Write-Host "Checking TypeScript..." -ForegroundColor Cyan
Push-Location $frontend
try {
    & npx tsc --noEmit
    if ($LASTEXITCODE -ne 0) { Fail 'TypeScript check failed.' }

    Write-Host "Checking frontend build (60 second timeout)..." -ForegroundColor Cyan
    $buildInfo = [System.Diagnostics.ProcessStartInfo]::new()
    $buildInfo.FileName = $npmCommand
    $buildInfo.Arguments = 'run build'
    $buildInfo.WorkingDirectory = $frontend
    $buildInfo.UseShellExecute = $false

    $build = [System.Diagnostics.Process]::new()
    $build.StartInfo = $buildInfo

    if (-not $build.Start()) {
        Fail 'npm run build process could not be started.'
    }

    if (-not $build.WaitForExit(60000)) {
        & taskkill.exe /PID $build.Id /T /F 2>&1 | Out-Null
        Fail 'npm run build timed out after 60 seconds.'
    }

    $build.WaitForExit()
    $capturedExitCode = $build.ExitCode
    if ($null -eq $capturedExitCode) {
        Fail 'npm run build finished but exit code was not captured.'
    }

    $buildExitCode = [int]$capturedExitCode
    $build.Dispose()
    if ($buildExitCode -ne 0) {
        Fail "npm run build failed with exit code $buildExitCode."
    }
    Write-Host "Frontend build check passed." -ForegroundColor Green
}
finally {
    Pop-Location
}

Write-Host "Starting backend development server..." -ForegroundColor Cyan
$backendCommand = "Set-Location -LiteralPath '$backend'; go run ./cmd/server"
Start-Process -FilePath 'powershell.exe' -ArgumentList '-NoExit', '-Command', $backendCommand -WorkingDirectory $backend | Out-Null
Start-Sleep -Seconds 2

Write-Host "Starting Wails desktop development environment..." -ForegroundColor Cyan
Push-Location $desktop
try {
    & wails dev
    if ($LASTEXITCODE -ne 0) { Fail "wails dev failed with exit code $LASTEXITCODE." }
}
finally {
    Pop-Location
}
