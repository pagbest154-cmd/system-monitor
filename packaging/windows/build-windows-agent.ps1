param(
    [Parameter(Mandatory = $true)]
    [string]$Version
)

$ErrorActionPreference = "Stop"

$PackagingDir = $PSScriptRoot
$Root = Resolve-Path (Join-Path $PackagingDir "..\..")
$BuildDir = Join-Path $PackagingDir "build"
$DistDir = Join-Path $PackagingDir "dist"
$OutputDir = Join-Path $PackagingDir "output"
$ThirdPartyDir = Join-Path $PackagingDir "third_party"
$NssmDir = Join-Path $ThirdPartyDir "nssm\win64"
$NssmExe = Join-Path $NssmDir "nssm.exe"
$SpecFile = Join-Path $PackagingDir "system-monitor-agent.spec"
$IssFile = Join-Path $PackagingDir "installer.iss"

function Ensure-Nssm {
    if (Test-Path $NssmExe) {
        return
    }

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
        throw "Inno Setup 6 not found. Install from https://jrsoftware.org/isinfo.php or use: choco install innosetup -y"
    }

    return $iscc
}

Write-Host "Building system-monitor-agent $Version for Windows..."

Remove-Item -Recurse -Force $BuildDir, $DistDir, $OutputDir -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

Ensure-Nssm

Push-Location $Root
python -m PyInstaller `
    --noconfirm `
    --clean `
    --distpath $DistDir `
    --workpath $BuildDir `
    $SpecFile
Pop-Location

$agentDist = Join-Path $DistDir "system-monitor-agent"
if (-not (Test-Path (Join-Path $agentDist "system-monitor-agent.exe"))) {
    throw "PyInstaller output not found: $agentDist\system-monitor-agent.exe"
}

$iscc = Ensure-InnoSetup
& $iscc `
    "/DAppVersion=$Version" `
    "/DSourceDir=$agentDist" `
    "/DRepoRoot=$Root" `
    "/DOutputDir=$OutputDir" `
    $IssFile

$setup = Get-ChildItem -Path $OutputDir -Filter "system-monitor-agent_*_setup.exe" | Select-Object -First 1
if (-not $setup) {
    throw "Installer exe was not produced in $OutputDir"
}

Write-Host "Done: $($setup.FullName)"
