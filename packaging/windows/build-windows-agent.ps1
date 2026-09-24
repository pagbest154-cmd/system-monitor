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
    Write-Host "Generating app-icon.ico from branding..."
    Push-Location $Root
    go run ./cmd/gen-icons
    Pop-Location
    if (-not (Test-Path $IconFile)) {
        throw "Icon not generated: $IconFile"
    }
}

function Ensure-WindowsResources {
    $manifest = Join-Path $PackagingDir "app.manifest"
    $rsrcOut = Join-Path $Root "cmd\system-monitor-agent\rsrc.syso"
    if (-not (Get-Command rsrc -ErrorAction SilentlyContinue)) {
        go install github.com/akavel/rsrc@v0.10.2
    }
    Write-Host "Embedding Windows manifest and icon..."
    rsrc -manifest $manifest -ico $IconFile -o $rsrcOut
}

Write-Host "Building system-monitor-agent $Version for Windows (Go)..."

Remove-Item -Recurse -Force $BuildDir, (Join-Path $PackagingDir "dist"), $OutputDir -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $DistDir, $OutputDir | Out-Null

Ensure-Nssm
Ensure-Icon
Ensure-WindowsResources

$agentExe = Join-Path $DistDir "system-monitor-agent.exe"
Push-Location $Root
$env:CGO_ENABLED = "0"
$ldflags = "-H windowsgui -s -w -X github.com/pagbest154-cmd/system-monitor/internal/version.Version=$Version"
go build -ldflags $ldflags -o $agentExe ./cmd/system-monitor-agent
Remove-Item (Join-Path $Root "cmd\system-monitor-agent\rsrc.syso") -ErrorAction SilentlyContinue
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
