# SimFox Worker Agent

`worker-agent` is the private-cloud execution node for SimFox.

It replaces the old `sim_server` WebSocket-oriented runner with a background worker that:

- registers itself with the SimFox platform
- sends periodic heartbeats
- claims queued simulation tasks
- runs SUMO locally
- reports completion or failure back to the platform

## Minimal local run

Create a config file from the example:

```powershell
Copy-Item E:\SimFox\worker-agent\config.example.json E:\SimFox\worker-agent\config.local.json
```

Then run:

```powershell
python E:\SimFox\worker-agent\main.py --config E:\SimFox\worker-agent\config.local.json
```

## Windows helpers

Start the worker-agent with the default local config:

```powershell
powershell -ExecutionPolicy Bypass -File E:\SimFox\worker-agent\start-windows.ps1
```

Start it with a different config:

```powershell
powershell -ExecutionPolicy Bypass -File E:\SimFox\worker-agent\start-windows.ps1 -ConfigPath E:\SimFox\worker-agent\config.example.json
```

Stop any running worker-agent process for this directory:

```powershell
powershell -ExecutionPolicy Bypass -File E:\SimFox\worker-agent\stop-windows.ps1
```

## Current assumptions

- the worker can reach `simulation_api`
- the worker has a valid bearer token
- task `Config.ConfigPath` points to a path visible from the worker host
- SUMO is installed locally and available as `sumo` or the configured `sumo_bin`

This first version intentionally uses the existing shared-path model. Artifact download and upload can be layered in later.
