# echo-local-dev MCP

Cursor MCP server that wraps Echo local dev Makefile targets so agents can check cluster status, run release validation, and probe the backend without memorizing commands.

## Tools

| Tool | Wraps |
|------|--------|
| `cluster_status` | `make dev-status` |
| `run_release_check` | `make release-check` |
| `run_validate_phase1` | `make validate-phase1` |
| `run_metagraph_test` | `make metagraph-test` |
| `run_ios_phase3_tests` | `swift test --filter EchoPhase3Tests` (in `ios/Echo`) |
| `health_backend` | `GET /health` (default `http://localhost:8000`) |

## Setup

Requires **Python 3.10+** (macOS: `brew install python@3.12`). The launcher script picks Homebrew Python when `/usr/bin/python3` is too old.

```bash
cd tools/echo-local-dev-mcp
./run.sh   # creates .venv, installs mcp, starts stdio server
```

Optional with uv:

```bash
cd tools/echo-local-dev-mcp
uv sync
uv run echo-local-dev-mcp
```

## Cursor configuration

Project config lives at [`.cursor/mcp.json`](../../.cursor/mcp.json). After pulling, reload MCP servers in Cursor (Settings → MCP → refresh).

If you prefer a user-global entry instead, add to `~/.cursor/mcp.json`:

```json
"echo-local-dev": {
  "command": "bash",
  "args": ["/absolute/path/to/echoapp/tools/echo-local-dev-mcp/run.sh"],
  "env": {
    "ECHO_REPO_ROOT": "/absolute/path/to/echoapp"
  }
}
```

## Manual smoke test

```bash
cd tools/echo-local-dev-mcp
ECHO_REPO_ROOT=/path/to/echoapp ./run.sh
```

Then invoke tools from Cursor or an MCP client over stdio.

## Notes

- `run_validate_phase1` requires Docker, the Euclid metagraph cluster, and JDK 21 — same constraints as the shell script.
- `run_ios_phase3_tests` may need full Xcode (not Command Line Tools alone) for Preview macros in some targets.
- Set `ECHO_REPO_ROOT` if the server is started outside the repo checkout.
