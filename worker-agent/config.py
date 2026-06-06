from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Optional


@dataclass
class WorkerConfig:
    platform_url: str
    worker_id: Optional[int]
    worker_name: str
    worker_secret: str
    address: str
    queue_name: str
    labels_json: str
    max_concurrency: int
    sumo_bin: str
    work_dir: Path
    workspace_root: Path
    poll_interval_sec: int
    heartbeat_interval_sec: int


def load_config(path: str) -> WorkerConfig:
    config_path = Path(path).resolve()
    raw = json.loads(config_path.read_text(encoding="utf-8"))
    workspace_root = Path(raw.get("workspace_root", config_path.parent.parent)).resolve()
    return WorkerConfig(
        platform_url=str(raw["platform_url"]).rstrip("/"),
        worker_id=int(raw["worker_id"]) if raw.get("worker_id") not in (None, "") else None,
        worker_name=str(raw["worker_name"]).strip(),
        worker_secret=str(raw["worker_secret"]).strip(),
        address=str(raw.get("address", "")).strip(),
        queue_name=str(raw.get("queue_name", "default")).strip() or "default",
        labels_json=str(raw.get("labels_json", "")).strip(),
        max_concurrency=max(1, int(raw.get("max_concurrency", 1))),
        sumo_bin=str(raw.get("sumo_bin", "sumo")).strip() or "sumo",
        work_dir=Path(raw.get("work_dir", "./workdir")).resolve(),
        workspace_root=workspace_root,
        poll_interval_sec=max(1, int(raw.get("poll_interval_sec", 3))),
        heartbeat_interval_sec=max(1, int(raw.get("heartbeat_interval_sec", 10))),
    )
