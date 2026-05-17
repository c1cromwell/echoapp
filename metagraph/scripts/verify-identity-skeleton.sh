#!/usr/bin/env bash
# metagraph/scripts/verify-identity-skeleton.sh
# WO-276 acceptance (static): verify Identity Metagraph modules are present in
# euclid.json, build.sbt, and source tree — runnable in CI without sbt/Docker.
#
# Does NOT replace: `cd metagraph && sbt compile` or `hydra start-genesis`
# (those need JDK 21 + sbt + Euclid on a developer machine).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
METAGRAPH_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$METAGRAPH_ROOT"

RED='\033[0;31m'; GRN='\033[0;32m'; NC='\033[0m'
fail() { echo -e "${RED}FAIL:${NC} $*" >&2; exit 1; }
ok()   { echo -e "${GRN}OK:${NC}  $*"; }

command -v jq >/dev/null 2>&1 || fail "jq is required (brew install jq)"

EUCLID_JSON="$METAGRAPH_ROOT/euclid.json"
[[ -f "$EUCLID_JSON" ]] || fail "missing $EUCLID_JSON"

# framework.modules
jq -e '.framework.modules.identity_l0 == "modules/identity_l0"' "$EUCLID_JSON" >/dev/null \
  || fail "euclid.json: framework.modules.identity_l0 must be modules/identity_l0"
jq -e '.framework.modules.identity_l1 == "modules/identity_l1"' "$EUCLID_JSON" >/dev/null \
  || fail "euclid.json: framework.modules.identity_l1 must be modules/identity_l1"
ok "euclid.json framework.modules.identity_l0 / identity_l1"

# docker.default_containers
jq -e '(.docker.default_containers | index("identity-l0")) != null' "$EUCLID_JSON" >/dev/null \
  || fail "euclid.json: docker.default_containers must include identity-l0"
jq -e '(.docker.default_containers | index("identity-l1")) != null' "$EUCLID_JSON" >/dev/null \
  || fail "euclid.json: docker.default_containers must include identity-l1"
ok "euclid.json docker.default_containers includes identity-l0, identity-l1"

# ports (WO-276: no conflicts with 9000/9200/9300/9400)
[[ "$(jq -r '.layers.identity_l0.ports.public' "$EUCLID_JSON")" == "9600" ]] \
  || fail "identity_l0 public port must be 9600"
[[ "$(jq -r '.layers.identity_l1.ports.public' "$EUCLID_JSON")" == "9500" ]] \
  || fail "identity_l1 public port must be 9500"
ok "euclid.json identity_l0=9600 identity_l1=9500"

BUILD_SBT="$METAGRAPH_ROOT/build.sbt"
grep -q 'lazy val identityL0' "$BUILD_SBT" || fail "build.sbt: missing identityL0 project"
grep -q 'lazy val identityL1' "$BUILD_SBT" || fail "build.sbt: missing identityL1 project"
grep -q '.aggregate(.*identityL0.*identityL1' "$BUILD_SBT" \
  || grep -q ', identityL0, identityL1' "$BUILD_SBT" \
  || fail "build.sbt: root must aggregate identityL0 and identityL1"
ok "build.sbt defines and aggregates identityL0 / identityL1"

MAIN_L0="modules/identity_l0/src/main/scala/com/echo/identity_l0/Main.scala"
MAIN_L1="modules/identity_l1/src/main/scala/com/echo/identity_l1/Main.scala"
[[ -f "$MAIN_L0" ]] || fail "missing $MAIN_L0"
[[ -f "$MAIN_L1" ]] || fail "missing $MAIN_L1"
ok "Identity L0/L1 Main.scala present"

if grep -RIn '00000000-0000-0000-0000-000000000000' --include='*.scala' "$METAGRAPH_ROOT" 2>/dev/null | grep -q .; then
  fail "zero UUID still present in metagraph/**/*.scala (WO-276 zero-UUID sweep)"
fi
ok "no zero-UUID cluster IDs in Scala sources"

echo ""
echo "All WO-276 static checks passed."
echo "Next (developer machine): cd metagraph && sbt compile && ./scripts/setup-euclid.sh"
