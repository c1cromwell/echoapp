# Week A & B — Launch execution

Companion to [`E2E_QUICK_START.md`](E2E_QUICK_START.md), [`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md), and [`COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md`](COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md) Wave 0.

---

## Wiring status (code vs E2E)

**Legend:** ✅ wired in app · 🔜 manual E2E required · ⬜ stub / not started

### Week A — Messaging

| Capability | Code | E2E (you) |
|------------|:----:|:---------:|
| `dm:` thread id via `ContactThreadHelper` / `ConversationID.direct` | ✅ | 🔜 |
| `ChatView` + `ChatDetailViewModel` (send, WS connect) | ✅ | 🔜 |
| Tab-level shared WS + per-chat conversation handler | ✅ | 🔜 |
| Local thread history (`ConversationThreadStore`) | ✅ | 🔜 A10 |
| Inbound text → open chat + hub preview | ✅ | 🔜 A4–A5 |
| Unread bump (skipped when chat open / own sender) | ✅ | 🔜 |
| Typing indicator + privacy merge (global + persona) | ✅ | 🔜 A6, A8 |
| Read receipts + checkmarks on sent bubbles | ✅ | 🔜 A9 |
| Reactions (REST + WS, picker/chips) | ✅ | 🔜 A7 |
| Glacial tab bar hidden on chat; composer `safeAreaInset` | ✅ | — |
| Copy message action | ✅ | — |
| Reply / forward / pin / edit | ✅ local | 🔜 |
| Chat profile “groups in common” | ✅ API | 🔜 (Week B B3 overlap) |

### Week B — Contacts (after Week A green)

| Capability | Code | E2E (you) |
|------------|:----:|:---------:|
| `echo://invite` deep link | ✅ | 🔜 B1 |
| Profile QR add contact | ✅ | 🔜 B2 |
| `ContactDetailView` block / favorite / Message | ✅ | 🔜 B3 |
| Mutual groups + mutual contacts on contact detail | ✅ | 🔜 B5 |
| SMS phone backup (`SMSOTPSetupView`) | ✅ | 🔜 B4 |
| WO-288 link device QR + login scan | ✅ | 🔜 B6 |
| Live OPRF PSI (`EchoOPRF.xcframework`) | 🔜 embed | 🔜 optional |
| OIDC4VC wallet enrollment | ✅ iOS | 🔜 `OIDC4VC_ENABLED` |

---

## Week A — Messaging go/no-go (critical path)

**Goal:** Two clients can DM with usable chat UI; headless gates green.

### Day 1 — Agent / Xcode prep

```bash
make dev
make phase3-signals-proof
make ios-preflight BUILD=1
```

**Code landed (chat):** hide `GlacialTabBar` on pushed chat; composer in `safeAreaInset` per [`design-previews/phaseA-chat.html`](design-previews/phaseA-chat.html). Thread persistence + inbound relay (May 2026).

### Day 2 — Two-simulator / TestFlight manual (you)

Use [`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md) steps **A1–A10** (same table, tester-friendly).

| Step | User A | User B | Pass | Code |
|------|--------|--------|------|:----:|
| 1 | Complete onboarding | Complete onboarding | Both on `make dev` | ✅ |
| 2 | New conversation → search B's @username | — | Thread opens | ✅ |
| 3 | Send plaintext "hello" | — | A shows sent | ✅ |
| 4 | — | Open same thread (search A or hub) | B sees message in chat | ✅ |
| 5 | B replies | A has chat open | A sees reply | ✅ |
| 6 | Type in field | Other chat open | Typing label | ✅ |
| 7 | Long-press message → 👍 | — | Reaction chip | ✅ |
| 8 | Settings → Privacy toggles | Repeat | Typing/receipts respect off | ✅ |
| 9 | Reopen chat after messages | — | History persists | ✅ |
| 10 | — | — | (TestFlight script A10) | ✅ |

**Thread id:** both sides must use `ConversationID.direct` → `dm:{sorted-did}:{sorted-did}` (automatic via `ContactThreadHelper`).

### Day 3 — Sign-off

- [ ] Week A manual A1–A10 passed ([`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md))
- [ ] `make regression` green
- [ ] Internal TestFlight build from `main`
- [ ] Software Factory: WO-192, WO-10 → `in_review` after you confirm E2E (not before)

---

## Week B — Contacts credibility (after Week A green)

**Goal:** Growth path without groups backend UI.

Use [`TESTFLIGHT_WEEK_A_B.md`](TESTFLIGHT_WEEK_A_B.md) steps **B1–B6**.

### Track (WO-222 + WO-39 partial)

| Item | Verify | Code |
|------|--------|:----:|
| `echo://invite?code=` and `echo://invite/CODE` | Cold start stashes; post-login sheet | ✅ |
| Profile QR → add contact | `POST /v3/contacts/add` | ✅ |
| Contacts tab → tap row → **View profile** | `ContactDetailView` with block / favorite | ✅ |
| Contacts → **Message** (detail or swipe) | Hub thread + chat | ✅ |
| **Add phone number** | Settings → Privacy or Account → `SMSOTPSetupView` | ✅ |
| **Groups in common** | Contact detail shows mutual groups when both users share group membership | ✅ |
| Block contact | Contact detail → Block → `POST /v3/contacts/block` | ✅ |
| Favorites | Context menu / swipe → star; **Favorites only** filter | ✅ |
| PSI (optional) | `make echooprf-ios` + embed framework; two devices | 🔜 |

### WO-288 multi-device (third sim optional)

| Item | Verify | Code |
|------|--------|:----:|
| Link QR | Settings → Devices → **Link new device** | ✅ |
| Scan | Login → **Sign in on new device** | ✅ |
| API | `/v1/login/link-device/*` aliases | ✅ |

### Pick one for TestFlight 2 story

- **WO-100** — `OIDC4VC_ENABLED=true` wallet enrollment E2E, or  
- **WO-221** — live OPRF PSI two-device match

---

## Week B+ — Message actions (landed on `main`)

| Action | Behavior |
|--------|----------|
| **Reply** | Composer quote strip; stores reply metadata on sent message |
| **Forward** | Sheet → pick another thread; `↪` prefix + local + WS send |
| **Pin** | One pinned message per chat (banner); toggle to unpin |
| **Edit** | Own messages within 15 minutes; local content update |

No server-side edit/delete sync yet — local thread + relay text only.

---

## Out of scope this sprint

- Groups create UI beyond coming-soon sheet
- Hub Channels segment implementation
- Cross-device edit/delete reconciliation
