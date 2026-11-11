#!/usr/bin/env bash
# scripts/metagraph/gen-prod-keys.sh
# Generate PRODUCTION p12 node keys for the Constellation metagraph and print their peer IDs.
#
# ⚠️  These are PRODUCTION signing keys. Treat them like root credentials:
#     - Use a STRONG, unique password (not the "password" default local dev uses).
#     - Back the .p12 files up securely (offline). Losing them can mean losing the node identity.
#     - For MainNet you must register each printed peer ID on the metagraph seedlist, and the token
#       treasury must meet the DAG staking requirement, BEFORE `hydra remote-start`.
#     See docs/ZERO_TO_PRODUCTION.md §4.
#
# Usage:
#   CL_PASSWORD='<strong-pass>' scripts/metagraph/gen-prod-keys.sh [--euclid <dir>] [--out <dir>] [--count 3]
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
EUCLID_DIR="${EUCLID_DIR:-$REPO_DIR/../euclid-development-environment}"
OUT_DIR=""
COUNT=3

while [ $# -gt 0 ]; do
  case "$1" in
    --euclid) EUCLID_DIR="$2"; shift 2;;
    --out)    OUT_DIR="$2"; shift 2;;
    --count)  COUNT="$2"; shift 2;;
    *) echo "Unknown arg: $1"; exit 1;;
  esac
done

OUT_DIR="${OUT_DIR:-$EUCLID_DIR/source/p12-files}"
JARS="$EUCLID_DIR/infra/shared/jars"
KEYTOOL="$JARS/cl-keytool.jar"
WALLET="$JARS/cl-wallet.jar"

command -v java >/dev/null || { echo "ERROR: java (JDK 21) is required."; exit 1; }
if [ ! -f "$KEYTOOL" ] || [ ! -f "$WALLET" ]; then
  echo "ERROR: Constellation jars not found in $JARS"
  echo "  Run 'scripts/hydra install' inside the Euclid clone first (it downloads cl-keytool.jar / cl-wallet.jar)."
  exit 1
fi
if [ -z "${CL_PASSWORD:-}" ]; then
  read -r -s -p "Enter a STRONG keystore password: " CL_PASSWORD; echo
  [ -n "$CL_PASSWORD" ] || { echo "ERROR: empty password."; exit 1; }
fi

mkdir -p "$OUT_DIR"
echo "Generating $COUNT production node key(s) into $OUT_DIR"
echo

for i in $(seq 1 "$COUNT"); do
  if [ "$i" -eq 1 ]; then base="token-key"; else base="token-key-$((i-1))"; fi
  keystore="$OUT_DIR/${base}.p12"
  alias="$base"

  if [ -f "$keystore" ]; then
    echo "⚠️  $keystore already exists — refusing to overwrite. Move it aside first."; continue
  fi

  CL_KEYSTORE="$keystore" CL_KEYALIAS="$alias" CL_PASSWORD="$CL_PASSWORD" \
    java -jar "$KEYTOOL" generate
  peer_id="$(CL_KEYSTORE="$keystore" CL_KEYALIAS="$alias" CL_PASSWORD="$CL_PASSWORD" \
    java -jar "$WALLET" show-id 2>/dev/null | tail -1)"

  echo "  node-$i:"
  echo "    keystore : $keystore  (alias: $alias)"
  echo "    peer id  : ${peer_id:-<run cl-wallet show-id manually>}"
  echo
done

cat <<'EOF'
Next:
  1. Confirm metagraph/euclid.json nodes[].key_file names/aliases match the files above,
     and set a real password (do NOT keep "password") — see docs/ZERO_TO_PRODUCTION.md §4.
  2. Back up the .p12 files offline.
  3. MainNet only: register the peer IDs on the seedlist and ensure DAG staking is satisfied
     BEFORE hydra remote-start, or the network will reject your snapshots.
EOF
