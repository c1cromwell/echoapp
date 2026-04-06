# Echo UX Specifications — v2 (Fuse-Inspired Update)
## Feature 1: Onboarding & Registration

### Overview
The onboarding flow is designed to be simple for mainstream users (phone-based) while offering advanced options for crypto-native users (verifiable credentials). Inspired by Fuse Wallet's progressive security model, the flow lets users in fast with basic security, then nudges them to complete setup over time. Security feels like a feature, not a chore.

### Key Changes from v1
- **NEW** Intro Carousel (3 animated value-proposition slides)
- **UPDATED** Screen 4: Passkey — rewritten copy using plain-language analogies
- **NEW** Screen 4B: Recovery Method setup
- **UPDATED** Screen 6: Combined Trust Score + Security Dashboard
- **NEW** Home nudge card for skipped setup steps

---

## Screen 0: Intro Carousel (NEW)

### Layout
```
┌─────────────────────────────────┐
│                                 │
│                                 │
│         [Animated Icon]         │
│                                 │
│      "Private by Default"       │
│                                 │
│   Your messages are encrypted   │
│   end-to-end. Not even Echo     │
│   can read them.                │
│                                 │
│          ● ○ ○                  │
│                                 │
│                                 │
│        [Get Started →]          │
│                                 │
└─────────────────────────────────┘
```

### Slide Content
| Slide | Icon | Title | Body |
|-------|------|-------|------|
| 1 | Animated lock/shield (pulse) | Private by Default | Your messages are encrypted end-to-end. Not even Echo can read them. |
| 2 | Animated trust ring (filling) | Trust, Not Just Talk | Build real reputation through verified actions. See who you can trust. |
| 3 | Animated key/fingerprint | Your Identity, Your Control | Own your digital identity. No big tech. No data harvesting. Just you. |

### Specifications
| Element | Spec |
|---------|------|
| Frame | 393 x 852px |
| Background | Bg/Primary |
| Icon container | 120x120px, centered, Y: 220 |
| Icon | 64px, animated, Primary/500 |
| Title | Heading/H1, Text/Primary, centered |
| Body | Body/Default, Text/Secondary, centered, max-width 300px |
| Page dots | 8px diameter, 12px gap, centered |
| Active dot | Primary/500, filled |
| Inactive dots | Border/Default, filled |
| "Get Started" button | Primary button, only on slide 3 |
| Skip link | "Skip" top-right, Text/Link, Body/Small |

### Interactions
- Swipe left/right to navigate slides
- Dots update with active state
- Auto-advance disabled (user-controlled)
- Skip link on all slides → jumps to Welcome
- Get Started on slide 3 → Welcome screen
- Respect reduced motion: disable animations

### Animations (300ms ease-out)
| Slide | Animation |
|-------|-----------|
| 1 | Shield scales up from 0.8 → 1.0, lock icon fades in |
| 2 | Trust ring fills from 0% → 65% with number counting up |
| 3 | Key rotates 0° → 15° → 0° gentle wobble, fingerprint pulses |

---

## Screen 1: Welcome / Splash (Unchanged)

### Layout
```
┌─────────────────────────────────┐
│                                 │
│         [Echo Logo]             │
│     "Messaging with Trust"      │
│                                 │
│   ┌─────────────────────────┐   │
│   │   Continue with Phone   │   │ ← Primary CTA
│   └─────────────────────────┘   │
│   ┌─────────────────────────┐   │
│   │  Use Verifiable ID      │   │ ← Secondary
│   └─────────────────────────┘   │
│      Already have account?      │
│          [Sign In]              │
│  ─────────────────────────────  │
│   🔒 Your identity stays yours  │
│      No data sold. Ever.        │
└─────────────────────────────────┘
```

### Specifications
| Element | Spec |
|---------|------|
| Logo | 120x120px, centered, gradient Primary/600 → Primary/500, Logo Shadow |
| Tagline | Heading/H2, Text/Primary |
| Primary Button | "Continue with Phone", 56px, Primary/500 |
| Secondary Button | "Use Verifiable ID", 56px, outlined |
| Trust badge | Body/Small, Text/Tertiary, lock icon 20px |

---

## Screen 2: Phone Number Entry (Unchanged)

### Layout
```
┌─────────────────────────────────┐
│ ←                               │
│     Enter your phone number     │
│  We'll send a verification code │
│  ┌──────┬──────────────────┐    │
│  │ +1 ▼ │ (555) 123-4567   │    │
│  └──────┴──────────────────┘    │
│   ┌─────────────────────────┐   │
│   │      Send Code          │   │
│   └─────────────────────────┘   │
│   By continuing, you agree to   │
│   our Terms and Privacy Policy  │
└─────────────────────────────────┘
```

### Specifications
| Element | Spec |
|---------|------|
| Back arrow | 24px, 44x44 tap target |
| Title | Heading/H1, Text/Primary |
| Subtitle | Body/Default, Text/Secondary |
| Country picker | 90px, Bg/Secondary, Border/Default |
| Phone input | Fill remaining, 56px |
| Button | Disabled until valid: Bg/Tertiary, Text/Tertiary |
| Legal text | Caption, Text/Tertiary, links in Text/Link |

---

## Screen 3: OTP Verification (Unchanged)

### Layout
```
┌─────────────────────────────────┐
│ ←                               │
│      Enter verification code    │
│   Sent to +1 (555) 123-4567    │
│     ┌───┬───┬───┬───┬───┬───┐  │
│     │ 1 │ 2 │ 3 │ _ │   │   │  │
│     └───┴───┴───┴───┴───┴───┘  │
│        Resend code (0:45)       │
│   ┌─────────────────────────┐   │
│   │       Verify            │   │
│   └─────────────────────────┘   │
│      Wrong number? [Change]     │
└─────────────────────────────────┘
```

### Specifications
| Element | Spec |
|---------|------|
| OTP boxes | 48x56px each, 8px gap, 328px total |
| Active box | Border/Focus, 2px |
| Filled box | OTP/FilledBorder, Bg/Secondary |
| Timer | Body/Small, Text/Secondary |
| Verify button | Disabled until 6 digits |

---

## Screen 4: Create Passkey (UPDATED — Plain Language)

### Layout
```
┌─────────────────────────────────┐
│                                 │
│         [Face ID Icon]          │
│                                 │
│     Secure your account         │
│                                 │
│   Sign in with your face —      │
│   no passwords, no codes.       │
│   Your biometric data never     │
│   leaves your device.           │
│                                 │
│   ┌─────────────────────────┐   │
│   │    Enable Face ID       │   │
│   └─────────────────────────┘   │
│                                 │
│        [Skip for now]           │
│                                 │
│  ─────────────────────────────  │
│                                 │
│   ✓ No passwords to remember    │
│   ✓ Your face stays on-device   │
│   ✓ Move to a new phone easily  │
│                                 │
└─────────────────────────────────┘
```

### Specifications
| Element | Spec |
|---------|------|
| Icon | 100x100px, gradient Primary/600 → Primary/400, animated pulse |
| Title | Heading/H1, Text/Primary |
| Body | Body/Default, Text/Secondary, center, max-width 300px |
| Benefits list | Body/Small, Text/Secondary, ✓ icon 16px Success/500 |

### Updated Benefit Copy (from v1 → v2)
| v1 (Technical) | v2 (Plain Language) |
|---|---|
| Creates your secure DID | No passwords to remember |
| No passwords to remember | Your face stays on-device |
| Works across devices | Move to a new phone easily |

---

## Screen 4B: Recovery Method (NEW)

### Layout
```
┌─────────────────────────────────┐
│ ←                               │
│                                 │
│       [Shield + Key Icon]       │
│                                 │
│     Protect your account        │
│                                 │
│   Add a backup so you never     │
│   lose access — even if you     │
│   lose your phone.              │
│                                 │
│   ┌─────────────────────────┐   │
│   │ 📧  Recovery Email      │   │ ← Recommended
│   │     Easiest to set up   │   │
│   └─────────────────────────┘   │
│                                 │
│   ┌─────────────────────────┐   │
│   │ 🔑  Recovery Wallet     │   │
│   │     Most secure         │   │
│   └─────────────────────────┘   │
│                                 │
│   ┌─────────────────────────┐   │
│   │ 👤  Trusted Contact     │   │
│   │     Social recovery     │   │
│   └─────────────────────────┘   │
│                                 │
│        [Skip for now]           │
│                                 │
│  Your account is secured with   │
│  Face ID. A recovery method     │
│  adds a safety net.             │
│                                 │
└─────────────────────────────────┘
```

### Specifications
| Element | Spec |
|---------|------|
| Frame | 393 x 852px, Bg/Primary |
| Nav Header | Back arrow, no right action |
| Icon | 80x80px, shield with key overlay, Primary/500 stroke |
| Title | Heading/H1, Text/Primary, "Protect your account" |
| Body | Body/Default, Text/Secondary, center, max-width 300px |
| Option cards | W: Fill, H: 72px, Radius: 16px, Bg: Surface/Card |
| Card border | 2px solid Border/Default |
| Card selected | 2px solid Primary/500, bg Primary/100 |
| Card icon | 32px emoji or icon, left-aligned |
| Card title | Label/Default (14px Medium), Text/Primary |
| Card subtitle | Caption (12px), Text/Tertiary |
| Recommended badge | "Recommended" pill on first card, 10px, Primary/500 bg, white text, radius 6 |
| Skip link | Ghost button, Text/Secondary, "Skip for now" |
| Footer note | Caption, Text/Tertiary, centered |
| Card gap | 12px between cards |
| Padding | 24px horizontal |

### Option Details
| Option | Icon | Title | Subtitle | Action |
|--------|------|-------|----------|--------|
| Email | 📧 | Recovery Email | Easiest to set up | → Email entry screen |
| Wallet | 🔑 | Recovery Wallet | Most secure | → WalletConnect / QR scan |
| Contact | 👤 | Trusted Contact | Social recovery | → Contact picker |

### Interactions
- Tap card → highlight selected → Continue button appears at bottom
- Skip → proceeds to Profile Setup with reminder
- Only one method required (can add more later in Settings)
- Email: sends verification code, non-custodial (Turnkey-style)
- Wallet: WalletConnect QR or auto-detect installed wallets
- Contact: Select from contacts, they approve recovery requests

---

## Screen 5: Profile Setup (Unchanged)

### Layout
```
┌─────────────────────────────────┐
│ ←                         Skip  │
│        [Avatar Placeholder]     │
│           + Add photo           │
│   Display Name                  │
│   ┌─────────────────────────┐   │
│   │ Alex                    │   │
│   └─────────────────────────┘   │
│   Username (optional)           │
│   ┌─────────────────────────┐   │
│   │ @alex_echo              │   │
│   └─────────────────────────┘   │
│   ✓ Available                   │
│   Bio (optional)                │
│   ┌─────────────────────────┐   │
│   │                         │   │
│   └─────────────────────────┘   │
│                         0/150   │
│   ┌─────────────────────────┐   │
│   │     Continue            │   │
│   └─────────────────────────┘   │
└─────────────────────────────────┘
```

### Specifications
| Element | Spec |
|---------|------|
| Avatar | 100x100px, circular, dashed border |
| Input labels | Label/Default, Text/Secondary |
| Inputs | 56px height, Body/Default, Bg/Secondary |
| Username check | Real-time, debounced 300ms, ✓ Success/500 |
| Bio | Multi-line 80px, 150 char limit |
| Character count | Caption, Text/Tertiary, right-aligned |

---

## Screen 6: Trust + Security Dashboard (UPDATED)

### Layout
```
┌─────────────────────────────────┐
│                                 │
│      🛡️ Your Trust Score        │
│                                 │
│          ┌─────────┐            │
│          │   15    │            │
│          │  /100   │            │
│          └─────────┘            │
│            Newcomer             │
│                                 │
│   ─── Your Security Setup ───   │
│                                 │
│   ✅ Phone verified             │
│   ✅ Passkey enabled            │
│   ⬜ Recovery method  [Add →]   │
│   ⬜ Identity verified [+50 →]  │
│                                 │
│   ─── Unlock with Trust ─────   │
│                                 │
│   ○ Create groups up to 30      │
│   ○ Earn 2× message rewards     │
│   ○ Verified badge              │
│   ○ Bank integrations           │
│                                 │
│   ┌─────────────────────────┐   │
│   │   Start messaging →     │   │
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### Specifications
| Element | Spec |
|---------|------|
| Frame | 393 x 852px, Bg/Primary |
| Shield icon | 32px, Text/Link |
| Title | Heading/H1, Text/Primary |
| Trust ring | 140x140px, 10px stroke |
| Score | 48px Bold, Primary/500 |
| Level | Body/Default Semibold, Text/Secondary |
| Section divider | 1px, Border/Default, with label centered |
| Section label | Label/Small, Text/Tertiary, uppercase tracking 1px |

### Security Setup Section (NEW)
| Status | Icon | Text | Action |
|--------|------|------|--------|
| Complete | ✅ 20px, Success/500 | "Phone verified" Body/Small, Text/Primary | None |
| Complete | ✅ 20px, Success/500 | "Passkey enabled" Body/Small, Text/Primary | None |
| Incomplete | ⬜ 20px, Border/Default | "Recovery method" Body/Small, Text/Secondary | "Add →" Text/Link |
| Incomplete | ⬜ 20px, Border/Default | "Identity verified" Body/Small, Text/Secondary | "+50 →" Text/Link |

| Element | Spec |
|---------|------|
| Row height | 44px (meets tap target) |
| Row gap | 4px |
| Check icon | 20px, Success/500 (complete) or Border/Default circle (incomplete) |
| Row text | Body/Small, Text/Primary (done) or Text/Secondary (pending) |
| Action link | Body/Small, Text/Link, right-aligned |
| Row padding | 0 24px horizontal |

### Unlock Benefits Section (Updated Copy)
| v1 (Vague) | v2 (Tangible) |
|---|---|
| Larger groups (30+) | Create groups up to 30 people |
| Higher reward multipliers | Earn 2× message rewards |
| Verified badge | Verified badge on your profile |
| Bank integrations | Connect bank for instant transfers |

### CTA
| Element | Spec |
|---------|------|
| Primary button | "Start messaging →", Primary/500, full width, 56px |
| Position | Bottom, 24px padding, 34px safe area |

---

## Updated Flow Diagram

```
Intro Carousel (3 slides) ──→ Skip ──┐
    │ "Get Started"                    │
    ▼                                  ▼
Screen 1: Welcome ────────────────────────
    │ "Continue with Phone"   │ "Use Verifiable ID"
    ▼                         ▼
Screen 2: Phone Entry     Screen 2B: Wallet Connect
    │                         │
    ▼                         ▼
Screen 3: OTP             Screen 3B: Select Credential
    │                         │
    ├─────────────────────────┘
    ▼
Screen 4: Passkey (plain language)
    │ Enable    │ Skip ──→ [flagged incomplete]
    ▼           │
Screen 4B: Recovery Method (NEW)
    │ Add       │ Skip ──→ [flagged incomplete]
    ▼           │
Screen 5: Profile Setup
    │
    ▼
Screen 6: Trust + Security Dashboard
    │ "Start messaging"
    ▼
[Home] ──→ "Complete Setup" nudge card if steps skipped
```

---

## Design Tokens (Unchanged)

### Colors — use variable modes
See `figma-design-spec.md` for full Light/Dark token table.

### Typography
```
H1: 28px / Bold / -0.5 tracking
H2: 24px / Semi-bold / -0.3 tracking
H3: 20px / Semi-bold / -0.2 tracking
Body: 16px / Regular / 24px line-height
Body Small: 14px / Regular / 20px line-height
Caption: 12px / Regular / 16px line-height
Label: 14px / Medium / 18px line-height
```

### Spacing
```
xs: 4px    sm: 8px    md: 16px
lg: 24px   xl: 32px   2xl: 48px
```

### Components
```
Button Height: 56px    Input Height: 56px
Border Radius: 12px    Card Radius: 16px
Icon Size: 24px        Option Card Height: 72px
```

---

## Animations

### Transitions
- Screen transitions: 300ms ease-out slide
- Button press: Scale 0.98, 100ms
- Input focus: Border color 200ms
- Carousel swipe: 300ms ease-out

### Intro Carousel Animations
- Slide 1: Shield scale 0.8→1.0 (400ms spring)
- Slide 2: Ring fill 0→65% (800ms ease-out), counter 0→65
- Slide 3: Key wobble rotation (600ms)

### Loading States
- OTP verification: Spinner in button
- DID creation: Progress indicator with "Securing your identity..."
- Username check: Subtle spinner next to input
- Recovery setup: "Connecting..." state in selected card

---

## Accessibility

- All tap targets: minimum 44x44px
- Color contrast: WCAG AA minimum (verified in dark mode)
- Screen reader labels on all interactive elements
- Reduce motion: Respect system preference, skip carousel animations
- Dynamic type support: Scale with system font size
- Recovery options: All methods accessible via VoiceOver
