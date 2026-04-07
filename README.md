# SimFox

SimFox is a multi-module traffic simulation workspace that combines Go backend services, a Vue-based management console, a React-based road editor, and a Python simulation server.

## Repository Layout

| Path | Purpose | Stack |
| --- | --- | --- |
| `config/` | Shared database and configuration helpers | Go |
| `config_Sever/` | Configuration service API and Swagger docs | Go |
| `logger/` | Logging service and tests | Go |
| `longConnection/` | Long-connection messaging helpers | Go |
| `middleware/` | Shared middleware such as auth helpers | Go |
| `models/` | Shared data models | Go |
| `port/` | Port allocation manager | Go |
| `SimulationScheduling/` | Simulation scheduling entrypoint | Go |
| `user_auth/` | User and project API service | Go |
| `simulation_manager/` | Management console | Vue 3 + TypeScript + Vite |
| `road-editor/` | 3D road editing frontend | React + TypeScript + Vite + Three.js |
| `sim_server/` | Simulation runner / websocket push service | Python |
| `SumoConfig/`, `typeDef/`, `utils/` | Shared support packages | Go |

## Getting Started

### Frontends

`simulation_manager/`

```powershell
cd E:\SimFox\simulation_manager
npm install
npm run dev
```

`road-editor/`

```powershell
cd E:\SimFox\road-editor
npm install
npm run dev
```

### Python simulation service

The Python service entrypoint is `sim_server/sim_server.py`. It expects a SUMO config path plus runtime parameters.

```powershell
cd E:\SimFox\sim_server
python sim_server.py --config path\to\sim.sumocfg --duration 60 --speed 1.0 --ports 9001,9002 --projectId 1
```

### Go services

This repository is organized as multiple Go modules instead of a single root module. Run commands inside the specific service directory you want to work on.

Example:

```powershell
cd E:\SimFox\user_auth
go run .
```

## Notes

- `config_Sever/` appears to be the configuration server module; the directory name is kept as-is to avoid breaking imports and scripts.
- The root `.gitignore` excludes local dependencies, caches, build artifacts, and local database files so day-to-day generated files do not pollute commits.
