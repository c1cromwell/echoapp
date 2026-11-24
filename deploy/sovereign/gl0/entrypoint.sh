#!/usr/bin/env bash
# Echo sovereign Global L0 node. One script for all nodes; GL0_ROLE selects genesis or validator.
# Environment `dev` is deliberate: see docs/OWN_L0_PLAN.md §9 (spike S2). The closed validator
# set is enforced by the seedlist, which Tessellation checks in every environment.
set -euo pipefail

: "${GL0_NAME:?}" "${GL0_IP:?}" "${GL0_ROLE:?}" "${GL0_NODE_COUNT:?}" "${CL_PASSWORD:?}"
SHARED=/shared
WORK=/data/$GL0_NAME
JARS=/code/global-l0
mkdir -p "$WORK" "$SHARED/ids" "$SHARED/addresses"
cd "$WORK"

export CL_KEYSTORE="$WORK/node.p12" CL_KEYALIAS="${CL_KEYALIAS:-echo-gl0}"
export CL_PUBLIC_HTTP_PORT=9000 CL_P2P_HTTP_PORT=9001 CL_CLI_HTTP_PORT=9002
export CL_APP_ENV=dev CL_COLLATERAL=0

if [ ! -f node.p12 ]; then
  java -jar $JARS/cl-keytool.jar generate >/dev/null
fi
java -jar $JARS/cl-wallet.jar show-id | tail -1 > "$SHARED/ids/$GL0_NAME"
java -jar $JARS/cl-wallet.jar show-address | tail -1 > "$SHARED/addresses/$GL0_NAME"

# The seedlist must be identical on every node, so wait until all nodes have published an id.
until [ "$(ls "$SHARED/ids" | wc -l)" -ge "$GL0_NODE_COUNT" ]; do sleep 1; done
sort "$SHARED"/ids/* > seedlist

info() { curl -fsS "http://$1:9000/node/info" 2>/dev/null || true; }
state() { info "$1" | sed -n 's/.*"state":"\([A-Za-z]*\)".*/\1/p'; }

if [ "$GL0_ROLE" = genesis ] && [ ! -d data ]; then
  # First boot only. Every validator address gets a balance so fees, if ever enabled, can be paid.
  for f in "$SHARED"/addresses/*; do echo "$(cat "$f"),${GL0_GENESIS_BALANCE:-100000000000000}"; done > genesis.csv
  exec java -jar $JARS/global-l0.jar run-genesis genesis.csv --ip "$GL0_IP" \
    --seedlist seedlist --prioritySeedlist seedlist
fi

# Validator, or a genesis node restarting with existing data: join through any Ready peer.
java -jar $JARS/global-l0.jar run-validator --ip "$GL0_IP" \
  --seedlist seedlist --prioritySeedlist seedlist &
PID=$!

(
  until [ "$(state "$GL0_IP")" = ReadyToJoin ]; do sleep 2; done
  while [ "$(state "$GL0_IP")" = ReadyToJoin ]; do
    for peer in $GL0_PEERS; do
      name=${peer%%=*}; ip=${peer##*=}
      [ "$name" = "$GL0_NAME" ] && continue
      [ "$(state "$ip")" = Ready ] || continue
      curl -fsS -X POST "http://127.0.0.1:9002/cluster/join" -H 'Content-Type: application/json' \
        -d "{\"id\":\"$(cat "$SHARED/ids/$name")\",\"ip\":\"$ip\",\"p2pPort\":9001}" && break
    done
    sleep 5
  done
) &
wait $PID
