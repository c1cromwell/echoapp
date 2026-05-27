---
name: echo-phase3-ios-wire
description: >-
  Xcode wiring checklist for Echo Phase 3 messaging UI — typing, read receipts,
  reactions. Use when wiring ChatView, adding SwiftUI to EchoApp.xcodeproj,
  manual two-client E2E, or finishing Phase 3 after the agent service layer.
---

# Echo Phase 3 iOS wire (Xcode)

Full spec: [`docs/PHASE3_IOS_UI_SPEC.md`](../../docs/PHASE3_IOS_UI_SPEC.md).

Agent layer is **landed** (`EchoPhase3Tests` green). This skill is for **Xcode-only** finish work.

Also read: skill **`echo-ios-agent-vs-xcode`**.

## Agent layer (already in repo — do not re-implement)

| File | Role |
|------|------|
| `Core/Networking/ConversationSignal.swift` | `WSEnvelope`, payloads |
| `Services/ConversationSignalService.swift` | WS send/receive |
| `Services/ReactionsAPI.swift` | REST reactions |
| `Features/Messaging/ChatDetailViewModel.swift` | Typing, receipts, reactions |
| `Features/Messaging/ConversationSignalLogic.swift` | Privacy + debounce helpers |
| `Core/DI/Container.swift` | `conversationSignalService`, `reactionsAPI` |

## Xcode UI files (in progress — add to `EchoApp.xcodeproj`)

| File | Status |
|------|--------|
| `Features/Messaging/TypingIndicatorView.swift` | Created — add to app target |
| `Features/Messaging/ReactionPickerView.swift` | Created — add to app target |
| `Features/Messaging/ReactionChipsView.swift` | Created — add to app target |
| `Presentation/Screens/Messaging/MessagingScreens.swift` | `ChatView` refactor in progress |

**Critical:** SPM alone does not update the app target. Every new `.swift` file → **Target Membership** in Xcode.

## ChatView wiring checklist

Current `ChatView` (`MessagingScreens.swift`) already includes much of this — verify each item:

- [ ] `@Bindable var viewModel: ChatDetailViewModel` (not mock-only state)
- [ ] `.task { configure(...); connect(accessToken:) }` with real `conversationId`, `peerDID`, `currentUserDID`
- [ ] `onChange(of: messageText)` → `viewModel.onInputChanged`
- [ ] Send button → `viewModel.onSendTapped()` after outbound send
- [ ] Peer bubble `onAppear` → `viewModel.onMessageAppeared(messageId:senderDID:)`
- [ ] `TypingIndicatorView` when `viewModel.peerIsTyping`
- [ ] Long-press → `ReactionPickerView`; chips → `ReactionChipsView`
- [ ] `toggleReaction` wired to picker/chips
- [ ] `mapDeliveryStatus` at view boundary only — logic uses `DeliveryStatus`
- [ ] `.onDisappear` / scene phase → `await viewModel.disconnect()`

## Navigation (often still missing)

`ConversationListView` uses mock data + `onSelectConversation(id)` callback. Wire parent (tab/coordinator) to:

1. Resolve `peerDID`, `conversationId`, `currentUserDID` from domain `Conversation` / auth session.
2. Factory `ChatDetailViewModel` from `DIContainer.shared` (inject `ConversationSignalService`).
3. `NavigationLink` or coordinator push → `ChatView(viewModel:..., contactName:, ...)`.

Do not pass empty `peerDID` / `conversationId` — WS signals require peer **`did:key`**.

## Privacy settings

Pass merged preferences into `configure(..., privacy:)`:

- Global: `EnhancedPrivacySettings.typingIndicators`, `.readReceipts`
- Persona: `PersonaPrivacySettings.sendTypingIndicators`, `.sendReadReceipts`
- Merge via `MessagingPrivacyPreferences.merged` in `ConversationSignalLogic.swift`

## Backend contracts (must match)

- WS: `WSEnvelope` with mandatory `to` = peer DID — **not** `WSRelayMessage`
- Types: `typing`, `read_receipt`, `reaction`
- Auth: `?token=` or Bearer on `/ws`
- REST: `POST/GET /v3/messages/react(ions)` via `PasskeySigningInterceptor`

## Verify in Xcode

```bash
# Unit tests (agent layer)
cd ios/Echo && swift test --filter EchoPhase3Tests

# App build (requires Xcode.app)
# Product → Build (EchoApp scheme)
```

MCP: `run_ios_phase3_tests`.

## Manual E2E (Step 5 — required before ship)

Prerequisites: `make dev`, `API_URL` → LAN backend, **two signed-in accounts**.

| Test | Pass criteria |
|------|---------------|
| Typing | A types → B sees dots; stop/send clears; 6s safety clear |
| Read receipts | B opens → A status advances to read; privacy off → stays delivered |
| Reactions | Long-press 👍; toggle off; GET matches UI; offline reconcile |
| Privacy leak | Account C not in thread receives **no** ephemeral signals |

Full steps: [`docs/E2E_LAUNCH_AND_TESTING.md`](../../docs/E2E_LAUNCH_AND_TESTING.md) §6.4 and `PHASE3_IOS_UI_SPEC.md` Step 5.

## Types — avoid dual paths

| Use in wired chat | Avoid |
|-------------------|-------|
| `ChatDetailMessage` / domain `Message` | Mock `ChatMessage` in same flow |
| `DeliveryStatus` | `MessageStatus` except at bubble boundary |

## Agent should NOT

- Re-write `ConversationSignalService` or envelope codec
- Run live two-client WebSocket E2E headlessly
- Commit partial Xcode project file edits without user review (`.pbxproj` conflicts)

## Related WOs

WO-192 (typing), WO-10/59 (reactions), WO-28 (messaging UI)
