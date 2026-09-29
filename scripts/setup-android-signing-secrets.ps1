# Upload android/release.keystore (or sysmon-release.jks) to GitHub Actions secrets.
# Run once from repo root after: gh auth login
param(
    [string]$KeystorePath = "android/sysmon-release.jks",
    [string]$StorePassword = "sysmon-release-store",
    [string]$KeyAlias = "sysmon",
    [string]$KeyPassword = "sysmon-release-key"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $root

gh auth status *> $null
if ($LASTEXITCODE -ne 0) {
    Write-Error "GitHub CLI is not logged in. Run: gh auth login"
}

if (-not (Test-Path $KeystorePath)) {
    Write-Error "Keystore not found: $KeystorePath. Generate with keytool first."
}

$b64 = [Convert]::ToBase64String([IO.File]::ReadAllBytes((Resolve-Path $KeystorePath)))
$b64 | gh secret set ANDROID_KEYSTORE_BASE64
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$StorePassword | gh secret set ANDROID_KEYSTORE_PASSWORD
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$KeyAlias | gh secret set ANDROID_KEY_ALIAS
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$KeyPassword | gh secret set ANDROID_KEY_PASSWORD
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Copy-Item $KeystorePath "android/release.keystore" -Force
Write-Host "GitHub secrets updated. Copied keystore to android/release.keystore for local release builds."
