#!/usr/bin/env bash
# scripts/setup-cardano-testnet.sh — Creates a Cardano testnet wallet and requests test ADA.
# Usage: ./scripts/setup-cardano-testnet.sh
#
# Prerequisites: cardano-cli installed (brew install cardano-cli or from IOHK releases)
set -euo pipefail

WALLET_DIR="configs/cardano-testnet"
NETWORK="--testnet-magic 2"  # Preview testnet

mkdir -p "$WALLET_DIR"

echo "=== Echo — Cardano Testnet Wallet Setup ==="

# Check for cardano-cli
if ! command -v cardano-cli &>/dev/null; then
  echo "cardano-cli not found. Install options:"
  echo "  macOS:  brew install cardano-cli"
  echo "  Linux:  Download from https://github.com/IntersectMBO/cardano-cli/releases"
  echo ""
  echo "Alternatively, use an online wallet:"
  echo "  1. Go to https://eternl.io or https://lace.io"
  echo "  2. Create a wallet on Preview testnet"
  echo "  3. Fund from faucet: https://docs.cardano.org/cardano-testnets/tools/faucet/"
  echo "  4. Save the address to $WALLET_DIR/payment.addr"
  exit 1
fi

# Generate keys if they don't exist
if [ ! -f "$WALLET_DIR/payment.skey" ]; then
  echo "Generating payment key pair..."
  cardano-cli address key-gen \
    --verification-key-file "$WALLET_DIR/payment.vkey" \
    --signing-key-file "$WALLET_DIR/payment.skey"

  echo "Generating stake key pair..."
  cardano-cli stake-address key-gen \
    --verification-key-file "$WALLET_DIR/stake.vkey" \
    --signing-key-file "$WALLET_DIR/stake.skey"

  echo "Building payment address..."
  cardano-cli address build \
    --payment-verification-key-file "$WALLET_DIR/payment.vkey" \
    --stake-verification-key-file "$WALLET_DIR/stake.vkey" \
    $NETWORK \
    --out-file "$WALLET_DIR/payment.addr"

  echo "✓ Wallet created"
else
  echo "✓ Wallet already exists"
fi

ADDR=$(cat "$WALLET_DIR/payment.addr")
echo ""
echo "Testnet address: $ADDR"
echo ""
echo "Next steps:"
echo "  1. Request test ADA from the faucet:"
echo "     https://docs.cardano.org/cardano-testnets/tools/faucet/"
echo "  2. Paste your address: $ADDR"
echo "  3. Select 'Preview' network"
echo "  4. You'll receive 10,000 test ADA"
echo ""
echo "Verify balance (requires a running node or API):"
echo "  cardano-cli query utxo --address $ADDR $NETWORK"
echo ""
echo "Files created in $WALLET_DIR/:"
ls -la "$WALLET_DIR/"
echo ""
echo "IMPORTANT: Never commit .skey files to git!"
