# Screen-catalog review loop

The screen catalog renders every registered screen to a PNG gallery so UI can be reviewed
without a running device. It is the primary visual-verification tool for Echo iOS work.

- Renderer: `ios/Echo/Sources/Features/ScreenCatalog/ScreenCatalogRenderer.swift`
  (`ImageRenderer`, 393×852 @3x, appends `manifest.jsonl`, gated by `SCREEN_CATALOG_GENERATE=1`).
- Registry (the test that lists every screen): `ios/Echo/ScreenCatalog/ScreenCatalogGeneratorTests.swift`.
- Mock data / composite previews for screens that don't rasterize cleanly (List/Form/
  NavigationStack): `ios/Echo/ScreenCatalog/ScreenCatalogFixtures.swift`.
- Runner: `make screen-catalog` (→ `scripts/screen-catalog/generate.sh` + `build-index.sh`).
- Output (git-tracked): `docs/screen_catalog/index.html`, `manifest.jsonl`, PNGs per journey.

## Register a new screen

Add a `render(...)` call in `ScreenCatalogGeneratorTests.testExportFullCatalog()`. Signature:

```swift
try ScreenCatalogRenderer.render(
    MyNewView(/* real or fixture args */),
    journey: "rewards",            // groups the screen in the gallery
    stepId: "07-leaderboard",      // ordered, unique within the journey
    title: "Leaderboard",
    e2eRef: "E2E §X.Y"             // trace back to the spec/flow
)
```

Guidance:
- If the real view relies on `@Environment`, DI, or `List`/`Form`/`NavigationStack` that
  `ImageRenderer` can't rasterize, add a `Catalog<Name>Preview` composite in
  `ScreenCatalogFixtures.swift` and render that instead — follow the existing composites
  (e.g. `CatalogRewardsPreview`, `CatalogMessagesHubPopulatedPreview`). Reuse the standard
  fixture cast (Jordan/Sam/Riley, self "Alex") for consistency.
- Keep `stepId` numbering contiguous within a journey; journeys already present include
  onboarding, auth, tabs, messaging, vip, device, calls, wallet, governance, privacy,
  contacts, profile, utility.

## Run and review

```
make screen-catalog          # boots iPhone 17 sim, runs the one XCTest, rsyncs PNGs out
open docs/screen_catalog/index.html
```

If the simulator build is unavailable in this environment, note that the catalog can't be
regenerated here and the screen must be reviewed on a machine with the iOS platform
installed; still register the screen so it renders on the next successful run.
