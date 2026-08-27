# Echo — Two-device E2E playbook

**When:** full A+B messaging test on two simulators or two iPhones.  
**Canonical gates:** [`E2E_TESTING.md`](E2E_TESTING.md) §1 (automated) then this file (manual).  
**Traceability:** [`knowledge-graph/TRACEABILITY.md`](knowledge-graph/TRACEABILITY.md) · Factory [8090](https://factory.8090.ai)

Frozen UX: validate `FirstRunCoordinator` / `GlacialLoginScreen` only — do not redesign.

---

## 0. What this run proves

| # | Customer path | Pass means |
|---|----------------|-------------|
| D0 | Connectivity | Both clients reach the same backend; WS upgrades |
| D1 | Find each other | `@username` invite or QR → both in Contacts |
| D2 | Live chat | A sends, B sees without refresh |
| D3 | Offline delivery | A sends while B is killed; B gets it after open (push if APNs) |
| D4 | Signals | Typing, delivered→read, reaction |
| D5 | New phone | Phrase restore on B → chats return (if backup exists) |
| D6 | Group | Create, add peer, bidirectional messages |

Skip D5 if you are not testing restore this session. Mark N/A, do not fail the run.

---

## 1. Prep (once per machine)

### 1.1 Stack

```bash
make dev
make dev-status          # Backend :8000 must be ✓
make phase3-signals-proof
make ios-preflight BUILD=1
```

Identity L0/L1 (`make start-identity`) only if you need username chain anchors or `validate-phase1` step 3.

| Need | Env / flag | If missing |
|------|------------|------------|
| Durable offline queue | `DATABASE_URL` (Compose sets this) | Queue is **in-memory**; lost on API restart |
| Lock-screen wake | `APNS_KEY_FILE` or `APNS_KEY_PATH`, `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_BUNDLE_ID`, `APNS_ENVIRONMENT=sandbox` | D3 still passes **on app open**; no banner while locked |
| Cloud backup blobs | MinIO/S3 from Compose | Backup push/pull fails |
| Wallet restore-did | User completed `/v3/wallet/link` (or `/v1/identity/link-wallet`) | restore-did cannot resolve DID |

Confirm API:

```bash
curl -fsS http://localhost:8000/health
# {"status":"operational", ...}
```

### 1.2 Two clients

| Setup | API_URL | Notes |
|-------|---------|--------|
| Two Simulators on this Mac | `http://localhost:8000` | Scheme **EchoApp** / product **EchoMessaging** |
| Two physical iPhones | `http://<Mac-LAN-IP>:8000` | Mac and phones on same Wi-Fi; allow incoming :8000 |
| TestFlight | HTTPS staging | TLS + WSS; APNs production if TF build |

Boot two sims:

```bash
xcrun simctl list devices available
# Xcode → Product → Destination → pick iPhone A, Run
# Duplicate scheme run: hold Option, pick iPhone B (or Window → Manage Runs)
```

Give devices distinct display names / `@usernames` in first-run (e.g. `alice` / `bob`).

### 1.3 Logging — open before you tap

**Backend (terminal A)**

```bash
make dev-logs
# or
docker logs -f echoapp-testnet --tail 200
```

Useful greps (no ciphertext should ever appear):

| Grep | Meaning |
|------|---------|
| `APNs push sender enabled` | Real `.p8` loaded at boot |
| `APNs push sender disabled` | D3 will not wake a locked phone |
| `ws upgrade` | Handshake failures |
| `flushOffline` / `message_queue` / `queued` | Durable drain |
| `restore-did` / `restore-challenge` | D5 |

**Postgres queue (optional)**

```bash
docker exec -it echo-postgres-testnet psql -U echo -d echo -c \
  "SELECT id, recipient_did, created_at FROM message_queue ORDER BY created_at DESC LIMIT 10;"
```

(Adjust user/db if your Compose differs. Blobs are opaque WS envelopes — T3 — do not print them.)

**iOS (Console.app or Xcode)**

1. Console.app → select Simulator / device → Start.
2. Filter subsystem: bundle id (`com.echo.app` or the EchoMessaging id).
3. Categories: `echo` (`EchoLogger`), `identity`, `provisioning`.

Xcode: each run window → Debug area. Scheme env `API_URL` must match §1.2.

**T0–T7:** logs may keep `did:key:…`. They must **not** contain message plaintext, recovery phrases, or APNs payloads with ciphertext.

---

## 2. Connectivity checklist (D0) — 5 minutes

Do this before onboarding.

| # | Check | Command / where | Pass |
|---|--------|-----------------|------|
| C1 | Backend health | `curl -fsS http://localhost:8000/health` | `operational` |
| C2 | Global L0 (optional) | `curl -fsS http://localhost:9000/node/info` | JSON (skip if not running hydra) |
| C3 | Data L1 (optional) | `curl -fsS http://localhost:9400/node/info` | JSON |
| C4 | Identity L1 (optional) | `curl -fsS http://localhost:9500/node/info` | JSON if `start-identity` |
| C5 | Postgres up | `make dev-status` / `docker ps` `echo-postgres-testnet` | running |
| C6 | Redis up | `echo-redis-testnet` | running (JWT/refresh) |
| C7 | MinIO up | `echo-minio-testnet` | running (backup/media) |
| C8 | Simulator can see API | On sim: first-run or login hits API (Xcode network log) | No connection refused |
| C9 | Device LAN | From iPhone Safari: `http://<LAN-IP>:8000/health` | JSON (ATS may block http — use scheme exception already in project) |
| C10 | WS path | After login, Messages tab open; backend log shows WS connect | No `ws upgrade error` |
| C11 | APNs (optional) | Boot log `APNs push sender enabled` + iOS Settings → Notifications allowed | Needed for locked-phone D3 |

If C1 fails, stop. Do not debug chat UI.

---

## 3. Device A and B — first-run (A1)

On **each** device, separately:

1. Fresh install or reset first-run.
2. Complete **existing** Glacial / FirstRun flow (display name, `@username`, passkey / Face ID, recovery phrase).
3. **Write down** username + 24-word phrase for the device you will wipe in D5.
4. Land on Messages tab (triggers `MessageRelaySession.connect()`).
5. Grant notifications when asked (`PushRegistrationService` → `POST /v3/notifications/register`).

| Device | @username | Phrase saved | WS connected (log) | Push registered |
|--------|-----------|--------------|--------------------|-----------------|
| A | | ☐ | ☐ | ☐ |
| B | | ☐ | ☐ | ☐ |

Pass: both authenticated, Messages hub visible.

---

## 4. Find each other (D1)

**Preferred:** `@username` invite.

1. A: Contacts or Invite sheet → copy `echo://invite?u=alice` (or share `@alice`).
2. B: open the link (Notes → tap, or paste into Accept invite).
3. B: accept → search + `addContact` (`addedVia: username_invite`).
4. A: confirm B appears in Contacts (may need pull-to-refresh).

**Alternate:** Profile QR on A → B scans (`QRContactAddCoordinator`).

| Step | Pass |
|------|------|
| Deep link parses handle not opaque UUID | ☐ |
| `POST /v3/contacts/add` 2xx in Charles/Proxyman or backend log | ☐ |
| Thread can open via Message on contact detail | ☐ |

Legacy `echo://invite?code=` may still parse — prefer `u=`.

---

## 5. Live DM (D2) — Week A A2–A5

Thread id must be `dm:{sorted-did}:{sorted-did}`.

| Step | A | B | Pass |
|------|---|---|------|
| Open thread | New conversation → search `@bob` | — | Hub row exists |
| Send | “hello-d2” | Chat open | A shows sent |
| Receive | — | Sees “hello-d2” without pull | Live WS |
| Reply | Chat open | “hello-d2-ack” | A sees live |

Backend: no plaintext of `hello-d2` in `make dev-logs`.

---

## 6. Offline delivery (D3) — the customer-critical path

**Without APNs:** B still receives on next `MessageRelaySession.connect()` (open app / tap Messages).  
**With APNs:** B may get a **content-blind** notification (`You have a new message` + `conversationId` only).

| Step | Action | Expected | ☐ |
|------|--------|----------|---|
| 1 | Force-quit B (swipe up). Confirm no WS on B. | Backend: B disconnected | |
| 2 | A sends “hello-d3-offline” | Backend enqueues (`message_queue` or memory) | |
| 3 | Optional: `SELECT` queue row for B’s DID | One new opaque blob | |
| 4 | If APNs: banner/silent on B | No sender name, no body text | |
| 5 | Open B (or tap notification) | WS connects; flush delivers “hello-d3-offline” | |
| 6 | Restart API **without** Postgres | In-memory queue **lost** — know this limitation | N/A |

Fail if: message never appears after B is foregrounded for 15s **and** Postgres is up.

---

## 7. Signals (D4) — Week A A6–A8 / sign-off §7.1

Both in the same thread, both foreground.

| Step | Expected | ☐ |
|------|----------|---|
| A types, does not send | B typing indicator (~and clears ~6s after stop) | |
| B opens / views A’s last message | A: delivered → read | |
| A long-press → 👍 | Chip on both; REST `/v3/messages/react` + WS | |
| Privacy → typing/receipts off (if shipped) | Signals suppressed | |

Envelope must be `WSEnvelope` with `to` = peer DID (not `WSRelayMessage`).

---

## 8. New phone restore (D5)

Requires: A (or B) created a **cloud backup** with the phrase, and wallet linked.

1. On the device that will be “old phone”: Settings → Backup → upload / enable daily backup (`RecoveryCoordinator` first-backup on confirmation if that path ran).
2. Confirm `POST /v3/backup/push` 2xx.
3. Wipe B (or use a third simulator): FirstRun → **Restore from phrase**.
4. Enter 24 words → `restore-challenge` → device signs → `restore-did` JWT.
5. App pulls `/v3/backup/pull` and decrypts with phrase.
6. Messages hub shows D2/D3 threads.

| Log | Pass |
|-----|------|
| Backend `restore-did` issues `access_token` (not `pending:wallet:` stub DID) | ☐ |
| iOS `echo.backup.restoredAfterRecovery` / restore count > 0 | ☐ |
| Ciphertext not in logs | ☐ |

---

## 9. Groups (D6) — optional same day

| Step | Expected | ☐ |
|------|----------|---|
| A creates group, adds B | Thread opens for both | |
| A sends “hello-d6” | B decrypts | |
| B replies | A decrypts | |
| B force-quit, A sends | Delivers on B reconnect (queue) | |

Group typing/reactions may lag 1:1 parity — note, do not block D2–D3.

---

## 10. What **not** to treat as a D2 fail

| Symptom | Likely cause |
|---------|----------------|
| B never wakes while locked | APNs env missing — still pass D3 if open-app delivery works |
| restore-did unknown DID | Wallet never linked |
| Simulator A works, physical B fails | `API_URL` still `localhost` |
| Typing never shows | Privacy toggle off, or WS `to` not peer DID |
| Hub empty after restore | Backup never pushed; or phrase mismatch |
| `Color.Echo` / Icy chrome on chat | Design debt — not a functional fail |
| Channels / calls / PSI | Out of scope unless you scheduled them |

---

## 11. Sign-off (copy into notes)

**Date:** ________ **API:** ________ **Build:** ________ **APNs:** on / off **Postgres:** on / off

| Gate | Result |
|------|--------|
| D0 connectivity C1–C10 | ☐ |
| D1 @username or QR | ☐ |
| D2 live DM | ☐ |
| D3 offline (open-app) | ☐ |
| D3 APNs wake (if keys) | ☐ / N/A |
| D4 signals | ☐ |
| D5 restore | ☐ / N/A |
| D6 groups | ☐ / N/A |
| No T0 plaintext in logs | ☐ |

**Ship this session:** ☐ messaging MVP proven · ☐ hold (note blocker)

Automated companion: `make regression` and, before TestFlight, `make regression-with-phase1`.
