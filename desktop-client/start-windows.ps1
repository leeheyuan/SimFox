$ErrorActionPreference = "Stop"

$desktopClientDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$electronExe = Join-Path $desktopClientDir "node_modules\electron\dist\electron.exe"
if (-not (Test-Path $electronExe)) {
  throw "desktop-client electron.exe was not found. Run npm install in desktop-client first."
}

Push-Location $desktopClientDir
try {
  npm run build
  Start-Process -FilePath $electronExe -ArgumentList "." -WorkingDirectory $desktopClientDir
} finally {
  Pop-Location
}
