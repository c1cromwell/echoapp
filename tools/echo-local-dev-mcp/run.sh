#!/usr/bin/env bash
# Bootstrap venv (if needed) and start the echo-local-dev MCP server.
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

pick_python() {
  local candidate
  for candidate in \
    /opt/homebrew/bin/python3.14 \
    /opt/homebrew/bin/python3.13 \
    /opt/homebrew/bin/python3.12 \
    /opt/homebrew/bin/python3.11 \
    /opt/homebrew/bin/python3.10 \
    python3.14 python3.13 python3.12 python3.11 python3.10 python3; do
    if command -v "$candidate" >/dev/null 2>&1 \
      && "$candidate" -c 'import sys; raise SystemExit(0 if sys.version_info >= (3, 10) else 1)' 2>/dev/null; then
      command -v "$candidate"
      return 0
    fi
  done
  echo "echo-local-dev MCP requires Python 3.10+ (install via Homebrew: brew install python@3.12)" >&2
  exit 1
}

PYTHON="$(pick_python)"

if [[ ! -d .venv ]]; then
  "$PYTHON" -m venv .venv
fi

.venv/bin/python -m pip install -q --upgrade pip
.venv/bin/pip install -q "mcp>=1.2.0"

exec .venv/bin/python "$DIR/echo_local_dev_mcp/server.py" "$@"
