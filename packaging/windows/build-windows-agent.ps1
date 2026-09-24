param(
    [Parameter(Mandatory = $true)]
    [string]$Version
)

$ErrorActionPreference = "Stop"

$PackagingDir = $PSScriptRoot
$Root = Resolve-Path (Join-Path $PackagingDir "..\..")
$BuildDir = Join-Path $PackagingDir "build"
$DistDir = Join-Path $PackagingDir "dist\system-monitor-agent"
$OutputDir = Join-Path $PackagingDir "output"
$ThirdPartyDir = Join-Path $PackagingDir "third_party"
$NssmDir = Join-Path $ThirdPartyDir "nssm\win64"
$NssmExe = Join-Path $NssmDir "nssm.exe"
$IssFile = Join-Path $PackagingDir "installer.iss"
$IconFile = Join-Path $PackagingDir "app-icon.ico"

function Ensure-Nssm {
    if (Test-Path $NssmExe) { return }
    $zipPath = Join-Path $ThirdPartyDir "nssm.zip"
    New-Item -ItemType Directory -Force -Path $NssmDir | Out-Null
    Write-Host "Downloading NSSM..."
    Invoke-WebRequest -Uri "https://nssm.cc/release/nssm-2.24.zip" -OutFile $zipPath
    Expand-Archive -Path $zipPath -DestinationPath (Join-Path $ThirdPartyDir "nssm-src") -Force
    Copy-Item (Join-Path $ThirdPartyDir "nssm-src\nssm-2.24\win64\nssm.exe") $NssmExe
    Remove-Item $zipPath -Force
    Remove-Item (Join-Path $ThirdPartyDir "nssm-src") -Recurse -Force
}

function Ensure-InnoSetup {
    $iscc = @(
        "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
        "${env:ProgramFiles}\Inno Setup 6\ISCC.exe"
    ) | Where-Object { Test-Path $_ } | Select-Object -First 1
    if (-not $iscc) {
        throw "Inno Setup 6 not found"
    }
    return $iscc
}

function Ensure-Icon {
    if (Test-Path $IconFile) { return }
    Write-Host "Generating app-icon.ico..."
    Add-Type -AssemblyName System.Drawing
    $bmp = New-Object System.Drawing.Bitmap 32, 32
    $graphics = [System.Drawing.Graphics]::FromImage($bmp)
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $graphics.Clear([System.Drawing.Color]::FromArgb(37, 99, 235))
    $brush = New-Object System.Drawing.SolidBrush ([System.Drawing.Color]::White)
    $graphics.FillEllipse($brush, 6, 6, 20, 20)
    $graphics.Dispose()
    $icon = [System.Drawing.Icon]::FromHandle($bmp.GetHicon())
    $stream = [System.IO.File]::Create($IconFile)
    try {
        $icon.Save($stream)
    } finally {
        $stream.Close()
        $icon.Dispose()
        $bmp.Dispose()
    }
    Write-Host "Created: $IconFile"
}

Write-Host "Building system-monitor-agent $Version for Windows (Go)..."

Remove-Item -Recurse -Force $BuildDir, (Join-Path $PackagingDir "dist"), $OutputDir -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $DistDir, $OutputDir | Out-Null

Ensure-Nssm
Ensure-Icon

$agentExe = Join-Path $DistDir "system-monitor-agent.exe"
Push-Location $Root
$env:CGO_ENABLED = "0"
$ldflags = "-s -w -X github.com/pagbest154-cmd/system-monitor/internal/version.Version=$Version"
go build -ldflags $ldflags -o $agentExe ./cmd/system-monitor-agent
Pop-Location

if (-not (Test-Path $agentExe)) {
    throw "Go build output not found: $agentExe"
}

Write-Host "Smoke test: system-monitor-agent.exe --version"
$smoke = Start-Process -FilePath $agentExe -ArgumentList "--version" -Wait -PassThru -NoNewWindow
if ($smoke.ExitCode -ne 0) {
    throw "Agent exe failed smoke test (exit $($smoke.ExitCode))"
}

Copy-Item $IconFile (Join-Path $DistDir "app-icon.ico") -Force

$iscc = Ensure-InnoSetup
& $iscc `
    "/DAppVersion=$Version" `
    "/DSourceDir=$DistDir" `
    "/DRepoRoot=$Root" `
    "/DOutputDir=$OutputDir" `
    $IssFile

$setup = Get-ChildItem -Path $OutputDir -Filter "system-monitor-agent_*_setup.exe" | Select-Object -First 1
if (-not $setup) {
    throw "Installer exe was not produced in $OutputDir"
}

Write-Host "Done: $($setup.FullName)"
