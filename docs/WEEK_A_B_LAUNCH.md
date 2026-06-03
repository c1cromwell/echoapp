# Week A & B — Launch execution

Companion to [`E2E_QUICK_START.md`](E2E_QUICK_START.md) and [`COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md`](COMPETITIVE_AUDIT_IMPLEMENTATION_PLAN.md) Wave 0.

---

## Week A — Messaging go/no-go (critical path)

**Goal:** Two clients can DM with usable chat UI; headless gates green.

### Day 1 — Agent / Xcode prep

```bash
make dev
make phase3-signals-proof
make ios-preflight BUILD=1
```

**Code landed (chat):** hide `GlacialTabBar` on pushed chat; composer in `safeAreaInset` per [`design-previews/phaseA-chat.html`](design-previews/phaseA-chat.html).

### Day 2 — Two-simulator manual (you)

| Step | User A | User B | Pass |
|------|--------|--------|------|
| 1 | Complete onboarding | Complete onboarding | Both on `make dev` |
| 2 | New conversation → search B's @username | — | Thread opens |
| 3 | Send plaintext "hello" | — | A shows sent |
| 4 | — | Open same thread (search A or hub) | B sees message in chat |
| 5 | B replies | A has chat open | A sees reply |
| 6 | Type in field | Other chat open | Typing label |
| 7 | Long-press message → 👍 | — | Reaction chip |
| 8 | Settings → Privacy toggles | Repeat | Typing/receipts respect off |

**Thread id:** both sides must use `ConversationID.direct` → `dm:{sorted-did}:{sorted-did}` (automatic via `ContactThreadHelper`).

### Day 3 — Sign-off

- [ ] `make regression` green
- [ ] Internal TestFlight build from `main`
- [ ] Software Factory: WO-192, WO-10 → `in_review` after you confirm E2E (not before)

---

## Week B — Contacts credibility (after Week A green)

**Goal:** Growth path without groups backend.

### Track (WO-222 + WO-39 partial)

| Item | Verify |
|------|--------|
| `echo://invite?code=` and `echo://invite/CODE` | Cold start stashes; post-login sheet |
| Profile QR → add contact | `POST /v3/contacts/add` |
| Contacts tab → list → Message | Hub thread + chat |
| Block contact | Detail → Block |
| PSI (optional) | `make echooprf-ios` + embed framework; two devices |

**Defer:** Groups create, Channels, WO-288 device link, `/v1/login/link-device` aliases.

### Pick one for TestFlight 2 story

- **WO-100** — `OIDC4VC_ENABLED=true` wallet enrollment E2E, or  
- **WO-221** — live OPRF PSI two-device match

---

## Out of scope this sprint

- `DeviceLinkFlowViews` / WO-288 (Wave 1)
- Groups "New group" beyond coming-soon sheet
- Hub Channels segment implementation
