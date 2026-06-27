#!/usr/bin/env bash
# Generate docs/screen_catalog PNGs via iOS Simulator + SwiftUI ImageRenderer.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CATALOG="${SCREEN_CATALOG_ROOT:-$ROOT/docs/screen_catalog}"
SIMULATOR="${SCREEN_CATALOG_SIMULATOR:-iPhone 17}"
BUNDLE_ID="${SCREEN_CATALOG_BUNDLE_ID:-com.echo.app}"

mkdir -p "$CATALOG"
CATALOG="$(cd "$CATALOG" && pwd)"

echo "==> Screen catalog → $CATALOG"
echo "==> Simulator: $SIMULATOR"

cd "$ROOT/ios/Echo"

copy_from_simulator() {
  local container src
  container="$(xcrun simctl get_app_container booted "$BUNDLE_ID" data 2>/dev/null || true)"
  if [[ -z "$container" ]]; then
    return 1
  fi
  src="$container/Documents/screen_catalog"
  if [[ ! -f "$src/manifest.jsonl" ]]; then
    return 1
  fi
  rsync -a --delete "$src/" "$CATALOG/"
  echo "==> Copied catalog from Simulator app container"
  return 0
}

xcrun simctl boot "$SIMULATOR" 2>/dev/null || true

set +e
xcodebuild test \
  -project EchoApp.xcodeproj \
  -scheme EchoMessaging \
  -destination "platform=iOS Simulator,name=${SIMULATOR}" \
  -only-testing:EchoUnitTests/ScreenCatalogGeneratorTests/testExportFullCatalog \
  2>&1 | tee /tmp/echo-screen-catalog.log
BUILD_EXIT=$?
set -e

if [[ $BUILD_EXIT -ne 0 ]]; then
  echo "xcodebuild failed (exit $BUILD_EXIT). Last lines:"
  tail -40 /tmp/echo-screen-catalog.log
  exit "$BUILD_EXIT"
fi

if ! copy_from_simulator; then
  if [[ ! -f "$CATALOG/manifest.jsonl" ]]; then
    echo "error: manifest.jsonl not written"
    echo "Checked Simulator Documents/screen_catalog for $BUNDLE_ID"
    exit 1
  fi
  echo "==> Using existing $CATALOG (simulator copy unavailable)"
fi

"$ROOT/scripts/screen-catalog/build-index.sh"
echo "==> Open file://$CATALOG/index.html"
