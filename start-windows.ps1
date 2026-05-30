[CmdletBinding()]
param(
  [string]$MySqlDsn = ""
)

$ErrorActionPreference = "Stop"

$workspaceRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$desktopStartScript = Join-Path $workspaceRoot "desktop-client\start-windows.ps1"
$localConfigPath = Join-Path $workspaceRoot "simfox.local.ps1"

if (-not (Test-Path $desktopStartScript)) {
  throw "desktop-client start-windows.ps1 was not found at $desktopStartScript"
}

if (Test-Path $localConfigPath) {
  . $localConfigPath
}

function Resolve-SimFoxDsn {
  param(
    [string]$CliValue
  )

  if ($CliValue) {
    return $CliValue
  }

  if ($script:SimFoxLocalConfig -and $script:SimFoxLocalConfig.ContainsKey("MySqlDsn")) {
    return $script:SimFoxLocalConfig["MySqlDsn"]
  }

  if ($env:SIMFOX_MYSQL_DSN) {
    return $env:SIMFOX_MYSQL_DSN
  }

  throw "MySQL DSN is not configured. Create E:\SimFox\simfox.local.ps1 or pass -MySqlDsn."
}

$resolvedDsn = Resolve-SimFoxDsn -CliValue $MySqlDsn

function Start-SimFoxService {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Name,

    [Parameter(Mandatory = $true)]
    [string]$RelativePath,

    [Parameter(Mandatory = $true)]
    [int]$Port,

    [string]$ExtraEnv = ""
  )

  $servicePath = Join-Path $workspaceRoot $RelativePath
  if (-not (Test-Path $servicePath)) {
    throw "$Name path was not found: $servicePath"
  }

  $envPrefix = "`$env:SIMFOX_MYSQL_DSN='$resolvedDsn'; "
  if ($ExtraEnv) {
    $envPrefix += $ExtraEnv + "; "
  }
  $command = "cd /d $servicePath && powershell -Command `"& { $envPrefix go run . }`""
  Start-Process cmd.exe -WindowStyle Hidden -ArgumentList "/c", $command | Out-Null
  Write-Host "Started $Name (target port $Port)"
}

Start-SimFoxService -Name "user_auth" -RelativePath "user_auth" -Port 8080
Start-SimFoxService -Name "config_Sever" -RelativePath "config_Sever" -Port 8081
Start-SimFoxService -Name "simulation_api" -RelativePath "simulation_api" -Port 8082
Start-SimFoxService -Name "SimulationScheduling" -RelativePath "SimulationScheduling" -Port 8090 -ExtraEnv "`$env:SCHEDULER_MAX_WORKERS='0'"

Start-Sleep -Seconds 6

$servicePorts = @(
  @{ Name = "user_auth"; Port = 8080 }
  @{ Name = "config_Sever"; Port = 8081 }
  @{ Name = "simulation_api"; Port = 8082 }
  @{ Name = "SimulationScheduling"; Port = 8090 }
)

Write-Host ""
Write-Host "Service status:"
foreach ($service in $servicePorts) {
  $listening = Get-NetTCPConnection -LocalPort $service.Port -State Listen -ErrorAction SilentlyContinue
  if ($listening) {
    Write-Host ("  OK   " + $service.Name + " on :" + $service.Port)
  } else {
    Write-Host ("  FAIL " + $service.Name + " on :" + $service.Port)
  }
}

Write-Host "Starting desktop-client"
powershell -ExecutionPolicy Bypass -File $desktopStartScript

Write-Host ""
Write-Host "SimFox startup command completed."
Write-Host "Desktop client should now be opening."
