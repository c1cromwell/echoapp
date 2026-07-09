---
name: echo-ios-design
description: Build and review Echo iOS SwiftUI screens against Echo's own design system, guardrails, and screen-catalog review loop. Use when creating, editing, or reviewing any post-auth Echo iOS UI (SwiftUI views, screens, components) in ios/Echo — especially Rewards/wallet, messaging, contacts, profile, governance, and VIP surfaces.
license: MIT
metadata:
  author: Echo
  version: "1.0"
  scope: ios/Echo
---

# Echo iOS Design

You are building or reviewing SwiftUI for **Echo**, an iOS-first privacy messenger.
Echo has a mature, opinionated design system; your job is to make new UI look like it
was always part of it, obey the project's frozen-surface guardrails, and verify visually
through the screen catalog. Generic SwiftUI correctness/performance/accessibility is
**delegated** — this skill is the Echo-specific layer on top.

## Order of work

1. **Read the token source of truth first**: `ios/Echo/Sources/DesignSystem/Colors.swift`
   (plus `Typography.swift`, `Spacing.swift`, `Shadows.swift`, `Glacial/GlacialTheme.swift`).
   Never guess a color/spacing value — every one has a named token. See
   `references/tokens.md`.
2. **Check the guardrails** in `references/guardrails.md` before touching a screen:
   onboarding/login are **frozen**; `night` surfaces are for private/hidden vault UI only;
   the React prototype's "Icy" hex must be remapped to production tokens.
3. **Build with tokens, compose with existing components** from
   `ios/Echo/Sources/Presentation/Components/` (`EchoButton`, `EchoCard`, `EchoTextField`,
   `TrustBadge`, `AvatarView`, `EchoTabBar`, `EchoNavBar`, `ListItems`, `MessageBubbles`,
   …) and the Glacial component set. Prefer reusing a component over re-styling a raw view.
4. **Wire the file into the build** (`scripts/xcode-add-sources.py`) and **register the
   screen** for visual review in `ScreenCatalogGeneratorTests.swift`, then run
   `make screen-catalog`. See `references/screen-catalog.md`.
5. **Delegate generic SwiftUI quality**: for `@Observable` data flow, view identity/
   invalidation, Liquid Glass, deprecation migration, or Instruments analysis, use the
   **`swiftui-expert`** skill; for a HIG/accessibility/modern-API code review, use
   **`swiftui-pro`**. This skill does not re-derive those — it owns the Echo layer.

## Hard rules (do not violate)

- **No raw hex / no `Color(red:…)` / no magic paddings** in view code. Use `Color.echo*`,
  `Font.echo*` / `.typographyStyle(_:)`, and `Spacing.*`. If a needed token is missing,
  add it to the design-system file, don't inline a literal.
- **Do not redesign or re-map onboarding/login** (`FirstRunCoordinator`,
  `GlacialLoginScreen`, and related auth views). They are the correct, frozen design.
- **`night` palette only for private/hidden/vault** surfaces. Everything else uses
  `paper`/`ink`.
- **One accent**: `echoSignal` (blue `0x0E7AB8`). Trust is a single green ladder
  (`echoTrustGreen` at tier opacities / `trustColor(for:)`), not a rainbow.
- **Respect the launch guardrails already in code**: for Rewards/token UI, never add a
  withdraw/transfer/redeem control while `custody_mode == "interim"`; drive any "no cash
  value" copy from the `GET /v1/gamification/status` flags, and keep the UI free of
  "worth / value / equity / ownership / perk" language (see the tokenomics launch posture).

## Verification

- Real gate is an **Xcode build** of the `EchoMessaging` scheme + `make screen-catalog`.
- When the headless Xcode simulator build is unavailable, fall back to
  `xcrun swiftc -parse -target arm64-apple-ios17.0-simulator -D DEBUG -D ECHO_PRODUCT_MESSAGING <file>.swift`
  for syntax, and hand the diff to `swiftui-pro` for an API/accessibility review. State
  clearly that a full Xcode build is still required.

References: `references/tokens.md`, `references/guardrails.md`, `references/screen-catalog.md`.
