# Echo iOS UI — Implementation Spec (Design System v3)

**Source of truth (post-auth interaction + layout):** [`docs/Echo Design System Setup_latest/`](Echo%20Design%20System%20Setup_latest/) — React prototype for **messages, chat, groups, channels, contacts, settings** only. **Not** for onboarding/login (§0).

**Source of truth (visual tokens on iOS):** [`docs/ux-spec.md`](ux-spec.md) §4 — **warm paper / ink / signal** (`echoPaper`, `echoInk`, `echoSignal`, `echoNight`). Do **not** ship the prototype’s Icy Minimal slate/sky hex values on iOS unless explicitly migrating tokens.

**Reference apps:** [Signal](https://signal.org/) (privacy, chat settings, safety numbers, disappearing messages) · [Telegram](https://telegram.org/) (folder filters, channels, pinned chats, message actions sheet, compose speed).

**Companion specs:** [`PHASE3_IOS_UI_SPEC.md`](PHASE3_IOS_UI_SPEC.md) (typing · receipts · reactions) · [`E2E_QUICK_START.md`](E2E_QUICK_START.md) (manual QA).

---

## 0. Frozen scope — do not change onboarding or login

The **shipped iOS onboarding and login UX is correct and frozen.** Agents and implementers must **not** redesign, replace, or re-map these flows from the React design prototype.

| In scope for prototype → iOS work | Out of scope (frozen) |
|--------------------------------|------------------------|
| Messages hub, chat, folders, pins, actions | **All onboarding screens** |
| Groups, channels, contacts (post-auth) | **Login / unlock / recovery entry** |
| Settings (except auth entry redesign) | **Account locked, trust intro on login path** |
| Personas & hidden folders (post-auth) | Prototype `src/app/pages/onboarding/*` |
| | Prototype `Echo_v3_2_Figma_Design_Spec.md` §3–5 (carousel, login, create account) |

**Canonical iOS flows (implement only bugfixes; no visual/flow refactors):**

- **New user:** `FirstRunCoordinator` → `EchoWelcomeView` → `DisplayNameEntryView` → `OnboardingOptionsView` → Face ID → optional `VIPPathView` → `RecoverySetupView` → main app.
- **Returning user:** `GlacialLoginScreen` → Face ID / passkey / recover → main app.
- **Recovery:** `RecoveryCoordinator`, `SMSOTPSetupView`, phrase views as wired today.

Do **not** adopt prototype routes such as `intro-carousel`, `onboarding/login.tsx`, `phone-entry` / `otp-verify` as primary paths, or merge onboarding into a 3-page marketing carousel. E2E validation: [`E2E_QUICK_START.md`](E2E_QUICK_START.md) §2.

---

## 1. Design system review summary

The **Echo Design System Setup_latest** package is a full **web prototype** (Vite + React Router) that models production UX beyond what iOS has wired today. It includes:

| Area | Prototype routes / files | iOS today |
|------|--------------------------|-----------|
| Onboarding | `src/app/pages/onboarding/*` | **Frozen** — `FirstRunCoordinator` (ignore prototype) |
| Login | `onboarding/login.tsx`, auth specs | **Frozen** — `GlacialLoginScreen` (ignore prototype) |
| Messages hub | `pages/messages.tsx`, `conversation-item`, `pinned-section` | `MessagesTabView`, `ConversationListView` (partial) |
| 1:1 chat | `pages/chat.tsx`, `message-bubble`, `message-actions-sheet` | `ChatView` / `MessagingScreens` (mock types) |
| Chat settings | `chat-settings-sheet.tsx` | Partial in privacy settings |
| Folder filters | `chat-folder-filter.tsx` | Not implemented |
| Groups | `group-*`, `GROUPS_IMPLEMENTATION.md` | Backend service; minimal UI |
| Channels | `channel-*`, `BROADCAST_CHANNELS_IMPLEMENTATION_SPEC.md` | Not in iOS |
| Personas | `personas.tsx`, `persona-switcher.tsx` | `PersonasManagementView`, `PersonaGateView` |
| Hidden folders | `hidden-folders/*` | Night surface patterns exist |
| Contacts / trust / rewards / analytics | matching `pages/*` | Partial settings / profile |

**Color policy for iOS implementation:** Map prototype **layout, density, and flows** for **post-auth messaging and social features** only; map prototype CSS variables to **existing Swift tokens** in `GlacialTheme.swift` / `Theme.swift`, not to `#0EA5E9` Icy accent unless product approves a token migration. **Do not** remap onboarding/login screens from prototype CSS.

```text
Prototype token          →  iOS token (keep current scheme)
--echo-accent / sky      →  echoSignal (#0E7AB8)
--echo-bg-primary        →  echoPaper (#FBFBF7)
--echo-text-primary      →  echoInk
--echo-bg-elevated card  →  echoPaperDim
--echo-success           →  echoTrustGreen (#1F7A4C)
--echo-error             →  echoAlert
private / vault surfaces →  echoNight / echoNightHi
```

---

## 2. Signal & Telegram — pattern matrix

Use this when implementing or reviewing screens. **Echo differentiators** stay: DID identity, trust tiers, Secure Enclave, no phone required.

| Pattern | Signal | Telegram | Echo (target) |
|---------|--------|----------|---------------|
| Chat list density | Compact rows, avatar, preview, time | Same + folders/chips | Match Signal density; add Telegram-style **folder filter chips** (trust-based, not just unread) |
| Pinned chats | Pin icon, section at top | Pinned + saved messages | `PinnedSection` — max 5 pins, reorder (prototype) |
| New message / compose | FAB or header + | Pencil / new chat | Header `+` → new DM / new group (prototype) |
| Message long-press | Copy, reply, forward, delete, react | Rich grid sheet + react | **Bottom sheet** grid (`MessageActionsSheet` parity) — see §5.3 |
| Reply / thread | Quote reply, limited threads | Reply + topic threads | Phase 3+: inline reply quote; threads deferred |
| Reactions | Emoji react | Emoji react | `ReactionPickerView` + REST (PHASE3 spec) |
| Read receipts | Optional, double check | Optional | WS + `DeliveryStatus` (PHASE3) |
| Typing indicator | Optional dots | Optional | WS ephemeral (PHASE3) |
| Chat info / settings | Per-chat mute, media, safety number | Mute, TTL, members | `ChatSettingsSheet`: silent, disappearing timer, verify (§5.4) |
| Disappearing messages | Timer per chat | Secret chat TTL | Off / 30s / 5m / 1h / 24h / 7d (prototype) |
| Contact discovery | Phone hash (optional) | Phone + username | PSI + username search (WO-221) |
| Channels / broadcast | N/A | Channels, unlimited subs | Echo **Broadcast Channels** (spec in prototype imports) |
| Groups | Groups v2 | Supergroups | Echo **Groups** RBAC (backend `internal/services/groups`) |
| App lock | Screen lock, registration lock | Passcode lock | Face ID + `StorageLockedView` |
| Night / private UI | N/A (minimal) | Secret chat dark theme | **echoNight** for hidden persona + hidden folders |

---

## 3. Navigation & app shell

### 3.1 Tab bar (Signal-style primary destinations)

| Tab | Prototype | iOS target | Notes |
|-----|-----------|------------|-------|
| Messages | `/` | `MessagesTabView` | Default tab; hosts folder filters + lists |
| Contacts | `/contacts` | Contacts stack | Discovery, search, invites |
| Personas | `/personas` | Personas | Public + hidden (gated) |
| Profile / Settings | `/profile`, `/settings` | Profile tab | Identity card, privacy hub |

**Deferred tabs** (prototype only until backend ready): Rewards, Analytics, Trust dashboard as **settings deep links**, not tab bar items (Signal keeps wallet out of primary tabs).

### 3.2 Messages hub layout (Telegram-informed)

Vertical order (from `messages.tsx`):

1. **Persona header** — avatar, display name, **trust badge** (tap → trust explainer). *Signal: no persona; Echo: multi-persona.*
2. **Search** — filters conversations (Signal global search parity).
3. **Pinned section** — collapsible; pin icon on rows.
4. **Folder filter chips** — `All | Verified | Trusted | Unverified` (map to trust tier opacities, not rainbow colors).
5. **Conversation list** — 1:1 and groups mixed; badges for mute, disappearing, scheduled, delivery failure.
6. **Secondary lists** (segmented or tabs): **Groups** · **Channels** · **Hidden folders** (last requires biometric gate).

SwiftUI: prefer single `MessagesHubView` coordinating `@Observable` view models; avoid duplicating mock `MessagingScreens` types.

---

## 4. Feature implementation map

Priority: **P0** TestFlight · **P1** Phase 2–3 · **P2** post-launch.

### 4.1 Onboarding & login — **frozen (no prototype work)**

| Feature | iOS (canonical — do not change) | Prototype | Agent rule |
|---------|-----------------------------------|-----------|------------|
| Welcome | `EchoWelcomeView` | `onboarding/welcome.tsx` | **Ignore prototype** |
| Username | `DisplayNameEntryView` | `profile-setup` | **Ignore prototype** |
| Face ID + VIP opt-in | `OnboardingOptionsView` | `onboarding-options` | **Ignore prototype** |
| VIP path | `VIPPathView`, enrollment | `verify-credentials` | Bugfixes only |
| Recovery | `RecoverySetupView`, `SMSOTPSetupView` | `recovery`, `otp-verify` | Bugfixes only |
| Login | `GlacialLoginScreen` | `login.tsx` | **Ignore prototype** |
| Lockout / locked | `BiometricLockoutView`, `AccountLockedView` | `account-locked` | Bugfixes only |
| Device recover | `RecoveryCoordinator` | `recover-*` | Bugfixes only |

**Flow authority:** `FirstRunCoordinator.swift` + [`ux-spec.md`](ux-spec.md) §2.1–2.4, §3.1–3.2. **Not** `Echo_v3_2_Figma_Design_Spec.md` onboarding/login sections.

### 4.2 Messaging core (P0–P1)

| Feature | Prototype component | iOS / service | Priority |
|---------|---------------------|---------------|----------|
| Conversation list row | `conversation-item.tsx` | Unify with domain models | P0 |
| Chat bubbles | `message-bubble` | `MessageBubble` + `DeliveryStatus` | P0 |
| Send pipeline | chat input bar | `ChatDetailViewModel` | P0 |
| Secure thread bar | (global) | `SecureThreadIndicator` | ✅ |
| Typing / receipts / reactions | — | PHASE3 spec | P1 |
| Message actions sheet | `message-actions-sheet.tsx` | New `MessageActionsSheet.swift` | P1 |
| Chat settings sheet | `chat-settings-sheet.tsx` | New `ChatSettingsSheet.swift` | P1 |
| Silent / mute chat | toggle in settings | Per-conversation prefs store | P1 |
| Disappearing messages | timer options | Local + policy flag; backend TTL TBD | P2 |
| Pin message | pin action | Local pin + optional sync | P2 |
| Forward / copy | actions | UIPasteboard + forward picker | P2 |
| Scheduled messages | schedule action | Deferred (UI mock only) | P3 |

### 4.3 Messages hub enhancements (P1)

| Feature | Prototype | iOS work |
|---------|-----------|----------|
| Folder filter bar | `chat-folder-filter.tsx` | `ChatFolderFilterView` — filter by `TrustBadge` tier |
| Persona switcher | `persona-switcher.tsx` | Wire `PersonaSwitcher` in hub header |
| Groups list embed | `groups-list.tsx` | Navigation to group detail |
| Channels list embed | `channels-list.tsx` | Placeholder until API |
| Hidden folders entry | `hidden-folder-data` | Biometric gate → night theme list |

### 4.4 Contacts & discovery (P0–P1)

| Feature | Prototype | iOS | Backend |
|---------|-----------|-----|---------|
| Contact list | `contacts.tsx` | `ContactsScreens` | `/v3/contacts` |
| PSI discovery | — | `ContactDiscoveryView` | `/v3/contacts/psi` |
| Username search | — | `UsernameSearchView` | `/v3/contacts/search` |
| Invite links | — | `InviteLinkSheet` | invite API |
| Secure contacts online | `secure-contacts-online` | Optional strip on hub | presence TBD |

### 4.5 Groups (P2)

**Spec:** `Echo Design System Setup_latest/src/imports/GROUPS_IMPLEMENTATION.md`  
**Backend:** `internal/services/groups/`

| Screen | Route | iOS (new) |
|--------|-------|-----------|
| Create group | `/groups/create` | `GroupCreateView` |
| Discover | `/groups/discover` | `GroupDiscoverView` |
| Group detail | `/groups/:id` | `GroupDetailView` |
| Group chat | `/group-chat/:id` | Reuse `ChatView` with `conversationType = .group` |

Signal parity: group info → members, mute, media gallery. Telegram parity: large member list, admin badges.

### 4.6 Broadcast channels (P3)

**Spec:** `BROADCAST_CHANNELS_IMPLEMENTATION_SPEC.md` (prototype imports)

| Screen | Route | Notes |
|--------|-------|-------|
| Discover | `/channels/discover` | Telegram channel directory analog |
| Create | `/channels/create` | Creator flow |
| Channel detail | `/channels/:id` | Subscribe, mute, description |
| Channel feed | read-only chat UI | Admin-only post; comment threads optional |

### 4.7 Personas & hidden folders (P1–P2)

| Feature | Prototype | iOS |
|---------|-----------|-----|
| Persona list | `personas.tsx` | `PersonasManagementView` |
| Create persona | `persona-create` | New |
| Hidden folder list | `hidden-folders` | Night theme + `PersonaGateView` |
| Hidden folder chat | `hidden-folder-chat` | Same chat components on `echoNight` |

Signal reference: no direct analog — Echo differentiator. Telegram reference: **folder lock** / archived chats aesthetic only.

### 4.8 Settings & identity (P0–P1)

| Screen | Prototype | iOS |
|--------|-----------|-----|
| Settings hub | `settings.tsx` | Settings stack |
| Privacy / discovery | toggles | `PrivacyHubView`, `ContactDiscoverySettingsView` |
| Devices | `device-management` | `DeviceManagementView` |
| Login audit | `login-audit` | Security log (when API exists) |
| Appearance | theme picker in v3.2 spec | **iOS:** Light / Dark / System using paper/ink dark equivalents — no multi-theme picker |

### 4.9 Wallet, trust, rewards, analytics (P2+)

Prototype pages exist; keep as **deep links** from Profile until DAG wallet TestFlight is in scope. Trust **badge** surfaces on Messages hub (P1); full trust dashboard P2.

---

## 5. Component specifications (iOS)

All components use **paper/ink** tokens and `SpringPressStyle` unless noted.

### 5.1 `ConversationRowView` (list cell)

**Reference:** Signal chat list row · Telegram dialog row.

| Element | Spec |
|---------|------|
| Height | min 72pt |
| Avatar | 48pt circle; online dot 10pt `echoTrustGreen` bottom-trailing |
| Title | 16pt semibold `echoInk`; trust badge 16pt after name |
| Preview | 14pt `echoInk55`, one line, tail truncate |
| Time | 12pt `echoInk40` trailing |
| Unread | `echoSignal` filled pill, max "99+" |
| Icons | mute (bell.slash), pin, disappearing timer, failed delivery (exclamation) |
| Swipe actions | Pin, Mute, Delete (Signal pattern) |

### 5.2 `ChatFolderFilterView`

**Reference:** Telegram folder tabs (horizontal chips).

| Filter | Maps to trust |
|--------|----------------|
| All | no filter |
| Verified | tier ≥ T2 |
| Trusted | tier ≥ T3 |
| Unverified | tier T0–T1 |

Active chip: `echoSignal` fill, white text. Inactive: `echoPaperDim` fill, `echoInk55` text. Unread count badge on chip optional.

### 5.3 `MessageActionsSheet`

**Reference:** Telegram long-press action grid · Signal action list.

| Action | Condition | Icon |
|--------|-----------|------|
| Reply | always | arrowshape.turn.up.left |
| Copy | always | doc.on.doc |
| Forward | always | arrowshape.turn.up.right |
| Pin | always | pin |
| Edit | sent, &lt; 15 min | pencil |
| Delete for everyone | sent | trash (echoAlert) |

Presentation: `.sheet` or custom bottom detent 40%; 4-column grid; message preview strip on top (prototype pattern).

### 5.4 `ChatSettingsSheet`

**Reference:** Signal contact/conversation settings · Telegram chat settings.

| Row | Control |
|-----|---------|
| Silent notifications | toggle |
| Disappearing messages | segmented: Off, 30s, 5m, 1h, 24h, 7d |
| Safety / verify | navigation → fingerprint / safety number (Signal **Safety Number** analog for DID) |
| Media & files | navigation (future) |
| Block / report | destructive |

### 5.5 `PinnedSectionView`

**Reference:** Telegram pinned chats.

Max 5 conversations; collapse chevron; show pin icon on row; long-press unpin.

### 5.6 Phase 3 overlays (existing spec)

- `TypingIndicatorView` — Signal three-dot bounce
- `ReactionPickerView` — Telegram emoji bar above keyboard
- `ReactionChipsView` — under-message chips

---

## 6. Build ownership (agent vs Xcode)

| Work | Agent (SPM) | Xcode (EchoApp target) |
|------|-------------|-------------------------|
| ViewModels, services, models | ✅ | — |
| New SwiftUI views from §5 | Compile in SPM where possible | Add to `EchoApp.xcodeproj` |
| Unify `ChatMessage` → domain `Message` | ✅ refactor | Wire `ChatView` |
| Design tokens | Document only | `GlacialTheme` / assets |
| Groups / channels API UI | Stubs + mocks | E2E when APIs live |

Skill: **`echo-ios-agent-vs-xcode`**, **`echo-phase3-ios-wire`**.

---

## 7. Phased delivery checklist

### Phase A — TestFlight messaging (current)

- [x] Onboarding & login — **frozen** (correct design; no prototype refactors)
- [x] `ChatDetailViewModel` wired in `ChatView` (typing, receipts, reactions, WS text relay)
- [x] PHASE3 typing / receipts / reactions (privacy toggles via `PrivacySettingsStore`)
- [x] Profile QR scan → add contact (`QRContactAddCoordinator`)
- [x] Encrypted text payloads (Kinnami + `/identity/resolve`; sim decrypt)
- [ ] Contact discovery + add contact
- [ ] Paper/ink on **post-auth** screens (onboarding/login already paper/ink)

### Phase B — Hub parity (design system)

- [ ] `MessagesHubView` with folder filters + pinned
- [ ] `MessageActionsSheet` + `ChatSettingsSheet`
- [ ] Persona header + switcher
- [ ] Per-chat mute / silent

### Phase C — Social graph

- [ ] Groups UI wired to `groups` service
- [ ] Invite + username search polish

### Phase D — Broadcast & advanced

- [ ] Channels read-only feed
- [ ] Disappearing messages policy
- [ ] Scheduled send (if product confirms)

---

## 8. Related files in design package

| Document | Purpose |
|----------|---------|
| `src/imports/Echo_v3_2_Figma_Design_Spec.md` | **Ignore §3–5** (onboarding/login) — messaging/theme tokens only if needed |
| `src/imports/Echo_v3.1.1_Messages_Figma_Spec.md` | Messages hub deltas |
| `src/imports/Echo_Auth_Figma_UX_Spec.md` | Auth flows |
| `src/imports/BROADCAST_CHANNELS_IMPLEMENTATION_SPEC.md` | Channels backend/UI spec |
| `src/imports/GROUPS_IMPLEMENTATION.md` | Groups backend summary |
| `src/app/routes.ts` | Complete route map |

---

## 9. Open decisions

1. **Dark mode:** Prototype supports Icy dark; iOS paper/ink dark palette needs explicit tokens in `Theme.swift` (follow system).
2. **Trust filter colors:** Prototype uses indigo/green chips; iOS uses **green opacity ladder** only — keep ux-spec §4.1.
3. **Reply threads:** Signal-style quote reply P1; full threads P3.
4. **Channels vs groups:** Channels are read-only broadcast; groups are participatory — do not merge UI.
