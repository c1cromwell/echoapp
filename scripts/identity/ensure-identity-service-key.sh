#!/usr/bin/env bash
# Generate (if missing) and export the gateway's signing material:
#   IDENTITY_SERVICE_KEY_PEM  secp256k1 key the gateway signs chain writes with
#   IDENTITY_SERVICE_DID      the service's did:key (P-256), used as issuer and authorised sender
set -euo pipefail

if [ -n "${ECHOAPP_ROOT:-}" ]; then
  ROOT="$ECHOAPP_ROOT"
elif [ -n "${BASH_SOURCE[0]:-}" ]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
else
  echo "ensure-identity-service-key: run with bash (source or ECHOAPP_ROOT=...)" >&2
  exit 1
fi

KEY_DIR="${IDENTITY_KEY_DIR:-$ROOT/.keys}"
KEY_PEM="$KEY_DIR/identity-service.pem"
DID_PEM="$KEY_DIR/identity-service-did.pem"
DID_PUB_PEM="$KEY_DIR/identity-service-did.pub.pem"

mkdir -p "$KEY_DIR"

if [ ! -f "$KEY_PEM" ]; then
  echo "[identity-service-key] Generating the secp256k1 chain signing key at $KEY_PEM"
  openssl ecparam -name secp256k1 -genkey -noout -out "$KEY_PEM"
  chmod 600 "$KEY_PEM"
fi

# Phase-1 authorized sender DID (did:key P-256) is independent of the secp256k1 proof key.
if [ ! -f "$DID_PEM" ]; then
  echo "[identity-service-key] Generating P-256 did:key material at $DID_PEM"
  openssl ecparam -name prime256v1 -genkey -noout -out "$DID_PEM"
  chmod 600 "$DID_PEM"
fi
if [ ! -f "$DID_PUB_PEM" ]; then
  openssl ec -in "$DID_PEM" -pubout -out "$DID_PUB_PEM" 2>/dev/null
fi

DIDKEY_OUT="$(go run "$ROOT/cmd/didkey" -pem "$DID_PUB_PEM")"
IDENTITY_SERVICE_DID="$(printf '%s\n' "$DIDKEY_OUT" | awk -F= '/^did=/ {print $2}')"

if [ -z "$IDENTITY_SERVICE_DID" ]; then
  echo "ensure-identity-service-key: could not derive the service DID" >&2
  exit 1
fi

export IDENTITY_SERVICE_KEY_PEM="$KEY_PEM"
export IDENTITY_SERVICE_DID
