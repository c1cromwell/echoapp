#!/usr/bin/env bash
# metagraph/scripts/setup-euclid.sh — Installs Euclid SDK dependencies and builds local cluster.
# Usage: cd metagraph && ./scripts/setup-euclid.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== Echo Metagraph — Euclid SDK Setup ==="

# 1. Check prerequisites
check_cmd() {
  if ! command -v "$1" &>/dev/null; then
    echo "ERROR: $1 is required but not installed."
    echo "  Install: $2"
    exit 1
  fi
}

check_cmd docker "https://docs.docker.com/get-docker/"
check_cmd java "brew install --cask temurin (JDK 21)"
check_cmd sbt "brew install sbt"
check_cmd scala "cs install scala:2.13.10 (via Coursier)"
check_cmd jq "brew install jq"
check_cmd yq "brew install yq"
check_cmd argc "cargo install argc"

echo "✓ All prerequisites found"

# 2. Verify enough RAM (need ≥8GB)
# Docker Desktop (macOS/Windows) reports via MemTotal; Docker CE on Linux always
# returns 0 for that field — fall back to reading /proc/meminfo directly.
if [ "$(uname)" = "Darwin" ]; then
  DOCKER_MEM=$(docker info --format '{{.MemTotal}}' 2>/dev/null | tr -d '[:space:]')
  DOCKER_MEM=${DOCKER_MEM:-0}
  DOCKER_MEM_GB=$((DOCKER_MEM / 1073741824))
else
  DOCKER_MEM_GB=$(awk '/MemTotal/{printf "%d", $2/1048576}' /proc/meminfo)
fi
if [ "${DOCKER_MEM_GB:-0}" -lt 8 ]; then
  echo "WARNING: ${DOCKER_MEM_GB}GB RAM detected. Euclid needs ≥8GB."
  if [ "$(uname)" = "Darwin" ]; then
    echo "  → Docker Desktop → Preferences → Resources → Memory → 8GB+"
  else
    echo "  → Ensure the host machine has at least 8GB RAM."
  fi
fi

# 3. Clone Euclid dev environment if not present
EUCLID_DIR="$PROJECT_DIR/../euclid-development-environment"
if [ ! -d "$EUCLID_DIR" ]; then
  echo "Cloning Euclid Development Environment..."
  git clone https://github.com/Constellation-Labs/euclid-development-environment "$EUCLID_DIR"
fi

cd "$EUCLID_DIR"

# 3b. Verify the remote-deploy Ansible playbooks are present. They ship with the upstream clone and are
# required by euclid.json -> deploy.ansible for `hydra remote-deploy`/`remote-start` (production).
# If a prior operation removed them from the working tree, restore from git so prod deploys stay runnable.
# See docs/ZERO_TO_PRODUCTION.md §4 (metagraph → integrationnet/mainnet).
if [ ! -d "$EUCLID_DIR/infra/ansible/remote" ]; then
  echo "WARNING: infra/ansible/remote is missing from the Euclid checkout (needed for remote deploy)."
  if git -C "$EUCLID_DIR" rev-parse --git-dir &>/dev/null; then
    echo "  → Restoring from git..."
    git -C "$EUCLID_DIR" checkout -- infra/ansible 2>/dev/null || \
      git -C "$EUCLID_DIR" restore infra/ansible 2>/dev/null || true
  fi
  if [ -d "$EUCLID_DIR/infra/ansible/remote" ]; then
    echo "✓ Restored infra/ansible from git"
  else
    echo "  → Could not restore. Local dev is unaffected, but remote metagraph deploy will not work"
    echo "    until these playbooks exist. Re-clone euclid-development-environment to recover them."
  fi
fi

# 4. Copy our euclid.json config (carries the deploy/nodes/snapshot_fees blocks for remote deploy)
cp "$PROJECT_DIR/euclid.json" "$EUCLID_DIR/euclid.json"
echo "✓ Copied euclid.json"

# 5. Link our metagraph source
SOURCE_DIR="$EUCLID_DIR/source/project/echo-metagraph"
if [ ! -d "$SOURCE_DIR" ]; then
  mkdir -p "$(dirname "$SOURCE_DIR")"
  ln -sf "$PROJECT_DIR" "$SOURCE_DIR"
  echo "✓ Linked metagraph source → $SOURCE_DIR"
fi

# 6. Install (downloads Tessellation framework JARs)
echo ""
echo "Running hydra install..."
echo "  This downloads Tessellation ${TESSELLATION_VERSION:-4.0.0-rc.0} framework JARs."
echo "  First run may take several minutes."
scripts/hydra install || {
  echo "NOTE: hydra install may fail if project already installed. Continuing..."
}

echo ""
echo "=== Setup complete ==="
echo ""
echo "Next steps:"
echo "  1. Build containers:    cd $EUCLID_DIR && scripts/hydra build"
echo "  2. Start cluster:       scripts/hydra start-genesis"
echo "  3. Check status:        scripts/hydra status"
echo ""
echo "Cluster endpoints (after start-genesis):"
echo "  Global L0:    http://localhost:9000/node/info"
echo "  Metagraph L0: http://localhost:9200/node/info"
echo "  Currency L1:  http://localhost:9300/node/info"
echo "  Data L1:      http://localhost:9400/node/info"
echo ""
echo "Identity Metagraph (SEPARATE metagraph, NOT managed by hydra):"
echo "  Start with:   make start-identity   (needs sbt assembly JARs)"
echo "  Identity L0:  http://localhost:9600/node/info  (VC / trust tier / StatusList2021 anchor)"
echo "  Identity L1:  http://localhost:9500/node/info  (VC / StatusList / org-role + device-key L1 submissions)"
