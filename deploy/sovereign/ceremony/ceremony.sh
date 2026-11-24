#!/usr/bin/env bash
# Key ceremony for an Echo network. Run on an offline machine; see docs/KEY_CEREMONY.md.
#
#   ceremony.sh <network> <secrets-dir>        e.g.  ceremony.sh staging /Volumes/ECHO-KEYS/staging
#
# Creates every key the network needs, then writes two things:
#   <secrets-dir>/                       PRIVATE. <name>.p12 per key, passwords.yml, the gateway's
#                                        signing key. Never copied into the repository.
#   deploy/sovereign/networks/<network>/ PUBLIC, committed. Node ids, seedlists, genesis files.
#
# Re-running is safe: a key that already exists is kept, never regenerated, and the public files
# are rebuilt from the keys present. To add a node later, raise GL0_NODES or MG_NODES and re-run.
set -euo pipefail

NETWORK="${1:?usage: ceremony.sh <network> <secrets-dir>}"
SECRETS="${2:?usage: ceremony.sh <network> <secrets-dir>}"
GL0_NODES="${GL0_NODES:-3}"
MG_NODES="${MG_NODES:-3}"

HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/../../.." && pwd)"
TOOLS="${ECHO_TESSELLATION_DIR:-$ROOT/metagraph/vendor/tessellation/v4.0.7}"
OUT="$ROOT/deploy/sovereign/networks/$NETWORK"

for f in cl-keytool.jar cl-wallet.jar; do
  [ -f "$TOOLS/$f" ] || { echo "missing $TOOLS/$f — run metagraph/vendor/fetch.sh first" >&2; exit 1; }
done
case "$SECRETS" in "$ROOT"/*) echo "the secrets directory must be outside the repository" >&2; exit 1 ;; esac

umask 077
mkdir -p "$SECRETS"
SECRETS="$(cd "$SECRETS" && pwd)"
PASSWORDS="$SECRETS/passwords.yml"
[ -f "$PASSWORDS" ] || echo "key_passwords:" > "$PASSWORDS"

password_of() { awk -v k="  $1:" '$0 ~ "^"k" " {print $2}' "$PASSWORDS" | tr -d '"'; }
tool() { # tool <key name> <jar> <args...>
  local name=$1 jar=$2; shift 2
  local pw; pw="$(password_of "$name")"
  ( cd "$SECRETS" && CL_KEYSTORE="$name.p12" CL_KEYALIAS="$name" CL_PASSWORD="$pw" \
      CL_STOREPASS="$pw" CL_KEYPASS="$pw" java -jar "$TOOLS/$jar" "$@" | tail -1 )
}
make_key() { # make_key <name>
  if [ -f "$SECRETS/$1.p12" ]; then echo "  kept     $1"; return; fi
  # Letters and digits only, so the password survives every file format it passes through.
  local pw; pw="$(openssl rand -hex 24)"
  echo "  $1: \"$pw\"" >> "$PASSWORDS"
  tool "$1" cl-keytool.jar generate >/dev/null
  echo "  created  $1"
}

GL0=(); for i in $(seq 1 "$GL0_NODES"); do GL0+=("gl0-$i"); done
MG=();  for i in $(seq 1 "$MG_NODES");  do MG+=("mg-$i");   done
# ECHO supply at genesis (docs/OWN_L0_PLAN.md §10): name=whole ECHO. 1 ECHO = 100,000,000 units.
ALLOCATION=(alloc-community-rewards=400000000 alloc-treasury=220000000 alloc-founders=180000000
            alloc-team-advisors=100000000 alloc-ecosystem=100000000)

echo "Keys in $SECRETS"
for k in "${GL0[@]}" "${MG[@]}" echo-owner identity-owner; do make_key "$k"; done
for a in "${ALLOCATION[@]}"; do make_key "${a%%=*}"; done

# The gateway signs Data L1 and Identity L1 updates with this key; the chain accepts no other.
eval "$(IDENTITY_KEY_DIR="$SECRETS" ECHOAPP_ROOT="$ROOT" bash -c \
  'source "$ECHOAPP_ROOT/scripts/identity/ensure-identity-service-key.sh" >&2
   echo "SIGNER=$IDENTITY_SERVICE_PUBLIC_KEY_HEX; SERVICE_DID=$IDENTITY_SERVICE_DID"')"

umask 022
mkdir -p "$OUT"
{
  echo "# Public facts about the Echo '$NETWORK' network, written by deploy/sovereign/ceremony/ceremony.sh."
  echo "# Node ids are public keys. No secret belongs in this file."
  echo "manifest:"
  echo "  network: $NETWORK"
  echo "  gateway_signer: \"$SIGNER\""
  echo "  identity_service_did: \"$SERVICE_DID\""
  echo "  nodes:"
  for k in "${GL0[@]}" "${MG[@]}"; do echo "    $k: \"$(tool "$k" cl-wallet.jar show-id)\""; done
  echo "  addresses:"
  for k in "${GL0[@]}" "${MG[@]}" echo-owner identity-owner; do echo "    $k: $(tool "$k" cl-wallet.jar show-address)"; done
  for a in "${ALLOCATION[@]}"; do echo "    ${a%%=*}: $(tool "${a%%=*}" cl-wallet.jar show-address)"; done
} > "$OUT/manifest.yml"

id_of() { awk -v k="    $1:" '$0 ~ "^"k" " {print $2}' "$OUT/manifest.yml" | tr -d '"' | head -1; }
address_of() { awk -v k="    $1:" '$0 ~ "^"k" DAG" {print $2}' "$OUT/manifest.yml"; }

# Global L0 trusts the majority of gl0.priority. It admits metagraph snapshots only from signers
# on gl0.seedlist, which is why the metagraph nodes are on it too.
for k in "${GL0[@]}"; do id_of "$k"; done | sort > "$OUT/gl0.priority"
for k in "${MG[@]}"; do id_of "$k"; done | sort > "$OUT/mg.seedlist"
cat "$OUT/gl0.priority" "$OUT/mg.seedlist" | sort > "$OUT/gl0.seedlist"

# Global L0's own token is not used by Echo (fees are zero); each node gets a nominal balance so
# a fee could be paid if that ever changes.
for k in "${GL0[@]}" "${MG[@]}"; do echo "$(address_of "$k"),100000000000000"; done > "$OUT/global-l0-genesis.csv"
for a in "${ALLOCATION[@]}"; do echo "$(address_of "${a%%=*}"),${a##*=}00000000"; done > "$OUT/echo-genesis.csv"
# The Identity metagraph has no token of its own; a genesis needs one balance, so the owner gets 1.
echo "$(address_of identity-owner),100000000" > "$OUT/identity-genesis.csv"

total=$(awk -F, '{s+=$2} END {printf "%.0f", s/100000000}' "$OUT/echo-genesis.csv")
[ "$total" = 1000000000 ] || { echo "ECHO genesis totals $total, expected 1000000000" >&2; exit 1; }

echo
echo "Public files written to $OUT (commit them):"
ls "$OUT"
echo
echo "ECHO at genesis: $total across ${#ALLOCATION[@]} addresses."
echo "Next: back up $SECRETS twice, offline (docs/KEY_CEREMONY.md), before any node is started."
