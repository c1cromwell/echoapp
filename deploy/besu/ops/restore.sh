#!/usr/bin/env bash
# Restores one validator's chain data from an archive made by backup.sh.
#
#   restore.sh <container> <data: docker volume name or host directory> <backup.tar.gz>
#
# The validator must be stopped. Whatever is in the data location is replaced. The archive's
# checksum is checked first; a backup that does not match its checksum is refused.
set -euo pipefail
[ $# -eq 3 ] || { echo "usage: restore.sh <container> <volume-or-directory> <backup.tar.gz>" >&2; exit 2; }
container="$1"; data="$2"; in="$3"
case "$in" in /*) ;; *) in="$PWD/$in" ;; esac
[ -f "$in" ] || { echo "no such backup: $in" >&2; exit 1; }
[ -f "$in.sha256" ] || { echo "no checksum beside $in; refusing to restore an unverified backup" >&2; exit 1; }
( cd "$(dirname "$in")" && shasum -a 256 -c "$(basename "$in").sha256" >/dev/null ) \
  || { echo "checksum mismatch for $in; refusing to restore" >&2; exit 1; }
if [ "$(docker inspect -f '{{.State.Running}}' "$container" 2>/dev/null || echo false)" = true ]; then
  echo "$container is running; stop it first" >&2; exit 1
fi
docker run --rm -v "$data:/data" -v "$(dirname "$in"):/in:ro" alpine:3.20 \
  sh -c 'find /data -mindepth 1 -delete && tar -C /data -xzf "/in/$0"' "$(basename "$in")"
echo "restored $container from $in"
