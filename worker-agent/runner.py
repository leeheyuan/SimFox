from __future__ import annotations

import subprocess
import threading
from dataclasses import dataclass
from pathlib import Path
from time import monotonic, sleep


@dataclass
class RunResult:
    ok: bool
    log_path: str
    error: str = ""


def run_sumo_task(
    sumo_bin: str,
    config_path: str,
    work_dir: Path,
    task_id: int,
    expected_duration_sec: float,
    progress_callback=None,
) -> RunResult:
    work_dir.mkdir(parents=True, exist_ok=True)
    task_dir = work_dir / f"task-{task_id}"
    task_dir.mkdir(parents=True, exist_ok=True)
    log_path = task_dir / "sumo.log"

    command = [sumo_bin, "-c", config_path]
    started = monotonic()
    with log_path.open("w", encoding="utf-8") as log_file:
        try:
            process = subprocess.Popen(
                command,
                cwd=task_dir,
                stdout=log_file,
                stderr=subprocess.STDOUT,
                text=True,
            )
        except FileNotFoundError as exc:
            return RunResult(ok=False, log_path=str(log_path), error=f"SUMO executable not found: {sumo_bin}")
        except Exception as exc:  # noqa: BLE001
            return RunResult(ok=False, log_path=str(log_path), error=str(exc))

        progress_thread = None
        stop_event = threading.Event()
        if progress_callback is not None and expected_duration_sec > 0:
            progress_thread = threading.Thread(
                target=_report_progress,
                args=(started, expected_duration_sec, progress_callback, stop_event),
                daemon=True,
            )
            progress_thread.start()

        completed = process.wait()
        stop_event.set()
        if progress_thread is not None:
            progress_thread.join(timeout=2)

    elapsed = monotonic() - started
    if completed != 0:
        return RunResult(
            ok=False,
            log_path=str(log_path),
            error=f"SUMO exited with code {completed} after {elapsed:.1f}s",
        )

    return RunResult(ok=True, log_path=str(log_path))


def _report_progress(started: float, expected_duration_sec: float, progress_callback, stop_event: threading.Event) -> None:
    last_progress = -1
    while not stop_event.is_set():
        elapsed = monotonic() - started
        progress = int(min(95, max(1, (elapsed / expected_duration_sec) * 100)))
        if progress != last_progress:
            try:
                progress_callback(progress)
                last_progress = progress
            except Exception:
                pass
        sleep(2)
