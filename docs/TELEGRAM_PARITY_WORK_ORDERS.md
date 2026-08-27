# Telegram Parity — Software Factory work orders

**Epic:** [WO-335](https://factory.8090.ai) — Telegram Parity Waves T0–T3 (post-Sept)  
**Parent of epic:** WO-321 (Sept 1 go-live) — post-launch only; do not block soft launch  
**Sibling:** WO-330 (Signal parity) — shared S1 history/backup (WO-333)  
**Assignee:** Chad Cromwell  
**FEATURE:** Telegram-class Messaging UX (`3ddcb0c4-d216-47c6-84f2-3a45dd35dd73`)  
**Canonical matrix:** [`TELEGRAM_ECHO_PARITY.md`](TELEGRAM_ECHO_PARITY.md)

| Wave | WO | Title | Status | Scope |
|------|-----|-------|--------|-------|
| **Epic** | **WO-335** | Telegram Parity Epic — Waves T0–T3 | `in_progress` | Program umbrella |
| **T0** | **WO-336** | Durable channels + admin | `ready` | `/v3/broadcasts/*` soak + ChannelAdmin Xcode E2E |
| **T1** | **WO-337** | Saved Messages, drafts, forward, search, quotes | `in_progress` | Logic landed; two-device E2E + hub search wiring |
| **T2** | **WO-338** | Folders, scheduled/silent HTTP, bots | `in_progress` | `/v3/chat-folders`, `/v3/messages/schedule` mounted (in-memory); iOS still local-first |
| **T3** | **WO-339** | Stickers, translate, import | `backlog` | Later; skip Stories/Stars/Nearby |

## Implementation landed (validated 2026-08-27)

See [`TELEGRAM_ECHO_PARITY.md`](TELEGRAM_ECHO_PARITY.md) §2. T1 sources are in `EchoApp.xcodeproj`. Remaining productization is two-device E2E, folder sync soak, scheduled Postgres, and the Xcode items below.

## Xcode-only (do not re-implement in SPM)

Scheme **EchoApp** / product **EchoMessaging**. New files: `python3 scripts/xcode-add-sources.py`. Frozen onboarding: do not restyle login.

| # | Feature | Why Xcode | Files / notes |
|---|---------|-----------|----------------|
| 1 | **Notification Service Extension** | New app-extension **target** (not in pbxproj) | `ios/Echo/NotificationService/` — decrypt preview; App Group + Keychain |
| 2 | **Push / APNs** | Capabilities + entitlements | `PushRegistrationService`; `API_URL` LAN not localhost on device |
| 3 | **CallKit + live WebRTC** | Signing, microphone/camera usage strings, release vs stub engine | `CallKitCoordinator`, `WebRTCLiveCallEngine` |
| 4 | **ChatView live session** | Confirm production navigation injects real DIDs | `MessagingScreens.ChatView` already binds `ChatDetailViewModel`; parent must not pass empty `peerDID` |
| 5 | **Typing / receipts / reactions** | Two-device manual | `TypingIndicatorView`, `ReactionPickerView`, `ReactionChipsView` (already in target) |
| 6 | **Saved Messages pin** | Hub vs empty-state both must open it | `SavedMessagesStore`, `MessagesHubView` |
| 7 | **Drafts** | Leave/reopen chat | `ComposerDraftStore` already wired in `ChatView` |
| 8 | **Multi-forward** | Selection mode sheet | `ForwardMessageSheet` |
| 9 | **Global search** | Replace dead hub preview callback | `MessagesHubView` preview still has `onOpenMessageSearch: { _ in }`; production tab uses `MessagesTabView` |
| 10 | **Folders UI** | Settings / hub chrome | `ChatFoldersView`, `ChatFolderFilterView` |
| 11 | **Schedule sheet** | Optional: POST `/v3/messages/schedule` | Today local `ScheduledMessageStore` only |
| 12 | **Safety numbers** | Chat settings sheet | `SafetyNumberCompareView` already presented from `ChatView` |
| 13 | **Message requests** | First-contact banner | `MessageRequestStore` in `ChatView` |
| 14 | **View-once** | Media send toggle + burn on open | `ViewOnceMedia` + composer toggle |
| 15 | **GIF picker** | Composer accessory | `GifSearchService` |
| 16 | **Wallpaper** | Chat settings | `ChatWallpaperStore` |
| 17 | **Group typing** | Group thread | `GroupChatView` already shows `TypingIndicatorView` |
| 18 | **Channel admin** | Channel owner flow | `ChannelAdminView` |
| 19 | **Backup / restore polish** | Phrase restore E2E | `BackupView`, `RecoveryService` (Signal S1) |
| 20 | **ZKCommitmentProof.swift** | Added to EchoApp target 2026-08-27 | `Sources/Core/Security/ZKCommitmentProof.swift` (WO-236) |

Headless agents land SPM logic + tests. SwiftUI/target membership for **new** files still uses `scripts/xcode-add-sources.py`.
