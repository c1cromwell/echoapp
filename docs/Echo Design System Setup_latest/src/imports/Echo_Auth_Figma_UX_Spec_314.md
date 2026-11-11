# Echo Authentication — Figma UX Design Spec

> **Version:** 2.2 · **Date:** March 2026 · **Status:** Ready for Design
> **Figma File:** Echo Design System · **Pages:** Screens / Onboarding, Components / Overlays, Screens / Settings

---

## Table of Contents

1. [Overview](#1-overview)
2. [New Screens Summary](#2-new-screens-summary)
3. [Onboarding Carousel (First Launch)](#3-onboarding-carousel-first-launch)
4. [Login Screen (Default Landing)](#4-login-screen-default-landing)
5. [OTP Verification Screen (Standardized)](#5-otp-verification-screen-standardized)
6. [Passkey Setup Screen (Standardized)](#6-passkey-setup-screen-standardized)
7. [Step-Up Verification Modal](#7-step-up-verification-modal)
8. [Device Management Screen](#8-device-management-screen)
9. [New Device Detected Alert](#9-new-device-detected-alert)
10. [Account Recovery Flow](#10-account-recovery-flow)
11. [Account Locked Screen](#11-account-locked-screen)
12. [Login Audit Log Screen](#12-login-audit-log-screen)
13. [Verification Badges on Messages Screen](#13-verification-badges-on-messages-screen)
14. [Tab Bar Restructure & Rewards Tab](#14-tab-bar-restructure--rewards-tab)
15. [New Component Variants](#15-new-component-variants)
16. [Error States & Feedback](#16-error-states--feedback)
17. [Animation Specifications](#17-animation-specifications)
18. [Theme Application](#18-theme-application)
19. [Accessibility Checklist](#19-accessibility-checklist)

---

## 1. Overview

The current Echo Figma design system has components for onboarding screens (Phone Entry, OTP, Passkey, Profile, Trust Intro) and the alternative wallet flow. However, the **first-launch onboarding experience**, the **primary login landing screen**, and several security-related screens are missing or need restructuring.

**Key flow change (v2.1):** The old Welcome screen (with "Continue with Phone" / "Use Verifiable ID") is **removed**. It is replaced by a 3-page onboarding carousel that shows only on first app launch, followed by a Login screen that serves as the **permanent default landing page** with a "Set Up Account" button for new users.

This spec defines all new screens, updated screens, and new component variants required to support the full authentication system. Every screen must exist in all three theme variants (Electric Violet, Teal + Warm Neutral, Icy Minimal).

### Design Principles for Auth

- **One primary action per screen** — the user always knows what to do next
- **Biometric-first** — Face ID / Touch ID is always the primary CTA for returning users
- **Minimal text** — headlines are 5 words or fewer, body text is 2 lines max
- **Forgiving errors** — error states suggest next steps, never blame the user
- **Progressive disclosure** — security details (DID, tokens) are hidden unless the user seeks them

---

## 2. New Screens Summary

| # | Screen | Figma Page | Status | Priority |
|---|---|---|---|---|
| 1 | Onboarding Carousel (3 pages) | Screens / Onboarding | **NEW** — replaces old Welcome | P0 |
| 2 | Login (Default Landing) | Screens / Onboarding | **NEW** — primary entry point | P0 |
| 3 | OTP Verification (standardized) | Screens / Onboarding | UPDATE — standardize component | P0 |
| 4 | Passkey Setup (standardized) | Screens / Onboarding | UPDATE — add Touch ID variant | P0 |
| 5 | Step-Up Verification Modal | Components / Overlays | **NEW** | P0 |
| 6 | Conversations List (badges) | Screens / Messages | UPDATE — add verification badges | P0 |
| 7 | Chat Thread Header (badges) | Screens / Messages | UPDATE — add verification badge | P0 |
| 8 | Rewards Dashboard | Screens / Rewards | **NEW** — replaces Trust tab | P0 |
| 9 | Tab Bar | Components / Navigation | UPDATE — Rewards replaces Trust | P0 |
| 10 | Privacy & Security (Trust section) | Screens / Settings | UPDATE — Trust & Identity section added at top | P0 |
| 11 | Device Management | Screens / Settings | **NEW** | P1 |
| 12 | New Device Detected Alert | Components / Overlays | **NEW** | P1 |
| 13 | Account Recovery — Method Selection | Screens / Onboarding | **NEW** | P1 |
| 14 | Account Recovery — Trusted Contacts | Screens / Onboarding | **NEW** | P1 |
| 15 | Account Recovery — Waiting | Screens / Onboarding | **NEW** | P1 |
| 16 | Account Locked | Screens / Onboarding | **NEW** | P2 |
| 17 | Login Audit Log | Screens / Settings | **NEW** | P2 |

**Total: 10 new screens (including 3 carousel pages + Rewards Dashboard), 6 screen updates, 9 new component variants (including badge system)**
**Removed: Old Welcome screen (replaced by Onboarding Carousel + Login), Trust tab (replaced by Rewards tab, Trust moved to Settings)**

---

## 3. Onboarding Carousel (First Launch)

### Purpose

The onboarding carousel is shown **only on the very first app launch** on a device. It introduces Echo's three core value propositions in a swipeable 3-page walkthrough. After completion (or skip), the user proceeds to the Login screen and never sees the carousel again.

### Layout — Page 1: Privacy

```
┌─────────────────────────────────┐
│                          [Skip] │  ← 16px, textSecondary, top-right
│                                 │
│                                 │
│                                 │
│       [Illustration]            │  ← 240x240px, centered
│       shield + lock             │
│                                 │
│                                 │
│  Your messages, your rules      │  ← 28px, Bold, textPrimary, centered
│                                 │
│  End-to-end encrypted           │  ← 16px, textSecondary, centered
│  messaging that no one —        │
│  not even us — can read.        │
│                                 │
│                                 │
│         ●  ○  ○                 │  ← Page indicator (custom capsules)
│                                 │
│   ┌─────────────────────────┐   │
│   │         Next             │   │  ← Primary CTA (gradient)
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### Layout — Page 2: Trust

```
┌─────────────────────────────────┐
│                          [Skip] │
│                                 │
│       [Illustration]            │
│       trust score badge         │
│                                 │
│  Trust you can see              │
│                                 │
│  Every contact has a trust      │
│  score built from real          │
│  verification, not just a       │
│  blue check.                    │
│                                 │
│         ○  ●  ○                 │
│                                 │
│   ┌─────────────────────────┐   │
│   │         Next             │   │
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### Layout — Page 3: Identity

```
┌─────────────────────────────────┐
│                          [Skip] │
│                                 │
│       [Illustration]            │
│       blockchain / key          │
│                                 │
│  Own your identity              │
│                                 │
│  Your identity lives on the     │
│  blockchain, controlled by      │
│  you. No central server,        │
│  no data harvesting.            │
│                                 │
│         ○  ○  ●                 │
│                                 │
│   ┌─────────────────────────┐   │
│   │      Get Started         │   │  ← Button text changes on last page
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### Specifications

| Element | Spec |
|---|---|
| Skip button | 16px, Semi-bold, `textSecondary`, top-right, 24px padding, tap target 44x44 |
| Illustration area | 240x240px, centered horizontally, vertically centered in upper half |
| Headline | 28px, Bold, `textPrimary`, centered, max 2 lines |
| Subtitle | 16px, Regular, `textSecondary`, centered, max 3 lines, 4px line spacing |
| Page indicator | Custom capsules: active = 24x8px `primary` fill, inactive = 8x8px `border` fill |
| Indicator spacing | 8px gap between capsules |
| Primary button | Full width, 56px height, `primaryGradient` fill, 12px corner radius |
| Button text | Pages 1-2: "Next" / Page 3: "Get Started" — 16px, Semi-bold, `textOnPrimary` |
| Bottom padding | 40px from button bottom to screen edge |
| Horizontal padding | 24px on all content, 32px on subtitle for narrower text column |

### Interactions

| Action | Behavior |
|---|---|
| Swipe left | Advance to next page (spring animation, 300ms) |
| Swipe right | Return to previous page |
| Tap "Next" | Advance to next page (same animation as swipe) |
| Tap "Get Started" (page 3) | Transition to Login screen (slide-left, 300ms) |
| Tap "Skip" (any page) | Immediately transition to Login screen |

### Illustration Assets Needed

| Page | Asset Name | Description |
|---|---|---|
| 1 | `onboarding-privacy` | Shield with lock icon, message bubbles with encryption symbols |
| 2 | `onboarding-trust` | Trust score circle with verification badge, connected people |
| 3 | `onboarding-identity` | Blockchain key / DID icon, user silhouette with ownership arrows |

All illustrations should follow the existing Echo illustration style (see `assets/illustrations/`). Use the theme's `primary` and `accent` colors as highlight colors, with `textTertiary` for supporting elements.

### First-Launch Gate

- The carousel is shown **only when** `hasCompletedOnboarding` is `false` (UserDefaults)
- After tapping "Get Started" or "Skip", the flag is set to `true`
- The carousel **never shows again** — not after logout, not after session expiry
- If the app is deleted and reinstalled on the same device, UserDefaults is cleared and the carousel shows again (expected behavior)

---

## 4. Login Screen (Default Landing)

This is the **primary screen** of the entire auth experience. After the one-time onboarding carousel, this screen is shown every time the app is opened and the user is not authenticated. It serves two audiences: returning users (Sign In with biometric passkey) and new users ("Set Up Account" button).

### Layout — Returning User (existing credentials on device)

```
┌─────────────────────────────────┐
│                                 │
│                                 │
│         [Echo Logo]             │
│          100x100px              │
│                                 │
│           Echo                  │  ← 32px, Bold, textPrimary
│    Messaging with Trust         │  ← 16px, textSecondary
│                                 │
│                                 │
│   ┌─────────────────────────┐   │
│   │  [FaceID]  Sign In      │   │  ← Primary CTA (gradient fill)
│   └─────────────────────────┘   │
│                                 │
│        @alex_echo               │  ← 14px, textSecondary
│     +1 (5**) ***-4567           │  ← 14px, textTertiary, masked
│                                 │
│                                 │
│                                 │
│       ────── or ──────          │  ← Divider with "or" text
│                                 │
│   ┌─────────────────────────┐   │
│   │    Set Up Account        │   │  ← Secondary CTA (outlined)
│   └─────────────────────────┘   │
│                                 │
│   Lost access? [Recover]        │  ← 14px, textSecondary link
│                                 │
│   🔒 Your identity stays yours  │  ← 12px, textTertiary
│      No data sold. Ever.        │
│                                 │
└─────────────────────────────────┘
```

### Layout — New User (no credentials on device)

When no existing account is detected on the device, the layout adjusts: the account info section is hidden and the Sign In button uses a generic icon instead of the biometric icon.

```
┌─────────────────────────────────┐
│                                 │
│                                 │
│         [Echo Logo]             │
│          100x100px              │
│                                 │
│           Echo                  │
│    Messaging with Trust         │
│                                 │
│                                 │
│   ┌─────────────────────────┐   │
│   │  [person]  Sign In      │   │  ← Primary CTA (for existing users
│   └─────────────────────────┘   │     on a new device)
│                                 │
│                                 │
│       ────── or ──────          │
│                                 │
│   ┌─────────────────────────┐   │
│   │    Set Up Account        │   │  ← Secondary CTA (outlined)
│   └─────────────────────────┘   │
│                                 │
│   Lost access? [Recover]        │
│                                 │
│   🔒 Your identity stays yours  │
│      No data sold. Ever.        │
│                                 │
└─────────────────────────────────┘
```

### Figma Component Variants

Create a component set named `LoginScreen` with two variants:

| Variant | Description |
|---|---|
| `Account=Existing` | Shows biometric icon (Face ID / Touch ID), account info (username + masked phone) below Sign In |
| `Account=None` | Shows generic person icon, no account info, same button layout |

### Specifications

| Element | Spec |
|---|---|
| Logo | 100x100px, 24px corner radius, centered |
| App name | 32px, Bold, `textPrimary`, centered, 4px below logo |
| Tagline | 16px, Regular, `textSecondary`, centered, 4px below app name |
| Sign In Button | Full width, 56px height, `primaryGradient` fill, 12px corner radius |
| Sign In icon (existing) | SF Symbol `faceid` or `touchid`, 20px, `textOnPrimary`, 12px left of text |
| Sign In icon (new device) | SF Symbol `person.crop.circle.badge.checkmark`, 20px, `textOnPrimary` |
| Account info | Centered, 14px, `textSecondary` for username, `textTertiary` for phone |
| Phone masking | Country code + first digit + asterisks + last 4: `+1 (5**) ***-4567` |
| Divider | 1px `border` lines with "or" text (14px, `textTertiary`) centered between them |
| Set Up Account Button | Full width, 56px height, 2px `primary` border, no fill, 12px corner radius |
| Set Up Account Text | 16px, Semi-bold, `primary` color |
| Recovery link | 14px, Regular, `textSecondary`, centered, tap → Account Recovery |
| Trust badge | 12px, `textTertiary`, lock icon (SF Symbol `lock.fill`, 12px) + text, centered |
| Bottom padding | 32px from trust badge to screen edge |

### Interaction States

| State | Visual |
|---|---|
| Default | Sign In button with gradient, Set Up Account outlined |
| Authenticating | Sign In button shows centered spinner, gradient dims to 80% opacity |
| Success | Brief checkmark animation (200ms) → transition to Home |
| Failed | Sign In returns to default, error text appears below account info (14px, `error`) |
| Account locked | Entire screen transitions to Account Locked screen |

### Auto-Login Behavior

If the app launches and detects a stored refresh token + valid biometric state:
1. Show this Login screen briefly (200ms)
2. Automatically trigger the Face ID / Touch ID prompt
3. On success → transition to Home with no user interaction required
4. On cancel → stay on Login screen, user can tap Sign In manually

### "Set Up Account" Tap

Tapping "Set Up Account" navigates to the Phone Entry screen. The Phone Entry screen gets a back arrow that returns to this Login screen (not onboarding).

### No Back Arrow

This screen has **no back navigation**. It is the root of the unauthenticated state. The only exits are: Sign In (→ Home), Set Up Account (→ Phone Entry), or Recover (→ Recovery flow).

---

## 5. OTP Verification Screen (Standardized)

The OTP screen already exists in the design spec but needs standardization as a reusable component, since OTP verification is used in three contexts: registration, step-up authentication, and account recovery.

### Layout (Unchanged, Standardized)

```
┌─────────────────────────────────┐
│ ←                               │
│                                 │
│      Enter verification code    │
│                                 │
│   Sent to +1 (555) 123-4567    │
│                                 │
│     ┌───┬───┬───┬───┬───┬───┐   │
│     │ _ │ _ │ _ │ _ │ _ │ _ │   │
│     └───┴───┴───┴───┴───┴───┘   │
│                                 │
│        Resend code (0:45)       │
│                                 │
│   [Error message area]          │  ← Only visible on error
│                                 │
│      Wrong number? [Change]     │  ← Only in registration context
│                                 │
└─────────────────────────────────┘
```

### Component Variants (Figma)

Create a component set named `OTPVerification` with these variants:

| Variant | Subtitle | Bottom Link | Context |
|---|---|---|---|
| `Context=Registration` | "Sent to {phone}" | "Wrong number? Change" | New account registration |
| `Context=StepUp` | "We sent a code to your phone" | (hidden) | Step-up authentication |
| `Context=Recovery` | "Verify your phone to recover" | (hidden) | Account recovery |

### Input Box States

| State | Border | Background | Text |
|---|---|---|---|
| Empty (unfocused) | 1px `border` | `surface` | — |
| Empty (focused) | 2px `primary` | `surface` | Cursor blink |
| Filled | 1px `border` | `primarySubtle` | 24px Semi-bold `textPrimary` |
| Error | 2px `error` | `errorLight` | — (cleared on error) |

### Animations

- **Auto-advance:** On digit entry, cursor moves to next box (instant)
- **Error shake:** Horizontal shake 6px amplitude, 2 cycles, 400ms
- **Haptic:** Light impact on each digit entry; error haptic on invalid code

---

## 6. Passkey Setup Screen (Standardized)

### Component Variants

Create variants for Face ID and Touch ID, with automatic selection based on device:

| Variant | Icon | Button Text | Available |
|---|---|---|---|
| `Biometric=FaceID` | `faceid` SF Symbol (80x80) | "Enable Face ID" | iPhone X+ |
| `Biometric=TouchID` | `touchid` SF Symbol (80x80) | "Enable Touch ID" | iPhone SE, iPad |

Both share identical layout, benefits list, and skip link.

---

## 7. Step-Up Verification Modal

A bottom sheet (half-screen) that appears when the user performs a sensitive action. Presented as a `.sheet` with `.presentationDetents([.fraction(0.4)])`.

### Layout

```
┌─────────────────────────────────┐
│                                 │
│          ───────                │  ← Drag indicator (36x5px, border color)
│                                 │
│      Verify it's you            │  ← 24px, Bold, textPrimary
│                                 │
│   This action requires          │  ← 16px, textSecondary
│   additional verification.      │
│                                 │
│   ┌─────────────────────────┐   │
│   │  [FaceID]  Verify now   │   │  ← Primary CTA (gradient)
│   └─────────────────────────┘   │
│                                 │
│     [Use verification code]     │  ← 14px, primary color link
│                                 │
└─────────────────────────────────┘
```

### Specifications

| Element | Spec |
|---|---|
| Container | Bottom sheet, 40% screen height, 24px corner radius (top), `background` fill |
| Drag indicator | 36x5px, `border` color, 8px rounded, centered, 8px top padding |
| Title | 24px, Bold, `textPrimary` |
| Subtitle | 16px, Regular, `textSecondary`, max 2 lines |
| Primary button | Same as Login screen biometric button |
| Fallback link | 14px, Semi-bold, `primary` color |
| Backdrop | 50% black overlay behind the sheet |

### Contextual Subtitles

The subtitle changes based on the triggering action:

| Action | Subtitle |
|---|---|
| Revoke device | "Removing a device requires your verification." |
| Large payment | "Payments over 100 ECHO require verification." |
| Change phone | "Changing your phone number requires verification." |
| Export recovery | "Your recovery phrase is sensitive. Verify to continue." |
| Delete account | "Deleting your account is permanent. Verify to continue." |

---

## 8. Device Management Screen

Accessed from Settings → Devices. Shows all registered devices with the ability to revoke access.

### Layout

```
┌─────────────────────────────────┐
│ ←              Devices          │
├─────────────────────────────────┤
│                                 │
│  This Device                    │  ← 14px, Semi-bold, textSecondary
│  ┌─────────────────────────────┐│
│  │ 📱 iPhone 15 Pro            ││  ← 16px, Semi-bold, textPrimary
│  │    iOS 17.4 • Active now    ││  ← 14px, textSecondary
│  │    New York, NY        🟢   ││  ← 14px, textTertiary + green dot
│  └─────────────────────────────┘│
│                                 │
│  Other Devices                  │  ← 14px, Semi-bold, textSecondary
│  ┌─────────────────────────────┐│
│  │ 📱 iPad Air                 ││
│  │    iPadOS 17.2 • 2 hrs ago  ││
│  │    New York, NY    [Remove] ││  ← 14px, error color text button
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ 💻 MacBook Pro              ││
│  │    macOS 14.3 • 3 days ago  ││
│  │    Brooklyn, NY    [Remove] ││
│  └─────────────────────────────┘│
│                                 │
│  ─────────────────────────────  │
│                                 │
│  ┌─────────────────────────────┐│
│  │    Log Out All Devices      ││  ← Destructive button (error color fill)
│  └─────────────────────────────┘│
│                                 │
└─────────────────────────────────┘
```

### Specifications

| Element | Spec |
|---|---|
| Navigation | Standard nav bar, "Devices" title, back arrow |
| Section headers | 14px, Semi-bold, `textSecondary`, 16px bottom padding |
| Device card | `Card` component, 16px corner radius, `surfaceElevated` fill, `border` 1px stroke |
| Device icon | Platform-specific emoji or SF Symbol, 24px |
| Device name | 16px, Semi-bold, `textPrimary` |
| Device details | 14px, Regular, `textSecondary`, "OS • Time" format |
| Location | 14px, Regular, `textTertiary` |
| Active indicator | 8px green dot (`success` color) for current device |
| Remove button | 14px, Semi-bold, `error` color, tap → Step-Up modal |
| Log Out All | Full width, 48px height, `error` color fill, white text, 12px radius |
| Card spacing | 12px gap between cards |

### Empty State

If no other devices are registered:

```
┌─────────────────────────────────┐
│                                 │
│  Other Devices                  │
│                                 │
│        [Device icon]            │
│                                 │
│    No other devices             │  ← 16px, textSecondary, centered
│    registered                   │
│                                 │
│    You can add devices by       │  ← 14px, textTertiary
│    scanning a QR code from      │
│    your primary device.         │
│                                 │
└─────────────────────────────────┘
```

### Remove Device Confirmation

Tapping "Remove" triggers the Step-Up Verification Modal (Section 7). After successful verification, show a confirmation alert:

```
┌─────────────────────────────────┐
│                                 │
│   Device removed                │  ← 20px, Semi-bold
│                                 │
│   iPad Air has been logged      │  ← 16px, textSecondary
│   out and can no longer         │
│   access your account.          │
│                                 │
│   ┌─────────────────────────┐   │
│   │         Done             │   │
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

---

## 9. New Device Detected Alert

An overlay alert that appears when a login succeeds but the device fingerprint doesn't match any registered device. This is **not** a bottom sheet — it's a centered modal alert.

### Layout

```
┌─────────────────────────────────┐
│                                 │
│   ┌───────────────────────────┐ │
│   │                           │ │
│   │    ⚠️ New device           │ │  ← 20px, Semi-bold, warning color
│   │                           │ │
│   │  We noticed you're         │ │  ← 16px, textSecondary
│   │  signing in from a         │ │
│   │  new device.               │ │
│   │                           │ │
│   │  📱 iPhone 16 Pro          │ │  ← 16px, Semi-bold, textPrimary
│   │     iOS 18.0               │ │  ← 14px, textTertiary
│   │     Chicago, IL            │ │  ← 14px, textTertiary
│   │                           │ │
│   │  ┌───────────────────┐    │ │
│   │  │  Verify & Add     │    │ │  ← Primary CTA → starts step-up
│   │  └───────────────────┘    │ │
│   │                           │ │
│   │    [That's not me]         │ │  ← 14px, error color → revoke + alert
│   │                           │ │
│   └───────────────────────────┘ │
│                                 │
└─────────────────────────────────┘
```

### Specifications

| Element | Spec |
|---|---|
| Backdrop | 50% black overlay, full screen |
| Alert card | 320px width, auto height, 20px corner radius, `surfaceElevated` fill |
| Alert card shadow | `shadow-lg` (0px 8px 32px rgba(0,0,0,0.15)) |
| Warning icon | ⚠️ emoji or SF Symbol `exclamationmark.triangle.fill`, 32px, `warning` color |
| Title | 20px, Semi-bold, `textPrimary` |
| Body | 16px, Regular, `textSecondary` |
| Device info | Card-in-card, `surface` background, 12px corner radius, 12px padding |
| Verify button | Full width within card, 48px height, `primary` gradient |
| "That's not me" | 14px, Semi-bold, `error` color, centered |

### "That's not me" Flow

If the user taps "That's not me":
1. Revoke the suspicious session immediately
2. Show: "We've blocked this device. Consider changing your password if you have one, and review your trusted contacts."
3. Navigate to Device Management screen

---

## 10. Account Recovery Flow

Three screens for the recovery process:

### Screen 10A: Recovery Method Selection

```
┌─────────────────────────────────┐
│ ←                               │
│                                 │
│      Recover your account       │  ← 28px, Bold
│                                 │
│   Choose how to verify your     │  ← 16px, textSecondary
│   identity.                     │
│                                 │
│   ┌─────────────────────────┐   │
│   │ 🔑 Recovery Phrase       │   │
│   │    Enter your 12-word    │   │
│   │    backup phrase         │   │
│   └─────────────────────────┘   │
│                                 │
│   ┌─────────────────────────┐   │
│   │ 👥 Trusted Contacts      │   │
│   │    Ask friends to help   │   │
│   │    verify your identity  │   │
│   └─────────────────────────┘   │
│                                 │
│   ┌─────────────────────────┐   │
│   │ 📱 Phone Verification    │   │
│   │    Verify via your       │   │
│   │    registered phone      │   │
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

**Card specifications:** 16px corner radius, `surfaceElevated` fill, 1px `border` stroke, 16px internal padding. Icon 32px left-aligned. Title 16px Semi-bold, subtitle 14px `textSecondary`. Tap → navigate to specific recovery flow.

### Screen 10B: Trusted Contact Recovery

```
┌─────────────────────────────────┐
│ ←                               │
│                                 │
│      Waiting for friends        │  ← 28px, Bold
│                                 │
│   We've sent a request to       │  ← 16px, textSecondary
│   your trusted contacts.        │
│   2 of 3 need to confirm.       │
│                                 │
│   ┌─────────────────────────┐   │
│   │ 😊 Sarah M.        ✅    │   │  ← Confirmed
│   └─────────────────────────┘   │
│   ┌─────────────────────────┐   │
│   │ 😊 James K.        ⏳    │   │  ← Pending
│   └─────────────────────────┘   │
│   ┌─────────────────────────┐   │
│   │ 😊 Maria L.        ⏳    │   │  ← Pending
│   └─────────────────────────┘   │
│                                 │
│   ── Progress: 1 of 2 needed ──│  ← Progress bar, primary color
│                                 │
│   This request expires in       │
│   23 hours, 45 minutes.         │  ← 14px, textTertiary
│                                 │
└─────────────────────────────────┘
```

**Status icons:** ✅ = confirmed (`success` color), ⏳ = pending (`textTertiary`), ❌ = declined (`error` color)

**Progress bar:** Full width, 4px height, `surface` background, `primary` fill proportional to confirmations.

### Screen 10C: Recovery Complete

```
┌─────────────────────────────────┐
│                                 │
│                                 │
│            ✅                    │  ← 64px success icon
│                                 │
│      Account recovered!         │  ← 28px, Bold
│                                 │
│   Set up a new passkey to       │  ← 16px, textSecondary
│   secure your account.          │
│                                 │
│   ┌─────────────────────────┐   │
│   │  [FaceID] Set Up Now    │   │  ← Primary CTA → Passkey Setup
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

---

## 11. Account Locked Screen

Shown when the account is temporarily locked due to too many failed attempts or suspicious activity.

### Layout

```
┌─────────────────────────────────┐
│                                 │
│                                 │
│            🔒                    │  ← 64px lock icon
│                                 │
│      Account locked             │  ← 28px, Bold, textPrimary
│                                 │
│   Too many failed attempts.     │  ← 16px, textSecondary
│   Try again in:                 │
│                                 │
│          47:32                   │  ← 48px, Bold, primary, countdown
│                                 │
│                                 │
│   ─────────────────────────     │
│                                 │
│   If this wasn't you,           │  ← 14px, textSecondary
│   [recover your account]        │  ← 14px, primary link
│   to secure it.                 │
│                                 │
│   ┌─────────────────────────┐   │
│   │    Contact Support       │   │  ← Secondary button (outlined)
│   └─────────────────────────┘   │
│                                 │
└─────────────────────────────────┘
```

### Specifications

| Element | Spec |
|---|---|
| Lock icon | SF Symbol `lock.fill`, 64px, `textTertiary` |
| Title | 28px, Bold, `textPrimary` |
| Countdown timer | 48px, Bold, `primary` color, MM:SS format, updates every second |
| Recovery link | 14px, `primary` color, navigates to Account Recovery |
| Support button | Full width, 48px, outlined in `border`, `textSecondary` text |

### Locked Reasons

| Reason | Subtitle Text |
|---|---|
| Too many attempts | "Too many failed attempts. Try again in:" |
| Suspicious activity | "Unusual activity detected. Try again in:" |
| Account suspended | "Your account has been suspended. Contact support for help." (no timer) |

---

## 12. Login Audit Log Screen

Accessed from Settings → Security → Login History. Shows a chronological list of login events.

### Layout

```
┌─────────────────────────────────┐
│ ←           Login History       │
├─────────────────────────────────┤
│                                 │
│  Today                          │  ← Section header
│  ┌─────────────────────────────┐│
│  │ ✅ Successful login         ││  ← 16px, Semi-bold
│  │    iPhone 15 Pro • Face ID  ││  ← 14px, textSecondary
│  │    New York, NY • 2:34 PM   ││  ← 14px, textTertiary
│  └─────────────────────────────┘│
│                                 │
│  Yesterday                      │
│  ┌─────────────────────────────┐│
│  │ ✅ Successful login         ││
│  │    iPhone 15 Pro • Face ID  ││
│  │    New York, NY • 11:20 AM  ││
│  └─────────────────────────────┘│
│  ┌─────────────────────────────┐│
│  │ ❌ Failed login             ││  ← error color icon
│  │    Unknown device           ││
│  │    Chicago, IL • 8:15 AM    ││  ← Unfamiliar location highlighted
│  └─────────────────────────────┘│
│                                 │
│  March 8                        │
│  ┌─────────────────────────────┐│
│  │ ⚠️ New device added         ││  ← warning color icon
│  │    iPad Air registered      ││
│  │    New York, NY • 3:45 PM   ││
│  └─────────────────────────────┘│
│                                 │
└─────────────────────────────────┘
```

### Event Types

| Event | Icon | Icon Color |
|---|---|---|
| Successful login | ✅ (checkmark.circle.fill) | `success` |
| Failed login | ❌ (xmark.circle.fill) | `error` |
| New device added | ⚠️ (exclamationmark.circle.fill) | `warning` |
| Device removed | 🔄 (arrow.triangle.2.circlepath) | `textTertiary` |
| Password/passkey changed | 🔑 (key.fill) | `primary` |
| Account recovery | 🛡️ (shield.fill) | `accent` |

### Suspicious Activity Highlight

If a login event comes from an unfamiliar location or unknown device, the list item gets a subtle highlight:

- Background: `warningLight` (light yellow)
- Left border: 3px `warning` color

---

## 13. Verification Badges on Messages Screen

The verification badge is a colored checkmark displayed next to user names on the Conversations List and Chat Thread header. The color communicates verification status at a glance. A special gold crown icon distinguishes VIP Pro subscribers.

### 13.1 Badge Tier System

| Status | Icon | Color | Hex | When Shown |
|---|---|---|---|---|
| None | — (hidden) | — | — | Unverified users, trust tier 0 with no badges |
| Phone Verified | ✓ (checkmark.seal.fill) | Gray | `#9CA3AF` | Phone verified, no identity verification |
| Identity Lite | ✓ (checkmark.seal.fill) | Indigo | `#6366F1` | KYC-lite via 3rd-party service (Prove, Daon) |
| Identity Full | ✓ (checkmark.seal.fill) | Green | `#10B981` | Government ID or Apple Digital ID verified |
| VIP Pro | 👑 (crown.fill) | Gold | `#F59E0B` | Paid VIP Pro subscriber + identity verified |

### 13.2 Conversations List — Badge Placement

The badge appears inline after the contact name on each conversation row, left of the timestamp.

```
┌─────────────────────────────────┐
│  Messages                    ✏️  │
├─────────────────────────────────┤
│ 🔍 Search                       │
├─────────────────────────────────┤
│ [Pinned section...]             │
├── ALL MESSAGES ─────────────────┤
│                                 │
│ ┌───┐                          │
│ │ AV│ Sarah M. ✓   2:34 PM    │  ← Green checkmark (identity full)
│ │   │ Hey, are you free...  2  │
│ └───┘                          │
│ ─────────────────────────────  │
│ ┌───┐                          │
│ │ AV│ James K. 👑   1:15 PM    │  ← Gold crown (VIP Pro)
│ │   │ Just sent the payment    │
│ └───┘                          │
│ ─────────────────────────────  │
│ ┌───┐                          │
│ │ AV│ Alex T. ✓   Yesterday   │  ← Indigo checkmark (identity lite)
│ │   │ Check out this link...   │
│ └───┘                          │
│ ─────────────────────────────  │
│ ┌───┐                          │
│ │ AV│ New Contact  11:30 AM    │  ← No badge (unverified)
│ │   │ Hi, I found you on...    │
│ └───┘                          │
│                                 │
└─────────────────────────────────┘
```

### Conversation Row Specifications

| Element | Spec |
|---|---|
| Badge position | 4px right of name text, vertically centered with name baseline |
| Badge size (list) | 14px (SF Symbol at 14pt weight semibold) |
| Badge color | Matches tier color from table above |
| Name + badge group | Treated as a single horizontal unit; truncates name if needed, badge never truncates |
| VIP Pro icon | SF Symbol `crown.fill`, 14px, gold `#F59E0B` — replaces checkmark entirely |
| Group conversations | No badge shown (groups don't have individual verification) |
| Spacing | Name text → 4px → badge → flexible space → timestamp |

### 13.3 Chat Thread Header — Badge Placement

The badge appears next to the contact name in the chat header, slightly larger than the list version.

```
┌─────────────────────────────────┐
│ ← [AV] Sarah M. ✓   Online 📞🎥│
├─────────────────────────────────┤
│                                 │
│                   ┌──────────┐  │
│                   │ Hey! How │  │
│                   │ are you? │  │
│                   └──────────┘  │
│  ┌──────────────┐               │
│  │ I'm great,   │               │
│  │ thanks!      │               │
│  └──────────────┘               │
│                                 │
├─────────────────────────────────┤
│ [+] Message...          [Send]  │
└─────────────────────────────────┘
```

### Chat Header Specifications

| Element | Spec |
|---|---|
| Badge position | 4px right of name text, vertically centered |
| Badge size (header) | 16px (SF Symbol at 16pt weight semibold) |
| Badge color | Same tier colors as list |
| Tap behavior | Tapping name + badge area opens contact profile |
| VIP Pro | Gold crown, 16px, same position |

### 13.4 Group Chat — Sender Badge

In group conversations, the badge appears next to the sender name above each message bubble.

```
┌─────────────────────────────────┐
│ ← [GRP] Team Echo       📞 🎥  │
├─────────────────────────────────┤
│                                 │
│  Sarah M. ✓                     │  ← Green check next to sender name
│  ┌──────────────┐               │
│  │ Meeting at 3 │               │
│  └──────────────┘               │
│                                 │
│  James K. 👑                     │  ← Gold crown next to sender name
│  ┌──────────────┐               │
│  │ I'll be there│               │
│  └──────────────┘               │
│                                 │
└─────────────────────────────────┘
```

| Element | Spec |
|---|---|
| Sender name | 13px, Semi-bold, `textSecondary` |
| Badge position | 3px right of sender name, vertically centered |
| Badge size | 12px (smaller to match 13px sender text) |

### 13.5 Badge on Pinned Section

For pinned contacts, the badge manifests as a colored ring around the avatar instead of an inline checkmark (since there's no name text in the compact pinned view).

```
    ┌─────┐    ┌─────┐    ┌─────┐
    │     │    │     │    │     │
    │ AV  │    │ AV  │    │ AV  │
    │     │    │     │    │     │
    └─────┘    └─────┘    └─────┘
     Sarah      James      Alex
      ✓          👑
```

| Status | Pinned Avatar Ring |
|---|---|
| None | Standard `border` color ring (1px) |
| Phone Verified | `#9CA3AF` gray ring (2px) |
| Identity Lite | `#6366F1` indigo ring (2px) |
| Identity Full | `#10B981` green ring (2px) |
| VIP Pro | `#F59E0B` gold ring (2px) + tiny crown icon (10px) bottom-right of avatar |

### 13.6 VIP Pro Badge — Special Treatment

VIP Pro gets distinct visual treatment beyond just color:

| Context | Rendering |
|---|---|
| Conversation list | Gold crown icon (`crown.fill`) replaces the checkmark seal |
| Chat header | Gold crown icon, same 16px size |
| Group sender name | Gold crown, 12px |
| Pinned avatar | Gold ring + tiny crown overlay at bottom-right |
| Contact profile | Gold pill badge: "👑 VIP Pro" with `#FEF3C7` background |
| About section | "VIP Pro Member" label with gold accent border |

### 13.7 Figma Component: Verification Badge

Create a component set in `Components / Badges`:

**Component Name:** `Badge / Verification`

| Variant | Properties |
|---|---|
| `Status=None` | Hidden (0x0, no content) |
| `Status=PhoneVerified, Size=Small` | `checkmark.seal.fill`, 14px, `#9CA3AF` |
| `Status=PhoneVerified, Size=Medium` | `checkmark.seal.fill`, 16px, `#9CA3AF` |
| `Status=PhoneVerified, Size=Large` | `checkmark.seal.fill`, 20px, `#9CA3AF` |
| `Status=IdentityLite, Size=Small` | `checkmark.seal.fill`, 14px, `#6366F1` |
| `Status=IdentityLite, Size=Medium` | `checkmark.seal.fill`, 16px, `#6366F1` |
| `Status=IdentityLite, Size=Large` | `checkmark.seal.fill`, 20px, `#6366F1` |
| `Status=IdentityFull, Size=Small` | `checkmark.seal.fill`, 14px, `#10B981` |
| `Status=IdentityFull, Size=Medium` | `checkmark.seal.fill`, 16px, `#10B981` |
| `Status=IdentityFull, Size=Large` | `checkmark.seal.fill`, 20px, `#10B981` |
| `Status=VipPro, Size=Small` | `crown.fill`, 14px, `#F59E0B` |
| `Status=VipPro, Size=Medium` | `crown.fill`, 16px, `#F59E0B` |
| `Status=VipPro, Size=Large` | `crown.fill`, 20px, `#F59E0B` |

**Component Name:** `Badge / VerificationPill`

| Variant | Properties |
|---|---|
| `Status=PhoneVerified` | Gray icon + "Phone Verified" text, `#F3F4F6` bg, 24px height |
| `Status=IdentityLite` | Indigo icon + "Verified" text, `#EEF2FF` bg, 24px height |
| `Status=IdentityFull` | Green icon + "Fully Verified" text, `#D1FAE5` bg, 24px height |
| `Status=VipPro` | Gold crown + "VIP Pro" text, `#FEF3C7` bg, 24px height |

---

## 14. Tab Bar Restructure & Rewards Tab

### 14.1 Updated Tab Bar

The Trust tab is removed from the bottom tab bar and replaced with Rewards. Trust moves to Settings > Privacy & Security.

```
BEFORE:
┌────────┬──────────┬────────┬─────────┐
│  💬    │   👥     │  🛡️    │   👤    │
│Messages│ Contacts │ Trust  │ Profile │
└────────┴──────────┴────────┴─────────┘

AFTER:
┌────────┬──────────┬────────┬─────────┐
│  💬    │   👥     │  ⭐    │   👤    │
│Messages│ Contacts │Rewards │ Profile │
└────────┴──────────┴────────┴─────────┘
```

### Tab Bar Specifications

| Tab | Icon (SF Symbol) | Label | Color (active) |
|---|---|---|---|
| Messages | `bubble.left.and.bubble.right.fill` | Messages | `primary` |
| Contacts | `person.2.fill` | Contacts | `primary` |
| Rewards | `star.circle.fill` | Rewards | `primary` |
| Profile | `person.crop.circle.fill` | Profile | `primary` |

| Element | Spec |
|---|---|
| Tab bar height | 49px (iOS standard) + safe area |
| Tab bar background | `surfaceElevated` with 1px `border` top stroke |
| Active icon | 24px, `primary` color |
| Inactive icon | 24px, `textTertiary` color |
| Label | 10px, Medium, matches icon color |
| Badge (unread) | Red dot (8px) on Messages icon for unread count |

### 14.2 Rewards Tab Screen

The Rewards Dashboard replaces the Trust tab as the third tab. It shows the user's ECHO token balance, earning multipliers, daily stats, quick actions, and recent earnings.

```
┌─────────────────────────────────┐
│         Rewards             📜  │  ← Large title, history icon
├─────────────────────────────────┤
│                                 │
│  ┌─────────────────────────────┐│
│  │  ╱ Gradient Card           ╲││
│  │                              ││
│  │    Available Balance         ││
│  │    1,247.50 ECHO             ││  ← 42px Bold white
│  │    ≈ $124.75 USD             ││  ← 16px white 70%
│  │                              ││
│  │   [Transfer] [Stake] [Convert]│  ← Frosted buttons
│  │                              ││
│  └─────────────────────────────┘│
│                                 │
│  ┌──────────┐ ┌──────────┐     │
│  │ 🛡️ Trust  │ │ 🔥 7-Day │     │  ← Multiplier badges
│  │   1.5x    │ │   1.2x   │     │
│  └──────────┘ └──────────┘     │
│                                 │
│  ┌──────┐ ┌──────┐ ┌──────┐   │
│  │  📅  │ │  📊  │ │  ⭐  │   │  ← Stat cards
│  │ 42.5 │ │ 892  │ │58/100│   │
│  │Today │ │Month │ │ Cap  │   │
│  └──────┘ └──────┘ └──────┘   │
│                                 │
│  QUICK ACTIONS                  │
│  ┌──────────┐ ┌──────────┐     │
│  │    🏆    │ │    💎    │     │
│  │Achieve-  │ │ Staking  │     │
│  │ments     │ │ 12% APY  │     │
│  └──────────┘ └──────────┘     │
│  ┌──────────┐ ┌──────────┐     │
│  │    👥    │ │    📜    │     │
│  │Referrals │ │ History  │     │
│  │50 ECHO ea│ │ View all │     │
│  └──────────┘ └──────────┘     │
│                                 │
│  RECENT EARNINGS      See All → │
│  ┌─────────────────────────────┐│
│  │ 💬 Messages Sent    +12.0  ││
│  │    24 messages today  2m   ││
│  ├─────────────────────────────┤│
│  │ ✓  Identity Verified +100  ││
│  │    Government ID      1h   ││
│  ├─────────────────────────────┤│
│  │ 👥 Referral Bonus    +50.0 ││
│  │    Sarah joined    Yester. ││
│  └─────────────────────────────┘│
│                                 │
├─────────────────────────────────┤
│  💬     👥      ⭐      👤     │
│Messages Contacts Rewards Profile│
└─────────────────────────────────┘
```

### Rewards Screen Specifications

| Element | Spec |
|---|---|
| Navigation | Large title "Rewards", history icon top-right |
| Balance card | Full width - 32px margin, `primaryGradient` fill, 24px corner radius, 24px internal padding |
| Balance label | 14px, white 90% opacity |
| Balance amount | 42px, Bold, white |
| Balance token | 20px, Medium, white 90%, 8px left of amount |
| USD value | 16px, white 70% opacity |
| Balance actions | Row of 3 frosted buttons (white 20% bg), 12px gap, 12px padding, 12px radius |
| Multiplier badges | Half-width cards, `surface` bg, 12px radius, icon (18px) + label (12px) + value (16px bold primary) |
| Stat cards | Third-width, `surface` bg, 16px radius, 16px padding, icon (24px) + value (20px bold) + label (12px) |
| Quick actions | 2x2 grid, half-width cards, `surface` bg + 1px `borderLight`, 16px radius, 20px padding, icon (32px) + label (14px bold) + sublabel (12px) |
| Earnings list | Full width rows, 16px horizontal padding, 14px vertical, 1px `borderLight` separator |
| Earning row | Icon (44x44, 12px radius, tinted bg) + title (15px medium) + subtitle (13px) + amount (16px bold `success`) + time (12px) |

### 14.3 Trust Moved to Privacy & Security

The existing Privacy & Security screen gains a new **Trust & Identity** section at the top, above the existing Profile Visibility section.

```
┌─────────────────────────────────┐
│ ←    Privacy & Security         │
├─────────────────────────────────┤
│                                 │
│  TRUST & IDENTITY               │  ← NEW SECTION (top of screen)
│                                 │
│  🛡️ Trust Score                  │
│     72/100 — Trusted        ›   │  ← Shows current score inline
│                                 │
│  🪪 Verify Identity              │
│     Government ID, Apple ID ›   │
│                                 │
│  📱 Devices                      │
│     2 devices registered    ›   │
│                                 │
│  📜 Login History                │
│     Last login: 2 min ago   ›   │
│                                 │
├─────────────────────────────────┤
│  PROFILE VISIBILITY             │  ← Existing section (unchanged)
│                                 │
│  Who can find me by username    │
│  Everyone                  ›    │
│                                 │
│  Who can see my online status   │
│  Contacts only             ›    │
│                                 │
│  Who can see my trust score     │
│  Everyone                  ›    │
│                                 │
├─────────────────────────────────┤
│  MESSAGING                      │  ← Existing (unchanged)
│  ...                            │
├─────────────────────────────────┤
│  CALLS                          │  ← Existing (unchanged)
│  ...                            │
├─────────────────────────────────┤
│  BLOCKCHAIN                     │  ← Existing (unchanged)
│  ...                            │
├─────────────────────────────────┤
│  SECURITY                       │  ← Existing (unchanged)
│  ...                            │
└─────────────────────────────────┘
```

### Trust & Identity Row Specifications

| Element | Spec |
|---|---|
| Section header | 12px, Semi-bold, `textTertiary`, uppercase, 1px tracking |
| Row height | 56px |
| Row icon | 20px emoji or SF Symbol, left-aligned, `primary` tint |
| Row title | 16px, Regular, `textPrimary` |
| Row subtitle | 14px, Regular, `textTertiary`, below title |
| Chevron | SF Symbol `chevron.right`, 14px, `textTertiary` |
| Trust Score row | Shows mini score badge: "72/100" in `primary` color + level label |
| Separator | 1px `borderLight`, inset 56px from left |

### Trust Score Row — Tap Behavior

Tapping "Trust Score" navigates to the **full Trust Dashboard** (existing screen, unchanged). The Trust Dashboard itself is identical to before — score circle, breakdown bars, verification status, trust actions. It simply lives one level deeper in navigation.

---

## 15. New Component Variants

Add these to the existing Figma component library:

### 15.1 Button Variants

| Variant Name | Properties |
|---|---|
| `Button / Biometric / FaceID` | Primary gradient fill, `faceid` icon (20px) left of text, 56px height |
| `Button / Biometric / TouchID` | Same as above with `touchid` icon |
| `Button / Secondary / Outlined` | 2px `primary` border, no fill, `primary` color text, 56px height — used for "Set Up Account" |
| `Button / Destructive / Filled` | `error` color fill, white text, 48px height |
| `Button / Destructive / Outlined` | `error` color border, `error` color text, transparent fill |

### 15.2 Input Variants

| Variant Name | Properties |
|---|---|
| `Input / OTP / 6-digit` | 6 boxes, 48x56px each, 8px gap, auto-advance behavior |
| `Input / OTP / Error` | Same layout with `error` border color and `errorLight` background |

### 15.3 Card Variants

| Variant Name | Properties |
|---|---|
| `Card / Device` | Icon (24px) + Name (16px bold) + Details (14px) + Location (14px) + Action button |
| `Card / Device / Current` | Same + green active dot indicator, no Remove button |
| `Card / Audit Log Entry` | Status icon (24px) + Event name (16px bold) + Details (14px) + Location + Time |

### 15.4 Overlay Variants

| Variant Name | Properties |
|---|---|
| `BottomSheet / StepUp` | 40% height, drag indicator, title, subtitle, biometric button, fallback link |
| `Alert / NewDevice` | Centered card, warning icon, device info, Verify button, "Not me" link |

### 15.5 Badge Variants

| Variant Name | Properties |
|---|---|
| `Badge / Session / Active` | Green dot (8px) + "Active now" (12px, `success`) |
| `Badge / Session / Inactive` | Gray dot (8px) + "2 hrs ago" (12px, `textTertiary`) |
| `Badge / Session / Revoked` | Red dot (8px) + "Revoked" (12px, `error`) |
| `Badge / Verification` | Checkmark or crown icon, 3 sizes (14/16/20px), 5 status variants — see Section 13.7 for full spec |
| `Badge / VerificationPill` | Icon + label in colored pill (24px height), 4 status variants — see Section 13.7 |

### 15.6 Onboarding Components (NEW)

| Variant Name | Properties |
|---|---|
| `PageIndicator / Carousel` | Horizontal stack of capsules. Active: 24x8px, `primary` fill. Inactive: 8x8px, `border` fill. 8px gap. |
| `OnboardingPage` | Illustration (240x240) + Headline (28px bold) + Subtitle (16px) + page indicator. Vertical auto-layout. |
| `Divider / WithLabel` | Two 1px `border` lines with centered label ("or"), 14px `textTertiary`. Used on Login screen. |
| `LoginScreen / Account=Existing` | Full Login layout with biometric icon, username, masked phone. |
| `LoginScreen / Account=None` | Full Login layout with generic icon, no account info. |

### 15.7 Tab Bar & Rewards Components (NEW)

| Variant Name | Properties |
|---|---|
| `TabBar / Default` | 4 tabs: Messages, Contacts, **Rewards** (`star.circle.fill`), Profile. Replaces old Trust tab. |
| `Card / Balance` | Gradient fill (`primaryGradient`), 24px radius. Balance (42px bold) + USD (16px) + 3 action buttons. |
| `Card / Multiplier` | Half-width, `surface` bg, 12px radius. Icon (18px) + label (12px) + value (16px bold primary). |
| `Card / Stat` | Third-width, `surface` bg, 16px radius. Icon (24px) + value (20px bold) + label (12px). |
| `Card / QuickAction` | Half-width, `surface` bg + 1px `borderLight`, 16px radius. Icon (32px) + label (14px) + sublabel (12px). |
| `Row / Earning` | Full-width. Icon (44px, tinted bg) + title + subtitle + amount (`success` bold) + time. |
| `Row / SettingsTrustIdentity` | Icon (20px) + title (16px) + subtitle (14px `textTertiary`) + chevron. For Privacy screen. |

---

## 16. Error States & Feedback

### Global Error Patterns

Every auth screen should have a designated error area that follows these rules:

| Rule | Spec |
|---|---|
| Position | Below the relevant input or action, 8px top spacing |
| Text | 14px, Regular, `error` color |
| Icon | Optional 16px `exclamationmark.circle` inline before text |
| Appearance | Fade in 200ms, slide down 4px |
| Disappearance | On next user input or after 5 seconds |
| Haptic | Error haptic (UINotificationFeedbackGenerator) |

### Error Messages (Design Text)

| Screen | Error | Display Text |
|---|---|---|
| Phone Entry | Invalid format | "Enter a valid phone number to continue." |
| Phone Entry | Rate limited | "Too many attempts. Try again in {time}." |
| OTP | Wrong code | "That code didn't work. Check and try again." |
| OTP | Expired | "Code expired. Tap Resend to get a new one." |
| OTP | Too many attempts | "Too many tries. Request a new code." |
| Login | Passkey failed | "Couldn't verify your identity. Try again." |
| Login | New device | "New device detected. We'll need to verify you." |
| Login | Account locked | (navigates to Account Locked screen) |

---

## 17. Animation Specifications

| Animation | Trigger | Duration | Easing | Details |
|---|---|---|---|---|
| Carousel page swipe | Swipe left/right | 300ms | spring(0.85) | Page slides, indicator capsule width animates |
| Carousel indicator | Page change | 200ms | ease-in-out | Active capsule expands 8→24px, previous shrinks 24→8px |
| Carousel → Login | "Get Started" or "Skip" | 300ms | ease-out | Slide from right (same as standard nav) |
| Screen transition | Navigate forward | 300ms | ease-out | Slide from right |
| Screen transition | Navigate back | 250ms | ease-in | Slide from left |
| Button press | Tap down | 100ms | ease-in | Scale to 0.98 |
| Button release | Tap up | 100ms | ease-out | Scale to 1.0 |
| OTP auto-advance | Digit entered | Instant | — | Focus jumps to next box |
| OTP error shake | Invalid code | 400ms | ease-in-out | 6px horizontal, 2 cycles |
| Login success | Auth complete | 200ms | ease-out | Checkmark scale from 0 → 1 |
| Progress spinner | Loading | Continuous | linear | Standard iOS activity indicator |
| Bottom sheet present | Step-up trigger | 300ms | spring(0.8) | Slide up from bottom |
| Alert appear | New device | 200ms | spring(0.9) | Scale from 0.9 → 1.0 + fade |
| Countdown timer | Every second | Instant | — | Text update only |
| Trusted contact confirm | Shard received | 300ms | spring(0.7) | ⏳ → ✅ with bounce |

### Reduce Motion

When the user has enabled "Reduce Motion" in system settings:
- Replace all slide transitions with cross-fade (200ms)
- Replace carousel swipe with instant page snap (no spring)
- Disable shake animation (show static red border instead)
- Disable spring animations (use linear)
- Keep countdown timer and progress spinner (functional, not decorative)

---

## 18. Theme Application

### Required Theme Variants

Every new screen must have **three Figma frames** — one per theme:

1. **Electric Violet** (default) — bold purple gradients, white backgrounds
2. **Teal + Warm Neutral** — teal accents, warm cream backgrounds
3. **Icy Minimal** — slate blue accents, cool gray backgrounds

### Using Figma Variables

All new screens must use Figma color variables, NOT hard-coded hex values. Reference the existing variable collections:

```
{Theme}/Primary/Default        → Button fills, active states
{Theme}/Primary/Gradient       → Primary CTA backgrounds
{Theme}/Background/Primary     → Screen backgrounds
{Theme}/Surface/Default        → Input backgrounds, cards
{Theme}/Text/Primary           → Headlines, input text
{Theme}/Text/Secondary         → Subtitles, descriptions
{Theme}/Text/Tertiary          → Hints, timestamps
{Theme}/Semantic/Error         → Error text, destructive buttons
{Theme}/Semantic/Success       → Active indicators, confirmations
{Theme}/Semantic/Warning       → Alerts, caution states
{Theme}/Border/Default         → Input borders, dividers
```

### Auto-Layout Requirements

All screens must use Figma Auto-Layout for responsive behavior:

- Screen frame: Fill container, vertical auto-layout, 24px horizontal padding
- Button groups: Vertical auto-layout, 12px gap
- Input groups: Vertical auto-layout, 8px gap
- Cards in lists: Vertical auto-layout, 12px gap

---

## 19. Accessibility Checklist

Before marking any auth screen as "Design Complete" in Figma, verify:

| Check | Requirement | Tool |
|---|---|---|
| Color contrast (text) | 4.5:1 minimum against background | Figma A11y plugin or Stark |
| Color contrast (large text) | 3:1 minimum for text ≥ 24px bold | Same |
| Color contrast (non-text) | 3:1 for icons, borders, focus indicators | Same |
| Tap targets | 44x44pt minimum on all interactive elements | Manual measurement |
| Focus order | Logical tab order annotated on each screen | Figma annotations |
| Screen reader labels | Every interactive element has a text annotation for VoiceOver | Figma annotations |
| Error identification | Errors identified by more than just color (icon + text) | Visual check |
| Text scalability | Layouts work at 200% Dynamic Type | Manual check in prototype |
| Motion | No auto-playing animations except functional (timers) | Visual check |

### Annotation Convention

Use a dedicated Figma layer called "♿ Accessibility" (hidden by default) on each screen with:
- Blue dotted outlines showing tap target areas
- Numbered labels showing tab order
- VoiceOver text annotations next to each element

---

*Last Updated: March 2026*
*Design System: Echo v2.0*
*Figma File: Echo Design System*
