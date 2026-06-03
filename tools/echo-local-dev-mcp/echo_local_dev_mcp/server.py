"""MCP server wrapping Echo Makefile and health-check commands for agents."""

from __future__ import annotations

import os
import subprocess
import urllib.error
import urllib.request
from pathlib import Path

from mcp.server.fastmcp import FastMCP

mcp = FastMCP(
    "echo-local-dev",
    instructions=(
        "Local Echo development helpers. Use cluster_status before validate-phase1. "
        "run_validate_phase1 needs Docker, Euclid metagraph, and JDK 21. "
        "run_ios_phase3_tests needs Swift toolchain (full Xcode for some targets)."
    ),
)


def repo_root() -> Path:
    env_root = os.environ.get("ECHO_REPO_ROOT")
    if env_root:
        return Path(env_root).resolve()
    here = Path(__file__).resolve()
    for parent in here.parents:
        if (parent / "Makefile").is_file() and (parent / "scripts" / "validate-phase1.sh").is_file():
            return parent
    raise RuntimeError("Could not locate echoapp repo root (set ECHO_REPO_ROOT)")


def format_result(
    label: str,
    command: list[str],
    cwd: Path,
    result: subprocess.CompletedProcess[str],
) -> str:
    cmd = " ".join(command)
    lines = [
        f"## {label}",
        f"**Command:** `{cmd}`",
        f"**Directory:** `{cwd}`",
        f"**Exit code:** {result.returncode}",
        "",
    ]
    if result.stdout.strip():
        lines.extend(["### stdout", "```", result.stdout.rstrip(), "```", ""])
    if result.stderr.strip():
        lines.extend(["### stderr", "```", result.stderr.rstrip(), "```", ""])
    if result.returncode != 0:
        lines.append("**Result:** failed")
    else:
        lines.append("**Result:** success")
    return "\n".join(lines)


def run_command(
    label: str,
    command: list[str],
    *,
    cwd: Path | None = None,
    timeout: int,
) -> str:
    workdir = cwd or repo_root()
    try:
        result = subprocess.run(
            command,
            cwd=workdir,
            capture_output=True,
            text=True,
            timeout=timeout,
            check=False,
        )
    except subprocess.TimeoutExpired as exc:
        stdout = exc.stdout or ""
        stderr = exc.stderr or ""
        return (
            f"## {label}\n"
            f"**Command:** `{' '.join(command)}`\n"
            f"**Directory:** `{workdir}`\n"
            f"**Result:** timed out after {timeout}s\n\n"
            f"### stdout\n```\n{stdout.rstrip()}\n```\n\n"
            f"### stderr\n```\n{stderr.rstrip()}\n```"
        )
    return format_result(label, command, workdir, result)


@mcp.tool()
def cluster_status() -> str:
    """Show Phase-1 testnet status (Docker backend + metagraph endpoint reachability)."""
    return run_command(
        "Cluster status",
        ["make", "dev-status"],
        timeout=120,
    )


@mcp.tool()
def run_release_check() -> str:
    """Run make release-check (Go build, race tests, vet, gofmt)."""
    return run_command(
        "Release check",
        ["make", "release-check"],
        timeout=900,
    )


@mcp.tool()
def run_validate_phase1() -> str:
    """Run scripts/validate-phase1.sh (WO-230 six-step go/no-go; needs full cluster)."""
    return run_command(
        "Validate Phase 1",
        ["make", "validate-phase1"],
        timeout=1800,
    )


@mcp.tool()
def run_metagraph_test() -> str:
    """Run make metagraph-test (Scala sharedData + identity L0/L1 tests; needs JDK 21)."""
    return run_command(
        "Metagraph tests",
        ["make", "metagraph-test"],
        timeout=900,
    )


@ mcp.tool()
def run_ios_phase3_tests() -> str:
    """Run swift test --filter EchoPhase3Tests in ios/Echo."""
    root = repo_root()
    return run_command(
        "iOS Phase 3 tests",
        ["swift", "test", "--filter", "EchoPhase3Tests"],
        cwd=root / "ios" / "Echo",
        timeout=600,
    )


@mcp.tool()
def run_regression(quick: bool = False, with_phase1: bool = False) -> str:
    """Run scripts/run-regression.sh (Go + iOS SPM). quick=Go only; with_phase1 adds validate-phase1."""
    cmd = ["./scripts/run-regression.sh"]
    if quick:
        cmd.append("--quick")
    if with_phase1:
        cmd.append("--with-phase1")
    return run_command(
        "Regression",
        cmd,
        timeout=1800,
    )


@mcp.tool()
def run_ios_preflight(build: bool = False, tests: bool = False) -> str:
    """Run scripts/ios-e2e-preflight.sh before Xcode manual E2E. Optional build (xcodebuild) and SPM tests."""
    cmd = ["./scripts/ios-e2e-preflight.sh"]
    if build:
        cmd.append("--build")
    if tests:
        cmd.append("--tests")
    return run_command(
        "iOS E2E preflight",
        cmd,
        timeout=1200,
    )


@mcp.tool()
def smoke_ios_backend(base_url: str = "http://localhost:8000") -> str:
    """Probe backend endpoints the iOS app uses (health, SMS register, OIDC4VC start)."""
    base = base_url.rstrip("/")
    lines = ["## iOS backend smoke", f"**Base:** `{base}`", ""]
    checks = [
        ("GET /health", "GET", f"{base}/health", None),
        (
            "POST /v1/auth/sms-recovery/register",
            "POST",
            f"{base}/v1/auth/sms-recovery/register",
            b'{"phone_hash":"sha256:00","phone_raw":"+15550001111","did":"did:key:zTest"}',
        ),
        (
            "POST /v1/enrollment/vc/start",
            "POST",
            f"{base}/v1/enrollment/vc/start",
            b'{"requested_claims":{"givenName":true,"familyName":true,"ageOver18":true}}',
        ),
    ]
    for label, method, url, body in checks:
        req = urllib.request.Request(url, method=method)
        if body is not None:
            req.data = body
            req.add_header("Content-Type", "application/json")
        try:
            with urllib.request.urlopen(req, timeout=8) as response:
                lines.append(f"- **{label}** → {response.status} OK")
                otp = response.headers.get("X-Dev-OTP") or response.headers.get("x-dev-otp")
                if otp and "sms-recovery" in url:
                    lines.append(f"  - `X-Dev-OTP`: {otp[:6]}…")
        except urllib.error.HTTPError as exc:
            lines.append(f"- **{label}** → HTTP {exc.code}")
            otp = exc.headers.get("X-Dev-OTP") or exc.headers.get("x-dev-otp")
            if otp and "sms-recovery" in url:
                lines.append(f"  - `X-Dev-OTP`: {otp[:6]}…")
        except urllib.error.URLError as exc:
            lines.append(f"- **{label}** → unreachable ({exc.reason})")
    lines.append("")
    lines.append("Run `make ios-preflight` for full Xcode + scheme checks.")
    return "\n".join(lines)


@mcp.tool()
def health_backend(base_url: str = "http://localhost:8000") -> str:
    """GET /health on the Echo backend (default http://localhost:8000)."""
    url = base_url.rstrip("/") + "/health"
    request = urllib.request.Request(url, method="GET")
    try:
        with urllib.request.urlopen(request, timeout=5) as response:
            body = response.read().decode("utf-8", errors="replace")
            return (
                f"## Backend health\n"
                f"**URL:** `{url}`\n"
                f"**Status:** {response.status}\n\n"
                f"### body\n```\n{body.rstrip()}\n```"
            )
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        return (
            f"## Backend health\n"
            f"**URL:** `{url}`\n"
            f"**Status:** {exc.code}\n\n"
            f"### body\n```\n{body.rstrip()}\n```"
        )
    except urllib.error.URLError as exc:
        return (
            f"## Backend health\n"
            f"**URL:** `{url}`\n"
            f"**Result:** unreachable ({exc.reason})\n\n"
            "Start the stack with `make dev` or `make testnet-up`."
        )


def main() -> None:
    mcp.run()


if __name__ == "__main__":
    main()
