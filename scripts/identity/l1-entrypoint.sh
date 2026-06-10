#!/usr/bin/env bash
#
# Identity Metagraph L1 — standalone initial-validator entrypoint.
#
# Mirrors Euclid's currency-l1 initial-validator flow for Tessellation 4.0.0-rc.0
# but runs standalone and peers with the identity-l0 container (this file) and
# the running Euclid Global L0.
#
# Reference: euclid-development-environment v0.19.0
#   infra/ansible/local/playbooks/start/currency-l1/initial-validator.ansible.yml
#
# NOTE (Phase D): runtime peering of L1 is validated separately once a live
# Euclid cluster + identity-l0 genesis are up. The wiring below matches the
# upstream template; iterate here if the L1 fails to reach Ready.
#
set -euo pipefail

SHARED=/shared-identity
WORKDIR="$SHARED/l1-workdir"
SELF_IP="${IDENTITY_L1_IP:-172.50.0.51}"

mkdir -p "$WORKDIR" "$SHARED"
cd "$WORKDIR"

cp "$(ls /jars/*assembly*.jar | head -1)" identity-l1.jar
cp /code/currency-l1/cl-wallet.jar cl-wallet.jar
cp /code/currency-l1/cl-keytool.jar cl-keytool.jar
export CL_STOREPASS="$CL_PASSWORD" CL_KEYPASS="$CL_PASSWORD"

# Wait for identity-l0 to generate the shared key (same key runs L0 + L1).
for i in $(seq 1 120); do
  if [ -f "$CL_KEYSTORE" ]; then break; fi
  echo "[identity-l1] waiting for shared identity key $CL_KEYSTORE ($i/120)…"
  sleep 2
done

# identity-l0 uses the same shared key, so its peer id is `show-id` of that key
# (CL_L0_PEER_ID — the L1's own metagraph L0).
export CL_L0_PEER_ID="$(java -jar cl-wallet.jar show-id)"
echo "[identity-l1] identity-l0 peer id:  $CL_L0_PEER_ID"
echo "[identity-l1] identity-l0 endpoint: ${CL_L0_PEER_HTTP_HOST}:${CL_L0_PEER_HTTP_PORT}"

# The Global L0 id comes from the live cluster, not a key (see l0-entrypoint).
GL0="${CL_GLOBAL_L0_PEER_HTTP_HOST}:${CL_GLOBAL_L0_PEER_HTTP_PORT}"
CL_GLOBAL_L0_PEER_ID="$(curl -s -m 5 "http://${GL0}/node/info" \
  | grep -o '"id":"[0-9a-f]\{128\}"' | head -1 | grep -o '[0-9a-f]\{128\}')"
export CL_GLOBAL_L0_PEER_ID
if [ -z "$CL_GLOBAL_L0_PEER_ID" ]; then
  echo "[identity-l1] ERROR: could not read Global L0 peer id from http://${GL0}/node/info" >&2
  exit 1
fi
echo "[identity-l1] global-l0 peer id:    $CL_GLOBAL_L0_PEER_ID"

# Wait for identity-l0 to publish its genesis address (the Identity metagraph id).
for i in $(seq 1 120); do
  if [ -s "$SHARED/identity.address" ]; then break; fi
  echo "[identity-l1] waiting for identity-l0 genesis address ($i/120)…"
  sleep 2
done
if [ ! -s "$SHARED/identity.address" ]; then
  echo "[identity-l1] ERROR: identity-l0 genesis address never appeared" >&2
  exit 1
fi
export CL_L0_TOKEN_IDENTIFIER="$(tr -d '\r\n' < "$SHARED/identity.address")"
echo "[identity-l1] identity metagraph id: $CL_L0_TOKEN_IDENTIFIER"

if [ ! -d data ]; then
  rm -rf data logs
fi
echo "[identity-l1] Starting run-initial-validator (foreground) on ${SELF_IP}…"
exec java -jar identity-l1.jar run-initial-validator --ip "$SELF_IP"
