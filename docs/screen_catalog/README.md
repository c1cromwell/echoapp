# ECHO Screen Catalog

Static PNG gallery of key customer journeys, generated from **real SwiftUI** via `ImageRenderer` (no manual Simulator screenshots).

## Quick start

```bash
make screen-catalog         # iOS Simulator + export PNGs + index.html
open docs/screen_catalog/index.html
```

Requires **Xcode** and an available **iOS Simulator** (default: `iPhone 17`; override with `SCREEN_CATALOG_SIMULATOR`).

## What gets exported

| Journey | Screens | E2E source |
|---------|---------|------------|
| `onboarding/` | Welcome, username, Face ID options | E2E §6.1 |
| `auth/` | Storage lock, biometric lockout | E2E §6.2 |
| `messaging/` | Empty hub, DM thread fixture, typing, reactions | TestFlight A1–A7 |
| `contacts/` | Invite accept, username search | TestFlight A2, B1 |
| `privacy/` | SOCKS proxy, PQ handshake, hidden folder | Privacy Hub |

Fixtures use **mock data** — not live backend state. The DM thread is a catalog composite (`CatalogChatThreadPreview`), not a full `ChatView` wiring.

## How it works

1. `EchoUnitTests/ScreenCatalogGeneratorTests` renders each view at **393×852 pt @3x**
2. PNGs land in the Simulator app `Documents/screen_catalog/` (sandbox)
3. `scripts/screen-catalog/generate.sh` copies them to `docs/screen_catalog/` via `simctl`
4. `manifest.jsonl` records metadata; `build-index.sh` writes tabbed `index.html`

Avoid wrapping views in `NavigationStack` for export — `ImageRenderer` cannot flatten UIKit navigation hosts.

## Environment

| Variable | Default |
|----------|---------|
| `SCREEN_CATALOG_ROOT` | `docs/screen_catalog` (host copy destination) |
| `SCREEN_CATALOG_SIMULATOR` | `iPhone 17` |
| `SCREEN_CATALOG_BUNDLE_ID` | `com.echo.app` |

## Adding a screen

1. Add a `ScreenCatalogRenderer.render(...)` call in `ios/Echo/ScreenCatalog/ScreenCatalogGeneratorTests.swift`
2. Run `make screen-catalog`
3. Commit PNGs + `manifest.jsonl` + `index.html` if you want them in git

## Limits

- Runs on **iOS Simulator** (headless via `xcodebuild` — no manual tapping)
- Does not replace two-client E2E or TestFlight sign-off
- Views that trigger Face ID / network on appear are avoided or use static fixtures
