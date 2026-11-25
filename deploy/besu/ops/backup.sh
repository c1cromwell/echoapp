#!/usr/bin/env bash
# Backs up one validator's chain data to a compressed archive, with a checksum beside it.
#
#   backup.sh <container> <data: docker volume name or host directory> <output.tar.gz>
#
# The validator is stopped while its data is copied, so the copy is consistent, and started again
# afterwards whatever happens. With four validators the chain keeps running on the other three:
# back up one at a time and never two at once.
#
# This archive holds chain data only. The node's key and the genesis file are small, never change,
# and are backed up once, at the key ceremony (docs/CHAIN_OPERATIONS.md).
set -euo pipefail
[ $# -eq 3 ] || { echo "usage: backup.sh <container> <volume-or-directory> <output.tar.gz>" >&2; exit 2; }
container="$1"; data="$2"; out="$3"
case "$out" in /*) ;; *) out="$PWD/$out" ;; esac
mkdir -p "$(dirname "$out")"

docker stop "$container" >/dev/null
trap 'docker start "$container" >/dev/null' EXIT
docker run --rm -v "$data:/data:ro" -v "$(dirname "$out"):/out" alpine:3.20 \
  tar -C /data -czf "/out/$(basename "$out").partial" .
mv "$out.partial" "$out"
( cd "$(dirname "$out")" && shasum -a 256 "$(basename "$out")" > "$(basename "$out").sha256" )
echo "backed up $container to $out ($(du -h "$out" | cut -f1))"
