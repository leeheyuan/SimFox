$ErrorActionPreference = "Stop"

$workspaceRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$rootStopScript = Join-Path $workspaceRoot "stop-windows.ps1"

if (-not (Test-Path $rootStopScript)) {
  throw "stop-windows.ps1 was not found at $rootStopScript"
}

powershell -ExecutionPolicy Bypass -File $rootStopScript
