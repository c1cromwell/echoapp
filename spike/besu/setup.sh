#!/usr/bin/env bash
# Spike setup: makes four throwaway validator keys and the genesis, and compiles the contract.
# Output goes to .network/ and build/ (both ignored by git).
set -euo pipefail
cd "$(dirname "$0")"
BESU_IMAGE="${BESU_IMAGE:-hyperledger/besu:26.11.2}"

if [ ! -f .network/genesis.json ]; then
  rm -rf .network && mkdir -p .network
  docker run --rm -u "$(id -u):$(id -g)" -v "$PWD/network:/in:ro" -v "$PWD/.network:/out" "$BESU_IMAGE" \
    operator generate-blockchain-config --config-file=/in/qbft-config.json --to=/out/gen --private-key-file-name=key >/dev/null
  cp .network/gen/genesis.json .network/genesis.json
  i=0; nodes=""
  for d in .network/gen/keys/*/; do
    i=$((i + 1)); mkdir -p ".network/node$i"; cp "$d/key" "$d/key.pub" ".network/node$i/"
    basename "$d" > ".network/node$i/address"
    pub="$(sed 's/^0x//' "$d/key.pub")"
    nodes="$nodes\"enode://$pub@172.63.0.1$i:30303\","
  done
  echo "[${nodes%,}]" > .network/static-nodes.json
  rm -rf .network/gen
  echo "created 4 validators: $(cat .network/node*/address | tr '\n' ' ')"
fi

[ -d node_modules/solc ] || npm install --silent
node compile.js
