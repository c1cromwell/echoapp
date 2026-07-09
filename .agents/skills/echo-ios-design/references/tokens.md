# Echo design tokens

Source of truth: `ios/Echo/Sources/DesignSystem/Colors.swift`,
`Typography.swift`, `Spacing.swift`, `Shadows.swift`, and the legacy bridge
`Glacial/GlacialTheme.swift` (maps `Color.Echo.*` / `Font.Echo.*` onto the new system —
still valid at ~51 call sites). Read the file before using a token; the list below is a
map, not a substitute.

Design thesis (from the Colors.swift header): **"Make the security visible. Make the rest
disappear."** Three surface families: **paper** (warm near-white, default), **ink**
(near-black text at opacity steps), **night** (reserved for private/hidden/vault only).

## Color (all `Color.echo…`)

- **Surfaces**: `echoPaper`, `echoPaperDim`, `echoPaperEdge`, `echoSurface`,
  `echoBackground`, `echoCardBackground`.
- **Ink / text**: `echoInk`, `echoInk70`, `echoInk55`, `echoInk40`, `echoInk20`,
  `echoInk10`, `echoInk05`, `echoHair`; convenience `echoPrimaryText`, `echoSecondaryText`.
- **Accent (one)**: `echoSignal` (`0x0E7AB8`), `echoSignalDim`; `echoPrimary` /
  `echoSecondary` are the semantic aliases used in components.
- **Trust (single green ladder)**: `echoTrustGreen`, `echoTrustGreenDim`, and the
  per-tier `echoTrustUnverified/Newcomer/Basic/Trusted/Verified/HighlyTrusted/Premium/Elite`.
  Prefer the helper `trustColor(for: level)` (level is a String, e.g. "verified",
  "trusted") over hand-picking a tier color.
- **Status/semantic**: `echoSuccess`, `echoWarning`, `echoError`, `echoAlert`, `echoInfo`.
- **Night (private/hidden ONLY)**: `echoNight`, `echoNightHi`, `echoNightInk`,
  `echoNightInk70`, `echoNightInk40`, `echoNightHair`.
- Grays `echoGray50…900` and `echoLightBg/Surface`, `echoDarkBg/Surface` exist for
  backward compat — prefer paper/ink tokens in new code.

Never write a raw hex, `Color(red:green:blue:)`, or `Color(hex:)` literal in a view. If a
value is genuinely missing, add a named token to `Colors.swift`.

## Typography (`ios/Echo/Sources/DesignSystem/Typography.swift`)

- Apply with the view modifier: `.typographyStyle(_ style: TypographyStyle, color: Color = .echoPrimaryText)`.
- `TypographyStyle` cases include `bodyLarge`, `body`, `bodySmall`, `caption`, plus title/
  heading styles — read the enum for the full set. Don't use `.font(.system(size:))` in
  Echo views; use a style.
- Monospace (data, codes, secure-comms labels): `Font.echomono(_ size:, weight:)`.

## Spacing (`ios/Echo/Sources/DesignSystem/Spacing.swift`)

- `Spacing` enum cases (use `.rawValue` where a CGFloat is needed):
  `xs, sm, md, lg, xl, xxl, xxxl, huge, massive`.
- Named layout constants: `Spacing.standard`, `.card`, `.horizontal`, `.vertical`, plus
  radii `sm/md/lg/xl/xxl/full`.

## Shadows / effects

- `.glacialShadow()`, `.ghostBorder()`, `LinearGradient.signature`, and
  `Animation.glacial` springs live in `Glacial/GlacialTheme.swift`. Use these instead of
  ad-hoc `.shadow(...)` / `.border(...)` for the Glacial look.

## Reusable components (compose, don't restyle)

`ios/Echo/Sources/Presentation/Components/`: `EchoButton`, `EchoCard`, `EchoTextField`,
`EchoTabBar`, `EchoNavBar`, `TrustBadge`, `TrustScoreView`, `AvatarView`, `EchoToast`,
`MessageBubbles`, `MediaBubbleView`, `ListItems`, `PinnedSection`, `SkeletonViews`,
`OTPInputView`, `NetworkStatusBanner`. Glacial set: `EchoLogo`, `GhostBorderCard`,
`GlacialNavigationBar`, `SignatureGradientButton`, `SecureThreadIndicator`.
