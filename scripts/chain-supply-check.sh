#!/usr/bin/env bash
# Guard for docs/OWN_L0_PLAN.md stage 1: the chain must be buildable without Constellation.
# Fails if a Constellation-hosted URL is back in the build path, or vendored binaries do not
# match the committed checksums.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EUCLID_DIR="${EUCLID_DIR:-$ROOT/../euclid-development-environment}"
fail=0

# Files that take part in building or starting the chain images.
paths=(
  "$ROOT/Makefile" "$ROOT/metagraph/scripts" "$ROOT/metagraph/euclid-overlay" "$ROOT/metagraph/vendor/fetch.sh"
  "$ROOT/metagraph/vendor/build-from-source.sh" "$ROOT/metagraph/euclid.json" "$ROOT/metagraph/build.sbt"
  "$ROOT/metagraph/project/plugins.sbt" "$ROOT/scripts/identity" "$ROOT/docker-compose.identity.yml"
  "$ROOT/deploy/sovereign"
)
[ -d "$EUCLID_DIR/infra/metagraph-ubuntu" ] && paths+=("$EUCLID_DIR/infra/metagraph-ubuntu/Dockerfile" "$EUCLID_DIR/infra/metagraph-base-image")

hits=$(grep -rnE "github\.com/Constellation-Labs|constellationnetwork\.io" "${paths[@]}" 2>/dev/null \
  | grep -vE "^\S+:[0-9]+:\s*#" || true)
if [ -n "$hits" ]; then
  echo "✗ Constellation-hosted URLs in the chain build path:"; echo "$hits"; fail=1
else
  echo "✓ no Constellation-hosted URL in the chain build path"
fi

VERSION="v$(jq -r .tessellation_version "$ROOT/metagraph/euclid.json")"
SUMS="$ROOT/metagraph/vendor/SHA256SUMS-$VERSION"
DIR="$ROOT/metagraph/vendor/tessellation/$VERSION"
if [ ! -f "$SUMS" ]; then
  echo "✗ no committed checksums for Tessellation $VERSION"; fail=1
elif [ ! -d "$DIR" ]; then
  echo "• vendored binaries not present locally (run metagraph/vendor/fetch.sh); checksums are committed"
elif (cd "$DIR" && shasum -a 256 -c "$SUMS" >/dev/null 2>&1); then
  echo "✓ vendored Tessellation $VERSION matches committed checksums"
else
  echo "✗ vendored Tessellation $VERSION does NOT match $SUMS"; fail=1
fi

dep=$(awk -F'"' '/val tessellation = /{print $2}' "$ROOT/metagraph/project/Dependencies.scala")
if [ "v$dep" = "$VERSION" ]; then echo "✓ SDK and node binaries are the same version ($dep)"
else echo "✗ SDK is $dep but euclid.json pins $VERSION"; fail=1; fi
exit $fail
