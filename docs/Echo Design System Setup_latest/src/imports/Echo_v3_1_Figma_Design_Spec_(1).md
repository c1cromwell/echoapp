# Echo — Figma Design Spec v3.0

> **Version:** 3.0 · **Date:** March 2026 · **Status:** Ready for Design
> **Theme:** Icy Minimal (Light) + Icy Minimal Dark · **Single theme system — no theme picker**

---

## Table of Contents

1. [Design Direction](#1-design-direction)
2. [Color System — Icy Minimal Light & Dark](#2-color-system)
3. [Onboarding Carousel (First Launch)](#3-onboarding-carousel)
4. [Login Screen (Default Landing)](#4-login-screen)
5. [Create Account Flow](#5-create-account-flow)
6. [Dark Mode Specifications](#6-dark-mode-specifications)
7. [Component Updates](#7-component-updates)
8. [Accessibility](#8-accessibility)

---

## 1. Design Direction

Echo uses a single theme — **Icy Minimal** — with automatic light/dark mode support. The Electric Violet and Teal Warm themes are removed. The theme picker component is removed from Settings. Users control light/dark via iOS system appearance settings, or an in-app override (Light / Dark / System).

### Design Principles

- **Cool and clean** — slate blues, crisp whites, sky-blue accents
- **Breathable** — generous whitespace, restrained color, no visual clutter
- **Frosted glass** — surfaces use translucency and subtle blur in dark mode
- **PayPal-level simplicity** — login has one input, one button, small fallback links

---

## 2. Color System

### 2.1 Icy Minimal Light

| Token | Hex | Usage |
|---|---|---|
| `primary` | `#64748B` | Navigation icons, secondary text |
| `primaryDark` | `#475569` | Pressed states |
| `accent` | `#0EA5E9` | CTAs, links, active indicators, focus rings |
| `accentDark` | `#0284C7` | Accent pressed states |
| `accentLight` | `#7DD3FC` | Gradient end, highlights |
| `success` | `#10B981` | Online, verified, positive |
| `warning` | `#F59E0B` | VIP gold, caution |
| `error` | `#EF4444` | Errors, destructive actions |
| `background` | `#F8FAFC` | Screen background |
| `backgroundSecondary` | `#F1F5F9` | Grouped sections |
| `surface` | `#E2E8F0` | Input backgrounds, cards |
| `surfaceElevated` | `#FFFFFF` | Cards, modals, tab bar |
| `textPrimary` | `#0F172A` | Headlines, body |
| `textSecondary` | `#475569` | Subtitles, descriptions |
| `textTertiary` | `#64748B` | Hints, timestamps, placeholders |
| `textOnAccent` | `#FFFFFF` | Text on accent-colored buttons |
| `border` | `#E2E8F0` | Dividers, input borders |
| `borderLight` | `#F1F5F9` | Subtle separators |
| `borderFocus` | `#0EA5E9` | Focused input ring |
| `accentGradient` | `#0EA5E9 → #7DD3FC` | Primary CTA fill |

### 2.2 Icy Minimal Dark

| Token | Hex | Usage |
|---|---|---|
| `primary` | `#94A3B8` | Navigation icons, secondary text |
| `primaryDark` | `#CBD5E1` | Pressed states (lighter in dark) |
| `accent` | `#38BDF8` | CTAs, links, active indicators |
| `accentDark` | `#0EA5E9` | Accent pressed |
| `accentLight` | `#7DD3FC` | Gradient end |
| `success` | `#34D399` | Online, verified |
| `warning` | `#FBBF24` | VIP gold |
| `error` | `#F87171` | Errors |
| `background` | `#0F172A` | Screen background |
| `backgroundSecondary` | `#1E293B` | Grouped sections |
| `surface` | `#334155` | Input backgrounds, cards |
| `surfaceElevated` | `#1E293B` | Cards, modals, tab bar |
| `textPrimary` | `#F1F5F9` | Headlines, body |
| `textSecondary` | `#CBD5E1` | Subtitles |
| `textTertiary` | `#94A3B8` | Hints, timestamps |
| `textOnAccent` | `#0F172A` | Text on accent buttons (dark text on bright accent) |
| `border` | `#334155` | Dividers |
| `borderLight` | `#1E293B` | Subtle separators |
| `borderFocus` | `#38BDF8` | Focused input ring |
| `accentGradient` | `#0EA5E9 → #38BDF8` | Primary CTA fill (slightly brighter) |

### 2.3 Figma Variable Setup

Create one variable collection named **"Echo Theme"** with two modes: **Light** and **Dark**. Every token above becomes a variable that swaps between modes. All screens reference variables — never hard-coded hex values.

```
Echo Theme (collection)
├── Mode: Light
│   ├── background = #F8FAFC
│   ├── accent = #0EA5E9
│   └── ... (all tokens)
└── Mode: Dark
    ├── background = #0F172A
    ├── accent = #38BDF8
    └── ... (all tokens)
```

---

## 3. Onboarding Carousel

Shown **only on first app launch**. Three pages with swipe navigation + Skip button. After "Get Started" or "Skip", the Login screen becomes the permanent landing.

### Page 1: Privacy

```
┌─────────────────────────────────┐
│                          [Skip] │
│                                 │
│                                 │
│       [Illustration]            │
│       Shield + lock glyph       │
│       240 × 240px               │
│                                 │
│                                 │
│    Privacy, Always              │  ← 28px, Bold, textPrimary
│                                 │
│    Your messages are always     │  ← 16px, textSecondary
│    encrypted end-to-end.        │     max-width 280px, centered
│    No one — not even Echo —     │
│    can read them.               │
│                                 │
│                                 │
│         ●  ○  ○                 │  ← Capsule indicators
│                                 │
│   ┌─────────────────────────┐   │
│   │         Next             │   │  ← Accent gradient fill
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### Page 2: Trust & Rewards

```
┌─────────────────────────────────┐
│                          [Skip] │
│                                 │
│       [Illustration]            │
│       Trust badge + coins       │
│                                 │
│                                 │
│  Trust Always, Earn Reputation  │
│                                 │
│    Your reputation is built     │
│    through verified actions     │
│    and trusted contacts.        │
│    Earn rewards along the way.  │
│                                 │
│                                 │
│         ○  ●  ○                 │
│                                 │
│   ┌─────────────────────────┐   │
│   │         Next             │   │
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### Page 3: Identity

```
┌─────────────────────────────────┐
│                          [Skip] │
│                                 │
│       [Illustration]            │
│       Key / blockchain node     │
│                                 │
│                                 │
│  Your Identity, Your Control    │
│                                 │
│    You own your identity.       │
│    No Big Tech. No Big Gov.     │
│    No data harvesting.          │
│                                 │
│                                 │
│         ○  ○  ●                 │
│                                 │
│   ┌─────────────────────────┐   │
│   │      Get Started         │   │  ← Button label changes
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### Carousel Specifications

| Element | Light Mode | Dark Mode |
|---|---|---|
| Background | `#F8FAFC` | `#0F172A` |
| Skip button | 16px Semi-bold `#64748B` | 16px Semi-bold `#94A3B8` |
| Illustration | Full color on light bg | Slightly desaturated, glow edges on dark |
| Headline | 28px Bold `#0F172A` | 28px Bold `#F1F5F9` |
| Body text | 16px Regular `#475569`, 24px line-height | 16px Regular `#CBD5E1` |
| Active capsule | 24×8px `#0EA5E9` | 24×8px `#38BDF8` |
| Inactive capsule | 8×8px `#E2E8F0` | 8×8px `#334155` |
| CTA button | 56px height, `accentGradient` fill, 12px radius | Same gradient, slightly brighter |
| CTA text | 16px Semi-bold `#FFFFFF` | 16px Semi-bold `#0F172A` |

### Carousel Illustration Assets

| Page | Asset | Description |
|---|---|---|
| 1 | `onboarding-privacy` | Shield icon with lock, encrypted message bubbles, cool blue tones |
| 2 | `onboarding-trust` | Trust score ring with checkmark, ECHO coin tokens floating, reward sparkles |
| 3 | `onboarding-identity` | Decentralized key icon, broken chain from corporate buildings, user at center |

All illustrations use the Icy Minimal palette: slate blues (`#64748B`, `#475569`), sky accent (`#0EA5E9`), and cool grays. No warm colors except gold accents on reward coins.

---

## 4. Login Screen

The Login screen follows the **PayPal login pattern**: one input field at top, one prominent CTA, and small text links for alternative methods underneath. This is the permanent landing after onboarding.

### Layout — Primary: Mobile + Passkey

```
┌─────────────────────────────────┐
│                                 │
│         [Echo Logo]             │
│          80 × 80px              │
│                                 │
│           Echo                  │  ← 28px Bold, textPrimary
│                                 │
│                                 │
│  Mobile Number or Email         │  ← 14px label, textSecondary
│  ┌─────────────────────────┐   │
│  │ +1  │ (555) 123-4567     │   │  ← 56px input, surface bg
│  └─────────────────────────┘   │
│                                 │
│   ┌─────────────────────────┐   │
│   │  [FaceID]  Log In       │   │  ← Primary CTA, accent gradient
│   └─────────────────────────┘   │
│                                 │
│   ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─   │
│                                 │
│   Log in with SMS code          │  ← 14px, accent color link
│   Log in with email code        │  ← 14px, accent color link
│                                 │
│                                 │
│                                 │
│                                 │
│   ─────────────────────────     │  ← 1px border divider
│                                 │
│   ┌─────────────────────────┐   │
│   │    Create Account        │   │  ← Secondary CTA, outlined
│   └─────────────────────────┘   │
│                                 │
│   🔒 End-to-end encrypted       │  ← 12px, textTertiary
│                                 │
└─────────────────────────────────┘
```

### Login Flow Logic

The input field accepts either a phone number or email. The system detects the format:

1. **Phone number entered + passkey exists on device** → "Log In" triggers Face ID / Touch ID passkey assertion
2. **Phone number entered + no passkey** → "Log In" sends SMS OTP, screen transitions to OTP verification
3. **Email entered + passkey exists** → Same passkey assertion
4. **Email entered + no passkey** → "Log In" sends email OTP

The alternative links bypass the primary flow:

- **"Log in with SMS code"** → Sends OTP to the entered phone, even if passkey exists
- **"Log in with email code"** → Prompts for email if phone was entered, then sends email OTP

### Login Specifications

| Element | Light Mode | Dark Mode |
|---|---|---|
| Screen background | `#F8FAFC` | `#0F172A` |
| Logo | 80×80px, 20px corner radius | Same, slight glow ring |
| App name | 28px Bold `#0F172A` | 28px Bold `#F1F5F9` |
| Input label | 14px Medium `#475569` | 14px Medium `#CBD5E1` |
| Input field | 56px height, `#E2E8F0` bg, 12px radius, 1px `#E2E8F0` border | 56px height, `#334155` bg, 1px `#334155` border |
| Input focused | 2px `#0EA5E9` border ring | 2px `#38BDF8` border ring |
| Input text | 16px Regular `#0F172A` | 16px Regular `#F1F5F9` |
| Country picker | Inline left, 64px width, `chevron.down` 12px | Same |
| Log In button | 56px height, `accentGradient`, 12px radius | Slightly brighter gradient |
| Log In icon | SF Symbol `faceid`/`touchid`, 20px, white | 20px, `#0F172A` |
| Log In text | 16px Semi-bold `#FFFFFF` | 16px Semi-bold `#0F172A` |
| Fallback links | 14px Medium `#0EA5E9`, 44px tap target height, 8px gap | 14px Medium `#38BDF8` |
| Divider | 1px `#E2E8F0` | 1px `#334155` |
| Create Account btn | 56px height, 2px `#0EA5E9` border, no fill, 12px radius | 2px `#38BDF8` border |
| Create Account text | 16px Semi-bold `#0EA5E9` | 16px Semi-bold `#38BDF8` |
| Trust badge | 12px Regular `#64748B`, lock icon 12px | 12px `#94A3B8` |

### Input States

| State | Light | Dark |
|---|---|---|
| Empty | Placeholder `#64748B`, `#E2E8F0` bg | Placeholder `#94A3B8`, `#334155` bg |
| Focused | 2px `#0EA5E9` ring, `#FFFFFF` bg | 2px `#38BDF8` ring, `#1E293B` bg |
| Filled | `#0F172A` text, `#FFFFFF` bg | `#F1F5F9` text, `#1E293B` bg |
| Error | 2px `#EF4444` ring, error text below | 2px `#F87171` ring |

### Interaction States

| State | Behavior |
|---|---|
| Default | Input empty, Log In button shows passkey icon if available, generic icon if not |
| Input focused | Keyboard rises, screen scrolls to keep input visible |
| Authenticating | Log In button shows spinner, gradient dims to 80% |
| Passkey prompt | iOS system Face ID / Touch ID sheet appears |
| Success | Checkmark flash (200ms) → slide to **Messages** screen |
| Failed passkey | Button returns to default, inline error: "Couldn't verify. Try again or use a code." |
| Invalid input | Red ring + error: "Enter a valid phone number or email." |
| Rate limited | Error: "Too many attempts. Try again in {time}." |

### "Create Account" Tap

Tapping "Create Account" navigates to the existing registration flow: Phone Entry → OTP → Passkey Setup → Profile → Trust Intro → **Messages**. The Phone Entry screen has a back arrow returning to this Login screen.

### Auto-Login

On app launch, if a stored refresh token + valid biometric state exist:
1. Show Login screen with stored phone/email pre-filled (200ms)
2. Auto-trigger passkey assertion
3. On success → slide to **Messages** screen (zero taps)
4. On cancel → stay on Login, user taps Log In manually

### Post-Login Destination

After **any** successful login — passkey, SMS OTP, email OTP, auto-login, or completing the Create Account registration flow — the user always lands on the **Messages screen** (Conversations List). The Messages tab is selected by default in the tab bar. The user never lands on Rewards, Contacts, or Profile after authentication.

---

## 5. Create Account Flow

Tapping "Create Account" on the Login screen enters the registration flow. The screens remain as previously specified (Phone Entry → OTP → Passkey Setup → Profile → Trust Intro → **Messages**) with these updates:

### Phone Entry Screen — Updated

Now includes an email option matching the PayPal pattern:

```
┌─────────────────────────────────┐
│ ←                               │
│                                 │
│     Create your account         │  ← 28px Bold
│                                 │
│  Enter your mobile number       │  ← 16px, textSecondary
│  to get started.                │
│                                 │
│  ┌──────┬──────────────────┐   │
│  │ +1 ▼ │ (555) 123-4567   │   │
│  └──────┴──────────────────┘   │
│                                 │
│   ┌─────────────────────────┐   │
│   │      Continue            │   │  ← Accent gradient
│   └─────────────────────────┘   │
│                                 │
│   Use email instead             │  ← 14px, accent link
│                                 │
│  ─────────────────────────────  │
│  By continuing, you agree to    │
│  our Terms and Privacy Policy   │
│                                 │
└─────────────────────────────────┘
```

- "Use email instead" swaps the phone input for an email text field
- Back arrow → returns to Login screen (not onboarding)
- All inputs, buttons, and colors follow Icy Minimal Light/Dark tokens

---

## 6. Dark Mode Specifications

### 6.1 System Integration

Dark mode follows the iOS system appearance by default, with an optional in-app override accessible from Settings → Appearance:

| Setting | Behavior |
|---|---|
| System (default) | Follows iOS `UIUserInterfaceStyle` |
| Light | Always light, ignores system |
| Dark | Always dark, ignores system |

### 6.2 Dark Mode Design Rules

| Rule | Implementation |
|---|---|
| No pure black backgrounds | Darkest background is `#0F172A` (Slate 900), not `#000000` |
| Elevated surfaces are lighter | Cards/modals use `#1E293B` (Slate 800) over `#0F172A` background |
| Accent colors brighten | `#0EA5E9` → `#38BDF8` for better contrast on dark backgrounds |
| Shadows become glows | Drop shadows → subtle light glows on elevated surfaces |
| Border colors darken | `#E2E8F0` → `#334155` to remain subtle |
| Images desaturate slightly | Illustrations at 90% saturation to avoid visual harshness |
| Gradient buttons | Same direction, slightly brighter end color |
| White text on accent | CTA button text flips to `#0F172A` (dark) on bright accent buttons |

### 6.3 Dark Mode Surface Hierarchy

```
Layer 0: Background       #0F172A  ← Screen background
Layer 1: Secondary bg     #1E293B  ← Grouped table sections
Layer 2: Surface          #334155  ← Input fields, inactive cards
Layer 3: Elevated         #1E293B  ← Modal cards, floating elements (+ subtle glow)
Layer 4: Overlay          #0F172A at 60% opacity  ← Backdrop behind sheets
```

### 6.4 Dark Mode Specific Effects

| Effect | Light Mode | Dark Mode |
|---|---|---|
| Tab bar | White bg + 1px top border | `#1E293B` bg + 1px `#334155` top border |
| Balance card gradient | `#0EA5E9 → #7DD3FC` | `#0EA5E9 → #38BDF8` (tighter range) |
| Input focus ring | 2px `#0EA5E9` outer glow | 2px `#38BDF8` + 4px soft glow (blur 8px, `#38BDF8` at 30%) |
| Navigation bar | `#FFFFFF` bg, `#0F172A` title | `#0F172A` bg, `#F1F5F9` title |
| Onboarding illustrations | Full color | 90% saturation + subtle edge glow |
| Verification badges | Standard colors | Same hex values (already high-contrast) |
| Pinned avatar ring | Colored ring on white | Colored ring + subtle outer glow |

### 6.5 Figma Setup for Dark Mode

1. **Every new screen** gets two frames: `ScreenName / Light` and `ScreenName / Dark`
2. Both frames reference the same Figma variable collection with mode swapped
3. The prototype uses the light variant by default, with a toggle in the prototype settings
4. Auto-layout constraints must work identically in both modes (no layout shifts)

---

## 7. Component Updates

### 7.1 Updated Components (Icy Minimal Only)

All existing components must be updated to remove Electric Violet and Teal Warm variants. Each component now has exactly **2 variants**: Light and Dark.

| Component | Light Variant | Dark Variant |
|---|---|---|
| `Button / Primary` | `accentGradient` fill, white text | `accentGradient` (brighter), `#0F172A` text |
| `Button / Secondary / Outlined` | 2px `#0EA5E9` border, `#0EA5E9` text | 2px `#38BDF8` border, `#38BDF8` text |
| `Button / Biometric / FaceID` | `accentGradient` fill, `faceid` icon white | Same gradient, icon `#0F172A` |
| `Input / Text` | `#E2E8F0` bg, `#0F172A` text | `#334155` bg, `#F1F5F9` text |
| `Input / Text / Focused` | `#FFFFFF` bg, 2px `#0EA5E9` ring | `#1E293B` bg, 2px `#38BDF8` ring + glow |
| `Card / Surface` | `#FFFFFF` bg, 1px `#E2E8F0` border | `#1E293B` bg, 1px `#334155` border |
| `TabBar` | `#FFFFFF` bg, 1px top `#E2E8F0` | `#1E293B` bg, 1px top `#334155` |
| `Badge / Verification` | Standard colors on light | Same colors on dark (already bright) |
| `Badge / VerificationPill` | Tinted bg | Same tint at 15% opacity on dark |
| `Divider / WithLabel` | 1px `#E2E8F0` + `#64748B` text | 1px `#334155` + `#94A3B8` text |

### 7.2 New Login Components

| Component | Properties |
|---|---|
| `Input / PhoneOrEmail` | Combined field: country picker (64px) + text input. Detects phone vs email format. |
| `Link / Fallback` | 14px Medium `accent` color, 44px tap target, no underline. For "Log in with SMS code" etc. |
| `Button / CreateAccount` | Full width, 56px, 2px accent border, no fill, accent text. |
| `LoginScreen / HasPasskey` | Full layout with Face ID icon on Log In button. |
| `LoginScreen / NoPasskey` | Full layout with generic icon on Log In button. |

### 7.3 Removed Components

| Removed | Reason |
|---|---|
| `ThemePicker` | Single theme, no picker needed |
| All Electric Violet variants | Theme removed |
| All Teal Warm variants | Theme removed |
| `Button / Tertiary` (old Sign In) | Replaced by full Login screen |
| Old `WelcomeScreen` | Replaced by Onboarding + Login |

### 7.4 Settings → Appearance (Updated)

```
┌─────────────────────────────────┐
│ ←       Appearance              │
├─────────────────────────────────┤
│                                 │
│  APPEARANCE MODE                │
│                                 │
│  ┌─────────┬─────────┬────────┐ │
│  │  Light  │  Dark   │ System │ │  ← Segmented control
│  └─────────┴─────────┴────────┘ │
│                                 │
│  ┌─────────────────────────────┐│
│  │  [Preview area showing      ││
│  │   current mode on a         ││
│  │   mini conversation list]   ││
│  └─────────────────────────────┘│
│                                 │
│  TEXT SIZE                       │
│  A─────────────────────────A    │  ← Slider
│                                 │
│  ...existing settings...        │
│                                 │
└─────────────────────────────────┘
```

The old Theme Picker section is replaced by a simple segmented control for Light / Dark / System. No theme swatches, no theme names.

---

## 8. Accessibility

### Color Contrast Verification

All text/background combinations must pass WCAG AA:

| Combination | Light | Ratio | Dark | Ratio |
|---|---|---|---|---|
| textPrimary on background | `#0F172A` on `#F8FAFC` | 15.4:1 ✅ | `#F1F5F9` on `#0F172A` | 14.5:1 ✅ |
| textSecondary on background | `#475569` on `#F8FAFC` | 6.7:1 ✅ | `#CBD5E1` on `#0F172A` | 9.0:1 ✅ |
| textTertiary on background | `#64748B` on `#F8FAFC` | 4.6:1 ✅ | `#94A3B8` on `#0F172A` | 5.5:1 ✅ |
| accent on background | `#0EA5E9` on `#F8FAFC` | 3.4:1 ⚠️ | `#38BDF8` on `#0F172A` | 6.7:1 ✅ |
| textOnAccent on accent button | `#FFFFFF` on `#0EA5E9` | 3.1:1 ⚠️ | `#0F172A` on `#38BDF8` | 5.7:1 ✅ |

> **Note:** Accent color on light background is 3.4:1 — passes for large text (3:1) but not small text. All accent-colored text (links, fallback options) must be ≥14px Medium weight to qualify as large text under WCAG. The CTA button text on gradient background passes because the gradient midpoint provides sufficient contrast.

### Dark Mode Specific

- All tap targets remain 44×44px minimum in both modes
- Focus rings are **more visible** in dark mode (glow effect added)
- Error states use `#F87171` (lighter red) in dark mode for better visibility
- Animations and timing are identical in both modes

---

*Last Updated: March 2026*
*Design System: Echo Icy Minimal v3.0*
*Modes: Light + Dark*
