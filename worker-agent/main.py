from __future__ import annotations

import argparse
import threading
import time
from pathlib import Path

from api_client import ApiClient
from config import WorkerConfig, load_config
from runner import run_sumo_task


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="SimFox worker agent")
    parser.add_argument("--config", required=True, help="Path to worker config JSON")
    return parser.parse_args()


class WorkerAgent:
    def __init__(self, config: WorkerConfig):
        self.config = config
        self.client = ApiClient(config.platform_url, config.token)
        self.worker_id: int | None = None
        self.running_tasks = 0
        self.running_lock = threading.Lock()

    def run(self) -> None:
        self.config.work_dir.mkdir(parents=True, exist_ok=True)
        self.worker_id = self.register()
        heartbeat_thread = threading.Thread(target=self.heartbeat_loop, daemon=True)
        heartbeat_thread.start()

        print(f"worker-agent started: worker_id={self.worker_id} name={self.config.worker_name}")
        while True:
            if self.current_running_tasks() >= self.config.max_concurrency:
                time.sleep(self.config.poll_interval_sec)
                continue

            try:
                claimed = self.claim_task()
            except Exception as exc:  # noqa: BLE001
                print(f"worker loop recoverable error: {exc}")
                time.sleep(self.config.poll_interval_sec)
                continue
            if not claimed:
                time.sleep(self.config.poll_interval_sec)
                continue

            worker_thread = threading.Thread(target=self.execute_task, args=(claimed,), daemon=True)
            worker_thread.start()
            time.sleep(1)

    def register(self) -> int:
        response = self.client.post(
            "/worker/register",
            {
                "name": self.config.worker_name,
                "address": self.config.address,
                "queueName": self.config.queue_name,
                "labelsJson": self.config.labels_json,
                "maxConcurrency": self.config.max_concurrency,
            },
        )
        worker_id = int(response["workerId"])
        return worker_id

    def heartbeat_loop(self) -> None:
        while True:
            try:
                self.client.post(
                    f"/worker/{self.worker_id}/heartbeat",
                    {
                        "status": "online" if self.current_running_tasks() == 0 else "busy",
                        "runningTasks": self.current_running_tasks(),
                    },
                )
            except Exception as exc:  # noqa: BLE001
                print(f"heartbeat failed: {exc}")
            time.sleep(self.config.heartbeat_interval_sec)

    def claim_task(self) -> dict | None:
        try:
            response = self.client.post(f"/worker/{self.worker_id}/tasks/next", {}, allow_no_content=True)
        except RuntimeError as exc:
            if "404" in str(exc):
                return None
            print(f"claim task failed: {exc}")
            return None
        if not response:
            return None
        return response.get("task")

    def execute_task(self, task: dict) -> None:
        self.change_running_tasks(1)
        task_id = int(task["ID"])
        try:
            config_path = resolve_config_path(task, self.config.workspace_root)
            if not config_path:
                raise RuntimeError("task config path is empty")
            if not config_path.exists():
                raise RuntimeError(f"task config path does not exist: {config_path}")

            expected_duration = extract_expected_duration(task)
            result = run_sumo_task(
                self.config.sumo_bin,
                str(config_path),
                self.config.work_dir,
                task_id,
                expected_duration_sec=expected_duration,
                progress_callback=lambda progress: self.report_progress(task_id, progress),
            )
            if result.ok:
                self.client.post(
                    f"/worker/{self.worker_id}/tasks/{task_id}/complete",
                    {"logUrl": result.log_path},
                )
                print(f"task {task_id} completed")
            else:
                self.client.post(
                    f"/worker/{self.worker_id}/tasks/{task_id}/fail",
                    {"error": result.error or "sumo failed", "logUrl": result.log_path},
                )
                print(f"task {task_id} failed: {result.error}")
        except Exception as exc:  # noqa: BLE001
            message = str(exc)
            try:
                self.client.post(
                    f"/worker/{self.worker_id}/tasks/{task_id}/fail",
                    {"error": message, "logUrl": ""},
                )
            except Exception as fail_exc:  # noqa: BLE001
                print(f"report fail task {task_id} failed: {fail_exc}")
            print(f"task {task_id} failed before execution: {message}")
        finally:
            self.change_running_tasks(-1)

    def current_running_tasks(self) -> int:
        with self.running_lock:
            return self.running_tasks

    def change_running_tasks(self, delta: int) -> None:
        with self.running_lock:
            self.running_tasks = max(0, self.running_tasks + delta)

    def report_progress(self, task_id: int, progress: int) -> None:
        try:
            self.client.post(
                f"/worker/{self.worker_id}/tasks/{task_id}/progress",
                {"progress": progress},
            )
        except Exception as exc:  # noqa: BLE001
            print(f"progress update failed for task {task_id}: {exc}")


def resolve_config_path(task: dict, workspace_root: Path) -> Path:
    config = task.get("Config") or task.get("config") or {}
    if isinstance(config, dict):
        for key in ("ConfigPath", "configPath"):
            value = config.get(key)
            if value:
                raw_path = Path(str(value))
                if raw_path.is_absolute():
                    return raw_path
                candidates = [
                    (workspace_root / raw_path).resolve(),
                    (workspace_root / "simulation_api" / raw_path).resolve(),
                ]
                for candidate in candidates:
                    if candidate.exists():
                        return candidate
                return candidates[0]
    return Path()


def extract_expected_duration(task: dict) -> float:
    for key in ("DurationSeconds", "durationSeconds"):
        value = task.get(key)
        if value is not None:
            try:
                return max(1.0, float(value))
            except (TypeError, ValueError):
                return 60.0
    return 60.0


def main() -> None:
    args = parse_args()
    agent = WorkerAgent(load_config(args.config))
    agent.run()


if __name__ == "__main__":
    main()
