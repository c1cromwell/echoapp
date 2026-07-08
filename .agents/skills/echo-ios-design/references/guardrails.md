# Echo iOS guardrails

Authority: `AGENTS.md` and `docs/ECHO_IOS_UI_IMPLEMENTATION_SPEC.md` §0. Violating these
produces "correct SwiftUI that's wrong for Echo."

## Frozen surfaces — do not redesign

- **Onboarding & login are frozen**: `FirstRunCoordinator`, `GlacialLoginScreen`, and
  related auth views are the correct design. Do **not** restyle them or map them from the
  React prototype's onboarding. (`AGENTS.md` line ~76.)
- Only **post-auth** UI may follow the React design prototype
  (`docs/Echo Design System Setup_latest/`), per `ECHO_IOS_UI_IMPLEMENTATION_SPEC.md` §0.

## Prototype token remap

The Figma/React prototype ("Icy Minimal") uses different hex — e.g. `#0EA5E9` — that must
be **remapped to production Echo tokens** (`echoSignal` = `0x0E7AB8`), not copied. See
`docs/ux-spec.md` (~line 289). When translating a prototype screen, convert every literal
to its nearest Echo token; never paste prototype CSS hex into Swift.

## Palette discipline

- `night` surfaces are for **private/hidden/vault** UI only. Default is paper/ink.
- One accent (`echoSignal`); trust is one green ladder (`trustColor(for:)`).

## Rewards / token UI (launch posture)

Echo launches the ECHO token **value-free** (gamified). When building any Rewards/wallet/
leaderboard UI:

- **No** withdraw / transfer / redeem / stake-for-yield control while custody is interim.
- Drive the "no cash value" disclaimer from `GET /v1/gamification/status`
  (`redeemable`/`transferable` flags + `custody_mode`); render it persistently in the hub.
- Keep copy free of **worth / value / equity / ownership / perk** language.
- See the project memory note "Tokenomics launch posture" for the three guardrails.

## pbxproj wiring

New Swift files must be registered in the classic project (no XcodeGen in use):
`python3 scripts/xcode-add-sources.py ios/Echo/EchoApp.xcodeproj/project.pbxproj <GROUPID>:File.swift`
(idempotent; writes a `.bak`). Find the target group id by grepping a sibling file in the
same folder. The Sources SwiftPM `Package.swift` globs automatically, but the app builds
from the `.xcodeproj`.

## Build-environment caveat

The headless Xcode simulator build may be unavailable in some environments (missing iOS
platform). Fallbacks:
- Syntax: `xcrun swiftc -parse -target arm64-apple-ios17.0-simulator -D DEBUG -D ECHO_PRODUCT_MESSAGING <file>.swift`
- Quality review: hand the diff to the `swiftui-pro` skill.
- Always state that a real Xcode build of `EchoMessaging` is still the gate.

## Figma MCP (optional token/component sync)

The Figma MCP is connected (`mcp__claude_ai_Figma__*`). Useful for keeping tokens in sync:
`get_variable_defs` (pull design variables), `get_screenshot`/`get_design_context`
(reference a frame), and Code Connect (`get_code_connect_map` / `add_code_connect_map`) to
bind Figma components to the Echo Swift components. Not required for a normal screen build;
reach for it when reconciling tokens or onboarding a new component from design.
