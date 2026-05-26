---
name: echo-ios-agent-vs-xcode
description: >-
  Splits Echo iOS work between headless agents (SPM logic, tests) and Xcode
  (SwiftUI, app target, manual E2E). Use for any ios/Echo task, Phase 3 messaging
  UI, onboarding screens, or when adding Swift files to the project.
---

# Echo iOS — agent vs Xcode

## Layout

| Location | Role |
|----------|------|
| `ios/Echo/Sources/` | SPM library `Echo` (shared code) |
| `ios/Echo/EchoApp.xcodeproj` | App target, schemes, signing |
| `ios/Echo/Package.swift` | SPM targets |
| `ios/Echo/Tests/` | `EchoTests` (may not compile — advisory CI) |
| `ios/Echo/SecurityTests/` | WO-208/211/223/224 — CI hard gate |
| `ios/Echo/Phase3Tests/` | Phase 3 signals — CI hard gate |

App entry `EchoApp.swift` is **excluded** from SPM; lives in Xcode project.

## Agent CAN do (no Xcode UI)

- Models, services, networking (`ConversationSignalService`, `ReactionsAPI`, …)
- ViewModels (`ChatDetailViewModel`, extend `MessagingViewModel`)
- Pure logic (`ConversationSignalLogic.swift`)
- Unit tests in **`EchoSecurityTests`** or **`EchoPhase3Tests`**
- `Package.swift` test target registration
- DI registration in `Core/DI/Container.swift` (`#if os(iOS)`)

Verify:

```bash
cd ios/Echo
DEVELOPER_DIR=/Applications/Xcode.app swift build --target Echo
swift test --filter EchoPhase3Tests
```

CI uses macOS + Xcode 16.x (see `.github/workflows/ios-ci.yml`).

## Xcode REQUIRED

- New Swift files → add to **`EchoApp.xcodeproj`** (SPM alone does not update app target)
- SwiftUI views: `TypingIndicatorView`, `ReactionPickerView`, wiring `ChatView`
- `@Bindable` / navigation integration with coordinators
- Code signing, TestFlight (`docs/PHASE1_LAUNCH.md`)
- Manual two-client tests vs `make dev` (see `PHASE3_IOS_UI_SPEC.md` Step 5)

## Canonical types (avoid duplicates)

| Use | Avoid in wired paths |
|-----|----------------------|
| `Domain/Models/Models.swift` → `Message`, `Reaction` | Mock `ChatMessage` in `MessagingScreens.swift` |
| `Features/Evidence/DeliveryStatus.swift` | `MessageStatus` in `MessageBubbles.swift` |

Refactor `ChatView` to consume `ChatDetailViewModel` + domain types when wiring Phase 3.

## Privacy settings (typing / receipts)

- Global: `EnhancedPrivacySettings.typingIndicators`, `.readReceipts`
- Persona: `PersonaPrivacySettings.sendTypingIndicators`, `.sendReadReceipts`
- Merge via `MessagingPrivacyPreferences.merged` in `ConversationSignalLogic.swift`

## Phase 3 wiring checklist (Xcode)

1. Add new source files to Xcode project
2. Inject `ConversationSignalService` + `ReactionsAPI` from `DIContainer`
3. On chat open: `configure(...)` + `connect(accessToken:)` from Keychain
4. Bind input → `onInputChanged`; message `onAppear` → read receipts
5. Long-press → `toggleReaction`
6. Show `TypingIndicatorView` when `peerIsTyping`
7. Run manual checklist in `docs/PHASE3_IOS_UI_SPEC.md`

## Feature specs

- Phase 3 messaging UI: `docs/PHASE3_IOS_UI_SPEC.md`
- UX / Glacial: `docs/ux-spec.md`
