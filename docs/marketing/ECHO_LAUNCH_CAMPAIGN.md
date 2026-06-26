# ECHO Launch — Guerrilla Campaign Brief

Scrappy / near-zero budget · bold & provocative · channels: X, TikTok, Instagram.
Goal: earn outsized attention with ~$0 and convert it to **verified beta sign-ups** (invite-only).

> **Integrity guardrail (read first).** Market only what ships. Mesh = "private beta / rolling out"
> (demo only with a real recording). Payments = don't promise real money movement (mock today). Never
> "unhackable / 100% secure"; say "end-to-end encrypted, no phone number, no servers reading your chats."
> Comparisons must be factual. This protects us from backlash + App Store/ad-policy issues.

## The one line
**"Messaging for humans. Not bots."**  (A/B: "No number. No email. No bots." · "Everyone here is verified.")

## Villains → hooks (each post attacks one)
1. **Bots & scams** → *verified humans only* (lead — most mainstream/viral)
2. **Dead zones** → *works with no internet* (mesh — most visual/earned-media)
3. **Surveillance/data harvesting** → *no phone number, no email* (privacy/sovereignty)

## Audiences (seed → mainstream)
Privacy/crypto X early adopters → scam-weary mainstream → activists/festival/outdoors/preppers → creators/journalists.

## 8-week phase plan
| Phase | Weeks | Lead hook | Big moves |
|---|---|---|---|
| 0 Teaser | −3→−1 | scarcity ("invite-only, verified humans") | waitlist live, seed 50 creators, build-in-public |
| 1 Launch | 0 | anti-scam / no-bots | coordinated drop, Stunt #1, X Space/AMA, Product Hunt |
| 2 Escalate | 1→4 | offline mesh | "Dead Zone" stunt clips, #NoSignalChallenge |
| 3 Sustain | 4→8 | no-phone/email privacy | reactive newsjacking, UGC flywheel |

## Channel roles
- **X** — war room: founder build-in-public, spicy threads, factual comparisons, reply-guy on big
  privacy/scam stories, weekly Spaces, the "real verification vs paid checkmark" angle.
- **TikTok** — virality engine: 7–15s hook clips, scam-bounce demos, POV skits, stitch/duet, trend-jack.
- **Instagram** — Reels (repurpose TikTok) + educational carousels (saves/shares) + Stories (countdowns/UGC).

## Content pillars (post daily, mix)
Enemy · Proof (demos) · Educate · Build-in-public · Community/UGC.  → full copy in `ECHO_CONTENT_PACK.md`.

## Guerrilla stunts (pilot one, scale winners)
Dead-Zone offline demo · "No bots allowed" sticker/QR drops · golden-ticket invites · "scammers hate us"
bounce demos · WhatsApp/Telegram comparison shots · #NoSignalChallenge · reactive newsjack template.

## Product-led growth loops (use what's built)
- Invite-only verified beta = the real **IDV gate** (scarcity is authentic).
- Referral rewards = wire to ECHO's **rewards/staking** system (both sides rewarded).
- "Verified" badge = real **trust tiers** (screenshot-worthy social proof).
- One-tap "add me" = `CurrentUserSession.identityShareURL(...)` deep links.

## KPIs
North-star: **verified activations**. Funnel: impressions → profile → waitlist → install → verify → D7.
Watch shares/saves (not likes), waitlist growth, invite **K-factor** (>1 = self-sustaining).
Tools: native analytics + UTM links + one sheet. Post small → double down on the top 10% of hooks.

## Budget
$0 organic + creator gifting. Tiny spend: stickers/props, CapCut/Canva, link-in-bio + scheduler.
Future paid $ only to amplify proven winners.

## Week-1 deliverables
Waitlist + invite mechanic · 20 posts (this pack) · 3 stunt clips · 50-creator seed list + DMs ·
handles/hashtags · newsjack template. See `ECHO_CONTENT_PACK.md`.

## Landing page + referral/invite — BUILD WHEN READY
> **Blocked on two prerequisites:** (1) a registered **legal entity** (for ToS/Privacy Policy,
> data processing, the email list, and App Store/ad accounts) and (2) a live **website URL/domain**.
> Until both exist, use a no-code waitlist (e.g. a hosted form) behind a placeholder link. When they're
> ready, build the real page **in the repo's `web/` app** with the spec below.

**Goal:** convert campaign traffic → waitlist → verified beta, with a viral referral loop tied to ECHO's
real rewards/trust systems.

**Page (real code, in `web/`):**
- Hero: "Messaging for humans. Not bots." + sub + single CTA (`Request an invite`).
- Capture: email **or** X handle (minimal friction); double opt-in for email (entity/CAN-SPAM/GDPR).
- Proof row: verified-human badge · no number/email · offline (beta) · private AI.
- Footer: ToS + Privacy Policy links (require the legal entity), social handles.

**Referral / invite mechanic:**
- On signup, issue a unique referral code/link (`?ref=CODE`).
- Track referrals server-side; show a **live position + "share to skip ahead"** (golden-ticket scarcity).
- Reward both sides on conversion — wire to ECHO's existing **rewards/staking** system, not a bespoke one.
- Deep-link continuity: the web `?ref=` should map to the in-app **`CurrentUserSession.identityShareURL(...)`**
  / `echo://invite?code=…` path so a referred user lands in the verified-invite flow.

**Compliance gates (need the legal entity):** Privacy Policy + ToS, email consent + unsubscribe,
cookie/analytics notice, accurate claims (mesh = beta, payments = not live), data-deletion path.

**Acceptance:** capture works end-to-end, referral attribution is correct (K-factor measurable),
`?ref=` → in-app invite verified, pages lint/build green in `web-ci`.
