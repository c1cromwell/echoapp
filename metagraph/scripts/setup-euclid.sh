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
# Must be the SAME checkout the Makefile runs hydra from (EUCLID_DIR, default: a
# sibling of the echoapp repo). These used to disagree — this script prepared
# echoapp/euclid-development-environment while `make dev` built the sibling,
# so the cluster silently ran the stock Euclid template instead of Echo's code.
EUCLID_DIR="${EUCLID_DIR:-$PROJECT_DIR/../../euclid-development-environment}"
if [ ! -d "$EUCLID_DIR" ]; then
  echo "Cloning Euclid Development Environment..."
  # Echo's own fork, pinned to the tag in euclid.json (docs/OWN_L0_PLAN.md stage 1).
  EUCLID_REPO_URL="${EUCLID_REPO_URL:-https://github.com/c1cromwell/euclid-development-environment}"
  EUCLID_TAG="$(jq -r '.github.tag // "v0.19.0"' "$PROJECT_DIR/euclid.json")"
  git clone --branch "$EUCLID_TAG" "$EUCLID_REPO_URL" "$EUCLID_DIR"
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

# 4. Generate hydra's euclid.json from ours. hydra (>= 0.4) needs `layers` as a
#    list, `framework.name`/`framework.modules` as the g8 template selectors and
#    `ref_type`; metagraph/euclid.json keeps its own shape for the skeleton
#    verifier and deploy docs, so translate rather than copy.
jq '
  .version = (.github.tag // "v0.19.0" | ltrimstr("v"))
  | .ref_type = "tag"
  | .framework = {name: "currency", modules: ["data"], version: .framework.version, ref_type: "tag"}
  | .layers = ["global-l0", "metagraph-l0", "currency-l1", "data-l1"]
  | .docker = {start_grafana_container: false}
  | del(.github)
' "$PROJECT_DIR/euclid.json" > "$EUCLID_DIR/euclid.json"
echo "✓ Wrote $EUCLID_DIR/euclid.json (project_name=$(jq -r .project_name "$EUCLID_DIR/euclid.json"))"

# 5. Sync the metagraph source into Euclid's Docker build context. It has to be
#    a real copy: the image build does `COPY project/<name>`, and Docker does not
#    follow a symlink that points outside the build context.
PROJECT_NAME="$(jq -r .project_name "$PROJECT_DIR/euclid.json")"
if [ -z "$PROJECT_NAME" ] || [ "$PROJECT_NAME" = "null" ]; then
  echo "ERROR: project_name is missing from $PROJECT_DIR/euclid.json" >&2
  exit 1
fi
SOURCE_DIR="$EUCLID_DIR/source/project/$PROJECT_NAME"
if [ -L "$SOURCE_DIR" ]; then
  rm "$SOURCE_DIR"
fi
mkdir -p "$SOURCE_DIR"
rsync -a --delete \
  --exclude 'target/' --exclude '.bsp/' --exclude '.metals/' --exclude '.bloop/' \
  --exclude 'io/' --exclude 'scripts/' \
  "$PROJECT_DIR/" "$SOURCE_DIR/"
echo "✓ Synced metagraph source → $SOURCE_DIR"

# 6. Echo's overlay on the Euclid checkout: the node image build takes Tessellation binaries
#    from Echo's vendored, checksum-verified copies instead of Constellation's GitHub.
rsync -a "$PROJECT_DIR/euclid-overlay/" "$EUCLID_DIR/"
TESS_TAG="v$(jq -r .tessellation_version "$PROJECT_DIR/euclid.json")"
"$PROJECT_DIR/vendor/fetch.sh" "$TESS_TAG"
mkdir -p "$EUCLID_DIR/infra/metagraph-ubuntu/vendor/$TESS_TAG"
rsync -a --delete --exclude "*.tar.gz" "$PROJECT_DIR/vendor/tessellation/$TESS_TAG/" "$EUCLID_DIR/infra/metagraph-ubuntu/vendor/$TESS_TAG/"
echo "✓ Applied Echo overlay and vendored Tessellation $TESS_TAG"

# `hydra install` is deliberately not run: its only jobs are to generate the stock template
# project from Constellation's GitHub (Echo's source is synced above instead) and to re-init git.

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
