#!/usr/bin/env bash
# Create a product-scoped release tag for independent launch trains.
#
# Usage:
#   ./scripts/release/product-tag.sh echo-messaging 1.0.0
#   ./scripts/release/product-tag.sh comply-portal 1.2.0
#   ./scripts/release/product-tag.sh echo-comply-ios 1.0.0
#   ./scripts/release/product-tag.sh echo-passport 1.0.0
#
# Valid products: echo-messaging | comply-portal | echo-comply-ios | echo-passport | passport-sdk

set -euo pipefail

PRODUCT="${1:-}"
VERSION="${2:-}"

usage() {
  sed -n '2,10p' "$0"
  exit 1
}

[[ -n "$PRODUCT" && -n "$VERSION" ]] || usage

case "$PRODUCT" in
  echo-messaging|comply-portal|echo-comply-ios|echo-passport|passport-sdk) ;;
  *) echo "Unknown product: $PRODUCT"; usage ;;
esac

TAG="${PRODUCT}@v${VERSION#v}"

if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "Tag already exists: $TAG"
  exit 1
fi

git tag -a "$TAG" -m "${PRODUCT} release v${VERSION#v}"
echo "Created tag $TAG"
echo "Push with: git push origin $TAG"
