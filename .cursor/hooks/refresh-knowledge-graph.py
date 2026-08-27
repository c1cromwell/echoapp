#!/usr/bin/env python3
"""Rebuild docs/knowledge-graph after relevant file edits. Fail open."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
BUILD = ROOT / "scripts" / "knowledge-graph" / "build.py"

WATCH = (
    "docs/phase-",
    "docs/echo-work-orders",
    "docs/PHASE2_GAP",
    "docs/PHASE4_7",
    "docs/ECHO_MESSAGING_LAUNCH",
    "docs/E2E_TESTING",
    "docs/E2E_TWO_DEVICE",
    "docs/knowledge-graph/",
    "scripts/knowledge-graph/",
)


def paths_from_stdin() -> list[str]:
    raw = sys.stdin.read()
    if not raw.strip():
        return []
    try:
        data = json.loads(raw)
    except json.JSONDecodeError:
        return []
    found: list[str] = []
    if isinstance(data, dict):
        for key in ("file_path", "filePath", "path"):
            val = data.get(key)
            if isinstance(val, str):
                found.append(val)
        files = data.get("files") or data.get("edits") or []
        if isinstance(files, list):
            for item in files:
                if isinstance(item, str):
                    found.append(item)
                elif isinstance(item, dict):
                    for key in ("file_path", "filePath", "path"):
                        val = item.get(key)
                        if isinstance(val, str):
                            found.append(val)
    return found


def relevant(path: str) -> bool:
    norm = path.replace("\\", "/")
    return any(token in norm for token in WATCH)


def main() -> int:
    paths = paths_from_stdin()
    if paths and not any(relevant(p) for p in paths):
        print("{}")
        return 0
    if not BUILD.is_file():
        print("{}")
        return 0
    subprocess.run([sys.executable, str(BUILD)], cwd=str(ROOT), check=False)
    print("{}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
