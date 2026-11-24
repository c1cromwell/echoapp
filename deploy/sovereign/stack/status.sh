#!/usr/bin/env bash
# Prints the state and peer count of every node in the rehearsal, read from inside its network.
set -euo pipefail
probe() { # probe <container> <ip> <port>
  docker exec "$1" bash -c "
    s=\$(curl -fsS -m 3 http://$2:$3/node/info 2>/dev/null | sed -n 's/.*\"state\":\"\([A-Za-z]*\)\".*/\1/p')
    n=\$(curl -fsS -m 3 http://$2:$3/cluster/info 2>/dev/null | grep -o '\"state\":\"Ready\"' | wc -l | tr -d ' ')
    echo \"\${s:-down} \${n:-0}\"" 2>/dev/null || echo "down 0"
}
printf '%-14s %-18s %-18s %-18s\n' layer node-1 node-2 node-3
for spec in global-l0:1:9000 metagraph-l0:2:9100 currency-l1:2:9200 data-l1:2:9300 identity-l0:2:9600 identity-l1:2:9500; do
  IFS=: read -r layer net port <<<"$spec"
  row="$layer"
  line=$(printf '%-14s' "$layer")
  for n in 1 2 3; do
    set -- $(probe "echo-sov-$layer-$n" "172.61.0.$net$n" "$port")
    line="$line $(printf '%-18s' "$1 ($2 ready)")"
  done
  echo "$line"
done
