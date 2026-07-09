---
name: echo-ios-design
description: >-
  Build and review Echo iOS SwiftUI screens against Echo's design system,
  frozen-surface guardrails, and the screen-catalog review loop. Use for any
  post-auth ios/Echo UI work — Rewards/wallet, messaging, contacts, profile,
  governance, VIP. Delegates generic SwiftUI quality to swiftui-expert / swiftui-pro.
---

# Echo iOS Design

Make new SwiftUI look like it was always part of Echo's system, obey the frozen-surface
guardrails, and verify visually via the screen catalog. Full reference bundle lives at
`.agents/skills/echo-ios-design/references/` (`tokens.md`, `guardrails.md`,
`screen-catalog.md`) — read those for depth; this file is the working summary.

## Rules

1. **Tokens, never literals.** Source of truth `ios/Echo/Sources/DesignSystem/Colors.swift`
   (+ `Typography.swift`, `Spacing.swift`, `Glacial/GlacialTheme.swift`). Use `Color.echo*`,
   `.typographyStyle(_:)` / `Font.echomono(_:)`, `Spacing.*`, `.glacialShadow()`,
   `.ghostBorder()`. No raw hex / `Color(red:…)` / magic paddings. Missing value → add a
   named token, don't inline.
2. **One accent** `echoSignal` (`0x0E7AB8`); trust is one green ladder
   (`trustColor(for: level)`); `night` palette only for private/hidden/vault.
3. **Frozen**: do NOT redesign onboarding/login (`FirstRunCoordinator`, `GlacialLoginScreen`).
   Only post-auth UI follows the React prototype; remap its "Icy" hex (`#0EA5E9`) → `echoSignal`.
4. **Compose existing components** from `ios/Echo/Sources/Presentation/Components/`
   (`EchoButton`, `EchoCard`, `TrustBadge`, `AvatarView`, `ListItems`, …) before restyling.
5. **Rewards/token UI launch posture**: no withdraw/transfer/redeem while custody is interim;
   drive the "no cash value" banner from `GET /v1/gamification/status`; no worth/value/
   equity/ownership/perk copy.
6. **Wire + review**: register the file with `scripts/xcode-add-sources.py`, add a
   `render(...)` entry in `ScreenCatalog/ScreenCatalogGeneratorTests.swift`, run
   `make screen-catalog`. Real gate = Xcode build of `EchoMessaging`; if the sim build is
   unavailable, `xcrun swiftc -parse -target arm64-apple-ios17.0-simulator -D DEBUG -D ECHO_PRODUCT_MESSAGING <file>.swift`.
7. **Delegate** `@Observable`/perf/Liquid Glass/Instruments to `swiftui-expert`; HIG/
   accessibility/API review to `swiftui-pro`. This skill owns only the Echo layer.
