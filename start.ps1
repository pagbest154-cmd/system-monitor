$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$python = Join-Path $root ".python\python.exe"

if (-not (Test-Path $python)) {
    Write-Error "Python не найден в .python\python.exe. Установите Python или запустите setup."
    exit 1
}

Start-Process -FilePath $python `
    -ArgumentList "-m system_monitor --mode standalone --host 127.0.0.1 --port 8080" `
    -WorkingDirectory $root `
    -WindowStyle Hidden

Write-Host "system-monitor запущен: http://127.0.0.1:8080"
