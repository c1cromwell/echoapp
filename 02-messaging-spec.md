# Echo UX Specifications
## Feature 2: Messaging & Conversations

---

# OVERVIEW

The messaging experience is the core of Echo. It combines familiar chat patterns with unique blockchain-verified features like provable messages, trust indicators, and encrypted communication.

---

# SCREEN SPECIFICATIONS

## 2.1 Conversations List

| Element | Specification |
|---------|--------------|
| Nav Title | "Messages", 20px Bold |
| Search Height | 36px, Gray 50 bg, 10px radius |
| Item Height | 76px |
| Avatar Size | 56px |
| Name | 16px SemiBold |
| Preview | 14px Regular, Gray 500, truncate |
| Timestamp | 13px Regular, Gray 400 |
| Unread Badge | Min 20px, Primary bg, 12px bold white |
| FAB | 56px, Primary bg, 20px from edges |

## 2.2 Chat Thread

### Header
| Element | Specification |
|---------|--------------|
| Height | 56px |
| Avatar | 40px |
| Name | 16px SemiBold |
| Status | 13px, Success if online |
| Actions | 40px tap targets |

### Messages
| Property | Sent | Received |
|----------|------|----------|
| Background | Primary | Gray 50 |
| Text | White | Gray 900 |
| Radius | 20/20/4/20 | 20/20/20/4 |
| Max Width | 75% | 75% |
| Padding | 10px 14px | 10px 14px |
| Font | 16px/1.4 | 16px/1.4 |

### Input Area
| Element | Specification |
|---------|--------------|
| Min Height | 52px |
| Input bg | Gray 50, 20px radius |
| Placeholder | "Message", Gray 400 |
| Send (empty) | Mic icon, Primary |
| Send (text) | Arrow, White on Primary |

---

# SPECIAL MESSAGE TYPES

- Voice: Play button + waveform + duration
- Image: 240px max, 16px radius
- File: Icon + name + size
- Link: Preview card with image/title/description
- Poll: Question + options with progress bars

---

# KEY INTERACTIONS

- Long press message → Reactions + Context menu
- Swipe conversation → Quick actions
- Pull down → Load earlier messages
- Tap proof badge → View blockchain verification

