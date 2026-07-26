# Signal Parity — Software Factory work orders

**Epic:** [WO-330](https://app.softwarefactory.dev) — Signal Parity Waves S1–S5 (post-Sept)  
**Parent of epic:** WO-321 (Sept 1 go-live) — post-launch only; do not block soft launch  
**Assignee:** Chad Cromwell  
**Canonical matrix:** [`SIGNAL_ECHO_PARITY.md`](SIGNAL_ECHO_PARITY.md)

| Wave | WO | Title | Status intent | Scope |
|------|-----|-------|---------------|-------|
| **Epic** | **WO-330** | Signal Parity Epic — Waves S1–S5 | `in_progress` | Program umbrella |
| **S1** | **WO-333** | History sync + cloud backup productization | `in_progress` | Postgres SyncStore for `/v3/backup/*`; restore-after-recovery; sync pull error surfacing; two-device E2E |
| **S2** | **WO-331** | Sealed-sender + Sender Keys + safety numbers | `ready` | Redact `sender_did` after deliver; `GroupSenderKeyStore` in group send path; `SafetyNumberCompareView` |
| **S3** | **WO-334** | Group Phase 3 signals + calls polish | `ready` | Group typing/reactions UI; `/v3/calls/relays`; WebRTC/TURN/CallKit polish |
| **S4** | **WO-332** | View-once, message requests, GIF, NSE, wallpaper | `ready` | ViewOnce, MessageRequestGate, GifSearchService, NSE skeleton, ChatWallpaper, screenshot notify |
| **S5** | — | Per-step PQ + federation (later) | `backlog` | Evolve Echo ratchet; WO-320 principles |

## Implementation landed (scaffold → wired, 2026-07-26)

See [`SIGNAL_ECHO_PARITY.md`](SIGNAL_ECHO_PARITY.md) §6. Waves S1–S4 have code paths in-repo; remaining productization is two-device E2E, Xcode NSE target, and CallKit/WebRTC release config polish.
