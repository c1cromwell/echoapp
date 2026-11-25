#!/usr/bin/env bash
# Key ceremony: creates everything a new network needs, once, on a machine you control.
#
#   ceremony.sh <output directory>
#
# Run it yourself, offline if you can. It refuses to write into a directory that already has
# anything in it, so a network's keys can never be overwritten by a second run.
#
# It creates:
#   genesis.json            the chain's first block, naming the four validators
#   node1..node4/key        each validator's private key        (secret: one per host, plus a backup)
#   node1..node4/key.pub    each validator's public key
#   node1..node4/address    each validator's address
#   validators.txt          the four addresses, comma-separated: what anyone verifying a proof trusts
#   owner.key               deploys the contracts and can change who may write   (secret: offline)
#   reserved.key            the address that holds the reserve                   (secret: offline)
#   MANIFEST.sha256         checksums of the public files, to confirm a copy is intact
#
# Afterwards (docs/CHAIN_OPERATIONS.md): back up the whole directory twice, in two places, encrypted;
# point Ansible's ceremony_dir at it to install the validators; then take owner.key and
# reserved.key offline. Publish validators.txt and genesis.json; they are not secret.
set -euo pipefail
[ $# -eq 1 ] || { echo "usage: ceremony.sh <output directory>" >&2; exit 2; }
ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
BESU_IMAGE="${BESU_IMAGE:-hyperledger/besu:26.11.2}"
OUT="$1"
mkdir -p "$OUT"
OUT="$(cd "$OUT" && pwd)"
[ -z "$(ls -A "$OUT")" ] || { echo "$OUT is not empty; refusing to overwrite an existing ceremony" >&2; exit 1; }
umask 077

docker run --rm -u "$(id -u):$(id -g)" -v "$ROOT/deploy/besu:/in:ro" -v "$OUT:/out" "$BESU_IMAGE" \
  operator generate-blockchain-config --config-file=/in/qbft-config.json --to=/out/gen --private-key-file-name=key >/dev/null
mv "$OUT/gen/genesis.json" "$OUT/genesis.json"
i=0
for d in "$OUT"/gen/keys/*/; do
  i=$((i + 1)); mkdir -p "$OUT/node$i"
  mv "$d/key" "$d/key.pub" "$OUT/node$i/"
  basename "$d" > "$OUT/node$i/address"
done
rm -rf "$OUT/gen"
[ "$i" = 4 ] || { echo "expected 4 validators, got $i" >&2; exit 1; }
cat "$OUT"/node*/address | paste -sd, - > "$OUT/validators.txt"

chain() { (cd "$ROOT" && go run ./cmd/echochain "$@"); }
chain newkey -out "$OUT/owner.key" >/dev/null
chain newkey -out "$OUT/reserved.key" >/dev/null
chmod 0400 "$OUT"/node*/key "$OUT/owner.key" "$OUT/reserved.key"
chmod 0444 "$OUT/genesis.json" "$OUT/validators.txt" "$OUT"/node*/key.pub "$OUT"/node*/address
( cd "$OUT" && shasum -a 256 genesis.json validators.txt node*/key.pub node*/address > MANIFEST.sha256 )

echo "ceremony complete in $OUT"
echo "  validators:        $(cat "$OUT/validators.txt")"
echo "  contract owner:    $(chain address -key "$OUT/owner.key")"
echo "  reserved address:  $(chain address -key "$OUT/reserved.key")"
echo "next: back this directory up twice, encrypted, in two places. See docs/CHAIN_OPERATIONS.md."
