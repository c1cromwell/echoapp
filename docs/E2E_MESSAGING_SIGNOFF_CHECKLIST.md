# Echo Messaging — E2E Sign-Off Checklist (§6.4–6.16)

> **Purpose:** Single run sheet for declaring M0–M7 messaging waves complete on real devices.
> Full detail lives in [`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md). Run in order; mark each row before moving on.

## Prerequisites (all sections)

| # | Check | Pass |
|---|--------|------|
| P1 | Backend up: `make dev` (or LAN cluster) | ☐ |
| P2 | Two iOS devices/simulators signed in (A + B) on same backend | ☐ |
| P3 | Xcode build succeeds (`EchoApp` scheme, iOS 17+) | ☐ |
| P4 | Automated pre-flight green (run once before manual matrix): | ☐ |

```bash
make release-check
go test ./internal/api/ -run 'Signal|Call|Media|Overflow|Archive|Poll|Sync|Backup'
cd ios/Echo && swift test --filter 'EchoPhase3Tests|EchoSecurityTests'
```

| P5 | WebRTC package linked in Xcode (M4c): add SPM `https://github.com/stasel/WebRTC.git` **131.0.0** to `EchoApp` target — see [`ios/WEBRTC_XCODE_SETUP.md`](../ios/WEBRTC_XCODE_SETUP.md) | ☐ |
| P6 | Microphone + camera permissions granted on both devices (Settings → Echo) | ☐ |

**Sign-off owner:** _______________ **Date:** _______________

---

## §6.4 — Phase 3 signals (M0 / WO-192, WO-10)

**Devices:** A + B, same `dm:` thread.

| # | Step | Expected | Pass |
|---|------|----------|------|
| 4.1 | A types in open chat | B sees typing indicator | ☐ |
| 4.2 | A stops typing ~6s | Indicator clears on B | ☐ |
| 4.3 | B opens chat, scrolls to A's message | A sees delivered → read | ☐ |
| 4.4 | A long-press → 👍 reaction | B sees reaction; REST/WS match | ☐ |
| 4.5 | Third user C (non-participant) | No WS signal received | ☐ |

**Notes:** _______________________________________________

---

## §6.4b — Message operations (M1 / WO-25, WO-84, WO-59)

**Devices:** A + B, same DM.

| # | Step | Expected | Pass |
|---|------|----------|------|
| 4b.1 | A edits a sent message | B sees updated text | ☐ |
| 4b.2 | A deletes a message | Tombstone on both clients | ☐ |
| 4b.3 | A pins message (≤5) | Pin visible on B after sync | ☐ |
| 4b.4 | A sets disappearing timer | Messages expire on schedule both sides | ☐ |
| 4b.5 | A replies / forwards | Thread metadata shows quote/source | ☐ |

**Notes:** _______________________________________________

---

## §6.9 — Group messaging (M2 / WO-207)

**Devices:** A (admin), B (member); optional C for remove/rekey.

| # | Step | Expected | Pass |
|---|------|----------|------|
| 9.1 | A: New Group → name + pick B | Group thread opens | ☐ |
| 9.2 | Wait ~5s on B | `group_key` received; B can send | ☐ |
| 9.3 | A sends group message | B decrypts plaintext | ☐ |
| 9.4 | B replies | A decrypts inbound | ☐ |
| 9.5 | A removes B (or C) | `requires_rekey`; survivors get new key | ☐ |
| 9.6 | Removed device tries new messages | Cannot decrypt post-rekey traffic | ☐ |
| 9.7 | B offline → A sends 2 msgs → B reconnects | Queued group text delivered | ☐ |

**Notes:** _______________________________________________

---

## §6.10 — Multi-device history sync (M3b / WO-CA3)

**Devices:** A (primary + history), B (new linked device via QR).

| # | Step | Expected | Pass |
|---|------|----------|------|
| 10.1 | A: Settings → Devices → Link new device → QR | Token issued | ☐ |
| 10.2 | B: Link this device → scan QR | Registration succeeds | ☐ |
| 10.3 | Within ~10s on A | “Message history sent to …” | ☐ |
| 10.4 | B: login → Messages | Prior DM threads appear | ☐ |
| 10.5 | A sends new DM while B online | Live WS; no duplicate merge | ☐ |
| 10.6 | B force-quit; A sends 2 DMs; B reopens | WS delivers new traffic | ☐ |
| 10.7 | A removes B's device | B gets `403 DEVICE_REVOKED` on sync | ☐ |

**Notes:** _______________________________________________

---

## §6.11 — Encrypted backup (M3c / WO-64 + WO-CA2)

**Device:** A (single device with DM history).

| # | Step | Expected | Pass |
|---|------|----------|------|
| 11.1 | Settings → Backup → Back Up Now + phrase | Local `.enc` + cloud upload OK | ☐ |
| 11.2 | Note thread count N | — | ☐ |
| 11.3 | Delete app / clear data; reinstall + login | Messages empty | ☐ |
| 11.4 | Restore from Cloud + same phrase | N threads restored | ☐ |
| 11.5 | Server logs / DB | Ciphertext only, no plaintext | ☐ |
| 11.6 | Enable auto-backup (daily) + phrase once | Keychain session key saved | ☐ |
| 11.7 | Foreground after due time | Auto-backup runs (wifi-only respected) | ☐ |
| 11.8 | Restore from Local | Documents backup restores without cloud | ☐ |

**Notes:** _______________________________________________

---

## §6.12 — Calls (M4c / WO-196 + WebRTC)

**Devices:** A + B, same LAN. **Requires WebRTC SPM linked (P5).**

| # | Step | Expected | Pass |
|---|------|----------|------|
| 12.1 | A: contact B → Voice call | CallKit outgoing; SDP offer sent | ☐ |
| 12.2 | B online (Messages tab open) | Incoming CallKit + full-screen call UI | ☐ |
| 12.3 | B accepts | Answer + ICE trickle; **audio connects** (hear tone/voice) | ☐ |
| 12.4 | Either party hangs up | `hangup` signal; both return to prior screen | ☐ |
| 12.5 | B declines incoming call | A sees "Declined"; missed in history | ☐ |
| 12.6 | A: Video call | Local PiP + remote video when connected | ☐ |
| 12.7 | Mute / speaker toggles | Audio route changes on device | ☐ |
| 12.8 | B offline; A calls | B gets missed-call push (`type: missed_call`) | ☐ |
| 12.9 | Optional: `ECHO_TURN_URL` set on backend | ICE config includes TURN entry | ☐ |

**Notes:** _______________________________________________

---

## §6.13 — Media messages (M5 / WO-194)

**Devices:** A + B; trust tier ≥2 for upload.

| # | Step | Expected | Pass |
|---|------|----------|------|
| 13.1 | A: photo in 1:1 DM | Optimistic UI; encrypted upload; thumbnail cached | ☐ |
| 13.2 | B online | Decrypt + preview renders | ☐ |
| 13.3 | A: voice note (hold mic) | Waveform bubble; B sees preview | ☐ |
| 13.4 | A: photo in **group** chat | All members see `MediaBubbleView` | ☐ |
| 13.5 | Dev: offline queue >1000 | Overflow manifest on reconnect | ☐ |

**Notes:** _______________________________________________

---

## §6.14 — Message search (M6a / WO-3 + WO-16)

**Device:** A (single device with DM history).

| # | Step | Expected | Pass |
|---|------|----------|------|
| 14.1 | Send DMs with unique keywords | Index updates within ~5s | ☐ |
| 14.2 | Messages hub → Search → keyword | Ranked hit + snippet | ☐ |
| 14.3 | Filter Voice / Photos | Results narrow by type | ☐ |
| 14.4 | Application Support | `search_index/index.enc` exists | ☐ |
| 14.5 | Recent searches + Clear | History works | ☐ |

**Notes:** _______________________________________________

---

## §6.15 — Archive, screenshots, polls, index sync (M6b)

**Devices:** A + B where noted.

| # | Step | Expected | Pass |
|---|------|----------|------|
| 15.1 | Archive conversation | Leaves main Chats list | ☐ |
| 15.2 | Hub → Archived | Thread listed; unarchive restores | ☐ |
| 15.3 | Search archived thread keyword | Excluded from default search | ☐ |
| 15.4 | A linked to B → DMs → index push | B search finds A's keywords | ☐ |
| 15.5 | Screenshot alert enabled; A screenshots | B gets `screenshot_alert` WS | ☐ |
| 15.6 | A creates poll | B sees bubble; vote syncs counts | ☐ |
| 15.7 | Chat settings archive + hub link | End-to-end archive UX | ☐ |

**Notes:** _______________________________________________

---

## §6.16 — On-device AI (M7a / WO-CA1)

**Device:** A (single device with DM history).

| # | Step | Expected | Pass |
|---|------|----------|------|
| 16.1 | Open DM with recent messages | `SmartReplyBar` shows 2–3 chips | ☐ |
| 16.2 | Tap chip | Composer fills suggestion text | ☐ |
| 16.3 | Disable smart replies (dev/consent) | Bar hidden | ☐ |
| 16.4 | 3+ message thread + summaries on | On-device summary available | ☐ |

**Notes:** _______________________________________________

---

## Final sign-off

| Wave | Section | Signed off |
|------|---------|------------|
| M0 signals | §6.4 | ☐ |
| M1 message ops | §6.4b | ☐ |
| M2 groups | §6.9 | ☐ |
| M3 multi-device | §6.10 | ☐ |
| M3 backup | §6.11 | ☐ |
| M4 calls | §6.12 | ☐ |
| M5 media | §6.13 | ☐ |
| M6 search | §6.14 | ☐ |
| M6 advanced | §6.15 | ☐ |
| M7 AI foundation | §6.16 | ☐ |

**Messaging parity (M0–M6) ready for TestFlight:** ☐ Yes ☐ No — blockers: _______________

---

## Related

- [`E2E_LAUNCH_AND_TESTING.md`](E2E_LAUNCH_AND_TESTING.md) — full launch matrix
- [`MESSAGING_COMPLETION_PLAN.md`](MESSAGING_COMPLETION_PLAN.md) — wave status
- [`ios/WEBRTC_XCODE_SETUP.md`](../ios/WEBRTC_XCODE_SETUP.md) — WebRTC.framework linking
- [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) — Phase 3 signal UI detail
