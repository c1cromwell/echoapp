#!/bin/bash
# Live check of the "username taken" prompt: the real app on one simulator against the local
# backend (`make dev` first).
#
#   scripts/e2e-username-taken.sh
#
#   1. erases the simulator, builds the app with simulator signing
#   2. creates an account through first run; the app registers its username
#   3. makes that username belong to someone else on the server, and clears the app's note that
#      it registered (the state of an account whose name was claimed between the availability
#      check and its first sign-in)
#   4. opens the app: it must ask for another name, refuse the taken one, register the new one,
#      and not ask again on the next launch
#   5. checks the server: the new name belongs to this account, the old one to the other
# Screenshots are saved under test-results/username-taken-live/.
set -uo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SIM="${SIM_A:-9E754BCF-A401-42A8-A5C4-034B05804579}"
WORK="${WORK:-$ROOT/test-results/.e2e-work-username}"
OUT="$ROOT/test-results/username-taken-live"
PG="${PG_CONTAINER:-echo-postgres-testnet}"
APP_ID="com.echo.app"
rm -rf "$WORK"; mkdir -p "$WORK"; cd "$ROOT/ios/Echo"
FAILED=0
fail() { echo "   FAIL: $1"; FAILED=1; }

sql() { docker exec "$PG" psql -U echoapp -d echoapp -AtF'|' -c "$1"; }

run() {
  local t="$1" n="$2"; shift 2
  xcrun simctl boot "$SIM" 2>/dev/null; xcrun simctl bootstatus "$SIM" -b >/dev/null 2>&1
  xcrun simctl spawn "$SIM" notifyutil -s com.apple.BiometricKit.enrollmentChanged 1
  xcrun simctl spawn "$SIM" notifyutil -p com.apple.BiometricKit.enrollmentChanged
  ( while true; do xcrun simctl spawn "$SIM" notifyutil -p com.apple.BiometricKit_Sim.pearl.match 2>/dev/null; sleep 2; done ) &
  local bio=$!; disown $bio 2>/dev/null
  env "$@" xcodebuild test-without-building -project EchoApp.xcodeproj -scheme EchoMessaging \
    -destination "platform=iOS Simulator,id=$SIM" -parallel-testing-enabled NO \
    "-only-testing:EchoUITests/EchoLiveE2EUITests/$t" -resultBundlePath "$WORK/$n.xcresult" > "$WORK/$n.log" 2>&1
  kill $bio 2>/dev/null
  xcrun xcresulttool export attachments --path "$WORK/$n.xcresult" --output-path "$WORK/$n-att" >/dev/null 2>&1
  grep -E "E2E-TAKEN|\*\* TEST|error: -\[" "$WORK/$n.log" | cut -c1-240
  grep -q "TEST EXECUTE SUCCEEDED" "$WORK/$n.log" || fail "$t did not pass"
}

echo "== erase the simulator and build"
xcrun simctl shutdown "$SIM" 2>/dev/null; xcrun simctl erase "$SIM" || exit 1
xcodebuild build-for-testing -project EchoApp.xcodeproj -scheme EchoMessaging -destination 'generic/platform=iOS Simulator' \
  CODE_SIGN_IDENTITY=- CODE_SIGNING_REQUIRED=NO > "$WORK/build.log" 2>&1
grep -E "\*\* TEST BUILD" "$WORK/build.log" || { grep -E "error:" "$WORK/build.log" | sort -u | head; exit 1; }

T=$(date +%H%M%S); OLD="carol$T"; NEW="carol${T}_b"
echo "== 1. create an account ($OLD)"
run testCreateAccount 1-create-account "TEST_RUNNER_E2E_USER=$OLD"; sleep 6
DID=$(sql "select did from users where lower(username) = lower('$OLD')")
[ -n "$DID" ] || { fail "the app did not register $OLD"; exit 1; }

echo "== 2. the name now belongs to someone else"
xcrun simctl terminate "$SIM" "$APP_ID" 2>/dev/null
OTHER="did:key:zSomeoneElse$T"
sql "update users set did = '$OTHER' where lower(username) = lower('$OLD')" >/dev/null || { fail "could not reassign the name"; exit 1; }
# The app's preferences live in its own data container, not the simulator-wide defaults.
PREFS="$(xcrun simctl get_app_container "$SIM" "$APP_ID" data)/Library/Preferences/$APP_ID"
xcrun simctl spawn "$SIM" defaults delete "$PREFS" echo.username.registered || { fail "could not clear the app's registration note"; exit 1; }
echo "   $OLD -> another identity; this account has no name on the server"

echo "== 3. the app asks for another name ($NEW)"
run testPickAnotherUsername 2-pick-another "TEST_RUNNER_E2E_USER=$OLD" "TEST_RUNNER_E2E_NEW_NAME=$NEW"

echo "== 4. what the server holds"
NEW_DID=$(sql "select did from users where lower(username) = lower('$NEW')")
OLD_DID=$(sql "select did from users where lower(username) = lower('$OLD')")
[ "$NEW_DID" = "$DID" ] && echo "   $NEW belongs to this account" || fail "$NEW is not registered to this account (found: ${NEW_DID:-nothing})"
[ "$OLD_DID" = "$OTHER" ] && echo "   $OLD still belongs to the other identity" || fail "$OLD changed hands (found: ${OLD_DID:-nothing})"

rm -rf "$OUT"; mkdir -p "$OUT"
python3 - "$WORK" "$OUT" <<'PY'
import json, os, re, shutil, sys
src, out = sys.argv[1], sys.argv[2]
count = 0
for d in sorted(os.listdir(src)):
    manifest = os.path.join(src, d, 'manifest.json')
    if not d.endswith('-att') or not os.path.exists(manifest):
        continue
    for test in json.load(open(manifest)):
        for a in test.get('attachments', []):
            f = a.get('exportedFileName', '')
            if f.endswith('.png'):
                label = re.sub(r'_\d+_[0-9A-F-]{36}\.png$', '', a.get('suggestedHumanReadableName', f))
                shutil.copy(os.path.join(src, d, f), os.path.join(out, d[:-4] + '__' + label + '.png'))
                count += 1
print('screenshots saved:', count, 'in', out)
PY
[ "$FAILED" = 0 ] && echo "RESULT: passed" || echo "RESULT: FAILED"
exit $FAILED
