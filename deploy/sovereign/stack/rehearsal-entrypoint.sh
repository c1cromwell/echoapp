#!/usr/bin/env bash
# Rehearsal only. Does on one machine what the key ceremony and Ansible do for real hosts:
# makes a key per host, publishes ids, builds the seedlists and genesis files, then hands over to
# the production launcher (deploy/sovereign/node/echo-node.sh). Keys here are throwaway.
#
#   HOST   gl0-1..3 | mg-1..3      LAYER  as ECHO_LAYER      IP  this container's address
set -euo pipefail
: "${HOST:?}" "${LAYER:?}" "${IP:?}" "${CL_PASSWORD:?}"
S=/shared
TOOLS=/code/global-l0
N=${HOST##*-}
mkdir -p $S/keys $S/ids $S/addresses $S/genesis $S/lock
wallet() { java -jar $TOOLS/cl-wallet.jar "$@" | tail -1; }
export CL_STOREPASS="$CL_PASSWORD" CL_KEYPASS="$CL_PASSWORD"
# Same environment as the launcher. Without it the tools assume Constellation's testnet rules.
export CL_APP_ENV=dev CL_COLLATERAL=0

# Several containers share one host key; mkdir is the lock.
once() { # once <name> <command...>: run the command in exactly one container, others wait for it
  local name=$1; shift
  if mkdir "$S/lock/$name" 2>/dev/null; then
    if "$@"; then touch "$S/lock/$name/done"; else rmdir "$S/lock/$name"; exit 1; fi
  fi
  until [ -f "$S/lock/$name/done" ]; do sleep 1; done
}
make_key() { # make_key <name>
  ( cd $S/keys && CL_KEYSTORE="$1.p12" CL_KEYALIAS="$1" java -jar $TOOLS/cl-keytool.jar generate >/dev/null
    # Write then rename: other containers act on these files the moment they exist.
    CL_KEYSTORE="$1.p12" CL_KEYALIAS="$1" wallet show-id > "$S/ids/.$1" && mv "$S/ids/.$1" "$S/ids/$1"
    CL_KEYSTORE="$1.p12" CL_KEYALIAS="$1" wallet show-address > "$S/addresses/.$1" && mv "$S/addresses/.$1" "$S/addresses/$1" )
}
once "key-$HOST" make_key "$HOST"
until [ "$(ls $S/addresses | grep -c -E '^(gl0|mg)-[0-9]+$')" -ge 6 ]; do sleep 1; done
# Global L0 admits metagraph snapshots only from signers on its own seedlist, so that list holds
# all six hosts; the priority list (whose majority metagraph nodes trust) is the Global L0 nodes.
once seedlists bash -c "cat $S/ids/gl0-* | sort > $S/gl0.priority; cat $S/ids/mg-* | sort > $S/mg.seedlist; cat $S/ids/gl0-* $S/ids/mg-* | sort > $S/gl0.seedlist"
export ECHO_PRIORITY_SEEDLIST=$S/gl0.priority

export ECHO_LAYER="$LAYER" ECHO_IP="$IP" ECHO_TOOLS=$TOOLS ECHO_HOME="/data/$HOST/$LAYER"
export CL_KEYSTORE="$S/keys/$HOST.p12" CL_KEYALIAS="$HOST"
[ "$N" = 1 ] && export ECHO_ROLE=lead || export ECHO_ROLE=validator
# Validators join one at a time: node 3 waits until the lead already has node 2.
export ECHO_JOIN_AFTER_READY=$((N - 1))

peers() { # peers <host prefix> <ip prefix> <public port> <p2p port>
  local i out=""
  for i in 1 2 3; do
    [ "$1-$i" = "$HOST" ] || out="$out $(cat "$S/ids/$1-$i")@$2$i:$3:$4"
  done
  echo "$out"
}
# The Global L0 node a metagraph host talks to: the one with the same number.
gl0() {
  export CL_GLOBAL_L0_PEER_HTTP_HOST="172.61.0.1$N" CL_GLOBAL_L0_PEER_HTTP_PORT=9000
  export CL_GLOBAL_L0_PEER_ID="$(cat "$S/ids/gl0-$N")"
}
# An L1 node uses the L0 of its own layer on its own host.
l0_peer() { # l0_peer <l0 public port> <address file>
  export CL_L0_PEER_HTTP_HOST="$IP" CL_L0_PEER_HTTP_PORT="$1" CL_L0_PEER_ID="$(cat "$S/ids/$HOST")"
  until [ -s "$2" ]; do sleep 1; done
  export CL_L0_TOKEN_IDENTIFIER="$(tr -d '\r\n' < "$2")"
}
ports() { export CL_PUBLIC_HTTP_PORT=$1 CL_P2P_HTTP_PORT=$(($1 + 1)) CL_CLI_HTTP_PORT=$(($1 + 2)); }
# Mirrors the ceremony: the genesis snapshot is created once, away from the running node.
make_genesis() { # make_genesis <name> <jar> <owner key name>
  [ -s "$S/addresses/$3" ] || make_key "$3"
  mkdir -p "$S/genesis/$1" && cd "$S/genesis/$1"
  echo "$(cat "$S/addresses/$3"),${GENESIS_AMOUNT:-100000000000000000}" > genesis.csv
  until CL_KEYSTORE="$S/keys/$3.p12" CL_KEYALIAS="$3" java -jar "$2" create-genesis genesis.csv >create-genesis.log 2>&1; do
    echo "[rehearsal] create-genesis for $1 failed; retrying: $(grep -m1 -E 'Exception|ERROR' create-genesis.log | cut -c1-200)"
    sleep 10
  done
}
l0_genesis() { # l0_genesis <name> <owner key name>
  if [ "$N" = 1 ]; then
    # Only this container makes the genesis, so no lock: a restart simply picks up where it stopped.
    [ -s "$S/genesis/$1/genesis.address" ] || make_genesis "$1" "$ECHO_JAR" "$2"
    export ECHO_GENESIS="$S/genesis/$1" ECHO_OWNER_KEYSTORE="$S/keys/$2.p12" ECHO_OWNER_KEYALIAS="$2"
  else
    until [ -s "$S/genesis/$1/genesis.address" ]; do sleep 1; done
    export CL_L0_TOKEN_IDENTIFIER="$(tr -d '\r\n' < "$S/genesis/$1/genesis.address")"
  fi
}

case "$LAYER" in
  global-l0)
    export ECHO_JAR=/code/global-l0/global-l0.jar ECHO_SEEDLIST=$S/gl0.seedlist
    ports 9000; export ECHO_PEERS="$(peers gl0 172.61.0.1 9000 9001)"
    if [ "$N" = 1 ]; then
      for f in $S/addresses/gl0-* $S/addresses/mg-*; do echo "$(cat "$f"),100000000000000"; done > $S/genesis/gl0.csv
      export ECHO_GENESIS=$S/genesis/gl0.csv
    fi ;;
  metagraph-l0)
    export ECHO_JAR=/jars/l0.jar ECHO_SEEDLIST=$S/mg.seedlist
    ports 9100; gl0; export ECHO_PEERS="$(peers mg 172.61.0.2 9100 9101)"
    l0_genesis echo echo-owner ;;
  currency-l1)
    export ECHO_JAR=/jars/l1.jar ECHO_SEEDLIST=$S/mg.seedlist
    ports 9200; gl0; l0_peer 9100 $S/genesis/echo/genesis.address
    export ECHO_PEERS="$(peers mg 172.61.0.2 9200 9201)" ;;
  data-l1)
    export ECHO_JAR=/jars/data_l1.jar ECHO_SEEDLIST=$S/mg.seedlist
    ports 9300; gl0; l0_peer 9100 $S/genesis/echo/genesis.address
    export ECHO_PEERS="$(peers mg 172.61.0.2 9300 9301)" ;;
  identity-l0)
    export ECHO_JAR=/jars/identity_l0.jar ECHO_SEEDLIST=$S/mg.seedlist
    ports 9600; gl0; export ECHO_PEERS="$(peers mg 172.61.0.2 9600 9601)"
    l0_genesis identity identity-owner ;;
  identity-l1)
    export ECHO_JAR=/jars/identity_l1.jar ECHO_SEEDLIST=$S/mg.seedlist
    ports 9500; gl0; l0_peer 9600 $S/genesis/identity/genesis.address
    export ECHO_PEERS="$(peers mg 172.61.0.2 9500 9501)" ;;
  *) echo "unknown LAYER $LAYER" >&2; exit 1 ;;
esac
exec /echo-node.sh
