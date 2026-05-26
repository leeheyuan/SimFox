$ErrorActionPreference = "Stop"

$workspaceRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$knownPorts = @(8080, 8081, 8082, 8090, 5173, 5174)
$servicePaths = @(
  (Join-Path $workspaceRoot "user_auth"),
  (Join-Path $workspaceRoot "config_Sever"),
  (Join-Path $workspaceRoot "simulation_api"),
  (Join-Path $workspaceRoot "SimulationScheduling"),
  (Join-Path $workspaceRoot "simulation_manager"),
  (Join-Path $workspaceRoot "desktop-client")
) | ForEach-Object { $_.ToLowerInvariant() }
$electronPath = (Join-Path $workspaceRoot "desktop-client\node_modules\electron\dist\electron.exe").ToLowerInvariant()

$pids = New-Object System.Collections.Generic.HashSet[int]

try {
  $connections = Get-NetTCPConnection -LocalPort $knownPorts -State Listen -ErrorAction Stop
  foreach ($connection in $connections) {
    [void]$pids.Add([int]$connection.OwningProcess)
  }
} catch {
  Write-Host "No listening SimFox ports were found."
}

$processes = Get-CimInstance Win32_Process -ErrorAction SilentlyContinue
foreach ($process in $processes) {
  $commandLine = ""
  $executablePath = ""
  if ($null -ne $process.CommandLine) {
    $commandLine = $process.CommandLine.ToLowerInvariant()
  }
  if ($null -ne $process.ExecutablePath) {
    $executablePath = $process.ExecutablePath.ToLowerInvariant()
  }

  $matchesWorkspace = $servicePaths | Where-Object { $commandLine.Contains($_) }
  $matchesElectron = $electronPath -and ($executablePath -eq $electronPath -or $commandLine.Contains($electronPath))

  if ($matchesWorkspace -or $matchesElectron) {
    [void]$pids.Add([int]$process.ProcessId)
  }
}

if ($pids.Count -eq 0) {
  Write-Host "No SimFox processes found."
  exit 0
}

$sortedPids = @($pids) | Sort-Object
Write-Host ("Stopping SimFox processes: " + ($sortedPids -join ", "))
Stop-Process -Id $sortedPids -Force
Write-Host "SimFox processes stopped."
