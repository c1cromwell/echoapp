#!/usr/bin/env bash
# Fetch Echo's vendored Tessellation node binaries and verify them against the checksums
# committed in this directory. Sources are Echo-controlled mirrors only, tried in order.
# Usage: metagraph/vendor/fetch.sh [version]   (default: version pinned in project/Dependencies.scala)
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
VERSION="${1:-v$(awk -F'"' '/val tessellation = /{print $2}' "$HERE/../project/Dependencies.scala")}"
DEST="$HERE/tessellation/$VERSION"
SUMS="$HERE/SHA256SUMS-$VERSION"
MIRRORS="${ECHO_VENDOR_MIRRORS:-https://github.com/c1cromwell/tessellation/releases/download/echo-$VERSION}"

[ -f "$SUMS" ] || { echo "ERROR: no committed checksums for $VERSION ($SUMS)" >&2; exit 1; }
mkdir -p "$DEST"
cp "$SUMS" "$DEST/SHA256SUMS"

verify() { (cd "$DEST" && shasum -a 256 -c SHA256SUMS >/dev/null 2>&1); }

if verify; then echo "✓ vendored Tessellation $VERSION present and verified"; exit 0; fi

for base in $MIRRORS; do
  echo "Fetching Tessellation $VERSION binaries from $base"
  # Only files that are missing or fail their checksum are downloaded, so an interrupted
  # run resumes instead of starting over.
  while read -r sum f; do
    if [ -f "$DEST/$f" ] && [ "$(shasum -a 256 "$DEST/$f" | awk '{print $1}')" = "$sum" ]; then continue; fi
    curl -fsSL --retry 3 --retry-delay 5 -o "$DEST/$f" "$base/$f" || break
  done < "$SUMS"
  if verify; then echo "✓ fetched and verified"; exit 0; fi
done

echo "ERROR: could not obtain binaries matching $SUMS from: $MIRRORS" >&2
echo "  Rebuild them from Echo's fork with metagraph/vendor/build-from-source.sh $VERSION" >&2
exit 1
