# ECHO — Figma Step-By-Step Implementation Guide
## Google Stitch Compatible | v3.2 | April 2026

---

## Table of Contents

1. [Design Audit Summary](#1-design-audit-summary)
2. [Design System Foundation](#2-design-system-foundation)
3. [Screen Inventory — Current vs Required](#3-screen-inventory)
4. [Figma Page & Frame Structure](#4-figma-page--frame-structure)
5. [Component Library Setup](#5-component-library-setup)
6. [Screen-by-Screen Implementation](#6-screen-by-screen-implementation)
7. [Missing Pages — Full Specifications](#7-missing-pages)
8. [Design Fixes for Existing Pages](#8-design-fixes-for-existing-pages)
9. [Interaction & Prototype Flows](#9-interaction--prototype-flows)
10. [Google Stitch Import Instructions](#10-google-stitch-import-instructions)
11. [Handoff Checklist](#11-handoff-checklist)

---

## 1. Design Audit Summary

### Current State (23 app screens + 15 onboarding screens)

| Category | Screens Built | Status |
|----------|--------------|--------|
| Onboarding | 15 screens | Complete — Intro, Welcome, Login, Phone, OTP, Passkey, Recovery, Profile, Verify Credentials, Trust Dashboard, Recover, Recover Phrase, Recover Contacts, Recover Complete, Account Locked |
| Messaging | 3 screens | Messages list, Chat view, Group Chat |
| Contacts | 1 screen | Contact list with trust tiers |
| Groups | 4 screens | Group Create, Group Detail, Group Discover, Group Chat |
| Channels | 3 screens | Channel Create, Channel Detail, Channel Discover |
| Personas | 3 screens | Persona List, Persona Create, Persona Detail |
| Hidden Folders | 3 screens | Hidden Folders list, Folder Detail, Folder Chat |
| Trust | 1 screen | Trust Dashboard (1474 lines — comprehensive) |
| Profile | 1 screen | Basic profile (126 lines — needs expansion) |
| Settings | 1 screen | Settings list |
| Rewards | 1 screen | Rewards page (should become Wallet) |
| Security | 2 screens | Device Management, Login Audit |
| Analytics | 1 screen | Onboarding funnel metrics |

### Critical Missing Screens (12 new pages required)

| Priority | Screen | PRD Feature Reference |
|----------|--------|----------------------|
| P0 | **Wallet Tab** | Tokenomics v2 §4.2 — replaces Rewards tab |
| P0 | **Voice/Video Call** | PRD §Voice and Video Calls |
| P0 | **Contact Detail** | PRD §Dynamic Trust Network |
| P1 | **Advanced Search** | PRD §Advanced Message Search |
| P1 | **Media Gallery** | PRD §Large File Sharing |
| P1 | **Notification Center** | Foundation Blueprint §Notification Badges |
| P1 | **QR Identity Share** | PRD §Decentralized Identity |
| P2 | **Governance/Voting** | Tokenomics v2 §3.3, IOS_IMPL §7 |
| P2 | **Backup & Export** | PRD §Recovery, IOS_IMPL §Persistence |
| P2 | **Enterprise Profile** | PRD §Enterprise Organization Profiles |
| P3 | **Bot Management** | PRD §Decentralized Bot Framework |
| P3 | **Staking/Delegation Detail** | Tokenomics v2 §4.3 |

### Design Violations Found

| Issue | Location | Fix |
|-------|----------|-----|
| `1px solid border-top` on tab bar | `tab-bar.tsx` line 22 | Replace with tonal shift or ghost border at 15% opacity |
| Profile page too thin (126 lines) | `profile.tsx` | Add DID badge, credential cards, wallet summary, QR share |
| Tab bar has 4 tabs instead of 3 | `tab-bar.tsx` | Change to: Messages, Wallet, Me |
| Some cards use flat backgrounds | Multiple pages | Add frosted glass layer with `backdrop-filter: blur(20px)` |
| Missing Secure Thread Indicator | All authenticated screens | Add 2px sky-blue pulsating line at top |
| Some pages use `#000000` pure black | Dark mode tokens | Replace with `#0F172A` (Deep Navy) or `#131B2E` (On Surface) |
| Hard divider lines in settings | `settings.tsx` | Replace with tonal shifts between surface tiers |

---

## 2. Design System Foundation

### The Glacial Interface — Token Reference

#### Color Tokens

```
Primary Palette:
  --echo-primary:           #006591    (Core brand — high-action states)
  --echo-primary-container: #0EA5E9    (Sky Blue — signature accent)
  --echo-deep-navy:         #0F172A    (Gradient start, dark backgrounds)
  --echo-on-surface:        #131B2E    (Text, never pure black)

Surface Tiers (Light Mode):
  --echo-surface:                  #FAF8FF    (Base layer)
  --echo-surface-container-lowest: #FFFFFF    (Action cards)
  --echo-surface-container-low:    #F1F5F9    (Grouped content)
  --echo-surface-container:        #E2E8F0    (Secondary panels)
  --echo-surface-container-high:   #C8D6E5    (Elevated content)
  --echo-surface-container-highest:#DAE2FD    (Inbound message bubbles)

Surface Tiers (Dark Mode):
  --echo-surface:                  #0F172A    (Base layer)
  --echo-surface-container-lowest: #1E293B    (Action cards)
  --echo-surface-container-low:    #1E293B    (Grouped content)
  --echo-surface-container:        #334155    (Secondary panels)
  --echo-surface-container-high:   #475569    (Elevated content)

Semantic:
  --echo-success: #10B981
  --echo-warning: #F59E0B
  --echo-error:   #EF4444

Trust Tier Colors:
  Inner Circle: Sky Blue (#0EA5E9)
  Trusted:      Emerald  (#10B981)
  Verified:     Amber    (#F59E0B)
  Peer:         Slate    (#64748B)
  Basic:        Grey     (#94A3B8)
```

#### Typography — Inter Font (Editorial Hierarchy)

```
Display-LG:   56px / Bold / -1.5% tracking   → "Hello" moments, empty states
Headline-SM:  24px / Bold / -0.5% tracking   → Screen titles
Title-MD:     20px / SemiBold / 0% tracking   → Section headers
Body-LG:      16px / Regular / 0% tracking    → Chat messages, body text
Body-MD:      14px / Regular / 0% tracking    → Secondary content
Label-MD:     12px / Medium / +2% tracking    → Timestamps, metadata
Label-SM:     10px / Bold / +10% tracking     → Micro labels, badges
```

#### Spacing System

```
spacing-1:  4px     (Inline padding)
spacing-2:  8px     (Tight grouping)
spacing-3:  12px    (Message bubble gap, same speaker)
spacing-4:  16px    (Standard padding)
spacing-5:  20px    (Section padding)
spacing-6:  24px    (Message gap, different speakers)
spacing-8:  32px    (Card padding)
spacing-10: 40px    (Major section breaks)
```

#### Corner Radius

```
radius-sm:   8px     (Inner elements, bottom-left of inbound bubbles)
radius-md:   16px    (Input fields, small cards)
radius-lg:   24px    (Message bubbles — all corners except speech pointer)
radius-xl:   32px    (Cards, buttons, containers — standard)
radius-full: 9999px  (Pills, avatars, primary CTAs)
```

#### Elevation & Shadows (Never Grey)

```
Ambient Shadow:   color: on-surface at 4%, blur: 24px, y: 4px
Card Shadow:      color: on-surface at 6%, blur: 32px, y: 8px
Deep Glacial:     color: deep-navy at 15%, blur: 32px, y: 12px   → Primary CTAs only
Sky Glow:         color: sky-blue at 25%, blur: 16px, y: 0px     → Active trust indicators
```

#### Animation Specification

```
Spring Standard:   mass: 1.0, stiffness: 300, damping: 0.85    → Page transitions, cards
Spring Gentle:     mass: 1.0, stiffness: 200, damping: 0.9     → Sheets, modals
Spring Snappy:     mass: 0.8, stiffness: 400, damping: 0.8     → Button press, toggles
Fade In:           duration: 200ms, ease: ease-out              → Content appearance
Pulse:             opacity 0.6→1.0, duration: 2s, repeat        → Secure Thread Indicator
```

---

## 3. Screen Inventory

### Complete App Map (35 screens total — 23 existing + 12 new)

```
ECHO App
├── Onboarding (15 screens) ✅ Complete
│   ├── Intro Carousel
│   ├── Welcome
│   ├── Login (Passkey + SMS)
│   ├── Phone Entry
│   ├── OTP Verify
│   ├── Passkey Setup
│   ├── Recovery Phrase
│   ├── Profile Setup
│   ├── Verify Credentials
│   ├── Trust Dashboard
│   ├── Recover (method selection)
│   ├── Recover Phrase
│   ├── Recover Contacts
│   ├── Recover Complete
│   └── Account Locked
│
├── Tab: Messages (8 screens)
│   ├── Messages List ✅
│   ├── Chat View ✅
│   ├── Group Chat ✅
│   ├── Voice/Video Call ❌ NEW
│   ├── Advanced Search ❌ NEW
│   ├── Media Gallery ❌ NEW
│   ├── Notification Center ❌ NEW
│   └── QR Identity Share ❌ NEW
│
├── Tab: Wallet (3 screens)
│   ├── Wallet Dashboard ❌ NEW (replaces Rewards)
│   ├── Staking/Delegation Detail ❌ NEW
│   └── Governance/Voting ❌ NEW
│
├── Tab: Me (4 screens)
│   ├── Profile ✅ (needs expansion)
│   ├── Contact Detail ❌ NEW
│   ├── Settings ✅
│   └── Backup & Export ❌ NEW
│
├── Feature Pages (accessed from Messages/Settings)
│   ├── Contacts ✅
│   ├── Groups (Create, Detail, Discover) ✅
│   ├── Channels (Create, Detail, Discover) ✅
│   ├── Personas (List, Create, Detail) ✅
│   ├── Hidden Folders (List, Detail, Chat) ✅
│   ├── Trust Dashboard ✅
│   ├── Device Management ✅
│   ├── Login Audit ✅
│   ├── Analytics ✅
│   ├── Enterprise Profile ❌ NEW
│   └── Bot Management ❌ NEW
```

---

## 4. Figma Page & Frame Structure

### Recommended Figma File Organization

```
Figma File: "ECHO iOS App v3.2"
│
├── 📄 Page: "Cover"
│   └── Frame: Cover card with app name, version, date
│
├── 📄 Page: "Design System"
│   ├── Frame: Color Tokens (light + dark swatches)
│   ├── Frame: Typography Scale (all 7 levels)
│   ├── Frame: Spacing & Grid (8px base grid)
│   ├── Frame: Corner Radius reference
│   ├── Frame: Shadow & Elevation reference
│   └── Frame: Animation specs (annotated)
│
├── 📄 Page: "Components"
│   ├── Frame: Navigation (Tab Bar, Nav Bar, Secure Thread Indicator)
│   ├── Frame: Buttons (Signature Gradient, Glass Secondary, Ghost)
│   ├── Frame: Cards (Ghost Border Card, Trust Card, Wallet Card)
│   ├── Frame: Input Fields (Text, Search, OTP, Phone)
│   ├── Frame: Message Bubbles (Inbound, Outbound, System, Scheduled, Disappearing)
│   ├── Frame: Avatars (Trust Ring, Online Dot, Persona Badge)
│   ├── Frame: Badges (Trust Tier, Verification, Unread Count)
│   ├── Frame: Sheets & Modals (Bottom Sheet, Biometric Auth, Alert)
│   ├── Frame: Lists (Conversation Item, Contact Item, Settings Row)
│   └── Frame: Special (QR Code, Call Controls, Wallet Actions)
│
├── 📄 Page: "Onboarding Flow"
│   └── 15 frames at 393×852 (iPhone 15 Pro)
│
├── 📄 Page: "Messages Tab"
│   ├── Frame: Messages List (with folder filters)
│   ├── Frame: Chat View (1:1)
│   ├── Frame: Group Chat
│   ├── Frame: Voice Call (active)
│   ├── Frame: Video Call (active)
│   ├── Frame: Advanced Search
│   ├── Frame: Media Gallery
│   ├── Frame: Notification Center
│   └── Frame: QR Identity Share
│
├── 📄 Page: "Wallet Tab"
│   ├── Frame: Wallet Dashboard
│   ├── Frame: Staking Detail
│   ├── Frame: Delegation Detail
│   └── Frame: Governance Voting
│
├── 📄 Page: "Me Tab"
│   ├── Frame: Profile (expanded)
│   ├── Frame: Contact Detail
│   ├── Frame: Settings
│   └── Frame: Backup & Export
│
├── 📄 Page: "Feature Screens"
│   ├── Frame: Contacts
│   ├── Frame: Groups (Create, Detail, Discover)
│   ├── Frame: Channels (Create, Detail, Discover)
│   ├── Frame: Personas (List, Create, Detail)
│   ├── Frame: Hidden Folders (List, Detail, Chat)
│   ├── Frame: Trust Dashboard
│   ├── Frame: Device Management
│   ├── Frame: Login Audit
│   ├── Frame: Enterprise Profile
│   └── Frame: Bot Management
│
├── 📄 Page: "Dark Mode"
│   └── All key screens duplicated in dark mode
│
└── 📄 Page: "Prototype Flows"
    ├── Flow: Onboarding → Home
    ├── Flow: Messages → Chat → Call
    ├── Flow: Wallet → Stake → Confirm
    └── Flow: Profile → Settings → Hidden Folders
```

### Frame Dimensions

```
Device:          iPhone 15 Pro
Frame Size:      393 × 852 px
Status Bar:      54px (included in frame)
Safe Area Top:   59px
Safe Area Bottom: 34px (home indicator)
Tab Bar Height:  82px (including bottom safe area)
Nav Bar Height:  44px (below status bar)
Content Area:    852 - 59 - 82 = 711px usable height
Grid:            8px base grid, 20px horizontal margins
```

---

## 5. Component Library Setup

### Step 1: Create Master Components

For each component below, create a Figma component with variants for state (default, hover, active, disabled, dark mode) and any configuration options.

#### 5.1 Tab Bar Component

```
Component Name: "TabBar"
Size: 393 × 82px
Background: surface-container-lowest (NO border-top — use tonal shift)
Bottom Padding: 20px (safe area)

Variants:
  - Active Tab: Messages | Wallet | Me
  
Tab Items (3 tabs):
  Messages:  MessageCircle icon (24px) + "Messages" label (10px)
  Wallet:    Wallet icon (24px) + "Wallet" label (10px)
  Me:        User icon (24px) + "Me" label (10px)

Active State:
  Icon: sky-blue (#0EA5E9), strokeWidth 2.2
  Label: sky-blue, fontWeight 600
  
Inactive State:
  Icon: text-tertiary (#64748B), strokeWidth 1.8
  Label: text-tertiary, fontWeight 500

Unread Badge (Messages tab):
  Position: top-right of icon, offset (-1.5px, -0.5px)
  Size: 8×8px circle
  Color: error (#EF4444)
```

#### 5.2 Secure Thread Indicator

```
Component Name: "SecureThreadIndicator"
Size: 393 × 2px
Position: Top of screen, below status bar
Color: primary-container (#0EA5E9)
Animation: opacity pulses 0.6 → 1.0 on 2-second cycle
Purpose: Indicates active encrypted connection
Present on: ALL authenticated screens
```

#### 5.3 Glacial Navigation Bar

```
Component Name: "GlacialNavBar"
Size: 393 × 44px (below Secure Thread)
Background: surface-variant at 60% opacity + backdrop-blur(20px)
Bottom Edge: ghost border (outline-variant at 15% opacity)
Shadow: ambient (on-surface 4%, blur 32px)

Variants:
  - Title Only: centered headline-sm text
  - Title + Back: left arrow + title
  - Title + Actions: title + right-side icon buttons
  - Messages Header: "ECHO" wordmark left + filter/compose right
```

#### 5.4 Signature Gradient Button

```
Component Name: "SignatureGradientButton"
Size: width auto / height 56px
Border Radius: 9999px (full pill)
Background: linear-gradient(135deg, #0F172A, #0EA5E9)
Shadow: deep glacial (navy 15%, blur 32px)
Text: white, 18px Inter Bold
Press State: scale(0.95) with spring animation

Variants:
  - Full Width (stretch)
  - Compact (hug content)
  - With Icon (left icon + text)
  - Loading (spinner replaces text)
  - Disabled (opacity 0.5, no shadow)
```

#### 5.5 Ghost Border Card

```
Component Name: "GhostBorderCard"
Size: auto
Background: surface-container-low
Border: outline-variant at 15% opacity (1px)
Border Radius: 32px
Padding: 24px
Shadow: ambient (on-surface 4%, blur 24px)

Variants:
  - Standard
  - Elevated (surface-container-lowest background)
  - Glass (60% opacity + backdrop blur)
  - Interactive (hover state brightens to surface-bright)
```

#### 5.6 Message Bubble

```
Component Name: "MessageBubble"
Max Width: 280px (71% of screen width)

Inbound Variant:
  Background: surface-container-highest
  Border Radius: 24px all sides, 8px bottom-left
  Text: on-surface, 16px Inter Regular
  
Outbound Variant:
  Background: signature gradient (deep navy → sky blue)
  Border Radius: 24px all sides, 8px bottom-right
  Text: white, 16px Inter Regular

States:
  - Default
  - With Reply Preview
  - With File Attachment
  - Scheduled (clock icon + scheduled time label)
  - Disappearing (timer icon + countdown)
  - Silent (mute icon badge)
  - Recalled (italic "This message was recalled" + slash icon)
  - Failed (red exclamation + "Retry" tap target)
  - Edited (pencil icon + "edited" label)

Spacing:
  Same speaker gap: 12px
  Different speaker gap: 24px
```

#### 5.7 Trust Ring Avatar

```
Component Name: "TrustRingAvatar"
Sizes: 40px (list), 56px (chat header), 80px (profile card), 140px (profile hero)

Structure:
  Outer Ring: 3px stroke, color based on trust tier
  Inner Image: circular crop, 4px white border inset
  
Trust Tier Ring Colors:
  Inner Circle: sky-blue (#0EA5E9) + glow shadow
  Trusted:      emerald (#10B981)
  Verified:     amber (#F59E0B)
  Peer:         slate (#64748B)
  Basic:        light grey (#CBD5E1)

Online Indicator:
  Position: bottom-right
  Size: 12px circle (10px at 40px avatar size)
  Color: success (#10B981)
  Border: 2px white

Verification Badge (optional):
  Position: bottom-right (replaces online dot)
  Size: 16px
  Types: checkmark-circle (verified), shield (VIP), building (enterprise)
```

#### 5.8 Conversation List Item

```
Component Name: "ConversationItem"
Size: 393 × 80px
Padding: 16px horizontal, 12px vertical

Layout:
  [Avatar 52px] [16px gap] [Name + Message stack] [Time + Badge]
  
Name: 16px Inter SemiBold, on-surface
Last Message: 14px Inter Regular, text-secondary (single line, ellipsis)
Time: 12px Inter Medium, text-tertiary, +2% tracking
Unread Badge: 20px circle, sky-blue fill, white text, 11px bold

States:
  - Default
  - Unread (bold name, badge visible)
  - Muted (mute icon next to time)
  - Pinned (pin icon overlay)
  - Disappearing (timer icon)
  - Silent (volume-off icon)
  - Group (stacked avatar)
  - Channel (radio icon avatar)
```

#### 5.9 Wallet Balance Card

```
Component Name: "WalletBalanceCard"
Size: 353 × 180px (full width minus margins)
Background: signature gradient (deep navy → sky blue)
Border Radius: 32px
Shadow: deep glacial
Padding: 24px

Content:
  "Total Balance" — 12px Label-MD, white 70%
  "24,830.00 ECHO" — 32px Display weight, white 100%
  "≈ $2,483.00 USD" — 14px Body-MD, white 60%
  "▲ 3.2% (24h)" — 12px Label-MD, success green on white-15% pill
```

#### 5.10 Wallet Action Button

```
Component Name: "WalletActionButton"
Size: 80 × 72px
Background: surface-container-low
Border: ghost border 15%
Border Radius: 20px

Content:
  Icon: 24px, sky-blue
  Label: 11px Label-SM, text-secondary
  
States: Default, Pressed (scale 0.95), Disabled (opacity 0.4)
Actions: Stake, Delegate, Swap, Bridge
```

---

## 6. Screen-by-Screen Implementation

### Implementation Order (by priority and dependency)

```
Phase 1 — Foundation (Week 1-2):
  1. Design System page (all tokens, components)
  2. Tab Bar update (3 tabs: Messages, Wallet, Me)
  3. Secure Thread Indicator (add to all screens)
  4. Fix existing design violations

Phase 2 — Core Missing Screens (Week 2-4):
  5. Wallet Dashboard (replaces Rewards)
  6. Contact Detail
  7. Voice/Video Call
  8. Advanced Search
  9. Profile expansion

Phase 3 — Secondary Screens (Week 4-6):
  10. Media Gallery
  11. Notification Center
  12. QR Identity Share
  13. Governance/Voting
  14. Staking/Delegation Detail

Phase 4 — Polish (Week 6-7):
  15. Backup & Export
  16. Enterprise Profile
  17. Bot Management
  18. Dark mode pass for all new screens
  19. Prototype flows and interactions
```

---

## 7. Missing Pages — Full Specifications

### 7.1 Wallet Dashboard (P0 — replaces Rewards tab)

**Route:** `/wallet`  
**Tab Position:** Center tab (2 of 3)  
**Source:** Tokenomics v2 §4.2, IOS_IMPL v4.2 §Wallet

```
┌─────────────────────────────────────────────┐
│ ▬▬▬▬▬▬▬ Secure Thread (2px sky) ▬▬▬▬▬▬▬▬▬ │
│                                              │
│  ECHO Wallet                          ⚙️    │
│                                              │
│  ┌─────────────────────────────────────┐    │
│  │  ◆ Signature Gradient Card          │    │
│  │                                      │    │
│  │  Total Balance                       │    │
│  │  24,830.00 ECHO                      │    │
│  │  ≈ $2,483.00 USD                    │    │
│  │  ▲ 3.2% (24h)                       │    │
│  └─────────────────────────────────────┘    │
│                                              │
│  ┌──────────────────────────────────────┐   │
│  │ Available    12,450.00               │   │
│  │ Staked        8,000.00  🔒 Gold      │   │
│  │ Delegated     Validator #7  →        │   │
│  │ Pending       4,380.00  [Claim All]  │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  ┌────────┬────────┬────────┬────────┐      │
│  │ 🔒     │ ↗      │ ⇄      │ 🔗     │      │
│  │ Stake  │Delegate│  Swap  │ Bridge │      │
│  └────────┴────────┴────────┴────────┘      │
│                                              │
│  TODAY'S REWARDS                             │
│  ┌──────────────────────────────────────┐   │
│  │ 💬 Messaging    4.2 / 50.0 ECHO     │   │
│  │ 🤝 Referrals    0.0 / 50.0 ECHO     │   │
│  │ 📊 Staking     12.8 ECHO (auto)     │   │
│  │ ░░░░░░░█████░░░░░░ 34% of daily cap │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  RECENT ACTIVITY                             │
│  ┌──────────────────────────────────────┐   │
│  │ ↓ +2.1 ECHO   Messaging     2m ago  │   │
│  │ ↓ +12.8 ECHO  Staking       6h ago  │   │
│  │ ↑ -500 ECHO   Staked Gold   2d ago  │   │
│  │ ↓ +50 ECHO    Referral      5d ago  │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  [ Messages ]  [ ★ Wallet ]  [  Me  ]       │
└─────────────────────────────────────────────┘
```

**Component Details:**
- Balance Card: signature gradient fill, 32px radius, deep glacial shadow
- Balance Breakdown: ghost border card, surface-container-low background
- Action Buttons: 4-column grid, 80×72px each, 20px radius
- Daily Rewards: progress bars with sky-blue fill, ghost border card
- Recent Activity: list items with directional arrows (green ↓ for received, red ↑ for sent)
- "Claim All" button: compact signature gradient pill

**Founder Vesting Section (conditional — only if DID has TokenLock):**
- Appears below Daily Rewards
- Shows: Role, Allocated, Vested (%), Locked, Next Unlock date
- Progress bar: sky-blue fill on surface-container background
- "Withdraw" button for vested unlocked amounts
- "View on DAG Explorer" link

### 7.2 Voice/Video Call Screen (P0)

**Route:** `/call/:id`  
**Source:** PRD §Voice and Video Calls, Blueprint §Call Establishment Flow

```
Voice Call (Active):
┌─────────────────────────────────────────────┐
│                                              │
│         Icy Background Atmosphere            │
│         (blurred gradient orbs)              │
│                                              │
│              ┌──────────┐                    │
│              │          │                    │
│              │  Avatar  │  (80px, trust ring)│
│              │          │                    │
│              └──────────┘                    │
│                                              │
│           Sarah Chen                         │
│         🟢 Trusted Node                      │
│                                              │
│            03:42                              │
│         (call duration)                       │
│                                              │
│     ┌─────────────────────────────┐          │
│     │  End-to-end encrypted       │          │
│     │  🔒 Noise Protocol          │          │
│     └─────────────────────────────┘          │
│                                              │
│                                              │
│    ┌────┐  ┌────┐  ┌────┐  ┌────┐           │
│    │Mute│  │Spkr│  │Scrn│  │Flip│           │
│    └────┘  └────┘  └────┘  └────┘           │
│                                              │
│              ┌──────────┐                    │
│              │  🔴 End  │  (64px red circle) │
│              └──────────┘                    │
│                                              │
└─────────────────────────────────────────────┘

Video Call (Active):
┌─────────────────────────────────────────────┐
│  Full-screen remote video feed               │
│                                              │
│  ┌──────┐  03:42  🔒 E2E   Sarah Chen      │
│  │ Self │                                    │
│  │ PiP  │         (frosted glass overlay)    │
│  └──────┘                                    │
│                                              │
│                                              │
│                                              │
│                                              │
│                                              │
│                                              │
│    ┌────┐  ┌────┐  ┌────┐  ┌────┐           │
│    │Mute│  │ Cam│  │Scrn│  │Flip│           │
│    └────┘  └────┘  └────┘  └────┘           │
│                                              │
│              ┌──────────┐                    │
│              │  🔴 End  │                    │
│              └──────────┘                    │
└─────────────────────────────────────────────┘
```

**Component Details:**
- Background: icy atmosphere with large blurred orbs (same as login)
- Avatar: 80px with trust ring, centered
- Name: headline-sm, bold
- Trust badge: pill with tier color
- Duration: display-lg weight, monospace appearance
- Encryption badge: frosted glass pill with lock icon
- Action buttons: 56px circles on frosted glass, icons in white
  - Mute: mic / mic-off toggle
  - Speaker/Camera: speaker / camera toggle
  - Screen Share: monitor icon
  - Flip Camera: camera-rotate icon
- End Call: 64px circle, error red fill, phone-off icon
- Self PiP (video mode): 100×140px, 16px radius, top-left, draggable

**States:**
- Incoming Call: pulsing avatar ring, Accept (green) / Decline (red) buttons
- Connecting: "Connecting..." with animated dots
- Active: timer running, controls visible
- On Hold: "Call on hold" label
- Screen Sharing: "Sharing your screen" banner at top with stop button
- Group Call: grid layout of participant avatars (2×2, 3×3)

### 7.3 Contact Detail (P0)

**Route:** `/contacts/:id`  
**Source:** PRD §Dynamic Trust Network, IOS_IMPL §Contacts

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Back                            ···     │
│                                              │
│              ┌──────────┐                    │
│              │  Avatar  │  140px trust ring  │
│              │          │                    │
│              └──────────┘                    │
│                                              │
│           Sarah Chen                         │
│         echo:sarah.eth                       │
│       🟢 Inner Circle · Online               │
│                                              │
│  ┌────────┬────────┬────────┬────────┐      │
│  │  💬    │  📞    │  📹    │  🔍    │      │
│  │Message │ Voice  │ Video  │ Search │      │
│  └────────┴────────┴────────┴────────┘      │
│                                              │
│  TRUST & IDENTITY                            │
│  ┌──────────────────────────────────────┐   │
│  │ Trust Score     87/100  ████████░░   │   │
│  │ DID             did:cardano:addr1... │   │
│  │ Verified Since  Jan 15, 2026         │   │
│  │ Mutual Groups   3                    │   │
│  │ Mutual Contacts 12                   │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  CREDENTIALS                                 │
│  ┌──────────────────────────────────────┐   │
│  │ ✓ Government ID Verified             │   │
│  │ ✓ Apple ID Connected                 │   │
│  │ ✓ Phone Number Verified              │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  SHARED MEDIA                                │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐  See All →    │
│  │    │ │    │ │    │ │    │               │
│  └────┘ └────┘ └────┘ └────┘               │
│                                              │
│  PRIVACY FOR THIS CONTACT                    │
│  ┌──────────────────────────────────────┐   │
│  │ Custom Notifications  On     →       │   │
│  │ Disappearing Messages Off    →       │   │
│  │ Block Contact                →       │   │
│  │ Report Contact               →       │   │
│  └──────────────────────────────────────┘   │
│                                              │
└─────────────────────────────────────────────┘
```

### 7.4 Advanced Search (P1)

**Route:** `/search`  
**Source:** PRD §Advanced Message Search and Archive System

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Cancel                                   │
│                                              │
│  ┌──────────────────────────────────────┐   │
│  │ 🔍  Search messages, files, links... │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  ┌──────┬───────┬──────┬───────┐            │
│  │ All  │ Files │Photos│ Links │            │
│  └──────┴───────┴──────┴───────┘            │
│                                              │
│  FILTERS                                     │
│  From: [Any Contact ▾]                       │
│  Date: [Any Time ▾]                          │
│  In:   [All Chats ▾]                         │
│                                              │
│  RECENT SEARCHES                             │
│  ┌──────────────────────────────────────┐   │
│  │ 🕐 quarterly report                  │   │
│  │ 🕐 travel plans                      │   │
│  │ 🕐 invoice                           │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  (results appear below as user types)        │
│                                              │
│  RESULTS                                     │
│  ┌──────────────────────────────────────┐   │
│  │ 👤 Sarah Chen · 2 days ago           │   │
│  │ "...the quarterly report looks..."   │   │
│  ├──────────────────────────────────────┤   │
│  │ 👤 Marcus J · 1 week ago             │   │
│  │ "...attached the final report..."    │   │
│  │ 📎 Q4_Report.pdf                     │   │
│  └──────────────────────────────────────┘   │
│                                              │
└─────────────────────────────────────────────┘
```

**Key Details:**
- Search is performed locally (client-side index for E2E encryption)
- Filter chips use ghost border style, sky-blue fill when active
- Results highlight matching keywords with sky-blue text
- File results show file type icon and size
- "All Chats" filter includes hidden folder chats if biometric authenticated

### 7.5 Media Gallery (P1)

**Route:** `/chat/:id/media`  
**Source:** PRD §Large File Sharing

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Back            Shared Media     Select  │
│                                              │
│  ┌──────┬───────┬──────┬───────┐            │
│  │Photos│Videos │ Files│ Links │            │
│  └──────┴───────┴──────┴───────┘            │
│                                              │
│  MARCH 2026                                  │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐               │
│  │    │ │    │ │    │ │    │               │
│  │img │ │img │ │img │ │img │               │
│  │    │ │    │ │    │ │    │               │
│  └────┘ └────┘ └────┘ └────┘               │
│  ┌────┐ ┌────┐ ┌────┐                       │
│  │    │ │    │ │    │                       │
│  │img │ │vid │ │img │                       │
│  │    │ │ ▶  │ │    │                       │
│  └────┘ └────┘ └────┘                       │
│                                              │
│  FEBRUARY 2026                               │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐               │
│  │    │ │    │ │    │ │    │               │
│  └────┘ └────┘ └────┘ └────┘               │
│                                              │
└─────────────────────────────────────────────┘
```

- Grid layout: 4 columns, 2px gap
- Image thumbnails fill grid cells
- Video thumbnails show play button overlay
- Files tab: list view with file icon, name, size, date
- Links tab: rich preview cards with thumbnail, title, domain

### 7.6 Notification Center (P1)

**Route:** `/notifications`  
**Source:** Blueprint §Notification Badges

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Back          Notifications    Mark All  │
│                                              │
│  TODAY                                       │
│  ┌──────────────────────────────────────┐   │
│  │ 🔔 Sarah Chen sent you a message    │   │
│  │    "That sounds great!"    2m ago    │   │
│  ├──────────────────────────────────────┤   │
│  │ 👥 Product Team — Alex mentioned you │   │
│  │    "@you check this PR"   15m ago    │   │
│  ├──────────────────────────────────────┤   │
│  │ 📞 Missed call from Marcus Johnson  │   │
│  │    Voice call · 3 min     1h ago     │   │
│  ├──────────────────────────────────────┤   │
│  │ 🏆 You earned 12.8 ECHO            │   │
│  │    Daily staking reward   6h ago     │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  YESTERDAY                                   │
│  ┌──────────────────────────────────────┐   │
│  │ 🛡️ Trust score updated to 74/100    │   │
│  │    +2 from peer verification        │   │
│  ├──────────────────────────────────────┤   │
│  │ 📢 ECHO Announcements posted        │   │
│  │    "v3.2 update available"           │   │
│  └──────────────────────────────────────┘   │
│                                              │
└─────────────────────────────────────────────┘
```

- Notification items: 80px height, left icon circle (40px), text stack
- Unread items: surface-container-lowest background with sky-blue left border
- Read items: surface background
- Categories: Messages, Calls, Groups, Channels, Trust, Wallet, System
- Swipe actions: Mark read, Mute, Delete

### 7.7 QR Identity Share (P1)

**Route:** `/profile/qr`  
**Source:** PRD §Decentralized Identity

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Back           My Identity        Share  │
│                                              │
│         Icy Background Atmosphere            │
│                                              │
│  ┌──────────────────────────────────────┐   │
│  │                                      │   │
│  │          ┌──────────────┐            │   │
│  │          │              │            │   │
│  │          │   QR Code    │            │   │
│  │          │  (DID + Key) │            │   │
│  │          │              │            │   │
│  │          └──────────────┘            │   │
│  │                                      │   │
│  │       echo:alice.eth                 │   │
│  │       did:cardano:addr1q8...         │   │
│  │       🔒 Trust Score: 72/100         │   │
│  │                                      │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  ┌────────────────┐  ┌────────────────┐     │
│  │  📷 Scan QR    │  │  📤 Share Link │     │
│  └────────────────┘  └────────────────┘     │
│                                              │
│  Scan another user's QR code to add them    │
│  as a trusted contact with verified          │
│  in-person attestation.                      │
│                                              │
└─────────────────────────────────────────────┘
```

- QR code: 200×200px, white background, contains DID URI + public key
- Card: frosted glass with ghost border
- "Scan QR" opens camera with QR detection overlay
- "Share Link" generates deep link for remote sharing
- In-person scan triggers "Verify Trust" flow (mutual trust attestation)

### 7.8 Governance/Voting (P2)

**Route:** `/governance`  
**Source:** Tokenomics v2 §3.3, IOS_IMPL v4.2 §7

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Back          Governance          Info   │
│                                              │
│  YOUR VOTING POWER                           │
│  ┌──────────────────────────────────────┐   │
│  │ Weight: 8,000  · Tier 4 · 1.8×      │   │
│  │ Effective Power: 14,400              │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  ACTIVE PROPOSALS                            │
│  ┌──────────────────────────────────────┐   │
│  │ PROP-007                             │   │
│  │ Increase staking rewards by 5%       │   │
│  │ ████████████░░░░░░ 67% For           │   │
│  │ Ends in 3 days · 234 voters          │   │
│  │ [Vote For] [Vote Against] [Abstain]  │   │
│  ├──────────────────────────────────────┤   │
│  │ PROP-006                             │   │
│  │ Add new validator node requirement   │   │
│  │ ██████████████████ 89% For           │   │
│  │ Ends in 1 day · 512 voters           │   │
│  │ ✅ You voted: For                    │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  COMPLETED                                   │
│  ┌──────────────────────────────────────┐   │
│  │ PROP-005 · Passed ✓                  │   │
│  │ Treasury allocation for Q2 marketing │   │
│  │ 92% For · 1,204 voters              │   │
│  └──────────────────────────────────────┘   │
│                                              │
└─────────────────────────────────────────────┘
```

- Voting Power card: ghost border, sky-blue accent for tier
- Proposal cards: ghost border, progress bar shows For/Against ratio
- Vote buttons: signature gradient for "Vote For", glass secondary for others
- Completed proposals: faded, with pass/fail badge
- Vote confirmation: modal with summary, "recorded on-chain" disclaimer

### 7.9 Backup & Export (P2)

**Route:** `/settings/backup`  
**Source:** PRD §Recovery, IOS_IMPL §Persistence

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Settings       Backup & Security         │
│                                              │
│  RECOVERY PHRASE                             │
│  ┌──────────────────────────────────────┐   │
│  │ ⚠️ Keep this phrase secret           │   │
│  │ Never share it with anyone           │   │
│  │                                      │   │
│  │ [View Recovery Phrase]               │   │
│  │  Requires biometric authentication   │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  ENCRYPTED BACKUP                            │
│  ┌──────────────────────────────────────┐   │
│  │ Last backup: Today, 9:14 AM          │   │
│  │ Size: 42.3 MB (encrypted)            │   │
│  │ Location: iCloud Keychain            │   │
│  │                                      │   │
│  │ [Back Up Now]                        │   │
│  │ [Change Backup Location]             │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  AUTO-BACKUP                                 │
│  ┌──────────────────────────────────────┐   │
│  │ Frequency     Daily        →         │   │
│  │ Include Media Yes          →         │   │
│  │ WiFi Only     Yes          →         │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  EXPORT DATA                                 │
│  ┌──────────────────────────────────────┐   │
│  │ [Export Chat History]                 │   │
│  │ [Export Contacts]                     │   │
│  │ [Export Identity (DID)]              │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  DANGER ZONE                                 │
│  ┌──────────────────────────────────────┐   │
│  │ [Delete All Data]   (error red)      │   │
│  └──────────────────────────────────────┘   │
│                                              │
└─────────────────────────────────────────────┘
```

### 7.10 Enterprise Organization Profile (P2)

**Route:** `/enterprise/:id`  
**Source:** PRD §Enterprise Organization Profiles

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Back                            ···     │
│                                              │
│  ┌──────────────────────────────────────┐   │
│  │  🏢 [Logo]                           │   │
│  │  Acme Financial Corp                 │   │
│  │  ✅ Verified Organization            │   │
│  │  🏛️ Financial Institution            │   │
│  │  Since: March 2026                   │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  VERIFICATION                                │
│  ┌──────────────────────────────────────┐   │
│  │ ✓ Business Registration Verified     │   │
│  │ ✓ Domain Ownership Confirmed         │   │
│  │ ✓ EV TLS Certificate Valid           │   │
│  │ ✓ Blockchain Identity Anchored       │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  OFFICIAL CHANNELS                           │
│  ┌──────────────────────────────────────┐   │
│  │ 📢 Acme Announcements    12K subs   │   │
│  │ 💬 Acme Support          Active      │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  AUTHORIZED REPRESENTATIVES                  │
│  ┌──────────────────────────────────────┐   │
│  │ 👤 Jane Smith  · VP, Client Rel.     │   │
│  │ 👤 Bob Lee     · Support Lead        │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  [Message Organization]                      │
│                                              │
└─────────────────────────────────────────────┘
```

### 7.11 Bot Management (P3)

**Route:** `/settings/bots`  
**Source:** PRD §Decentralized Bot Framework

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Settings        Bot Framework     + Add  │
│                                              │
│  MY BOTS                                     │
│  ┌──────────────────────────────────────┐   │
│  │ 🤖 Expense Tracker Bot              │   │
│  │    Active · 3 conversations          │   │
│  │    Last triggered: 2h ago            │   │
│  ├──────────────────────────────────────┤   │
│  │ 🤖 Daily Standup Bot                │   │
│  │    Active · Product Team group       │   │
│  │    Next run: Tomorrow 9:00 AM        │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  DISCOVER BOTS                               │
│  ┌──────────────────────────────────────┐   │
│  │ 🤖 Weather Bot          [+ Add]     │   │
│  │ 🤖 Translation Bot      [+ Add]     │   │
│  │ 🤖 Reminder Bot         [+ Add]     │   │
│  │ 🤖 Poll Bot             [+ Add]     │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  All bots run in sandboxed environments      │
│  with no access to your private keys.        │
│                                              │
└─────────────────────────────────────────────┘
```

### 7.12 Staking/Delegation Detail (P3)

**Route:** `/wallet/staking`  
**Source:** Tokenomics v2 §1, IOS_IMPL §Wallet

```
┌─────────────────────────────────────────────┐
│ ▬▬▬ Secure Thread ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬ │
│  ← Wallet          Staking            Info  │
│                                              │
│  YOUR STAKING POSITION                       │
│  ┌──────────────────────────────────────┐   │
│  │ Staked:      8,000 ECHO              │   │
│  │ Tier:        Gold (5,000+ ECHO)      │   │
│  │ APY:         12.5%                    │   │
│  │ Lock Period: 90 days remaining        │   │
│  │ Multiplier:  1.8×                     │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  STAKING TIERS                               │
│  ┌──────────────────────────────────────┐   │
│  │ Bronze  100+    8.0% APY    1.0×     │   │
│  │ Silver  1,000+ 10.0% APY   1.2×     │   │
│  │ Gold    5,000+ 12.5% APY   1.8× ←   │   │
│  │ Plat.   25,000+ 15.0% APY  2.5×     │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  DELEGATION                                  │
│  ┌──────────────────────────────────────┐   │
│  │ Delegated to: Validator #7           │   │
│  │ Validator uptime: 99.8%              │   │
│  │ Commission: 5%                        │   │
│  │ [Change Validator]                    │   │
│  └──────────────────────────────────────┘   │
│                                              │
│  [Stake More]  [Unstake]                     │
│                                              │
└─────────────────────────────────────────────┘
```

---

## 8. Design Fixes for Existing Pages

### 8.1 Tab Bar Fix

**Current:** 4 tabs (Messages, Contacts, Rewards, Profile) with `border-top: 1px solid`

**Fix:**
- Change to 3 tabs: **Messages**, **Wallet**, **Me**
- Remove `border-top` line — use tonal shift instead (tab bar bg = `surface-container-lowest`, content area bg = `surface`)
- Move Contacts to accessible from Messages header (people icon) and Settings
- Rewards becomes Wallet (full screen, not just a list)

### 8.2 Profile Page Expansion

**Current:** 126 lines — basic avatar, trust tier bars, persona switcher, 3 menu items

**Add:**
- DID display: `echo:alice.eth` and truncated Cardano DID with copy button
- Credential badges section: list of verified credentials with icons
- Wallet summary card: total ECHO balance with "View Wallet" link
- QR Code share button: navigates to QR Identity Share
- Bio/status text field
- Mutual connections count
- "Edit Profile" button with photo, name, bio editing
- Expand Account Configuration section: Personas, Hidden Folders, Trust Dashboard, Devices, Login History

### 8.3 Settings Page — Remove Hard Lines

**Current:** Uses implicit section dividers that may render as lines

**Fix:** 
- Use tonal shifts between sections (surface → surface-container-low)
- Section headers: 10px Label-SM, bold, +10% tracking, text-tertiary
- No `Divider()` or `border-bottom` between rows

### 8.4 Add Secure Thread Indicator

Add the 2px sky-blue pulsating line to ALL authenticated screens:
- Messages, Chat, all Group pages, all Channel pages, all Persona pages
- Hidden Folders, Trust, Profile, Settings
- Wallet, Contact Detail, Call screens
- Position: immediately below status bar, above navigation bar

### 8.5 Messages List — Header Update

**Current:** "ECHO" wordmark + right-side icons

**Add to header:**
- Notification bell icon (with unread badge) → opens Notification Center
- Search icon → opens Advanced Search
- Compose icon → new message
- Maintain persona switcher accessibility

---

## 9. Interaction & Prototype Flows

### Flow 1: Onboarding → Home

```
Intro Carousel → Welcome → Login (Passkey/SMS)
  → Phone Entry → OTP → Passkey Setup
  → Recovery Phrase → Profile Setup
  → Verify Credentials → Trust Dashboard
  → Messages List (home)
```

### Flow 2: Message → Chat → Call

```
Messages List → tap conversation → Chat View
  Chat View → tap phone icon → Voice Call
  Chat View → tap video icon → Video Call
  Chat View → tap user avatar → Contact Detail
  Chat View → tap search icon → Advanced Search (scoped)
  Chat View → swipe message → Message Actions Sheet
```

### Flow 3: Wallet → Stake → Confirm

```
Wallet Tab → tap "Stake" → Staking Detail
  → enter amount → select tier → confirm
  → biometric auth → transaction signed
  → success confirmation → back to Wallet (updated balance)
```

### Flow 4: Profile → Settings → Hidden Folders

```
Me Tab → Profile → tap Settings gear → Settings
  Settings → tap "Hidden Folders" → Biometric Auth Modal
  → success → Hidden Folders list → tap folder → Folder Detail
  → tap chat → Hidden Folder Chat
```

### Flow 5: Governance Voting

```
Wallet → "Governance" link → Governance page
  → tap proposal → Proposal Detail
  → tap "Vote For" → Vote Confirmation Modal
  → review summary → "Submit Vote"
  → on-chain recording → confirmation
```

---

## 10. Google Stitch Import Instructions

### What is Google Stitch?

Google Stitch (formerly Project IDX templates) accepts structured design specifications and can generate Figma-compatible component definitions. To load this guide into Stitch:

### Step 1: Prepare Component Definitions

Export each component specification from Section 5 as individual JSON component definitions:

```json
{
  "component": "TabBar",
  "type": "navigation",
  "dimensions": { "width": 393, "height": 82 },
  "variants": ["messages_active", "wallet_active", "me_active"],
  "tokens": {
    "background": "var(--echo-surface-container-lowest)",
    "active_color": "var(--echo-accent)",
    "inactive_color": "var(--echo-text-tertiary)"
  },
  "children": [
    { "type": "tab_item", "icon": "MessageCircle", "label": "Messages" },
    { "type": "tab_item", "icon": "Wallet", "label": "Wallet" },
    { "type": "tab_item", "icon": "User", "label": "Me" }
  ]
}
```

### Step 2: Create Stitch Project

1. Open Google Stitch at `stitch.google.com`
2. Create new project: "ECHO iOS App v3.2"
3. Set target platform: iOS (iPhone 15 Pro — 393×852)
4. Import design tokens from Section 2 as a theme file
5. Upload this markdown as the specification document

### Step 3: Generate Screens

For each screen in Section 7, Stitch can auto-generate layouts using:
- The wireframe ASCII specifications as layout guides
- Component library definitions from Section 5
- Design tokens from Section 2
- Content from the mock data already in the codebase

### Step 4: Apply Glacial Interface Theme

After generation, apply these Glacial Interface rules:
- Replace any `1px solid` borders with ghost borders (15% opacity)
- Ensure all shadows use `on-surface` tinted colors (never grey)
- Verify 32px corner radius on all cards and containers
- Add backdrop-blur(20px) to all floating elements
- Confirm Inter font at all text positions
- Add Secure Thread Indicator to all screens

### Step 5: Export to Figma

Stitch exports directly to Figma. After export:
1. Verify component instances are properly linked
2. Set up Auto Layout on all frames
3. Create dark mode variants
4. Build prototype connections per Section 9 flows
5. Run accessibility check (contrast ratios, touch targets)

---

## 11. Handoff Checklist

### Design Completeness

- [ ] All 35 screens designed (23 existing + 12 new)
- [ ] Dark mode variants for all screens
- [ ] All component variants documented
- [ ] Secure Thread Indicator on all authenticated screens
- [ ] Tab bar updated to 3 tabs (Messages, Wallet, Me)
- [ ] Profile page expanded with DID, credentials, wallet summary
- [ ] All ghost borders at 15% opacity (no 1px solid borders)
- [ ] All shadows tinted with on-surface (no grey shadows)
- [ ] 32px corner radius on all cards and containers
- [ ] Inter font at all text positions

### Interaction Completeness

- [ ] Prototype flows connected per Section 9
- [ ] All button states: default, pressed, disabled, loading
- [ ] All input states: empty, focused, filled, error
- [ ] Sheet/modal open and close animations
- [ ] Page transition animations (spring, 0.85 damping)
- [ ] Biometric auth modal integrated where needed
- [ ] Swipe gestures on list items
- [ ] Pull-to-refresh on scrollable lists

### Developer Handoff

- [ ] All spacing uses design system tokens
- [ ] All colors reference CSS variables
- [ ] Component documentation includes state management notes
- [ ] API endpoint references for dynamic content
- [ ] Accessibility annotations (VoiceOver labels, semantic roles)
- [ ] RTL layout considerations noted
- [ ] Safe area handling documented

---

*ECHO Figma Implementation Guide v3.2*  
*Aligned to: PRD v2.5.1, Tokenomics v2, IOS_IMPL v4.2, DESIGN.md v1.0*  
*Generated: April 2026*
