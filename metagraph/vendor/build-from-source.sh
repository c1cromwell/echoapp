#!/usr/bin/env bash
# Build the Tessellation node binaries from source (Echo's fork by default) into
# metagraph/vendor/tessellation/<version>/. Use when a new version is adopted or when every mirror
# is gone. Needs JDK 21 and sbt. Prints the checksums; commit them as SHA256SUMS-<version> only
# after reviewing what changed upstream.
# Usage: metagraph/vendor/build-from-source.sh v4.0.7 [path-to-existing-tessellation-clone]
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
VERSION="${1:?version tag, e.g. v4.0.7}"
SRC="${2:-}"
REPO="${TESSELLATION_REPO_URL:-https://github.com/c1cromwell/tessellation.git}"
DEST="$HERE/tessellation/$VERSION"

if [ -z "$SRC" ]; then
  SRC="$(mktemp -d)/tessellation"
  git clone --depth 1 --branch "$VERSION" "$REPO" "$SRC"
fi
cd "$SRC"
git checkout --quiet "$VERSION"
RELEASE_TAG="$VERSION" SBT_OPTS="${SBT_OPTS:--Xss8m -Xmx6g}" \
  sbt -Dversion.override="${VERSION#v}" wallet/assembly keytool/assembly dagL0/assembly dagL1/assembly

mkdir -p "$DEST"
cp modules/wallet/target/scala-2.13/tessellation-wallet-assembly-*.jar "$DEST/cl-wallet.jar"
cp modules/keytool/target/scala-2.13/tessellation-keytool-assembly-*.jar "$DEST/cl-keytool.jar"
cp modules/dag-l0/target/scala-2.13/tessellation-dag-l0-assembly-*.jar "$DEST/cl-node.jar"
cp modules/dag-l1/target/scala-2.13/tessellation-dag-l1-assembly-*.jar "$DEST/cl-dag-l1.jar"
(cd "$DEST" && shasum -a 256 cl-node.jar cl-dag-l1.jar cl-wallet.jar cl-keytool.jar | tee SHA256SUMS)
echo "Source commit: $(git rev-parse HEAD)"
