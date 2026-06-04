# TestFlight — Week A & B tester script

**One page** for internal testers. Backend must be reachable from every device (see setup). Full detail: [`WEEK_A_B_LAUNCH.md`](WEEK_A_B_LAUNCH.md), [`E2E_QUICK_START.md`](E2E_QUICK_START.md).

---

## Before you start (organizer)

| Requirement | Action |
|-------------|--------|
| API running | Mac: `make dev` (or shared staging URL) |
| Device `API_URL` | Xcode scheme → Run → Environment → `API_URL=http://<Mac-LAN-IP>:8000` (not `localhost` on physical iPhone) |
| Same Wi‑Fi | Phone and Mac on one network |
| Dev SMS OTP | `.env` → `DEV_MODE=true`; restart API; use `X-Dev-Otp` header after register (see E2E quick start) |
| Two accounts | User A and User B — complete onboarding on each install |

**Build:** Internal TestFlight from `main` after organizer runs `make phase3-signals-proof`.

---

## Week A — Messaging (required)

Use **User A** and **User B**. Record pass/fail per row.

| # | User A | User B | Pass if |
|---|--------|--------|---------|
| A1 | Install → onboarding → Messages tab | Same | Both reach main app; no crash |
| A2 | Tap **+** → search **@username** of B → open chat | — | Chat opens; name matches B |
| A3 | Send **hello** | — | Message appears on A (sent checkmark) |
| A4 | — | Open same thread (hub row or search **@username** of A) | B sees **hello** in the thread (not only hub preview) |
| A5 | Keep chat open | Send **reply** | A sees reply without leaving chat |
| A6 | Type in composer (don’t send) | B’s chat open | B sees **typing…** / **{name} is typing…** |
| A7 | Long-press A’s message → **👍** | — | Reaction chip on message |
| A8 | Settings → Privacy → turn **typing** OFF | Repeat A6 | Typing label does **not** appear |
| A9 | Turn **read receipts** OFF on B | A sends; B opens chat | A’s checkmarks do **not** advance to read |
| A10 | Force-quit → reopen chat | — | Prior messages still visible |

**Thread rule:** Both sides must share the same `dm:…` thread (created via New conversation or Contacts → Message). If B sees nothing, confirm both use search/@username flow, not a stale demo thread.

**Report blockers with:** device model, `API_URL`, User A/B @usernames, step #, screenshot.

---

## Week B — Contacts (after Week A passes)

| # | Action | Pass if |
|---|--------|---------|
| B1 | Open `echo://invite?code=TEST` (Notes/Safari) cold or warm | Invite sheet or stash after login |
| B2 | Profile → QR → Scan other user’s code | Contact added; can Message |
| B3 | Contacts → tap row → **View profile** | Block / favorite / Message work |
| B4 | Settings → Privacy → **Add phone number** | SMS OTP completes |
| B5 | Star a contact → filter **Favorites only** | List filters |
| B6 | (Optional) Account → Devices → **Link new device** + second install **Sign in on new device** | QR link completes |

**Optional flagship (pick one for TestFlight 2):**

- **WO-100:** `.env` `OIDC4VC_ENABLED=true` → onboarding **Digital ID** / wallet enrollment completes.
- **WO-221:** Two phones with SMS backup + PSI discovery → shared phone contact appears as match.

---

## Out of scope (do not file as Week A/B bugs)

- Creating groups UI (coming soon sheet only)
- Hub **Channels** tab
- Reply / forward / pin / edit (not implemented)

---

## Sign-off (organizer only)

- [ ] Week A table A1–A10 passed on TestFlight build __________
- [ ] `make regression` green on commit shipped to TestFlight
- [ ] Software Factory WO-192, WO-10 → `in_review` (after Week A E2E only)
