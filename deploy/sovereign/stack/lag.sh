#!/usr/bin/env bash
# One line: how many Global L0 nodes are Ready, the newest global snapshot and its signer count,
# and for each metagraph its own newest snapshot against the newest one Global L0 has included.
# A metagraph snapshot is final for outsiders only once Global L0 includes it, so "behind" is the lag.
#   lag.sh [global l0 node to ask, default 3]
set -euo pipefail
via="${1:-3}"
ready=0
for n in 1 2 3; do
  s=$(docker exec "echo-sov-global-l0-$n" curl -fsS -m 3 "http://172.61.0.1$n:9000/node/info" 2>/dev/null | sed -n 's/.*"state":"\([A-Za-z]*\)".*/\1/p' || true)
  [ "$s" = Ready ] && ready=$((ready + 1))
done
own() { docker exec echo-sov-mg-2 curl -fsS -m 3 "http://172.61.0.22:$1/snapshots/latest/ordinal" | tr -dc 0-9; }
addr() { docker exec echo-sov-metagraph-l0-2 cat "/shared/genesis/$1/genesis.address" | tr -d '\r\n'; }
docker exec "echo-sov-global-l0-$via" curl -fsS -m 5 -H 'Accept: application/json' "http://172.61.0.1$via:9000/global-snapshots/latest/combined" |
  python3 -c "
import sys, json
d = json.load(sys.stdin)
seen = {k: v['Right'][0]['value']['ordinal'] for k, v in d[1]['lastCurrencySnapshots'].items()}
out = ['$(date -u +%T)', 'gl0 $ready/3 Ready', 'global %d (%d signers)' % (d[0]['value']['ordinal'], len(d[0]['proofs']))]
for name, addr, own in (('echo', '$(addr echo)', $(own 9100)), ('identity', '$(addr identity)', $(own 9600))):
    out.append('%s own %d included %d behind %d' % (name, own, seen.get(addr, -1), own - seen.get(addr, -1)))
print(' | '.join(out))"
