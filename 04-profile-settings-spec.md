# Echo UX Specifications
## Feature 4: Profile & Settings

### Overview
The Profile section allows users to manage their identity, personas, and app preferences. Echo's unique personas feature lets users present different identities to different contacts while maintaining a single verified identity.

---

## Information Architecture

```
Profile Tab
├── My Profile Header
│   ├── Avatar
│   ├── Name & Username
│   ├── Trust Score Badge
│   └── Edit Profile CTA
├── My Personas (max 5)
│   ├── Professional
│   ├── Personal
│   ├── Family
│   ├── Gaming
│   └── + Create New
├── Quick Actions
│   ├── QR Code
│   ├── Invite Friends
│   └── Echo Rewards
└── Settings
    ├── Account
    ├── Privacy
    ├── Notifications
    ├── Appearance
    ├── Storage & Data
    ├── Help & Support
    └── About
```

---

## Screen 1: Profile Tab

### Layout
```
┌─────────────────────────────────┐
│          My Profile             │
├─────────────────────────────────┤
│                                 │
│         ┌─────────┐             │
│         │ [Photo] │             │
│         │         │             │
│         └─────────┘             │
│                                 │
│        Alex Echo ✓              │
│        @alexecho                │
│                                 │
│     ┌─────────────────┐         │
│     │ 🛡️ Trust: 72    │         │
│     │   Trusted       │         │
│     └─────────────────┘         │
│                                 │
│     [Edit Profile]              │
│                                 │
├─────────────────────────────────┤
│  MY PERSONAS                    │
│                                 │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐   │
│  │ 👔 │ │ 🏠 │ │ 👨‍👩‍👧 │ │ ＋ │   │
│  │Pro │ │Pers│ │Fam │ │Add │   │
│  └────┘ └────┘ └────┘ └────┘   │
│                                 │
├─────────────────────────────────┤
│  ┌─────────────────────────────┐│
│  │ 📱 My QR Code            ›  ││
│  └─────────────────────────────┘│
│  ┌─────────────────────────────┐│
│  │ 👥 Invite Friends        ›  ││
│  └─────────────────────────────┘│
│  ┌─────────────────────────────┐│
│  │ 🪙 Echo Rewards          ›  ││
│  │    142.5 ECHO available     ││
│  └─────────────────────────────┘│
│                                 │
├─────────────────────────────────┤
│  SETTINGS                       │
│                                 │
│  👤 Account                  ›  │
│  🔒 Privacy & Security       ›  │
│  🔔 Notifications            ›  │
│  🎨 Appearance               ›  │
│  💾 Storage & Data           ›  │
│  ❓ Help & Support           ›  │
│  ℹ️  About Echo               ›  │
│                                 │
├─────────────────────────────────┤
│  💬     👥      🛡️      👤     │
└─────────────────────────────────┘
```

---

## Screen 2: Edit Profile

### Layout
```
┌─────────────────────────────────┐
│ Cancel   Edit Profile      Save │
├─────────────────────────────────┤
│                                 │
│         ┌─────────┐             │
│         │ [Photo] │             │
│         │   📷    │             │
│         └─────────┘             │
│       Change Photo              │
│                                 │
├─────────────────────────────────┤
│                                 │
│  Display Name                   │
│  ┌─────────────────────────────┐│
│  │ Alex Echo                   ││
│  └─────────────────────────────┘│
│                                 │
│  Username                       │
│  ┌─────────────────────────────┐│
│  │ @alexecho                   ││
│  └─────────────────────────────┘│
│  ✓ Available                    │
│                                 │
│  Bio                            │
│  ┌─────────────────────────────┐│
│  │ Product designer & crypto   ││
│  │ enthusiast. Building the    ││
│  │ future of trusted comms.    ││
│  └─────────────────────────────┘│
│                              78/150│
│                                 │
│  Status                         │
│  ┌─────────────────────────────┐│
│  │ 🚀 Shipping new features    ││
│  └─────────────────────────────┘│
│                                 │
├─────────────────────────────────┤
│  LINKS                          │
│                                 │
│  Website                        │
│  ┌─────────────────────────────┐│
│  │ https://alexecho.dev        ││
│  └─────────────────────────────┘│
│                                 │
│  [+ Add Link]                   │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 3: Personas Management

### Layout
```
┌─────────────────────────────────┐
│ ←        My Personas            │
├─────────────────────────────────┤
│                                 │
│  Personas let you show          │
│  different sides of yourself    │
│  to different contacts.         │
│                                 │
├─────────────────────────────────┤
│  ACTIVE PERSONAS (3/5)          │
│                                 │
│  ┌─────────────────────────────┐│
│  │ 👔 Professional        ★    ││
│  │    Alex Echo                ││
│  │    Product Lead @ Echo      ││
│  │    Default persona          ││
│  │                        Edit ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ 🏠 Personal                 ││
│  │    Alex                     ││
│  │    Just vibing ✌️           ││
│  │    28 contacts can see      ││
│  │                        Edit ││
│  └─────────────────────────────┘│
│                                 │
│  ┌─────────────────────────────┐│
│  │ 👨‍👩‍👧 Family                  ││
│  │    Dad                      ││
│  │    Family stuff only        ││
│  │    12 contacts can see      ││
│  │                        Edit ││
│  └─────────────────────────────┘│
│                                 │
├─────────────────────────────────┤
│                                 │
│  ┌─────────────────────────────┐│
│  │   + Create New Persona      ││
│  └─────────────────────────────┘│
│                                 │
│  You can create up to 5         │
│  personas. 2 remaining.         │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 4: Create/Edit Persona

### Layout
```
┌─────────────────────────────────┐
│ Cancel   New Persona       Save │
├─────────────────────────────────┤
│                                 │
│  PERSONA TYPE                   │
│                                 │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐   │
│  │ 👔 │ │ 🏠 │ │ 👨‍👩‍👧 │ │ 🎮 │   │
│  │Pro │ │Pers│ │Fam │ │Game│   │
│  └────┘ └────┘ └────┘ └────┘   │
│                                 │
│  ┌────┐                         │
│  │ ✨ │ Custom                  │
│  └────┘                         │
│                                 │
├─────────────────────────────────┤
│                                 │
│  Persona Name                   │
│  ┌─────────────────────────────┐│
│  │ Gaming                      ││
│  └─────────────────────────────┘│
│                                 │
│  Display Name                   │
│  ┌─────────────────────────────┐│
│  │ NightOwl42                  ││
│  └─────────────────────────────┘│
│  This is what contacts see      │
│                                 │
│  Avatar                         │
│  ○ Use main avatar              │
│  ● Use different avatar         │
│    ┌─────────┐                  │
│    │ [Photo] │ Change           │
│    └─────────┘                  │
│                                 │
│  Bio                            │
│  ┌─────────────────────────────┐│
│  │ Competitive gamer. Top 500  ││
│  │ in Valorant.                ││
│  └─────────────────────────────┘│
│                                 │
├─────────────────────────────────┤
│  VISIBILITY                     │
│                                 │
│  Who can see this persona?      │
│                                 │
│  ○ All contacts                 │
│  ● Selected contacts only       │
│  ○ No one (hidden)              │
│                                 │
│  [Select Contacts] 0 selected   │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 5: Account Settings

### Layout
```
┌─────────────────────────────────┐
│ ←         Account               │
├─────────────────────────────────┤
│                                 │
│  IDENTITY                       │
│                                 │
│  Phone Number                   │
│  +1 •••••••4567            ›    │
│                                 │
│  Email (optional)               │
│  Not set                   ›    │
│                                 │
│  DID                            │
│  did:cardano:abc1...xyz    📋   │
│                                 │
├─────────────────────────────────┤
│  SECURITY                       │
│                                 │
│  Passkeys                       │
│  2 passkeys registered     ›    │
│                                 │
│  Two-Factor Authentication      │
│  Enabled                   ›    │
│                                 │
│  Active Sessions                │
│  3 devices                 ›    │
│                                 │
├─────────────────────────────────┤
│  RECOVERY                       │
│                                 │
│  Recovery Phrase                │
│  Set up                    ›    │
│                                 │
│  Trusted Recovery Contacts      │
│  2 contacts set up         ›    │
│                                 │
├─────────────────────────────────┤
│  DANGER ZONE                    │
│                                 │
│  ┌─────────────────────────────┐│
│  │  🗑️ Delete Account          ││
│  └─────────────────────────────┘│
│                                 │
└─────────────────────────────────┘
```

---

## Screen 6: Privacy & Security Settings

### Layout
```
┌─────────────────────────────────┐
│ ←    Privacy & Security         │
├─────────────────────────────────┤
│                                 │
│  PROFILE VISIBILITY             │
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
│  MESSAGING                      │
│                                 │
│  Who can message me             │
│  Contacts only             ›    │
│                                 │
│  Read receipts                  │
│  ────────────────────────── ◉   │
│                                 │
│  Typing indicators              │
│  ────────────────────────── ◉   │
│                                 │
├─────────────────────────────────┤
│  CALLS                          │
│                                 │
│  Who can call me                │
│  Trusted & Inner Circle    ›    │
│                                 │
├─────────────────────────────────┤
│  BLOCKCHAIN                     │
│                                 │
│  Anchor messages by default     │
│  ────────────────────────── ○   │
│  Store message proofs on chain  │
│                                 │
│  Show verification badges       │
│  ────────────────────────── ◉   │
│                                 │
├─────────────────────────────────┤
│  SECURITY                       │
│                                 │
│  Screen lock for app            │
│  Immediately               ›    │
│                                 │
│  Hide message previews          │
│  ────────────────────────── ◉   │
│                                 │
│  Screenshot notifications       │
│  ────────────────────────── ◉   │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 7: Notification Settings

### Layout
```
┌─────────────────────────────────┐
│ ←       Notifications           │
├─────────────────────────────────┤
│                                 │
│  MESSAGES                       │
│                                 │
│  Message notifications          │
│  ────────────────────────── ◉   │
│                                 │
│  Show previews                  │
│  Always                    ›    │
│                                 │
│  Sound                          │
│  Echo Default              ›    │
│                                 │
├─────────────────────────────────┤
│  GROUPS                         │
│                                 │
│  Group notifications            │
│  ────────────────────────── ◉   │
│                                 │
│  @Mentions only                 │
│  ────────────────────────── ○   │
│                                 │
├─────────────────────────────────┤
│  CALLS                          │
│                                 │
│  Call notifications             │
│  ────────────────────────── ◉   │
│                                 │
│  Ring sound                     │
│  Reflection                ›    │
│                                 │
├─────────────────────────────────┤
│  OTHER                          │
│                                 │
│  Contact requests               │
│  ────────────────────────── ◉   │
│                                 │
│  Trust score changes            │
│  ────────────────────────── ◉   │
│                                 │
│  Reward notifications           │
│  ────────────────────────── ◉   │
│                                 │
├─────────────────────────────────┤
│  QUIET HOURS                    │
│                                 │
│  Enable quiet hours             │
│  ────────────────────────── ◉   │
│                                 │
│  From: 10:00 PM                 │
│  To:   7:00 AM                  │
│                                 │
│  Allow calls from Inner Circle  │
│  ────────────────────────── ◉   │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 8: Appearance Settings

### Layout
```
┌─────────────────────────────────┐
│ ←        Appearance             │
├─────────────────────────────────┤
│                                 │
│  THEME                          │
│                                 │
│  ┌────────┐┌────────┐┌────────┐ │
│  │ ☀️     ││ 🌙     ││ 📱     │ │
│  │ Light  ││ Dark   ││ System │ │
│  │   ✓    ││        ││        │ │
│  └────────┘└────────┘└────────┘ │
│                                 │
├─────────────────────────────────┤
│  ACCENT COLOR                   │
│                                 │
│  ● Indigo (default)             │
│  ○ Purple                       │
│  ○ Blue                         │
│  ○ Green                        │
│  ○ Orange                       │
│                                 │
├─────────────────────────────────┤
│  CHAT                           │
│                                 │
│  Chat wallpaper                 │
│  Default                   ›    │
│                                 │
│  Message corners                │
│  Rounded                   ›    │
│                                 │
│  Font size                      │
│  ──────●────────────────        │
│  Small    Medium    Large       │
│                                 │
├─────────────────────────────────┤
│  APP ICON                       │
│                                 │
│  ┌────┐ ┌────┐ ┌────┐ ┌────┐   │
│  │ 🟣 │ │ ⚫ │ │ ⚪ │ │ 🟡 │   │
│  │ ✓  │ │    │ │    │ │ ★  │   │
│  └────┘ └────┘ └────┘ └────┘   │
│   Default Dark  Light  Gold     │
│                   ★ Premium     │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 9: Storage & Data

### Layout
```
┌─────────────────────────────────┐
│ ←      Storage & Data           │
├─────────────────────────────────┤
│                                 │
│  STORAGE USED                   │
│                                 │
│  ┌─────────────────────────────┐│
│  │ ████████████░░░░░░░░░░░░░░ ││
│  │ 2.4 GB of 5 GB used         ││
│  └─────────────────────────────┘│
│                                 │
│  Photos & Videos      1.8 GB ›  │
│  Documents            420 MB ›  │
│  Voice Messages       180 MB ›  │
│  Other                 42 MB ›  │
│                                 │
├─────────────────────────────────┤
│  MANAGE STORAGE                 │
│                                 │
│  Auto-download media            │
│  Wi-Fi only               ›     │
│                                 │
│  Media quality                  │
│  Standard                 ›     │
│                                 │
│  Keep media                     │
│  Forever                  ›     │
│                                 │
│  ┌─────────────────────────────┐│
│  │   🗑️ Clear Cache            ││
│  │      156 MB                 ││
│  └─────────────────────────────┘│
│                                 │
├─────────────────────────────────┤
│  NETWORK                        │
│                                 │
│  Use less data for calls        │
│  ────────────────────────── ○   │
│                                 │
│  Proxy settings            ›    │
│                                 │
├─────────────────────────────────┤
│  BACKUP                         │
│                                 │
│  ┌─────────────────────────────┐│
│  │   ☁️ Back Up Now             ││
│  │      Last: Today, 2:30 PM   ││
│  └─────────────────────────────┘│
│                                 │
│  Auto backup                    │
│  Daily                    ›     │
│                                 │
│  Include media                  │
│  ────────────────────────── ◉   │
│                                 │
└─────────────────────────────────┘
```

---

## Screen 10: About & Support

### Layout
```
┌─────────────────────────────────┐
│ ←          About                │
├─────────────────────────────────┤
│                                 │
│         ┌─────────┐             │
│         │  ECHO   │             │
│         │  LOGO   │             │
│         └─────────┘             │
│                                 │
│           Echo                  │
│        Version 1.0.0            │
│                                 │
├─────────────────────────────────┤
│  SUPPORT                        │
│                                 │
│  Help Center                ›   │
│  Contact Support            ›   │
│  Report a Problem           ›   │
│                                 │
├─────────────────────────────────┤
│  LEGAL                          │
│                                 │
│  Terms of Service           ›   │
│  Privacy Policy             ›   │
│  Open Source Licenses       ›   │
│                                 │
├─────────────────────────────────┤
│  DEVELOPER                      │
│                                 │
│  API Documentation          ›   │
│  Create a Bot               ›   │
│                                 │
├─────────────────────────────────┤
│  SOCIAL                         │
│                                 │
│  𝕏 Follow @echoapp          ›   │
│  📰 Blog                     ›   │
│                                 │
├─────────────────────────────────┤
│                                 │
│  Made with 💜 by the Echo Team  │
│                                 │
│  🔗 Powered by Cardano          │
│  🌐 Constellation Hypergraph    │
│                                 │
└─────────────────────────────────┘
```

---

## Component Specifications

### Profile Header Card
```
Avatar: 100px, centered
Name: 24px Bold, flex with badge
Username: 16px Regular #6B7280
Trust Badge: Pill shape, level colored
Edit Button: Secondary style, 36px height
```

### Persona Card
```
Size: 72px × 88px
Background: #FFFFFF
Border: 1px solid #E5E7EB
Border Radius: 12px
Icon: 32px emoji
Label: 12px SemiBold
Selected: Border 2px #6366F1, bg #EEF2FF
```

### Settings Row
```
Height: 56px
Padding: 16px
Label: 16px Regular #1A1A1A
Value/Chevron: 16px #9CA3AF, right aligned
Divider: 1px #F3F4F6
Toggle: iOS-style, 51×31px
```

### Danger Button
```
Background: #FEE2E2
Text: #DC2626
Border: none
Height: 48px
Border Radius: 12px
Icon: Left aligned
```

---

## Interactions

### Edit Profile
- Tap avatar → Action sheet (Take Photo / Choose / Remove)
- Username change → Real-time availability check
- Save → Loading state → Success toast

### Persona Management
- Tap persona → Edit sheet
- Long press → Quick actions (Edit / Delete / Set Default)
- Drag to reorder

### Settings Toggles
- Tap → Immediate toggle with haptic
- Some toggles show confirmation sheet
- Destructive actions require confirmation

---

## Empty States

### No Personas Created
```
Icon: 🎭
Title: "Express different sides of yourself"
Subtitle: "Create personas to show different profiles to different people"
CTA: "Create Your First Persona"
```

### No Recovery Setup
```
Icon: ⚠️
Title: "Protect your account"
Subtitle: "Set up recovery options in case you lose access"
CTA: "Set Up Recovery"
```
