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

# 2. Verify Docker has enough memory (need ≥8GB)
DOCKER_MEM=$(docker info --format '{{.MemTotal}}' 2>/dev/null || echo "0")
DOCKER_MEM_GB=$((DOCKER_MEM / 1073741824))
if [ "$DOCKER_MEM_GB" -lt 8 ]; then
  echo "WARNING: Docker has ${DOCKER_MEM_GB}GB RAM allocated. Euclid needs ≥8GB."
  echo "  → Docker Desktop → Preferences → Resources → Memory → 8GB+"
fi

# 3. Clone Euclid dev environment if not present
EUCLID_DIR="$PROJECT_DIR/../euclid-development-environment"
if [ ! -d "$EUCLID_DIR" ]; then
  echo "Cloning Euclid Development Environment..."
  git clone https://github.com/Constellation-Labs/euclid-development-environment "$EUCLID_DIR"
fi

cd "$EUCLID_DIR"

# 4. Copy our euclid.json config
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
