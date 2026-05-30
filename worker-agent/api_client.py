from __future__ import annotations

import json
from typing import Any
from urllib import error, request


class ApiClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url.rstrip("/")
        self.token = token

    def post(self, path: str, payload: dict[str, Any] | None = None, allow_no_content: bool = False) -> Any:
        return self._request("POST", path, payload, allow_no_content=allow_no_content)

    def get(self, path: str) -> Any:
        return self._request("GET", path, None)

    def _request(
        self,
        method: str,
        path: str,
        payload: dict[str, Any] | None,
        allow_no_content: bool = False,
    ) -> Any:
        body = None
        headers = {
            "Authorization": self.token,
            "Accept": "application/json",
        }
        if payload is not None:
            body = json.dumps(payload).encode("utf-8")
            headers["Content-Type"] = "application/json"

        req = request.Request(f"{self.base_url}{path}", data=body, headers=headers, method=method)
        try:
            with request.urlopen(req, timeout=30) as response:
                status = response.status
                text = response.read().decode("utf-8").strip()
        except error.HTTPError as exc:
            if allow_no_content and exc.code == 204:
                return None
            detail = exc.read().decode("utf-8", errors="ignore")
            raise RuntimeError(f"{method} {path} failed with {exc.code}: {detail}") from exc
        except error.URLError as exc:
            raise RuntimeError(f"{method} {path} failed: {exc.reason}") from exc
        except OSError as exc:
            raise RuntimeError(f"{method} {path} failed: {exc}") from exc

        if allow_no_content and status == 204:
            return None
        if not text:
            return None
        return json.loads(text)
