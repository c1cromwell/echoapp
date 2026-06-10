#!/usr/bin/env bash
#
# Identity Metagraph L0 — standalone genesis/boot entrypoint.
#
# Mirrors Euclid's metagraph-l0 genesis flow for Tessellation 4.0.0-rc.0
# (create-genesis -> owner signing message -> run-genesis) but runs as a
# standalone container that peers with the already-running Euclid Global L0
# (the hydra `make dev` cluster), instead of being templated by hydra ansible.
#
# Reference: euclid-development-environment v0.19.0
#   infra/ansible/local/playbooks/start/metagraph-l0/genesis.ansible.yml
#
set -euo pipefail

SHARED=/shared-identity
WORKDIR="$SHARED/l0-workdir"
SELF_IP="${IDENTITY_L0_IP:-172.50.0.50}"

mkdir -p "$WORKDIR" "$SHARED"
cd "$WORKDIR"

# Stage the assembled fat JAR plus the Constellation CLI tools that ship inside
# the Euclid base image (currency-l1/ always exists for the Echo layer set).
cp "$(ls /jars/*assembly*.jar | head -1)" identity-l0.jar
cp /code/currency-l1/cl-wallet.jar cl-wallet.jar
cp /code/currency-l1/cl-keytool.jar cl-keytool.jar
cp /code/shared_genesis/genesis.csv genesis.csv

# Generate the dedicated Identity owner/node key once, persisted in the shared
# volume so the Identity metagraph id stays stable across restarts. Reusing the
# cluster's token-key.p12 here would fail with AddressAlreadyInUse.
export CL_STOREPASS="$CL_PASSWORD" CL_KEYPASS="$CL_PASSWORD"
if [ ! -f "$CL_KEYSTORE" ]; then
  echo "[identity-l0] Generating dedicated identity key at $CL_KEYSTORE…"
  java -jar cl-keytool.jar generate
fi
echo "[identity-l0] owner/node key:     $CL_KEYSTORE (alias=$CL_KEYALIAS)"

# Dial the live Global L0 (the running hydra cluster). Read its peer id straight
# from /node/info — the Euclid dev keys are regenerated independently of the host
# p12 we mount, so deriving the id from a key is unreliable.
GL0="${CL_GLOBAL_L0_PEER_HTTP_HOST}:${CL_GLOBAL_L0_PEER_HTTP_PORT}"
echo "[identity-l0] Global L0 endpoint:  $GL0"
CL_GLOBAL_L0_PEER_ID="$(curl -s -m 5 "http://${GL0}/node/info" \
  | grep -o '"id":"[0-9a-f]\{128\}"' | head -1 | grep -o '[0-9a-f]\{128\}')"
export CL_GLOBAL_L0_PEER_ID
if [ -z "$CL_GLOBAL_L0_PEER_ID" ]; then
  echo "[identity-l0] ERROR: could not read Global L0 peer id from http://${GL0}/node/info" >&2
  echo "[identity-l0] Is the Euclid dev cluster running ('make dev')?" >&2
  exit 1
fi
echo "[identity-l0] Global L0 peer id:   $CL_GLOBAL_L0_PEER_ID"

if [ -d data/incremental_snapshot ]; then
  export CL_L0_TOKEN_IDENTIFIER="$(tr -d '\r\n' < "$SHARED/identity.address")"
  echo "[identity-l0] Resuming from persisted snapshots (run-rollback)…"
  echo "[identity-l0] identity metagraph id: $CL_L0_TOKEN_IDENTIFIER"
  exec java -jar identity-l0.jar run-rollback --ip "$SELF_IP"
fi

# First boot only — mirrors Euclid metagraph-l0 genesis for Tessellation 4.x.
rm -rf data logs
echo "[identity-l0] Creating genesis from genesis.csv…"
java -jar identity-l0.jar create-genesis genesis.csv

ownerAddress="$(java -jar cl-wallet.jar show-address)"
metagraphId="$(tr -d '\r\n' < genesis.address)"
echo "[identity-l0] owner address:       $ownerAddress"
echo "[identity-l0] identity metagraph id: $metagraphId"

# run-genesis requires a signed owner message in Tessellation 4.x.
java -jar cl-wallet.jar create-owner-signing-message \
  --address "$ownerAddress" \
  --metagraphId "$metagraphId" \
  --parentOrdinal 0 > owner-message

# Publish the genesis address so identity-l1 can pick up CL_L0_TOKEN_IDENTIFIER.
cp genesis.address "$SHARED/identity.address"

echo "[identity-l0] Starting run-genesis (foreground) on ${SELF_IP}…"
exec java -jar identity-l0.jar run-genesis genesis.snapshot \
  --metagraph-owner-message ./owner-message \
  --ip "$SELF_IP"
