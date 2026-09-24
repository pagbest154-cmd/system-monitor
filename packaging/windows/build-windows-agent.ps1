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

$iconFile = Join-Path $PackagingDir "app-icon.ico"
Write-Host "Generating app-icon.ico..."
python (Join-Path $PackagingDir "generate_icon.py")
if (-not (Test-Path $iconFile)) {
    throw "Icon file was not generated: $iconFile"
}
$iconSize = (Get-Item $iconFile).Length
if ($iconSize -lt 500) {
    throw "Icon file looks invalid ($iconSize bytes): $iconFile"
}
Write-Host "Icon ready: $iconFile ($iconSize bytes)"

Push-Location $Root
python -m PyInstaller `
    --noconfirm `
    --clean `
    --distpath $DistDir `
    --workpath $BuildDir `
    $SpecFile
Pop-Location

$agentDist = Join-Path $DistDir "system-monitor-agent"
$agentExe = Join-Path $agentDist "system-monitor-agent.exe"
if (-not (Test-Path $agentExe)) {
    throw "PyInstaller output not found: $agentExe"
}

# Icon is embedded by PyInstaller (system-monitor-agent.spec). Do not run rcedit on the
# bootloader — it corrupts the exe and breaks startup with "Could not load PKG archive".
Copy-Item $iconFile (Join-Path $agentDist "app-icon.ico") -Force

Write-Host "Smoke test: system-monitor-agent.exe --version"
$smoke = Start-Process -FilePath $agentExe -ArgumentList "--version" -Wait -PassThru -NoNewWindow
if ($smoke.ExitCode -ne 0) {
    throw "Agent exe failed smoke test (exit $($smoke.ExitCode))"
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
