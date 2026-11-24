#!/usr/bin/env bash
# Spike failure tests against the running four-node network (docker compose up -d; echobesu deploy).
# Each one is something that needed an operator, or failed, on self-hosted Tessellation.
set -uo pipefail
cd "$(dirname "$0")"
B=./build/echobesu
bn() { # block height of node $1, or -1 if it does not answer
  local r; r=$(curl -s -m 3 -X POST -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' "http://localhost:$((8544 + $1))" \
    | sed -n 's/.*"result":"\(0x[0-9a-f]*\)".*/\1/p')
  [ -n "$r" ] && printf '%d' "$r" || printf -- '-1'
}
heights() { echo "n1=$(bn 1) n2=$(bn 2) n3=$(bn 3) n4=$(bn 4)"; }
anchor() { $B anchor 2>&1 | sed 's/^/  /' | cut -c1-110; }
wait_height() { # wait_height <node> <minimum>
  until [ "$(bn "$1")" -ge "$2" ]; do sleep 1; done
}

docker compose up -d >/dev/null 2>&1; wait_height 1 1
echo "== baseline: $(heights)"; anchor

echo "== one of four validators down (node4)"
docker stop echo-besu-4 >/dev/null; anchor; anchor

echo "== two of four down (node3 too): below quorum, the chain must stop, not fork"
docker stop echo-besu-3 >/dev/null; before=$(bn 1); sleep 20
echo "  node1 height $before -> $(bn 1) after 20s"

echo "== both restarted, nothing else done"
t=$(date +%s); docker start echo-besu-3 echo-besu-4 >/dev/null; wait_height 1 $((before + 1))
echo "  blocks resumed $(( $(date +%s) - t ))s after restart"; anchor; sleep 6; echo "  $(heights)"

echo "== node2 loses its disk and rejoins from nothing"
docker compose rm -sf node2 >/dev/null 2>&1; docker volume rm echo-besu-spike_data2 >/dev/null
t=$(date +%s); docker compose up -d node2 >/dev/null 2>&1; target=$(bn 1); wait_height 2 "$target"
echo "  node2 synced to block $target in $(( $(date +%s) - t ))s"; anchor

echo "== anchoring every ~5s for 3 minutes while every validator is restarted in turn"
ok=0; bad=0
for i in $(seq 1 36); do
  case $i in 4) docker restart echo-besu-1 >/dev/null & ;; 12) docker restart echo-besu-2 >/dev/null & ;;
             20) docker restart echo-besu-3 >/dev/null & ;; 28) docker restart echo-besu-4 >/dev/null & ;; esac
  # Submit through whichever node answers, as a gateway with several endpoints would.
  sent=0
  for n in 2 3 4 1; do
    if [ "$(bn $n)" -ge 0 ] && $B -rpc "http://localhost:$((8544 + n))" anchor >/dev/null 2>&1; then sent=1; break; fi
  done
  [ $sent = 1 ] && ok=$((ok + 1)) || bad=$((bad + 1))
  sleep 3
done
wait; sleep 8
echo "  anchors accepted=$ok failed=$bad; $(heights)"
