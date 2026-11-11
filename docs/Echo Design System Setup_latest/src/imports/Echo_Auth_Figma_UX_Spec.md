# Echo Authentication — Figma UX Design Spec

> **Version:** 2.0 · **Date:** March 2026 · **Status:** Ready for Design
> **Figma File:** Echo Design System · **Pages:** Screens / Onboarding, Components / Overlays, Screens / Settings

---

## Table of Contents

1. [Overview](#1-overview)
2. [New Screens Summary](#2-new-screens-summary)
3. [Updated Welcome Screen](#3-updated-welcome-screen)
4. [Login Screen (Returning User)](#4-login-screen-returning-user)
5. [OTP Verification Screen (Standardized)](#5-otp-verification-screen-standardized)
6. [Passkey Setup Screen (Standardized)](#6-passkey-setup-screen-standardized)
7. [Step-Up Verification Modal](#7-step-up-verification-modal)
8. [Device Management Screen](#8-device-management-screen)
9. [New Device Detected Alert](#9-new-device-detected-alert)
10. [Account Recovery Flow](#10-account-recovery-flow)
11. [Account Locked Screen](#11-account-locked-screen)
12. [Login Audit Log Screen](#12-login-audit-log-screen)
13. [New Component Variants](#13-new-component-variants)
14. [Error States & Feedback](#14-error-states--feedback)
15. [Animation Specifications](#15-animation-specifications)
16. [Theme Application](#16-theme-application)
17. [Accessibility Checklist](#17-accessibility-checklist)

---

## 1. Overview

The current Echo Figma design system has comprehensive coverage for onboarding (Welcome, Phone Entry, OTP, Passkey, Profile, Trust Intro) and the alternative wallet flow. However, the **returning user login experience** and several security-related screens are completely missing.

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
| 1 | Welcome (updated) | Screens / Onboarding | UPDATE — promote Sign In link | P0 |
| 2 | Login (Returning User) | Screens / Onboarding | **NEW** | P0 |
| 3 | OTP Verification (standardized) | Screens / Onboarding | UPDATE — standardize component | P0 |
| 4 | Passkey Setup (standardized) | Screens / Onboarding | UPDATE — add Touch ID variant | P0 |
| 5 | Step-Up Verification Modal | Components / Overlays | **NEW** | P0 |
| 6 | Device Management | Screens / Settings | **NEW** | P1 |
| 7 | New Device Detected Alert | Components / Overlays | **NEW** | P1 |
| 8 | Account Recovery — Method Selection | Screens / Onboarding | **NEW** | P1 |
| 9 | Account Recovery — Trusted Contacts | Screens / Onboarding | **NEW** | P1 |
| 10 | Account Recovery — Waiting | Screens / Onboarding | **NEW** | P1 |
| 11 | Account Locked | Screens / Onboarding | **NEW** | P2 |
| 12 | Login Audit Log | Screens / Settings | **NEW** | P2 |

**Total: 7 new screens, 5 screen updates, 7 new component variants**

---

## 3. Updated Welcome Screen

### Current vs. Updated

The existing Welcome screen has "Already have account? Sign In" as small link text at the bottom. This needs to be promoted to a visible tertiary button since returning users will encounter this screen frequently after app reinstalls or new device setups.

### Layout (Updated)

```
┌─────────────────────────────────┐
│                                 │
│                                 │
│         [Echo Logo]             │
│          120x120px              │
│                                 │
│     "Messaging with Trust"      │
│                                 │
│   ┌─────────────────────────┐   │
│   │   Continue with Phone   │   │ ← Primary CTA (filled)
│   └─────────────────────────┘   │
│                                 │
│   ┌─────────────────────────┐   │
│   │  Use Verifiable ID      │   │ ← Secondary (outlined)
│   └─────────────────────────┘   │
│                                 │
│   ┌─────────────────────────┐   │  ← NEW: Promoted from link
│   │       Sign In           │   │     to tertiary button
│   └─────────────────────────┘   │
│                                 │
│  ─────────────────────────────  │
│                                 │
│   🔒 Your identity stays yours  │
│      No data sold. Ever.        │
│                                 │
└─────────────────────────────────┘
```

### Changes

| Element | Before | After |
|---|---|---|
| Sign In | 14px link text, `#6B7280` | Full-width tertiary button, 48px height, 16px `#7C3AED` text, no fill, no border |
| Sign In position | Below secondary button, minimal spacing | Below secondary button with 12px gap |
| Button group spacing | 16px gap between primary and secondary | 12px gap between all three buttons |

### Specifications

| Element | Spec |
|---|---|
| Tertiary Button (Sign In) | Full width, 48px height, no background, no border |
| Tertiary Button Text | 16px, Semi-bold, Primary color (theme-dependent) |
| Tertiary Button Tap State | Text opacity 0.6, 100ms |

---

## 4. Login Screen (Returning User)

This is the most critical new screen. It appears when a returning user taps "Sign In" from Welcome, or when the app detects an existing session that needs biometric re-verification.

### Layout

```
┌─────────────────────────────────┐
│ ←                               │
│                                 │
│                                 │
│         [Echo Logo]             │
│          64x64px (smaller)      │
│                                 │
│         Welcome back            │
│                                 │
│                                 │
│   ┌─────────────────────────┐   │
│   │  [FaceID]  Sign In      │   │ ← Primary CTA (gradient fill)
│   └─────────────────────────┘   │
│                                 │
│     Signed in as @alex_echo     │  ← 14px, textSecondary
│     +1 (555) ***-4567           │  ← 14px, textTertiary, masked
│                                 │
│                                 │
│   Not you? [Switch account]     │  ← 14px, primary color link
│                                 │
│   ─────────────────────────     │
│                                 │
│   Lost access? [Recover]        │  ← 14px, textSecondary link
│                                 │
└─────────────────────────────────┘
```

### Specifications

| Element | Spec |
|---|---|
| Back arrow | 24px, tap target 44x44, navigates to Welcome |
| Logo | 64x64px (half of Welcome size), centered, same asset |
| Title | 28px, Bold, textPrimary, "Welcome back" |
| Sign In Button | Full width, 56px height, `primaryGradient` fill, 12px corner radius |
| Face ID icon | SF Symbol `faceid`, 20px, textOnPrimary, 12px left of text |
| Touch ID variant | SF Symbol `touchid`, same specs (auto-detect device capability) |
| Account info | Centered, 14px, `textSecondary` for username, `textTertiary` for phone |
| Phone masking | Show country code + first digit + asterisks + last 4 digits: `+1 (5**) ***-4567` |
| Switch account link | 14px, Semi-bold, primary color, tap → navigate to Welcome |
| Divider | 1px, `border` color, full width with 24px horizontal padding |
| Recovery link | 14px, Regular, `textSecondary`, tap → Account Recovery flow |

### Interaction States

| State | Visual |
|---|---|
| Default | Sign In button with gradient, Face ID icon |
| Authenticating | Button shows centered `ProgressView` (spinner), gradient dims to 80% opacity |
| Success | Brief checkmark animation (200ms) → transition to Home |
| Failed | Button returns to default, error text appears below account info (14px, `error` color) |
| Account locked | Entire screen transitions to Account Locked screen |

### Auto-Login Behavior

If the app launches and detects a stored refresh token + valid biometric state, it should:
1. Show this Login screen briefly (200ms)
2. Automatically trigger the Face ID prompt
3. On success → transition to Home with no user interaction required

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

## 13. New Component Variants

Add these to the existing Figma component library:

### 13.1 Button Variants

| Variant Name | Properties |
|---|---|
| `Button / Biometric / FaceID` | Primary gradient fill, `faceid` icon (20px) left of text, 56px height |
| `Button / Biometric / TouchID` | Same as above with `touchid` icon |
| `Button / Destructive / Filled` | `error` color fill, white text, 48px height |
| `Button / Destructive / Outlined` | `error` color border, `error` color text, transparent fill |
| `Button / Tertiary` | No fill, no border, primary color text, 48px height |

### 13.2 Input Variants

| Variant Name | Properties |
|---|---|
| `Input / OTP / 6-digit` | 6 boxes, 48x56px each, 8px gap, auto-advance behavior |
| `Input / OTP / Error` | Same layout with `error` border color and `errorLight` background |

### 13.3 Card Variants

| Variant Name | Properties |
|---|---|
| `Card / Device` | Icon (24px) + Name (16px bold) + Details (14px) + Location (14px) + Action button |
| `Card / Device / Current` | Same + green active dot indicator, no Remove button |
| `Card / Audit Log Entry` | Status icon (24px) + Event name (16px bold) + Details (14px) + Location + Time |

### 13.4 Overlay Variants

| Variant Name | Properties |
|---|---|
| `BottomSheet / StepUp` | 40% height, drag indicator, title, subtitle, biometric button, fallback link |
| `Alert / NewDevice` | Centered card, warning icon, device info, Verify button, "Not me" link |

### 13.5 Badge Variants

| Variant Name | Properties |
|---|---|
| `Badge / Session / Active` | Green dot (8px) + "Active now" (12px, `success`) |
| `Badge / Session / Inactive` | Gray dot (8px) + "2 hrs ago" (12px, `textTertiary`) |
| `Badge / Session / Revoked` | Red dot (8px) + "Revoked" (12px, `error`) |

---

## 14. Error States & Feedback

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

## 15. Animation Specifications

| Animation | Trigger | Duration | Easing | Details |
|---|---|---|---|---|
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
- Disable shake animation (show static red border instead)
- Disable spring animations (use linear)
- Keep countdown timer and progress spinner (functional, not decorative)

---

## 16. Theme Application

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

## 17. Accessibility Checklist

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
