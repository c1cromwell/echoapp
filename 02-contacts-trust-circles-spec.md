# Echo UX Specifications
## Feature 2: Contacts & Trust Circles

### Overview
The Contacts feature is central to Echo's trust-based messaging. Unlike traditional contact lists, Echo organizes contacts into Trust Circles that determine visibility, permissions, and interaction levels. Users can discover contacts via multiple privacy-preserving methods.

---

## Information Architecture

```
Contacts Tab
├── Search Bar
├── Contact Requests (if any)
├── Trust Circles
│   ├── Inner Circle (★)
│   ├── Trusted (✓)
│   └── Acquaintances (○)
├── All Contacts (A-Z)
└── FAB: Add Contact
    ├── Scan QR Code
    ├── Search Username
    ├── Search by DID
    └── Share My QR
```

---

## Screen 1: Contacts List

### Layout
```
┌─────────────────────────────────┐
│ ←        Contacts          ＋   │
├─────────────────────────────────┤
│  🔍 Search contacts...          │
├─────────────────────────────────┤
│  ┌─────────────────────────────┐│
│  │ 👤 3 contact requests    ›  ││
│  └─────────────────────────────┘│
├─────────────────────────────────┤
│  TRUST CIRCLES                  │
│                                 │
│  ⭐ Inner Circle          12 ›  │
│  ✓  Trusted               28 ›  │
│  ○  Acquaintances         45 ›  │
├─────────────────────────────────┤
│  ALL CONTACTS                   │
│                                 │
│  A ─────────────────────────    │
│  [Avatar] Alex Rivera           │
│           @alexr · Trust 72  ›  │
│                                 │
│  [Avatar] Amanda Chen      ⭐   │
│           @amandac · Trust 91›  │
│                                 │
│  B ─────────────────────────    │
│  [Avatar] Blake Thompson        │
│           @blaket · Trust 45 ›  │
│                                 │
├─────────────────────────────────┤
│  💬     👥      🛡️      👤     │
│ Messages Contacts Trust  Profile│
└─────────────────────────────────┘
```

### Specifications
| Element | Spec |
|---------|------|
| Search bar | 36px height, #F9FAFB bg, 10px radius |
| Request banner | #FEF3C7 bg, 12px radius, tap to expand |
| Section headers | 13px, SemiBold, #9CA3AF, uppercase |
| Circle counts | 14px, #6B7280, right-aligned |
| Alpha headers | 13px, SemiBold, #9CA3AF, sticky |
| Contact row | 64px height, 56px left for avatar |
| Trust indicator | Colored dot or badge based on circle |

### Interactions
- Tap search → Focus with keyboard
- Tap request banner → Request list screen
- Tap circle row → Filtered contact list
- Tap contact → Contact profile
- Long press contact → Quick actions sheet
- FAB → Add contact options

---

## Screen 2: Contact Requests

### Layout
```
┌─────────────────────────────────┐
│ ←    Contact Requests           │
├─────────────────────────────────┤
│  PENDING (3)                    │
│                                 │
│  ┌─────────────────────────────┐│
│  │ [Avatar] Jordan Lee      ✓  ││
│  │          @jordanlee         ││
│  │          Trust Score: 67    ││
│  │          "Hey! Met at the   ││
│  │          conference..."     ││
│  │                             ││
│  │  2 mutual contacts          ││
│  │                             ││
│  │  [Accept]  [Decline]  [···] ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ [Avatar] Unknown User       ││
│  │          @user8472          ││
│  │          Trust Score: 12    ││
│  │          No message         ││
│  │                             ││
│  │  ⚠️ Low trust score         ││
│  │                             ││
│  │  [Accept]  [Decline]  [···] ││
│  └─────────────────────────────┘│
│                                 │
│  SENT (2)                       │
│                                 │
│  ┌─────────────────────────────┐│
│  │ [Avatar] Sarah Chen      ✓  ││
│  │          Pending · 2 days   ││
│  │          [Cancel Request]   ││
│  └─────────────────────────────┘│
│                                 │
└─────────────────────────────────┘
```

### Request Card Specifications
| Element | Spec |
|---------|------|
| Card | White bg, 16px radius, 16px padding |
| Avatar | 56px, left aligned |
| Name | 16px, SemiBold |
| Username | 14px, Regular, #6B7280 |
| Trust score | 14px, colored by level |
| Message | 14px, #6B7280, max 2 lines |
| Mutual contacts | 13px, #6B7280, link style |
| Warning | 13px, #F59E0B, with icon |
| Accept button | Primary filled, 36px height |
| Decline button | Ghost style, 36px height |

### Accept Flow
```
Tap Accept → Bottom Sheet:
┌─────────────────────────────────┐
│  ━━━                            │
│                                 │
│  Add to Trust Circle            │
│                                 │
│  Choose where to add Jordan:    │
│                                 │
│  ○ ⭐ Inner Circle              │
│       Full access, see all      │
│       personas                  │
│                                 │
│  ● ✓  Trusted                   │
│       Standard access           │
│                                 │
│  ○ ○  Acquaintance              │
│       Limited visibility        │
│                                 │
│  ┌─────────────────────────────┐│
│  │       Add Contact           ││
│  └─────────────────────────────┘│
└─────────────────────────────────┘
```

---

## Screen 3: Trust Circle Detail

### Layout
```
┌─────────────────────────────────┐
│ ←      Inner Circle        Edit │
├─────────────────────────────────┤
│                                 │
│  ⭐ Your most trusted contacts  │
│     12 people                   │
│                                 │
│  ┌─────────────────────────────┐│
│  │  What Inner Circle sees:    ││
│  │  • All your personas        ││
│  │  • Online status            ││
│  │  • Read receipts            ││
│  │  • Can add to groups        ││
│  └─────────────────────────────┘│
│                                 │
├─────────────────────────────────┤
│  🔍 Search inner circle...      │
├─────────────────────────────────┤
│                                 │
│  [Avatar] Amanda Chen        ✓  │
│           Trust 91 · 3 years ›  │
│                                 │
│  [Avatar] Marcus Johnson     ✓  │
│           Trust 88 · 2 years ›  │
│                                 │
│  [Avatar] Sarah Chen         ✓  │
│           Trust 85 · 1 year  ›  │
│                                 │
│  [Avatar] David Park         ✓  │
│           Trust 82 · 6 months›  │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  [+ Add to Inner Circle]        │
│                                 │
└─────────────────────────────────┘
```

### Circle Permissions Summary
| Circle | See Personas | Online Status | Read Receipts | Add to Groups |
|--------|-------------|---------------|---------------|---------------|
| Inner Circle | All | Yes | Yes | Yes |
| Trusted | Selected | Yes | Yes | Ask |
| Acquaintance | Default only | No | No | No |

---

## Screen 4: Add Contact

### Layout - Method Selection
```
┌─────────────────────────────────┐
│ ×       Add Contact             │
├─────────────────────────────────┤
│                                 │
│  ┌─────────────────────────────┐│
│  │  📷                         ││
│  │  Scan QR Code               ││
│  │  Scan someone's Echo QR     ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │  @                          ││
│  │  Search Username            ││
│  │  Find by @username          ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │  🔗                         ││
│  │  Enter DID                  ││
│  │  Add by decentralized ID    ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │  📤                         ││
│  │  Share My QR Code           ││
│  │  Let others add you         ││
│  └─────────────────────────────┘│
│                                 │
│  ─────────────────────────────  │
│                                 │
│  🔒 Echo never accesses your    │
│     phone contacts              │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 5: QR Code Scanner

### Layout
```
┌─────────────────────────────────┐
│ ×       Scan QR Code            │
├─────────────────────────────────┤
│                                 │
│  ┌─────────────────────────────┐│
│  │                             ││
│  │                             ││
│  │      ┌─────────────┐        ││
│  │      │             │        ││
│  │      │   CAMERA    │        ││
│  │      │    FEED     │        ││
│  │      │             │        ││
│  │      └─────────────┘        ││
│  │                             ││
│  │                             ││
│  └─────────────────────────────┘│
│                                 │
│  Point camera at an Echo QR     │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  [🔦 Toggle Flash]              │
│                                 │
│  [Show My QR Instead]           │
│                                 │
└─────────────────────────────────┘
```

### QR Scanned Success
```
┌─────────────────────────────────┐
│                                 │
│         ┌─────────┐             │
│         │ [Avatar]│             │
│         │   SC    │             │
│         └─────────┘             │
│                                 │
│        Sarah Chen ✓             │
│        @sarahchen              │
│                                 │
│        Trust Score: 85          │
│        ████████████░░ 85/100    │
│                                 │
│        ✓ Identity Verified      │
│        🔗 DID Verified          │
│                                 │
│        3 mutual contacts        │
│                                 │
│  ┌─────────────────────────────┐│
│  │     Send Contact Request    ││
│  └─────────────────────────────┘│
│                                 │
│        [Cancel]                 │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 6: My QR Code

### Layout
```
┌─────────────────────────────────┐
│ ×        My QR Code             │
├─────────────────────────────────┤
│                                 │
│         ┌─────────┐             │
│         │ [Avatar]│             │
│         │   ME    │             │
│         └─────────┘             │
│                                 │
│          Alex Echo              │
│          @alexecho              │
│                                 │
│  ┌─────────────────────────────┐│
│  │                             ││
│  │    ▄▄▄▄▄▄▄ ▄▄▄▄▄ ▄▄▄▄▄▄▄   ││
│  │    █ ▄▄▄ █ ███▀█ █ ▄▄▄ █   ││
│  │    █ ███ █ ▀▄█▄▄ █ ███ █   ││
│  │    █▄▄▄▄▄█ █ ▄▀█ █▄▄▄▄▄█   ││
│  │    ▄▄▄▄▄ ▄▄▄█▀ ▄ ▄ ▄ ▄▄▄   ││
│  │    █▄█▀▄▀▄▀▄██▀▄▄▀█▄▀▄▀█   ││
│  │    ▄▄▄▄▄▄▄ █▄▀▀ ▄███ ▄ █   ││
│  │    █ ▄▄▄ █ ▄█▀▄▀  ▄▄▀▄▀▄   ││
│  │    █ ███ █ █ █▀▄▄▀ ▀ █▄█   ││
│  │    █▄▄▄▄▄█ █  ▀▄█ ▀▀▀▄▀    ││
│  │                             ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │    📤 Share QR Code         ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │    📷 Scan a Code           ││
│  └─────────────────────────────┘│
│                                 │
└─────────────────────────────────┘
```

---

## Screen 7: Username Search

### Layout
```
┌─────────────────────────────────┐
│ ←     Search Username           │
├─────────────────────────────────┤
│                                 │
│  ┌─────────────────────────────┐│
│  │ @ │ sarahc                  ││
│  └─────────────────────────────┘│
│                                 │
│  RESULTS                        │
│                                 │
│  ┌─────────────────────────────┐│
│  │ [SC] Sarah Chen          ✓  ││
│  │      @sarahchen             ││
│  │      Trust: 85 · 3 mutual   ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ [SC] Sarah C.               ││
│  │      @sarahc_design         ││
│  │      Trust: 42 · 0 mutual   ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ [S]  SarahCrypto            ││
│  │      @sarahcrypto           ││
│  │      Trust: 28 · 0 mutual   ││
│  └─────────────────────────────┘│
│                                 │
│  ─────────────────────────────  │
│                                 │
│  💡 Tip: Exact matches show     │
│     verification badges         │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 8: Contact Profile (Extended)

### Layout
```
┌─────────────────────────────────┐
│ ←                          ···  │
├─────────────────────────────────┤
│                                 │
│         ┌─────────┐             │
│         │ [Photo] │             │
│         │         │             │
│         └─────────┘             │
│                                 │
│      Sarah Chen ✓               │
│      @sarahchen                 │
│                                 │
│  ┌──────┐ ┌──────┐ ┌──────┐    │
│  │  💬  │ │  📞  │ │  📹  │    │
│  │ Chat │ │ Call │ │Video │    │
│  └──────┘ └──────┘ └──────┘    │
│                                 │
├─────────────────────────────────┤
│  TRUST                          │
│                                 │
│  Score: 85/100                  │
│  ████████████████░░░░ Verified  │
│                                 │
│  ⭐ Inner Circle           [▼]  │
│     Change trust circle         │
│                                 │
├─────────────────────────────────┤
│  VERIFICATION                   │
│                                 │
│  ✓ Identity Verified            │
│    Government ID · Jan 2024     │
│                                 │
│  🔗 DID Verified                │
│    did:cardano:abc1...7xyz      │
│                                 │
│  📱 Phone Verified              │
│    +1 ••••••• 4567              │
│                                 │
├─────────────────────────────────┤
│  PERSONAS VISIBLE TO YOU        │
│                                 │
│  [👔] Professional              │
│  [🏠] Personal                  │
│                                 │
├─────────────────────────────────┤
│  SHARED MEDIA              See All│
│                                 │
│  [img] [img] [img] [img]        │
│                                 │
├─────────────────────────────────┤
│  SHARED GROUPS                  │
│                                 │
│  [🏢] Product Team              │
│  [🎨] Design Guild              │
│                                 │
├─────────────────────────────────┤
│                                 │
│  [🔕 Mute Notifications]        │
│                                 │
│  [🚫 Block Contact]             │
│                                 │
│  [⚠️ Report]                    │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 9: Edit Trust Circle (Bottom Sheet)

### Layout
```
┌─────────────────────────────────┐
│  ━━━                            │
│                                 │
│  Change Trust Circle            │
│                                 │
│  Sarah Chen is currently in:    │
│  ⭐ Inner Circle                │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  ┌─────────────────────────────┐│
│  │ ⭐ Inner Circle          ✓  ││
│  │    See all personas, full   ││
│  │    access to your profile   ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ ✓  Trusted                  ││
│  │    See selected personas,   ││
│  │    standard messaging       ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ ○  Acquaintance             ││
│  │    Default persona only,    ││
│  │    limited visibility       ││
│  └─────────────────────────────┘│
│                                 │
│  ─────────────────────────────  │
│                                 │
│  [Remove from Contacts]         │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 10: Manage Visible Personas

### Layout
```
┌─────────────────────────────────┐
│ ←   Personas Visible to Sarah   │
├─────────────────────────────────┤
│                                 │
│  Choose which of your personas  │
│  Sarah Chen can see:            │
│                                 │
│  ─────────────────────────────  │
│                                 │
│  ┌─────────────────────────────┐│
│  │ 👔 Professional          ☑  ││
│  │    Alex Echo · Product Lead ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ 🏠 Personal              ☑  ││
│  │    Alex · Just me           ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ 👨‍👩‍👧 Family               ☐  ││
│  │    Dad · Family stuff       ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ 🎮 Gaming                ☐  ││
│  │    NightOwl42 · Gaming      ││
│  └─────────────────────────────┘│
│                                 │
│  ─────────────────────────────  │
│                                 │
│  💡 Inner Circle contacts can   │
│     see all personas by default │
│                                 │
│  ┌─────────────────────────────┐│
│  │         Save Changes        ││
│  └─────────────────────────────┘│
│                                 │
└─────────────────────────────────┘
```

---

## Component Specifications

### Contact List Item
```
Height: 64px
Padding: 12px 16px
Avatar: 44px
Name: 16px SemiBold #1A1A1A
Username: 14px Regular #6B7280
Trust indicator: 14px, colored by score
Chevron: 20px #D1D5DB
Divider: 1px #F3F4F6 (inset 72px)
```

### Trust Circle Badge
```
Inner Circle: ⭐ #F59E0B background #FEF3C7
Trusted: ✓ #10B981 background #D1FAE5
Acquaintance: ○ #6B7280 background #F3F4F6
```

### Contact Request Card
```
Background: #FFFFFF
Border: 1px solid #E5E7EB
Border Radius: 16px
Padding: 16px
Shadow: 0 1px 3px rgba(0,0,0,0.1)

Accept Button: 
  Height: 36px
  Background: #6366F1
  Color: white
  
Decline Button:
  Height: 36px
  Background: transparent
  Color: #6B7280
  Border: 1px solid #E5E7EB
```

### QR Code Container
```
Background: #FFFFFF
Border: 1px solid #E5E7EB
Border Radius: 16px
Padding: 24px
QR Size: 200x200px
```

---

## Animations & Transitions

### Contact Request Accept
1. Card scales to 0.98 on press
2. On accept, checkmark animation plays
3. Card slides out to right (300ms)
4. List items shift up (200ms)

### Trust Circle Change
1. Selection indicator animates (scale 0→1)
2. Previous selection fades (150ms)
3. Haptic feedback on selection

### QR Scanner
1. Scanning frame pulses subtly
2. On successful scan, frame turns green
3. Success haptic + sound
4. Profile card slides up from bottom

---

## Empty States

### No Contacts
```
Illustration: Two people connecting
Title: "Your contact list is empty"
Subtitle: "Add contacts by scanning QR codes or searching usernames"
CTA: "Add Your First Contact"
```

### No Search Results
```
Icon: 🔍 with X
Title: "No contacts found"
Subtitle: "Try a different search term or add a new contact"
CTA: "Add Contact"
```

### No Contact Requests
```
Icon: ✓ in circle
Title: "All caught up!"
Subtitle: "No pending contact requests"
```

---

## Error States

### QR Scan Failed
```
Icon: ⚠️
Title: "Couldn't read QR code"
Subtitle: "Make sure it's a valid Echo QR code"
CTA: "Try Again"
```

### User Not Found
```
Icon: 👤 with ?
Title: "User not found"
Subtitle: "Check the username and try again"
```

### Request Failed
```
Toast: "Couldn't send request. Try again."
Duration: 3 seconds
Action: "Retry"
```
