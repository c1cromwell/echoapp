#!/bin/bash
# End-to-end run of the real app on two simulators against the local backend (`make dev` first).
#
#   scripts/e2e-two-simulators.sh
#
# What it does, with no stubs anywhere:
#   1. erases both simulators (a deleted app leaves its Keychain behind; an erase does not)
#   2. builds the app with simulator signing (the Keychain needs entitlements)
#   3. creates an account on each simulator through first run
#   4. Bob finds Alice by username and sends two messages while she is offline
#   5. Alice opens the app, receives them, replies
#   6. Bob receives the replies and sends another
#   7. Alice keeps the chat open and receives a message live
#   8. Alice signs out and signs back in with Face ID
# Screenshots of every step are saved under test-results/e2e-two-simulators/.
#
# The backend should run with ECHO_ACCESS_LOG=true (the local stack does) so the summary can show
# what the app asked the server. ECHO_ACCESS_LOG_ID_TAILS=true adds delivery tracing; local only.
#
# One step is a fixture, not the product: the app does not yet register usernames with the
# server, so the two usernames are inserted into the local database after the accounts exist.
set -uo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SIM_A="${SIM_A:-9E754BCF-A401-42A8-A5C4-034B05804579}"   # iPhone 17 Pro, iOS 26.0
SIM_B="${SIM_B:-3E8BC487-28C9-4080-A918-4A52500559FB}"   # iPhone 17 Pro Max, iOS 26.0
WORK="${WORK:-$ROOT/test-results/.e2e-work}"
OUT="$ROOT/test-results/e2e-two-simulators"
PG="${PG_CONTAINER:-echo-postgres-testnet}"; API="${API_CONTAINER:-echoapp-testnet}"
mkdir -p "$WORK"; cd "$ROOT/ios/Echo"

sql() { docker exec "$PG" psql -U echoapp -d echoapp -AtF'|' -c "$1"; }
quiet() { grep -Ev '^ |Terminated|done \)|xcrun simctl|E2E-SCREEN|error: -\[' | cut -c1-240; }

# run <device> <test method> <result name> [TEST_RUNNER_VAR=value ...]
run() {
  local d="$1" t="$2" n="$3"; shift 3
  rm -rf "$WORK/$n.xcresult" "$WORK/$n-att"
  xcrun simctl boot "$d" 2>/dev/null; xcrun simctl bootstatus "$d" -b >/dev/null 2>&1
  # Face ID: enrolled, and every prompt answered with a match.
  xcrun simctl spawn "$d" notifyutil -s com.apple.BiometricKit.enrollmentChanged 1
  xcrun simctl spawn "$d" notifyutil -p com.apple.BiometricKit.enrollmentChanged
  ( while true; do xcrun simctl spawn "$d" notifyutil -p com.apple.BiometricKit_Sim.pearl.match 2>/dev/null; sleep 2; done ) &
  local bio=$!; disown $bio 2>/dev/null
  env "$@" xcodebuild test-without-building -project EchoApp.xcodeproj -scheme EchoMessaging \
    -destination "platform=iOS Simulator,id=$d" -parallel-testing-enabled NO \
    "-only-testing:EchoUITests/EchoLiveE2EUITests/$t" -resultBundlePath "$WORK/$n.xcresult" > "$WORK/$n.log" 2>&1
  kill $bio 2>/dev/null
  xcrun xcresulttool export attachments --path "$WORK/$n.xcresult" --output-path "$WORK/$n-att" >/dev/null 2>&1
  grep -E "E2E-|\*\* TEST" "$WORK/$n.log" | quiet
}

echo "== erase both simulators and build"
xcrun simctl shutdown "$SIM_A" 2>/dev/null; xcrun simctl shutdown "$SIM_B" 2>/dev/null
xcrun simctl erase "$SIM_A" && xcrun simctl erase "$SIM_B" || exit 1
xcodebuild build-for-testing -project EchoApp.xcodeproj -scheme EchoMessaging -destination 'generic/platform=iOS Simulator' \
  CODE_SIGN_IDENTITY=- CODE_SIGNING_REQUIRED=NO > "$WORK/build.log" 2>&1
grep -E "\*\* TEST BUILD" "$WORK/build.log" || { grep -E "error:" "$WORK/build.log" | sort -u | head; exit 1; }

T=$(date +%H%M%S); AL="alice$T"; BO="bob$T"
START=$(date -u +%Y-%m-%dT%H:%M:%SZ)
echo "== 1. Alice creates an account ($AL)"; run "$SIM_A" testCreateAccount 1-alice-creates-account "TEST_RUNNER_E2E_USER=$AL"; sleep 6
DA=$(sql "select did from identity_device order by created_at desc limit 1")
echo "== 2. Bob creates an account ($BO)";   run "$SIM_B" testCreateAccount 2-bob-creates-account "TEST_RUNNER_E2E_USER=$BO"; sleep 6
DB=$(sql "select did from identity_device where device_label='primary' order by created_at desc limit 1")
DA=$(sql "select did from identity_device where device_label='primary' and did <> '$DB' order by created_at desc limit 1")
sql "insert into users (user_id, did, username, trust_tier, display_name) values ('user-$AL','$DA','$AL',1,'$AL'), ('user-$BO','$DB','$BO',1,'$BO')" >/dev/null  # fixture

echo "== 3. Bob finds Alice by username and sends two messages (Alice is offline)"
run "$SIM_B" testConversation 3-bob-sends-first-messages "TEST_RUNNER_E2E_PEER=$AL" TEST_RUNNER_E2E_TAG=bob-1 "TEST_RUNNER_E2E_SEND=Hello Alice 👋|How are you today?"
echo "== 4. Alice opens the app, receives both, and replies"
run "$SIM_A" testConversation 4-alice-receives-and-replies "TEST_RUNNER_E2E_PEER=$BO" TEST_RUNNER_E2E_TAG=alice-1 "TEST_RUNNER_E2E_EXPECT=Hello Alice 👋|How are you today?" "TEST_RUNNER_E2E_SEND=Hi Bob! 😀|All good here 🎉"
echo "== 5. Bob receives the replies and sends another"
run "$SIM_B" testConversation 5-bob-receives-and-replies "TEST_RUNNER_E2E_PEER=$AL" TEST_RUNNER_E2E_TAG=bob-2 "TEST_RUNNER_E2E_EXPECT=Hi Bob! 😀|All good here 🎉" "TEST_RUNNER_E2E_SEND=Great 🚀"
echo "== 6/7. Live: Alice keeps the chat open while Bob sends"
( run "$SIM_A" testConversation 7-alice-receives-live "TEST_RUNNER_E2E_PEER=$BO" TEST_RUNNER_E2E_TAG=alice-2 "TEST_RUNNER_E2E_EXPECT=Great 🚀" "TEST_RUNNER_E2E_EXPECT_LIVE=Live ping ⚡" TEST_RUNNER_E2E_WAIT=150 > "$WORK/alice-live.out" 2>&1 ) &
sleep 45
run "$SIM_B" testConversation 6-bob-sends-live "TEST_RUNNER_E2E_PEER=$AL" TEST_RUNNER_E2E_TAG=bob-3 "TEST_RUNNER_E2E_SEND=Live ping ⚡"
wait; echo "-- Alice:"; cat "$WORK/alice-live.out"
echo "== 8. Alice signs out and signs back in with Face ID"
run "$SIM_A" testSignOutAndSignIn 8-alice-signs-out-and-in "TEST_RUNNER_E2E_USER=$AL"

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

echo "== what the app asked the server"
docker logs --since "$START" "$API" 2>&1 | grep " access " | sed 's/^.* access //' \
  | awk '{print $1, $2, $3, ($4 ~ /^[A-Z_]+$/ ? $4 : "")}' | sed -E 's/\{id[^}]*\}/{id}/g' | sort | uniq -c | sort -rn | head -40
echo "== what the server holds for the two accounts"
sql "select 'identities', count(*) from identity_device where did in ('$DA','$DB') and device_label='primary'
     union all select 'encryption keys published', count(*) from identity_device where did in ('$DA','$DB') and device_label<>'primary'
     union all select 'recovery links', count(*) from wallet_accounts where did in ('$DA','$DB')
     union all select 'delivery keys', count(*) from delivery_keys where did in ('$DA','$DB')
     union all select 'messages queued then delivered', count(*) from message_queue where recipient_did in ('$DA','$DB') and status='delivered'
     union all select 'queue rows still holding a message', count(*) from message_queue where recipient_did in ('$DA','$DB') and octet_length(encrypted_payload)>0
     union all select 'queue rows naming a sender or conversation', count(*) from message_queue where recipient_did in ('$DA','$DB') and (sender_did<>'' or conversation_id<>'')"
