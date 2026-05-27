# Competitive Audit — WhatsApp · Telegram · XChat (Grok) · Signal

**Date:** 2026-05-26
**Scope:** Review four messengers' current (2025–2026) direction and audit features Echo might add, filtered through Echo's privacy-first, decentralized, `did:key` + tokenomics thesis.
**Relationship to other docs:** complements the per-phase work orders (`docs/phase-{2..7}-work-orders.md`) and the Phase 2 gap audit (`docs/PHASE2_GAP_AUDIT.md`). Focuses on **net-new themes** beyond the already-tracked messaging backlog (typing/reactions/calls/search).

---

## 1. Competitor snapshots (2025–2026)

- **XChat (X / Grok)** — launched iOS Apr 2026. E2E encrypted (Rust), **no phone number** (X-account identity), voice/video, disappearing messages, screenshot notifications, zero ads/tracking, super-app ambition. Signature: **in-chat Grok AI assistant**. ⚠ Content sent to Grok is processed **unencrypted** server-side (retention/training unclear) — an explicit E2E exception.
- **Signal** — **Secure Backups** (Signal-hosted, user-held 64-char recovery key, free + paid to 100 GB, cross-platform restore); **multi-device message sync/transfer**; pinned messages; polls; **usernames** (Ristretto25519 + ZK proofs, no phone exposure); member labels. Strictly private, no ads.
- **Telegram** — super-app on TON: **Mini Apps** (≈500M MAU programmable in-chat app platform); **Stars** in-app currency (tips, paid content, gifts, subscriptions); collectible gifts/NFTs; channel ad-revenue share; in-chat **Wallet** (100M+ users).
- **WhatsApp (Meta)** — **usernames** decoupled from phone (~June 2026, Meta-account-verified) + **passkey login**; **Meta AI** (assistant, AI summaries, in-chat translation) with "Private Processing" confidential-compute aim; communities/channels, polls, chat lock; ads in Updates.

## 2. Cross-cutting themes (where the market is moving)
1. **In-chat AI is now table-stakes** for the giants (Grok, Meta AI) — but **both make a privacy tradeoff**.
2. **Identity is decoupling from phone numbers** (all four).
3. **User-held-key encrypted cloud backups + seamless multi-device.**
4. **In-chat economy:** payments, tips, gifts, mini-apps (Telegram leads; XChat super-app next).

## 3. Where Echo is already AHEAD (defend + surface — don't rebuild)
- **Identity:** did:key + Secure-Enclave passkeys + metagraph-anchored `@username` (D1) — phone-free, portable, *self-sovereign*, which the others are only now bolting on. Make onboarding/UX showcase it.
- **No ads / no tracking / E2E by design** — matches Signal/XChat; beats WhatsApp's ad direction.
- **Decentralized transport + metagraph integrity anchoring (D3/D4)** — provable integrity none of them offer.

## 4. Recommended additions (prioritized, privacy-first lens)

### Tier 1 — high impact, strong fit → folded into the roadmap (§6)
- **A. Privacy-preserving AI assistant** — the differentiator. AI that **never breaks E2E**: on-device/local models for summaries/smart-replies/translation; any server-side assist gated by confidential-compute/ZK + explicit per-message consent + zero retention.
- **B. Secure encrypted backups + multi-device sync** (Signal model) — user-held key reusing Echo's BIP-39 recovery; cross-platform restore; did:key-scoped multi-device message sync.
- **C. In-chat ECHO economy** (Telegram model, decentralized) — tips/gifts/paid content via the ECHO token + existing wallet.

### Tier 2 — parity / quick wins
On-device in-chat translation · local AI thread summaries · screenshot notifications · collectible/sticker gifts · member labels · pinned messages (WO-59) · polls (WO-23).

### Tier 3 — platform bets (later)
Mini-app / bot platform (Telegram's 500M-MAU proof) on the Decentralized Bot Framework as *verifiable, sandboxed, privacy-respecting* apps; super-app surface tying messaging + wallet + identity.

## 5. Explicitly DON'T adopt (conflicts with Echo's thesis)
- Ads / tracking / engagement-data monetization.
- AI that silently de-encrypts content (Grok's model) — only consented, private AI.
- Centralized phone-number or Meta-account-style identity verification.

## 6. Tier-1 → Roadmap integration (proposed work orders)

> Provisional IDs (`WO-CA*`) — final WO numbers to be assigned by Software Factory (phase docs are SF-synced; highest existing WO is 291). Each item is mirrored as a "Competitive Audit Additions" note in its phase doc.

| ID | Phase | Title | Extends / blueprint | Scope summary |
|----|-------|-------|---------------------|---------------|
| **WO-CA1** | 7 (Advanced) | Privacy-preserving in-chat AI assistant | Decentralized Bot Framework and Automation | On-device summaries/smart-replies/translation; server-side only via confidential-compute/ZK with explicit per-message consent + zero retention. Never sends plaintext to a cloud model by default. |
| **WO-CA2** | 5 (Privacy) | Consumer secure encrypted backups | WO-64; reuse BIP-39 recovery (WO-234) | User-held recovery key (the 24-word phrase), encrypted cloud backup of chats + opt-in media tiers, cross-platform restore. Server never holds the key. |
| **WO-CA3** | 3 (Messaging Core) | did:key-scoped multi-device message sync | WO-73 (cross-device search-index sync) | Sync message history across a user's registered devices keyed by did:key (controller pattern), E2E re-encrypted per device; no plaintext server copy. |
| **WO-CA4** | 7 (Advanced) + Tokenomics | In-chat ECHO economy (tips / gifts / paid content) | ECHO tokenomics + existing wallet | Decentralized analog to Telegram Stars: send ECHO tips/gifts in chat, paid posts/unlocks, all on-chain via the existing wallet. No custodial in-app currency. |

**Dependencies / sequencing:** WO-CA3 depends on the existing multi-device controller pattern (Phase 2 identity). WO-CA2 depends on recovery (WO-234, done) + backup storage (IPFS/S3 from D3). WO-CA1 depends on the Bot Framework (Phase 7) and an on-device model runtime decision. WO-CA4 depends on the wallet + tokenomics.

## Sources
- XChat / Grok: x-chats.org; IBTimes (Grok messages unencrypted); WinBuzzer (iOS launch / super-app).
- Signal: gHacks + TechCrunch (Secure Backups); AboutSignal (2026 roadmap: pins, polls, member labels, multi-device).
- Telegram: AInvest + TelegramGroups (Mini Apps on TON, Stars, monetization, Wallet).
- WhatsApp: MEF + Times Bull (usernames June 2026, passkeys, Meta AI summaries/translation).
