#!/usr/bin/env bash
# Launcher for one Echo chain node process. The same script runs every layer on every host:
# under systemd on real hosts (deploy/ansible) and in the Docker rehearsal (deploy/sovereign/stack).
# Everything is configured by environment; nothing is discovered from the network except peer state.
#
# Required:
#   ECHO_LAYER     global-l0 | metagraph-l0 | currency-l1 | data-l1 | identity-l0 | identity-l1
#   ECHO_ROLE      lead | validator. The lead creates the layer on first boot (genesis for an L0,
#                  initial validator for an L1). After that every node behaves the same.
#   ECHO_HOME      working directory; data/ and logs/ are created here
#   ECHO_JAR       this layer's jar
#   ECHO_TOOLS     directory holding cl-wallet.jar
#   ECHO_IP        address other nodes reach this node on
#   ECHO_SEEDLIST  file of allowed peer ids, one per line. For a metagraph layer: the nodes of
#                  that metagraph. For global-l0: the Global L0 nodes AND every metagraph node,
#                  because Global L0 accepts a metagraph snapshot only when one of its signers is
#                  on this list. (So a metagraph node key could also join Global L0; Echo holds
#                  all of these keys.)
#   ECHO_PRIORITY_SEEDLIST  the Global L0 node ids only, on every layer. Metagraph layers accept a
#                  global snapshot only when the majority of these nodes agree on it.
#   ECHO_PEERS     the other nodes of this layer: "id@host:publicPort:p2pPort ..." (may be empty)
#   CL_KEYSTORE CL_KEYALIAS CL_PASSWORD CL_PUBLIC_HTTP_PORT CL_P2P_HTTP_PORT CL_CLI_HTTP_PORT
# Every layer except global-l0:
#   CL_GLOBAL_L0_PEER_HTTP_HOST CL_GLOBAL_L0_PEER_HTTP_PORT CL_GLOBAL_L0_PEER_ID
# metagraph-l0 / identity-l0 validators, and every L1:
#   CL_L0_TOKEN_IDENTIFIER   the metagraph id (the lead L0 reads it from genesis.address instead)
# Every L1:
#   CL_L0_PEER_HTTP_HOST CL_L0_PEER_HTTP_PORT CL_L0_PEER_ID
# First boot of a lead:
#   global-l0:                ECHO_GENESIS = genesis.csv
#   metagraph-l0/identity-l0: ECHO_GENESIS = genesis.csv, or a directory already holding
#                             genesis.snapshot and genesis.address (from the key ceremony).
#                             ECHO_OWNER_KEYSTORE / _KEYALIAS / _PASSWORD sign the owner message;
#                             they default to the node key.
# Optional:
#   ECHO_ROLLBACK_HASH     global-l0 only, when no peer is Ready: restart the whole network from
#                          this snapshot hash. Set it on the one node holding the newest snapshot
#   ECHO_JOIN_AFTER_READY  join only once the peer reports this many Ready nodes (default 1)
#   ECHO_JOIN_TIMEOUT      seconds a join may take before the node is stopped for a restart (default 420)
#   ECHO_JAVA_OPTS         extra JVM options (heap, logback configuration)
#   ECHO_ALLOW_OPEN=1      local rehearsal only: permit a metagraph without signer allowlists
#
# CL_APP_ENV is fixed to `dev` and collateral to 0 on purpose: docs/OWN_L0_PLAN.md §10 (S2, S3).
# The closed validator set comes from the seedlist, which Tessellation enforces in every environment.
set -euo pipefail

log() { echo "[echo-node ${ECHO_LAYER:-?}] $*"; }
die() { log "ERROR: $*" >&2; exit 1; }

for v in ECHO_LAYER ECHO_ROLE ECHO_HOME ECHO_JAR ECHO_TOOLS ECHO_IP ECHO_SEEDLIST ECHO_PRIORITY_SEEDLIST \
         CL_KEYSTORE CL_KEYALIAS CL_PASSWORD CL_PUBLIC_HTTP_PORT CL_P2P_HTTP_PORT CL_CLI_HTTP_PORT; do
  [ -n "${!v:-}" ] || die "$v is not set"
done
ECHO_PEERS="${ECHO_PEERS:-}"

case "$ECHO_LAYER" in
  global-l0)                kind=gl0 ;;
  metagraph-l0|identity-l0) kind=l0 ;;
  currency-l1|data-l1|identity-l1) kind=l1 ;;
  *) die "unknown ECHO_LAYER '$ECHO_LAYER'" ;;
esac
case "$ECHO_ROLE" in lead|validator) ;; *) die "ECHO_ROLE must be lead or validator" ;; esac

[ -f "$ECHO_JAR" ] || die "jar not found: $ECHO_JAR"
[ -f "$CL_KEYSTORE" ] || die "keystore not found: $CL_KEYSTORE"
[ -s "$ECHO_SEEDLIST" ] || die "seedlist missing or empty: $ECHO_SEEDLIST"
[ -s "$ECHO_PRIORITY_SEEDLIST" ] || die "priority seedlist missing or empty: $ECHO_PRIORITY_SEEDLIST"

if [ "$kind" != gl0 ]; then
  for v in CL_GLOBAL_L0_PEER_HTTP_HOST CL_GLOBAL_L0_PEER_HTTP_PORT CL_GLOBAL_L0_PEER_ID; do
    [ -n "${!v:-}" ] || die "$v is not set"
  done
fi
if [ "$kind" = l1 ]; then
  for v in CL_L0_PEER_HTTP_HOST CL_L0_PEER_HTTP_PORT CL_L0_PEER_ID CL_L0_TOKEN_IDENTIFIER; do
    [ -n "${!v:-}" ] || die "$v is not set"
  done
fi

# Without these lists anyone who reaches a Data L1 or Identity L1 port can write to the chain.
if [ "${ECHO_ALLOW_OPEN:-0}" != 1 ]; then
  [ "$ECHO_LAYER" != metagraph-l0 ] || [ -n "${DATA_AUTHORIZED_SIGNERS:-}" ] \
    || die "DATA_AUTHORIZED_SIGNERS is not set (the gateway's signing key id)"
  [ "$ECHO_LAYER" != identity-l0 ] || [ -n "${IDENTITY_SERVICE_PUBLIC_KEY_HEX:-}" ] \
    || die "IDENTITY_SERVICE_PUBLIC_KEY_HEX is not set"
fi

export CL_APP_ENV=dev CL_COLLATERAL=0
export CL_STOREPASS="${CL_STOREPASS:-$CL_PASSWORD}" CL_KEYPASS="${CL_KEYPASS:-$CL_PASSWORD}"

mkdir -p "$ECHO_HOME"
cd "$ECHO_HOME"

SELF_ID="$(java -jar "$ECHO_TOOLS/cl-wallet.jar" show-id | tail -1)"
grep -qx "$SELF_ID" "$ECHO_SEEDLIST" || die "this node's id is not on the seedlist $ECHO_SEEDLIST"

# shellcheck disable=SC2086
run() { exec java ${ECHO_JAVA_OPTS:-} -jar "$ECHO_JAR" "$@" --ip "$ECHO_IP" \
          --seedlist "$ECHO_SEEDLIST" --prioritySeedlist "$ECHO_PRIORITY_SEEDLIST"; }

state() { curl -fsS -m 3 "http://$1/node/info" 2>/dev/null | sed -n 's/.*"state":"\([A-Za-z]*\)".*/\1/p'; }
ready_count() { curl -fsS -m 3 "http://$1/cluster/info" 2>/dev/null | grep -o '"state":"Ready"' | wc -l | tr -d ' '; }

# Prints "id host p2pPort" of the first peer that is Ready and has enough Ready nodes around it.
ready_peer() {
  local p id rest host pub p2p
  for p in $ECHO_PEERS; do
    id=${p%%@*}; rest=${p#*@}
    host=${rest%%:*}; rest=${rest#*:}
    pub=${rest%%:*}; p2p=${rest#*:}
    [ "$(state "$host:$pub")" = Ready ] || continue
    [ "$(ready_count "$host:$pub")" -ge "${ECHO_JOIN_AFTER_READY:-1}" ] || continue
    echo "$id $host $p2p"
    return 0
  done
  return 1
}

# An L0 has a chain to resume only once it has written a snapshot. A failed first boot leaves the
# directory behind but empty, and must be allowed to try genesis again.
has_data() {
  [ -n "$(ls -A data/incremental_snapshot 2>/dev/null)" ]
}

# A metagraph node cannot start before the layers beneath it answer.
wait_ready() {
  until [ "$(state "$2")" = Ready ]; do log "waiting for $1 at $2"; sleep 5; done
}
[ "$kind" = gl0 ] || wait_ready "Global L0" "$CL_GLOBAL_L0_PEER_HTTP_HOST:$CL_GLOBAL_L0_PEER_HTTP_PORT"
[ "$kind" != l1 ] || wait_ready "its L0" "$CL_L0_PEER_HTTP_HOST:$CL_L0_PEER_HTTP_PORT"

# A layer that already exists somewhere is always joined, whatever this node's role. Only when no
# peer is Ready does the lead create or restore it.
# Whole-network restore. Any Global L0 node may do it, because it must be the node that holds the
# newest snapshot: metagraph nodes have already seen that snapshot and cannot follow an older one.
if [ "$kind" = gl0 ] && [ -n "${ECHO_ROLLBACK_HASH:-}" ] && has_data && ! ready_peer >/dev/null; then
  log "restoring the network from snapshot $ECHO_ROLLBACK_HASH"
  run run-rollback "$ECHO_ROLLBACK_HASH"
fi

if [ "$ECHO_ROLE" = lead ] && ! ready_peer >/dev/null; then
  case "$kind" in
    gl0)
      if has_data; then
        die "data exists and no Global L0 peer is Ready. Restarting the whole network needs \
ECHO_ROLLBACK_HASH on the node with the newest snapshot (see the rollback playbook); refusing to guess."
      fi
      [ -f "${ECHO_GENESIS:-}" ] || die "first boot needs ECHO_GENESIS (genesis.csv)"
      log "first boot: genesis from $ECHO_GENESIS"
      run run-genesis "$ECHO_GENESIS"
      ;;
    l0)
      if has_data; then
        CL_L0_TOKEN_IDENTIFIER="${CL_L0_TOKEN_IDENTIFIER:-$(tr -d '\r\n' < genesis.address)}"
        export CL_L0_TOKEN_IDENTIFIER
        log "resuming metagraph $CL_L0_TOKEN_IDENTIFIER from its last snapshot"
        run run-rollback
      fi
      [ -e "${ECHO_GENESIS:-}" ] || die "first boot needs ECHO_GENESIS"
      if [ -d "$ECHO_GENESIS" ]; then
        cp "$ECHO_GENESIS/genesis.snapshot" "$ECHO_GENESIS/genesis.address" .
      else
        # create-genesis writes genesis.snapshot beside its input, so work on a copy here.
        log "creating genesis snapshot from $ECHO_GENESIS"
        [ "$ECHO_GENESIS" -ef genesis.csv ] || cp "$ECHO_GENESIS" genesis.csv
        java -jar "$ECHO_JAR" create-genesis genesis.csv >/dev/null
      fi
      metagraph_id="$(tr -d '\r\n' < genesis.address)"
      (
        export CL_KEYSTORE="${ECHO_OWNER_KEYSTORE:-$CL_KEYSTORE}" CL_KEYALIAS="${ECHO_OWNER_KEYALIAS:-$CL_KEYALIAS}"
        export CL_PASSWORD="${ECHO_OWNER_PASSWORD:-$CL_PASSWORD}"
        export CL_STOREPASS="$CL_PASSWORD" CL_KEYPASS="$CL_PASSWORD"
        owner="$(java -jar "$ECHO_TOOLS/cl-wallet.jar" show-address | tail -1)"
        java -jar "$ECHO_TOOLS/cl-wallet.jar" create-owner-signing-message \
          --address "$owner" --metagraphId "$metagraph_id" --parentOrdinal 0 > owner-message
      )
      log "first boot: genesis of metagraph $metagraph_id"
      run run-genesis genesis.snapshot --metagraph-owner-message ./owner-message
      ;;
    l1)
      log "starting the cluster as initial validator"
      run run-initial-validator
      ;;
  esac
fi

if [ "$kind" = l0 ]; then
  [ -n "${CL_L0_TOKEN_IDENTIFIER:-}" ] || CL_L0_TOKEN_IDENTIFIER="$(tr -d '\r\n' < genesis.address 2>/dev/null || true)"
  [ -n "$CL_L0_TOKEN_IDENTIFIER" ] || die "CL_L0_TOKEN_IDENTIFIER is not set"
  export CL_L0_TOKEN_IDENTIFIER
fi

# Join in the background, then become the node process so signals and the exit code are the JVM's.
(
  until [ "$(state "127.0.0.1:$CL_PUBLIC_HTTP_PORT")" = ReadyToJoin ]; do sleep 2; done
  while [ "$(state "127.0.0.1:$CL_PUBLIC_HTTP_PORT")" = ReadyToJoin ]; do
    if peer="$(ready_peer)"; then
      set -- $peer
      curl -fsS -m 10 -X POST "http://127.0.0.1:$CL_CLI_HTTP_PORT/cluster/join" \
        -H 'Content-Type: application/json' -d "{\"id\":\"$1\",\"ip\":\"$2\",\"p2pPort\":$3}" \
        && log "join requested through $2"
    fi
    sleep 5
  done
  # A join can stall for good. The joiner downloads the newest snapshot and then asks a peer for
  # that exact consensus round; if the cluster has moved on by one round the peer answers 409, and
  # Tessellation 4.0.7 retries the same stale round 20 times and then stays in Observing for ever.
  # Nothing recovers it but a restart, so stop the node and let its supervisor start it afresh.
  deadline=$((SECONDS + ${ECHO_JOIN_TIMEOUT:-420}))
  until [ "$(state "127.0.0.1:$CL_PUBLIC_HTTP_PORT")" = Ready ]; do
    if [ "$SECONDS" -ge "$deadline" ]; then
      log "ERROR: not Ready ${ECHO_JOIN_TIMEOUT:-420}s after joining; stopping so the node is restarted"
      kill $$
      exit 1
    fi
    sleep 10
  done
  log "Ready"
) &
log "starting as validator; will join a Ready peer"
run run-validator
