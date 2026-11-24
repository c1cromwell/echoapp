#!/usr/bin/env bash
# Prints "<ordinal> <hash>" of the newest snapshot a stopped or running L0 node holds on disk.
# Used to choose ECHO_ROLLBACK_HASH: run it on every Global L0 node and restore from the node
# with the highest ordinal (if several tie, their hashes must be equal).
#   latest-snapshot.sh /var/lib/echo/global-l0
set -euo pipefail
d="${1:?usage: latest-snapshot.sh <layer working directory>}/data/incremental_snapshot"
[ -d "$d/ordinal" ] || { echo "no snapshots under $d" >&2; exit 1; }
top="$(find "$d/ordinal" \( -type f -o -type l \) | awk -F/ '{print $NF}' | sort -n | tail -1)"
f="$(find "$d/ordinal" \( -type f -o -type l \) -name "$top" | head -1)"
# Each snapshot is stored once and linked under both its ordinal and its hash.
hash="$(find "$d/hash" -samefile "$f" | awk -F/ '{print $NF}' | head -1)"
[ -n "$hash" ] || { echo "no hash entry for ordinal $top" >&2; exit 1; }
echo "$top $hash"
