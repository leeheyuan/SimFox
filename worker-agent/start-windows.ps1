[CmdletBinding()]
param(
  [string]$ConfigPath = ""
)

$ErrorActionPreference = "Stop"

$workerAgentDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$resolvedConfigPath = $ConfigPath
if (-not $resolvedConfigPath) {
  $resolvedConfigPath = Join-Path $workerAgentDir "config.local.json"
}

if (-not (Test-Path $resolvedConfigPath)) {
  throw "worker-agent config was not found: $resolvedConfigPath"
}

$pythonCommand = "cd /d $workerAgentDir && python main.py --config $resolvedConfigPath"
Start-Process cmd.exe -WindowStyle Hidden -ArgumentList "/c", $pythonCommand | Out-Null

Write-Host "worker-agent started with config: $resolvedConfigPath"
