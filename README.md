# SimFox

SimFox is being refocused into a private-cloud traffic simulation management platform.

The core product is no longer a road-network renderer or a chat agent. It manages projects, input artifacts, task queues, cluster workers, logs, and result files. SUMO execution happens on private-cloud worker machines. The Electron client is only a local capability bridge for desktop file access, uploads, downloads, notifications, and controlled native-tool integration.

## Target Architecture

```text
Desktop Client
  Electron shell
  Web management UI
  Local file picker / upload / download / open folder / notifications
        |
        v
Platform API
  Auth / tenant / projects / artifacts / task queue / logs / results
        |
        v
Scheduler
  Assigns queued tasks to private-cloud workers
        |
        v
Worker Agent
  Runs on cluster machines
  Pulls tasks, downloads inputs, starts SUMO, uploads logs/results
```

## Repository Layout

| Path | Purpose | Status |
| --- | --- | --- |
| `desktop-client/` | Electron local capability bridge for files, folders, notifications, and controlled native actions | Main desktop shell |
| `simulation_manager/` | Vue management console for projects, tasks, results, cluster, settings | Main frontend |
| `simulation_api/` | Go platform API for projects, tasks, worker endpoints | Main backend |
| `SimulationScheduling/` | Existing local scheduler; to be replaced by worker-agent scheduling | Transitional |
| `worker-agent/` | Python private-cloud worker that registers, heartbeats, claims tasks, and runs SUMO locally | Main execution node |
| `sim_server/` | Deprecated local WebSocket-era SUMO runner kept only for short-term reference | Transitional |
| `models/` | Shared GORM models for tenants, projects, tasks, artifacts, workers, logs | Main backend |
| `SumoConfig/`, `typeDef/`, `utils/` | SUMO config parsing and shared helpers | Keep |

## Current Product Boundary

Keep in the main line:

- Project and input-file management
- Simulation task queue
- Worker registration and heartbeat
- Worker task claim/complete/fail API
- Result artifact and log management
- Private-cloud deployment
- Electron as a local desktop bridge

Move out of the main line:

- Three.js road-network rendering
- Mapbox/SUMO live visualization as the primary workflow
- Chat agent and tool-calling orchestration
- Direct SSH execution from Electron to cluster machines
- Browser-side unrestricted system command execution

## Local Startup

The Go services read MySQL from `SIMFOX_MYSQL_DSN` first. If it is not set, they fall back to the legacy built-in DSN in [db.go](/E:/SimFox/config/db.go:10).

Example:

```powershell
$env:SIMFOX_MYSQL_DSN = "root:password@tcp(127.0.0.1:3306)/simulationtraffic?charset=utf8mb4&parseTime=True&loc=Local"
```

### One-click Windows startup

Create a local private config once:

```powershell
Copy-Item E:\SimFox\simfox.local.ps1.example E:\SimFox\simfox.local.ps1
```

Then edit [simfox.local.ps1](/E:/SimFox/simfox.local.ps1.example) and set your real MySQL DSN:

```powershell
$script:SimFoxLocalConfig = @{
  MySqlDsn = "root:your-password@tcp(127.0.0.1:3306)/simulationtraffic?charset=utf8mb4&parseTime=True&loc=Local"
}
```

`simfox.local.ps1` is ignored by Git, so your local password stays on your machine.

To start the main local services and Electron desktop client together, run:

```powershell
powershell -ExecutionPolicy Bypass -File E:\SimFox\start-windows.ps1
```

The script will print a quick status table for:

- `user_auth` on `8080`
- `config_Sever` on `8081`
- `simulation_api` on `8082`
- `SimulationScheduling` on `8090`

If any row shows `FAIL`, that service did not stay up and the desktop app may show login or data-loading errors.

To override the database DSN:

```powershell
powershell -ExecutionPolicy Bypass -File E:\SimFox\start-windows.ps1 -MySqlDsn "root:your-password@tcp(127.0.0.1:3306)/simulationtraffic?charset=utf8mb4&parseTime=True&loc=Local"
```

### Install frontend dependencies

```powershell
cd E:\SimFox\simulation_manager
npm install
```

### Start each service in its own terminal

`user_auth` on `:8080`

```powershell
cd E:\SimFox\user_auth
go run .
```

`config_Sever` on `:8081`

```powershell
cd E:\SimFox\config_Sever
go run .
```

`simulation_api` on `:8082`

```powershell
cd E:\SimFox\simulation_api
go run .
```

`SimulationScheduling` on `:8090`

```powershell
cd E:\SimFox\SimulationScheduling
go run .
```

`simulation_manager` on `:5173`

```powershell
cd E:\SimFox\simulation_manager
npm run dev -- --host 127.0.0.1 --port 5173
```

Open:

```text
http://127.0.0.1:5173/
```

### Electron desktop entry

Run the web UI inside Electron instead of a browser:

```powershell
cd E:\SimFox\desktop-client
npm install
npm run dev
```

This starts the `simulation_manager` dev server and opens Electron automatically.

If the local Electron package is unavailable or you want a direct Windows launcher, run:

```powershell
powershell -ExecutionPolicy Bypass -File E:\SimFox\desktop-client\start-windows.ps1
```

When imported from Electron, the `Projects` page can select a single `.sumocfg` file and automatically resolve sibling dependency files from the local filesystem before upload.

### Stop all local processes

To stop all known SimFox backend, frontend, and Electron processes on Windows, run:

```powershell
powershell -ExecutionPolicy Bypass -File E:\SimFox\stop-windows.ps1
```

There is also a desktop-local wrapper:

```powershell
powershell -ExecutionPolicy Bypass -File E:\SimFox\desktop-client\stop-windows.ps1
```

The stop script targets:

- listeners on `8080`, `8081`, `8082`, `8090`, `5173`, `5174`
- Electron processes started from `desktop-client`
- processes whose command line points at the SimFox workspace service directories

### Optional background startup from PowerShell

```powershell
$env:SIMFOX_MYSQL_DSN = "root:password@tcp(127.0.0.1:3306)/simulationtraffic?charset=utf8mb4&parseTime=True&loc=Local"

Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd E:\SimFox\user_auth; go run ."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd E:\SimFox\config_Sever; go run ."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd E:\SimFox\simulation_api; go run ."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd E:\SimFox\SimulationScheduling; go run ."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd E:\SimFox\simulation_manager; npm run dev -- --host 127.0.0.1 --port 5173"
```

The scheduler is still transitional. It currently launches local Python/SUMO commands and remains only as a compatibility bridge while the worker-agent path is being built.

## Worker API Draft

Worker agents should call the platform API instead of being controlled directly by Electron:

```text
POST /worker/register
POST /worker/:workerId/heartbeat
POST /worker/:workerId/tasks/next
POST /worker/:workerId/tasks/:taskId/complete
POST /worker/:workerId/tasks/:taskId/fail
```

The first `worker-agent/` version is now in place and uses these endpoints as its control loop.
