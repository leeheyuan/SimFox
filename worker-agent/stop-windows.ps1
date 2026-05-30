$ErrorActionPreference = "Stop"

$workerAgentDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$escapedWorkerDir = [Regex]::Escape($workerAgentDir)

$processes = Get-CimInstance Win32_Process -ErrorAction SilentlyContinue |
  Where-Object {
    $_.Name -like "python*.exe" -and
    $_.CommandLine -match "main\.py" -and
    $_.CommandLine -match $escapedWorkerDir
  }

if (-not $processes) {
  Write-Host "No worker-agent process found."
  exit 0
}

$processIds = @($processes | Select-Object -ExpandProperty ProcessId)
Write-Host ("Stopping worker-agent process ids: " + ($processIds -join ", "))
Stop-Process -Id $processIds -Force
Write-Host "worker-agent stopped."
