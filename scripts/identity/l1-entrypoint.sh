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
# Role: "initial" (default) starts the L1 cluster; "validator" joins it. A data
# L1 needs 3 nodes — a lone node accepts POST /data but never forms a block
# ("Cannot start data own consensus: Not enough peers"), so nothing reaches L0.
ROLE="${IDENTITY_L1_ROLE:-initial}"
WORKDIR="$SHARED/${IDENTITY_L1_WORKDIR:-l1-workdir}"
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
# identity.p12 is the identity-l0 node key (and the initial L1's key).
export CL_L0_PEER_ID="$(java -jar cl-wallet.jar show-id)"
LEAD_L1_ID="$CL_L0_PEER_ID"

if [ "$ROLE" = "validator" ]; then
  # Each validator needs its own node key.
  export CL_KEYSTORE="$SHARED/${IDENTITY_L1_WORKDIR}.p12"
  export CL_KEYALIAS="${IDENTITY_L1_WORKDIR}"
  if [ ! -f "$CL_KEYSTORE" ]; then
    echo "[identity-l1] Generating validator key at $CL_KEYSTORE…"
    java -jar cl-keytool.jar generate
  fi
fi
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

if [ "$ROLE" = "validator" ]; then
  LEAD_IP="${IDENTITY_L1_LEAD_IP:-172.50.0.51}"
  LEAD_PUBLIC="${IDENTITY_L1_LEAD_PUBLIC_PORT:-9500}"
  LEAD_P2P="${IDENTITY_L1_LEAD_P2P_PORT:-9501}"
  echo "[identity-l1] Starting run-validator on ${SELF_IP}, joining ${LEAD_IP}:${LEAD_P2P}…"
  java -jar identity-l1.jar run-validator --ip "$SELF_IP" &
  JAVA_PID=$!
  trap 'kill "$JAVA_PID" 2>/dev/null || true' TERM INT
  # Join once the lead is Ready and this node is ReadyToJoin.
  for i in $(seq 1 120); do
    lead="$(curl -s -m 3 "http://${LEAD_IP}:${LEAD_PUBLIC}/node/info" | grep -o '"state":"[A-Za-z]*"' || true)"
    self="$(curl -s -m 3 "http://localhost:${CL_PUBLIC_HTTP_PORT}/node/info" | grep -o '"state":"[A-Za-z]*"' || true)"
    # Join one at a time: a joiner only learns the peers the lead already has,
    # so validators that join concurrently never discover each other and every
    # consensus round stalls with "Not enough peers".
    ready_peers="$(curl -s -m 3 "http://${LEAD_IP}:${LEAD_PUBLIC}/cluster/info" | grep -o '"state":"Ready"' | wc -l | tr -d ' ' || true)"
    if [ "$lead" = '"state":"Ready"' ] && [ "$self" = '"state":"ReadyToJoin"' ] \
       && [ "${ready_peers:-0}" -ge "${IDENTITY_L1_JOIN_AFTER_READY:-1}" ]; then
      curl -s -m 10 -X POST "http://localhost:${CL_CLI_HTTP_PORT}/cluster/join" \
        -H 'Content-Type: application/json' \
        -d "{\"id\":\"${LEAD_L1_ID}\",\"ip\":\"${LEAD_IP}\",\"p2pPort\":${LEAD_P2P}}" \
        && { echo "[identity-l1] join requested"; break; } \
        || echo "[identity-l1] join request failed; retrying"
    fi
    kill -0 "$JAVA_PID" 2>/dev/null || break
    sleep 3
  done
  wait "$JAVA_PID"
  exit $?
fi
echo "[identity-l1] Starting run-initial-validator (foreground) on ${SELF_IP}…"
exec java -jar identity-l1.jar run-initial-validator --ip "$SELF_IP"
