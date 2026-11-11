# Echo — Messages Screen Updates · Figma Design Spec

> **Version:** 3.1 · **Date:** March 2026 · **Applies to:** Screens / Messages
> **Theme:** Icy Minimal (Light + Dark)

---

## Table of Contents

1. [Overview of Changes](#1-overview-of-changes)
2. [Persona Header — Verification Badge](#2-persona-header--verification-badge)
3. [Chat Folder Filter Bar](#3-chat-folder-filter-bar)
4. [Updated Messages Screen (Full Layout)](#4-updated-messages-screen-full-layout)
5. [Trust Score — No Decay Policy](#5-trust-score--no-decay-policy)
6. [New Components](#6-new-components)

---

## 1. Overview of Changes

Three changes to the Messages screen and trust system:

| Change | What | Where |
|---|---|---|
| Persona header badge | User's own verification status badge shown next to their profile name | Top of Messages screen (PersonaHeaderView) |
| Chat folder filters | Horizontal filter tabs that sort conversations by trust level | Below Pinned section, replaces old "ALL MESSAGES" divider |
| No score decay | Trust scores never decrease from inactivity — only reports/blocks reduce score | Trust scoring engine (backend + specs) |

**Removed:** The old `SectionDivider(title: "ALL MESSAGES")` static label is removed and replaced by the interactive folder filter bar.

---

## 2. Persona Header — Verification Badge

The persona header sits at the top of the Messages screen and shows the user's active persona (name, avatar). The update adds the user's own verification badge inline after their display name.

### Before

```
┌─────────────────────────────────┐
│ [AV]  Alex Echo           ▼  ✏️ │
└─────────────────────────────────┘
```

### After

```
┌─────────────────────────────────┐
│ [AV]  Alex Echo ✓          ▼  ✏️ │  ← Badge after name
└─────────────────────────────────┘
```

### Specifications

| Element | Spec |
|---|---|
| Badge position | 4px right of display name text, vertically centered with name baseline |
| Badge size | 16px (SF Symbol, `.medium` size from badge system) |
| Badge icon | Matches user's own `VerificationStatus` — same `checkmark.seal.fill` / `crown.fill` system |
| Badge color | Same tier colors: gray (phone), indigo (lite), green (full), gold (VIP Pro) |
| No badge | If user is unverified (`VerificationStatus.none`), no icon is shown — name renders without extra space |
| Tap behavior | Tapping the badge opens a tooltip/popover: "Your trust status: Fully Verified" with a "Learn more" link → Settings > Privacy > Trust Score |

### Badge Color Reference (Same as Conversation Row Badges)

| Status | Color | Icon |
|---|---|---|
| Phone Verified | `#9CA3AF` (gray) | `checkmark.seal.fill` |
| Identity Lite | `#6366F1` (indigo) | `checkmark.seal.fill` |
| Identity Full | `#10B981` (green) | `checkmark.seal.fill` |
| VIP Pro | `#F59E0B` (gold) | `crown.fill` |

### Light / Dark Mode

| Element | Light | Dark |
|---|---|---|
| Badge colors | Standard hex values | Same hex values (already high-contrast) |
| Tooltip background | `#FFFFFF` | `#1E293B` |
| Tooltip text | `#0F172A` | `#F1F5F9` |
| Tooltip border | `#E2E8F0` | `#334155` |
| Tooltip shadow | Standard drop shadow | Subtle glow |

### Tooltip Layout

```
         ┌──────────────────────────┐
         │ Your trust status:       │
         │ ✓ Fully Verified         │  ← Badge color + label
         │                          │
         │ Learn more →             │  ← Accent link → Trust Score
         └──────────────────────────┘
                    ▽
              (points to badge)
```

| Element | Spec |
|---|---|
| Tooltip width | Auto (max 220px) |
| Tooltip padding | 12px |
| Tooltip corner radius | 12px |
| Tooltip arrow | 8px equilateral triangle pointing down |
| Label | 13px Regular `textSecondary` |
| Status line | 14px Semi-bold, badge icon (14px) + badge color text |
| "Learn more" | 13px Medium `accent` color |
| Dismiss | Tap anywhere outside tooltip |

---

## 3. Chat Folder Filter Bar

A horizontal scrollable filter bar that categorizes conversations by the other participant's trust/verification level. This replaces the old static "ALL MESSAGES" section divider.

### Layout

```
┌─────────────────────────────────┐
│                                 │
│  ┌─────┐ ┌──────────┐ ┌───────┐│
│  │ All │ │ Verified │ │Trusted││  ← Scrollable pill tabs
│  └─────┘ └──────────┘ └───────┘│
│                       ┌────────────┐
│                       │Unverified │
│                       └────────────┘
│                                 │
└─────────────────────────────────┘
```

The pills scroll horizontally if they don't all fit on screen (they will fit on most devices, but scroll ensures future-proofing).

### Filter Tabs

| Tab | Label | Filters conversations to | Badge indicator |
|---|---|---|---|
| All | "All" | No filter — shows every conversation | None |
| Verified | "Verified" | Contacts with `identityFull` or `vipPro` status (green/gold badge) | Green dot |
| Trusted | "Trusted" | Contacts with `identityLite` or `phoneVerified` status (indigo/gray badge) | Indigo dot |
| Unverified | "Unverified" | Contacts with `none` verification status (no badge) | Gray dot |

### Filter Tab Specifications

| Element | Active State (Light) | Inactive State (Light) |
|---|---|---|
| Background | `#0EA5E9` (accent) | `#F1F5F9` (backgroundSecondary) |
| Text color | `#FFFFFF` | `#475569` (textSecondary) |
| Text | 13px Semi-bold | 13px Medium |
| Height | 32px | 32px |
| Corner radius | 16px (pill) | 16px (pill) |
| Horizontal padding | 16px | 16px |
| Dot indicator | 6px circle, left of text, white | 6px circle, left of text, colored (see table) |
| Gap between pills | 8px | 8px |
| Container padding | 16px horizontal, 8px vertical | — |

| Element | Active State (Dark) | Inactive State (Dark) |
|---|---|---|
| Background | `#38BDF8` (accent) | `#1E293B` (backgroundSecondary) |
| Text color | `#0F172A` | `#CBD5E1` (textSecondary) |
| Dot indicator | 6px circle, `#0F172A` | 6px circle, colored |

### Dot Colors (Inactive State)

| Tab | Dot Color | Meaning |
|---|---|---|
| All | No dot | Shows everything |
| Verified | `#10B981` (green) | Matches green verification badge |
| Trusted | `#6366F1` (indigo) | Matches indigo verification badge |
| Unverified | `#9CA3AF` (gray) | Matches gray/no badge |

### Count Badges

Each filter tab (except "All") shows an unread count badge when that category has unread messages:

```
  ┌──────────────┐
  │ ● Verified 3 │  ← "3" = unread count in this category
  └──────────────┘
```

| Element | Spec |
|---|---|
| Count position | Right of label text, 4px gap |
| Count text | 12px Bold |
| Count color (active) | White (light) / `#0F172A` (dark) — matches pill text |
| Count color (inactive) | `accent` color |
| Count shown | Only when unread > 0; hidden when 0 |

### Interaction

| Action | Behavior |
|---|---|
| Tap "All" | Show all conversations (default) |
| Tap "Verified" | Filter to only conversations where the other participant has identity-full or VIP Pro status |
| Tap "Trusted" | Filter to conversations where participant has identity-lite or phone-verified status |
| Tap "Unverified" | Filter to conversations where participant has no verification |
| Scroll | Horizontal scroll if pills exceed screen width |
| Animation | Active pill slides accent bg with 200ms ease-in-out; conversation list cross-fades 150ms |

### Group Conversations

Group conversations appear in the **highest-tier filter** that any member belongs to. A group with one Verified and three Unverified members appears under "Verified". This encourages users to join groups with verified members.

### Empty State

If a filter has no conversations:

```
┌─────────────────────────────────┐
│                                 │
│        [Empty icon]             │
│                                 │
│    No verified contacts yet     │  ← 16px Semi-bold, textSecondary
│                                 │
│    Your verified contacts       │  ← 14px, textTertiary
│    will appear here.            │
│                                 │
└─────────────────────────────────┘
```

Each filter has contextual empty text:

| Filter | Empty Title | Empty Subtitle |
|---|---|---|
| Verified | "No verified contacts yet" | "Contacts who verify their identity will appear here." |
| Trusted | "No trusted contacts yet" | "Contacts who verify their phone will appear here." |
| Unverified | "No unverified contacts" | "All your contacts have some level of verification!" |

---

## 4. Updated Messages Screen (Full Layout)

### Complete wireframe with all changes applied:

```
┌─────────────────────────────────┐
│ [AV]  Alex Echo ✓          ▼  ✏️ │  ← Persona header + YOUR badge
├─────────────────────────────────┤
│ 🔍 Search messages...           │  ← Search bar
├─────────────────────────────────┤
│                                 │
│ ○ ○ ○ ○ ○ ○                    │  ← Pinned contacts (horizontal scroll)
│ Sara Jas Mar Ale Jam             │
│                                 │
├─────────────────────────────────┤
│                                 │
│ [All] [● Verified 2] [● Trusted]│  ← Folder filter pills
│                       [● Unver.]│
│                                 │
├─────────────────────────────────┤
│                                 │
│ ┌───┐                          │
│ │ AV│ Sarah M. ✓   2:34 PM    │  ← Green check (identity full)
│ │   │ Hey, are you free...  2  │
│ └───┘                          │
│ ─────────────────────────────  │
│ ┌───┐                          │
│ │ AV│ James K. 👑   1:15 PM    │  ← Gold crown (VIP Pro)
│ │   │ Just sent the payment    │
│ └───┘                          │
│ ─────────────────────────────  │
│ ┌───┐                          │
│ │ AV│ Alex T. ✓   Yesterday   │  ← Indigo check (identity lite)
│ │   │ Check out this link...   │
│ └───┘                          │
│ ─────────────────────────────  │
│ ┌───┐                          │
│ │ AV│ New Contact  11:30 AM    │  ← No badge (unverified)
│ │   │ Hi, I found you on...    │
│ └───┘                          │
│                                 │
│                            [+]  │  ← FAB (new message)
├─────────────────────────────────┤
│  💬     👥      ⭐      👤     │
│Messages Contacts Rewards Profile│
└─────────────────────────────────┘
```

### Layout Order (Top to Bottom)

1. **Persona Header** — avatar + name + verification badge + persona switcher + compose
2. **Search Bar** — full width search input
3. **Pinned Section** — horizontal scrollable pinned contacts/groups
4. **Folder Filter Bar** — horizontal pill tabs: All / Verified / Trusted / Unverified
5. **Conversation List** — filtered based on active folder tab
6. **FAB** — floating action button for new message
7. **Tab Bar** — Messages / Contacts / Rewards / Profile

---

## 5. Trust Score — No Decay Policy

### Policy Change

Trust scores **never decay from inactivity**. A user who verifies their identity, builds a trust score of 72, and then goes inactive for 6 months will still have a trust score of 72 when they return.

### What Reduces Trust Score

Trust scores only decrease from **negative actions by other users**:

| Event | Impact | Processing |
|---|---|---|
| Spam report received | -2 points per report (max -10) | Immediate |
| Fraud report received | -5 points per report (max -20) | Immediate |
| Blocked by another user | -1 point per block (max -10) | Batch (24hr) |

### What Does NOT Reduce Trust Score

| Scenario | Trust Score Impact | ECHO Reward Impact |
|---|---|---|
| Inactive for 30 days | No change | Daily reward cap paused |
| Inactive for 6 months | No change | Streak multiplier resets to 1.0x |
| Inactive for 1 year | No change | No rewards earned during inactivity |
| Low message volume | No change | Lower messaging rewards (proportional) |
| Stopped staking | No change | Staking rewards stop; staking points remain |

### Rewards vs Trust — Separation of Concerns

| System | Affected by Inactivity | Affected by Reports |
|---|---|---|
| **Trust Score** (0-100) | Never | Yes — reports/blocks reduce score |
| **ECHO Rewards** | Yes — daily cap pauses, streak resets | N/A |
| **Reward Multiplier** | Streak component resets after 7 days inactive | Trust multiplier stays (derived from trust score) |

### Updated Trust Score Formula

The formula from the blueprint remains the same — but with an explicit **no-decay guarantee**:

```
trust_score = min(100, max(0,
  verification_score +      // 0-30, permanent once earned
  interaction_score +        // 0-20, only increases (account age, messages, contacts, groups)
  behavior_score +           // 0-30, only increases (transactions, staking, governance)
  report_penalty             // 0 to -20, only from reports/blocks
))
```

**Critical rule:** `interaction_score` and `behavior_score` components are **monotonically increasing** — once a user reaches 500 messages, they keep the 5 points forever even if they stop messaging. Once they reach 10 transactions, they keep those 10 points forever.

### Design Implications

The "Trust Score" display in Settings > Privacy > Trust Score should include a reassurance line:

```
┌─────────────────────────────────┐
│         ┌─────────┐             │
│       ╱           ╲             │
│      │     72      │            │
│       ╲   /100    ╱             │
│         └─────────┘             │
│                                 │
│      ✓ TRUSTED                  │
│      1.2x reward multiplier     │
│                                 │
│  🔒 Your trust score never      │  ← NEW reassurance line
│     decreases from inactivity   │
│                                 │
└─────────────────────────────────┘
```

---

## 6. New Components

### 6.1 Figma Components to Create

| Component | Location | Variants |
|---|---|---|
| `PersonaHeader / WithBadge` | Components / Navigation | `Badge=None`, `Badge=Phone`, `Badge=Lite`, `Badge=Full`, `Badge=VipPro` × Light/Dark |
| `FilterPill / Active` | Components / Filters | `Label=All`, `Label=Verified`, `Label=Trusted`, `Label=Unverified` × Light/Dark |
| `FilterPill / Inactive` | Components / Filters | Same labels × Light/Dark |
| `FilterBar` | Components / Filters | Auto-layout horizontal stack of 4 pills |
| `Tooltip / TrustStatus` | Components / Overlays | Light/Dark |
| `EmptyState / Filter` | Components / States | `Filter=Verified`, `Filter=Trusted`, `Filter=Unverified` × Light/Dark |

### 6.2 Updated Screen Frames

Create these Figma frames:

| Frame | Description |
|---|---|
| `Messages / All / Light` | Full messages screen, "All" filter active, light mode |
| `Messages / All / Dark` | Same, dark mode |
| `Messages / Verified / Light` | Messages screen, "Verified" filter active, showing only verified contacts |
| `Messages / Verified / Dark` | Same, dark mode |
| `Messages / Verified / Empty / Light` | Messages screen, "Verified" active, empty state |
| `Messages / TrustBadge Tooltip / Light` | Messages screen with persona badge tooltip open |

---

*Applies to: Echo v3.0 Figma Design Spec*
*Sections affected: Screens / Messages, Components / Filters, Components / Navigation*
