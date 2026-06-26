# Echo — Corporate Structure & US Compliance Master Plan

**Prepared for:** Founder, Echo (the "you" throughout)
**Version:** 1.0 · **Date:** 2026-06-26
**Covers:** Echo Message · Echo Comply · Echo Passport · the ECHO token / Network State / Foundation layer

---

> ## ⚠️ Read this first — important disclaimer
>
> **This is an informational planning document. It is NOT legal advice or tax advice, and no
> attorney–client or accountant–client relationship is created by it.** It was prepared to organize
> your decisions and give you a concrete, sequenced plan — *not* to replace licensed counsel. Several
> things you are building (a token with staking and rewards, in-app payments, enterprise/health-data
> compliance, identity/biometric verification) sit on top of **securities, money-transmission, and
> privacy law** where the penalties for getting it wrong are severe. **Before you act on the
> token, payments, HIPAA, or any cross-border step, engage the specific licensed professionals named
> in §12.** Where this document says "do X now," it means the *low-risk formation and hygiene* steps;
> everything carrying real regulatory risk is explicitly routed to counsel.
>
> Laws change and some items below (notably the federal beneficial-ownership reporting rules) were in
> active flux as of mid-2026 — treat every date and rule as "verify before relying."

---

## How to use this document

1. **Read §2 (Executive summary)** for the whole picture in one page.
2. **Do §5 (Formation checklist) now** — that's the part you can safely execute this month. A
   stripped-down version lives in `FORMATION_CHECKLIST.md` for printing.
3. **Skim §6** for the per-product to-do lists and **§7** for what to *defer* (token/Foundation).
4. **Take §12 to your first attorney/CPA meeting** — it's the list of who to hire and what to ask.
5. Fill in the values in the **"Fill these in" box at the end of §5** once, and they flow everywhere.

---

## Table of contents

1. [Where you are today](#1-where-you-are-today)
2. [Executive summary](#2-executive-summary)
3. [Recommended entity architecture](#3-recommended-entity-architecture)
4. [Your four questions, answered](#4-your-four-questions-answered)
5. [Step-by-step formation checklist (DO NOW)](#5-step-by-step-formation-checklist-do-now)
6. [Per-product compliance packets](#6-per-product-compliance-packets)
7. [Token + Foundation track (DEFERRED, pre-architected)](#7-token--foundation-track-deferred-pre-architected)
8. [Anonymity & privacy operations playbook](#8-anonymity--privacy-operations-playbook)
9. [Tax overview](#9-tax-overview)
10. [Compliance calendar + document register](#10-compliance-calendar--document-register)
11. [Informational ideas (appendix)](#11-informational-ideas-appendix)
12. [Who to hire & what to ask](#12-who-to-hire--what-to-ask)
13. [Glossary + sources](#13-glossary--sources)

---

## 1. Where you are today

- You run **one privacy protocol with three products**: **Echo Message** (consumer E2E messaging),
  **Echo Comply** (enterprise compliance messaging / audit / eDiscovery), and **Echo Passport**
  (credential wallet + selective disclosure + in-chat payments).
- You have a **token economy on paper** — ECHO, hard-capped at 1B, with validator staking, per-message
  and referral rewards, trust-tier multipliers, and a proposed transparent fee-split — plus a
  **"Network State"** vision in which users own a piece of the network. Your own docs already name an
  **"Echo Protocol Foundation (Wyoming DUNA)"** as the protocol steward.
- **Legally, almost nothing is formed yet.** The only legal artifacts in the repo are *draft* Terms of
  Service and Privacy Policy (`docs/legal/`) with `[LEGAL ENTITY NAME]` placeholders. There is no
  entity, no EIN, no bank account, no IP assignment, no trademark.
- **Your stated goals:** (a) a real, bankable, tax-filing entity now; (b) stay as anonymous as legally
  possible; (c) keep the non-profit / token / Network State path open without triggering it early;
  (d) a clear per-product compliance plan; (e) a Wyoming LLC as the base.
- **Your decisions (captured 2026-06-26):** home base is a **no-income-tax state**; token timing is
  **undecided/exploring**; you want **minimal structure now, scale later**; deliverable is **one
  master doc + checklists**.

---

## 2. Executive summary

**The recommendation in one line:** *Form a single anonymous **Wyoming holding LLC** now, run all
three products inside it as divisions/DBAs, and stand up product subsidiaries and the non-profit
Foundation only when specific triggers fire.*

This gives you the "real name and entity" a bank and the IRS require, with the least cost and admin,
while preserving Wyoming's strong owner-anonymity and asset protection — and it keeps the token /
Network State path fully open without prematurely creating securities, money-transmission, or
tax-exempt obligations you're not ready to carry.

| Your question | Short answer |
|---|---|
| Wyoming LLC? | **Yes.** It's the right base for anonymity + asset protection. |
| Holding company in my son's (23) name? | **No** — not as a privacy device. It's not real anonymity, it can void liability protection / edge into fraud, and it creates gift-tax + his-creditor risk. Gift him a stake *transparently* instead, or use a trust. |
| Can I stay anonymous? | **Yes on the public record; no to your bank and the IRS.** The *entity* is the public face; your name is disclosed only to your bank (KYC) and the IRS (EIN). |
| Make it a non-profit (for Network State)? | **Not the whole thing.** Keep the products for-profit; create a *separate* non-profit Foundation later, only when the token actually moves. |
| What do I create now? | One WY LLC + EIN + operating agreement + bank account + **IP assignment** + ToS/Privacy filled in + trademark filings + insurance. See §5. |
| What do I defer? | Product subsidiaries, the Foundation, any token sale/listing, any custodial payment feature. See §7. |

> ### ▶ Do-this-now box
> 1. Pick a name; check Wyoming availability.
> 2. Hire a **commercial registered agent** in Wyoming.
> 3. File **Articles of Organization** (manager-managed, for privacy).
> 4. Sign an **Operating Agreement**.
> 5. Get an **EIN** (you are the "responsible party").
> 6. **Assign all existing code + brand into the LLC** (don't skip this).
> 7. Open a **business bank account**.
> 8. Fill ToS/Privacy with the entity name; file **trademarks**; get **insurance**; set up **bookkeeping**.
>
> Everything token-, payment-, or HIPAA-related: **talk to counsel first (§12).**

---

## 3. Recommended entity architecture

### 3.1 The structure

```
        YOU  (true owner / control person)
         │   — disclosed to your BANK (KYC) and the IRS (EIN) only, not to the public
         │   — optionally hold your membership interest through a REVOCABLE LIVING TRUST
         │     (adds privacy + clean estate transfer to your son; see §11)
         ▼
   ┌──────────────────────────────────────────────┐
   │  ECHO HOLDINGS, LLC   (Wyoming)               │   ◀── FORM NOW
   │  • owns ALL IP, the ECHO brand, the contracts │
   │  • holds the bank account; files the taxes    │
   │  • anonymous on the public WY record          │
   │  • commercial registered agent is the only    │
   │    public-facing name                         │
   │  • tax: single-member → disregarded entity;   │
   │    elect S-corp later when profit warrants    │
   └───────────────┬───────────────────────────────┘
                   │  products run as DBAs / internal divisions for now
        ┌──────────┼───────────────┬──────────────────────┐
        ▼          ▼               ▼                      │
  Echo Message  Echo Comply   Echo Passport               │  ◀── spin each into a wholly-owned
  (consumer)    (enterprise)  (credentials + pay)          │      subsidiary LLC WHEN a trigger fires
                   ▲ highest liability                     │      (see §3.3)
                   │ → first to spin out
                   │
   ┌───────────────┴────────────────────────────────┐
   │  ECHO PROTOCOL FOUNDATION        [LATER]        │   ◀── form ONLY when the token actually
   │  • non-profit steward of the protocol + token  │      moves toward distribution
   │  • Network State governance + Privacy Commons   │      (Wyoming DUNA is the lead-fit vehicle;
   │    Treasury (public goods)                      │       alternatives: 501(c)(4) / offshore)
   │  • separate from the for-profit HoldCo          │
   └─────────────────────────────────────────────────┘
```

### 3.2 Why this shape

- **One entity now = fastest path to "bankable + tax-filing."** A bank needs a single legal entity,
  an EIN, formation docs, and an operating agreement. You get all of that with one filing.
- **Holding-company framing protects the valuable thing — the IP.** The LLC *owns the code, the brand,
  and the contracts*. Later, operating subsidiaries license the IP from the HoldCo, so a lawsuit
  against an operating product can't easily reach the crown-jewel IP.
- **Wyoming gives anonymity + charging-order protection** that most states don't. Members and managers
  are not listed on the public Articles, and Wyoming's charging-order rules make it hard for a
  member's personal creditor to seize the company.
- **Divisions now, subsidiaries later** avoids paying for and maintaining four sets of filings,
  registered agents, bank accounts, and tax returns before there's revenue to justify them.
- **For-profit + non-profit two-track** is the standard, battle-tested crypto pattern (operating
  company ships the product; a separate foundation stewards the token and protocol). It also keeps the
  commercial revenue out of any tax-exempt entity, which is required, not optional (see §4 and §7).

### 3.3 Spin-out triggers (when a division becomes its own subsidiary)

Create a wholly-owned subsidiary LLC for a product when **any** of these is true for it:

- It signs **enterprise contracts** that demand a dedicated counterparty, indemnities, or a BAA
  (**Echo Comply** will hit this first — enterprise + potential health data = highest liability).
- It takes **outside investment** earmarked for that product line.
- It introduces a materially different **liability or regulatory profile** (e.g. **Echo Passport**
  turning on real money movement, or custody of financial credentials).
- It generates **standalone revenue** large enough that you want separate books / a separate sale or
  spin-off path.

Until a trigger fires, run the product as a **registered DBA / trade name** under Echo Holdings, LLC.

---

## 4. Your four questions, answered

### 4.1 "Should I put the holding company in my son's (23) name?" → **No — not to stay anonymous.**

Using your adult son as the owner-of-record to hide your involvement is a **nominee arrangement**, and
it backfires on every axis:

- **It isn't real anonymity.** Banks must identify the **beneficial owner** and the **control person**
  under their Customer Identification Program (CIP) / KYC rules, and beneficial-ownership rules look
  through nominal title to whoever actually *controls or economically benefits*. If you run it and earn
  from it, **you are a beneficial owner regardless of whose name is on the certificate** — so you'd
  have to disclose yourself to the bank anyway, defeating the purpose.
- **It can destroy the liability shield.** A company used as an "alter ego," or where ownership is a
  fiction, is exactly what lets a court **pierce the veil** and reach the real operator personally.
- **Done to mislead a bank or regulator, it can be a crime.** Putting a straw owner on bank/CIP
  paperwork to conceal the true owner can be **bank fraud / false statements**. Don't.
- **It creates tax and family-law exposure.** If he truly owns it, the income is taxed to *him* (or
  your funding it is a **taxable gift** with gift-tax-return consequences). And because the asset is
  legally *his*, it's exposed to **his** creditors, lawsuits, divorce, or simply his choice to sell or
  encumber it.

**What to do instead** (covered in §8 and §11):
- Own it yourself, and get practical anonymity from **Wyoming's privacy + a commercial registered
  agent + an organizer service** so your name never appears on a public filing.
- If your real goal is **estate planning / passing it to your son**, do that *transparently*: grant him
  a documented **membership interest**, or place your interest in a **revocable living trust** that
  names him as beneficiary. Both are legitimate, both keep your name off the public record, and a CPA
  handles the gift/valuation mechanics correctly.

### 4.2 "Can I stay anonymous?" → **Yes on the public record. No to your bank and the IRS.**

| Anonymous from… | Achievable? | Why |
|---|---|---|
| Public formation record | **Yes** | Wyoming doesn't list members/managers on the Articles. With a commercial registered agent + organizer service, your name appears nowhere public. |
| App users / ToS / App Store | **Yes (your name)** | The **entity** is named as operator everywhere; *you* are not. (Apple does require the entity's legal name + a D-U-N-S number — that's the entity, not you.) |
| Your bank | **No** | CIP/KYC requires the bank to identify and verify the true beneficial owner + control person. You give the bank your real identity. Non-negotiable. |
| The IRS | **No** | The EIN application (Form SS-4) names a **responsible party** — a real human (you). Taxes flow to real people. |
| FinCEN (CTA / beneficial-ownership) | **Status in flux** | A March 2025 FinCEN interim rule generally **exempted US-formed companies** from beneficial-ownership (BOI) reporting (leaving only certain foreign companies). So a Wyoming LLC likely has **no federal BOI filing today** — but this has been litigated and changed repeatedly; **verify current status before relying** (§8). |

**Bottom line:** you can be invisible to the public and to your users, but you cannot — and must not
try to — be invisible to your bank or the IRS. The strategy is *"the company is the public face; my
name lives only where the law requires it."*

### 4.3 "Should it be a non-profit (for the Network State)?" → **Not the whole thing — use a dual structure.**

- **The commercial products must stay for-profit.** Echo Message subscriptions/VIP, Echo Comply
  enterprise revenue, and Echo Passport rails are commercial income that *benefits you*. A U.S.
  charity (501(c)(3)) is barred from **private inurement** — it cannot exist to enrich its founder, so
  it's the wrong wrapper for your products and your income.
- **The Network State / token / public-goods layer is where a non-profit fits.** A **separate**
  foundation can legitimately steward the protocol, hold and distribute the token under governance,
  and run the **Privacy Commons Treasury** (legal defense of surveilled users, grants, research).
  Keeping it separate is exactly what prevents "user-owned" from quietly becoming "owned by whoever
  buys the company," which is the credibility of your whole ownership thesis.
- **This two-track model is the industry norm** (operating company + protocol foundation). Your own
  docs already assume it.
- **Form it later, on a trigger** (token moving toward distribution). Vehicle options are compared in
  §7; the lead candidate from your docs — a **Wyoming DUNA** — is purpose-built for a token/DAO
  steward.

### 4.4 "Wyoming LLC?" → **Yes.** Plus one nuance for your no-tax home base.

Wyoming is the right base: **no state income tax, strong privacy (no public member list), strong
charging-order protection, low fees, mature LLC case law.** The one decision to make with your
attorney:

- **You operate from a different no-income-tax state.** When you actively run the business from your
  home state, that state generally considers the LLC to be **"doing business"** there, which means you
  either (a) **form in Wyoming and foreign-qualify** the LLC in your home state, or (b) **form directly
  in your home state**.
  - *Form in WY + foreign-qualify at home* → keeps Wyoming as the legal home (privacy/protection) but
    adds a second registration + registered agent, and your home state's public record may show more.
  - *Form directly at home* → simpler/cheaper, but you may lose some of Wyoming's specific privacy and
    charging-order advantages depending on the state.
  - Because your home state has **no income tax**, the usual "Wyoming saves me state tax" argument is
    weak — the real reason to still choose Wyoming here is **privacy + asset protection**, not tax. Have
    your attorney weigh those against the cost of dual registration for *your specific home state*
    (some no-tax states, e.g. FL, publish more owner/manager detail than others).

---

## 5. Step-by-step formation checklist (DO NOW)

Do these in order. Costs are rough 2026 ballparks; confirm current figures.

| # | Step | Who does it | Est. cost | Est. time | Notes |
|---|------|-------------|-----------|-----------|-------|
| 1 | **Choose entity name** + check Wyoming availability; reserve if needed | You / registered agent | $0–60 | 1 day | Pick a name that *doesn't* tie to you personally. Suggested operating entity: **"Echo Holdings, LLC"** (final name is yours to choose). Confirm the matching `.com` + app handles are free. |
| 2 | **Engage a Wyoming commercial registered agent** | You | ~$50–200/yr | 1 day | The agent's address (not yours) is the public contact. Many offer an **organizer + mail/virtual-address** bundle so even the organizer isn't you. |
| 3 | **File Articles of Organization** (choose **manager-managed**) | Registered agent / attorney | $100–150 state fee | 1–3 days | Manager-managed keeps members off the doc and lets a manager (you, or a manager entity) be the operational face. WY does not list members. |
| 4 | **Adopt an Operating Agreement** | Attorney (recommended) | $0 (template) – $1,500 (drafted) | 1–5 days | The internal rulebook: ownership %, management, transfer restrictions, what happens on death (ties to your son/estate plan). Banks ask for it. **Don't rely on a free template for a multi-product, token-bearing company — have counsel draft it.** |
| 5 | **Get an EIN** (Form SS-4) | You (free, IRS) | $0 | Same day online | You are the **responsible party**. Needed for banking + taxes. Don't pay a service for this — it's free. |
| 6 | **Assign ALL existing IP into the LLC** | Attorney | included in §4 or ~$500 | 1–3 days | **Critical and easy to forget.** A signed **IP assignment** moving the codebase, the ECHO brand/logos, the domains, and any prior contractor work *from you personally into Echo Holdings, LLC*. Without this, the company doesn't actually own what it's built on, which breaks future investment, sale, or licensing. |
| 7 | **Open a business bank account** | You | $0–25/mo | 1–3 days | Bring: Articles, EIN letter, Operating Agreement, your ID. **You will disclose yourself as beneficial owner — that's normal.** Consider a crypto-friendly / tech-friendly bank given the token roadmap. |
| 8 | **Fill in ToS + Privacy** with the entity name | You + privacy counsel | reuse existing drafts | 1 day | Replace `[LEGAL ENTITY NAME]` → **Echo Holdings, LLC** in `docs/legal/TERMS_OF_SERVICE_DRAFT.md` and `PRIVACY_POLICY_DRAFT.md`; have privacy counsel finalize before publishing. |
| 9 | **Set up bookkeeping + engage a CPA** | You + CPA | ~$50–300/mo | ongoing | Separate books from day one (commingling is the #1 way to lose the liability shield). The CPA also advises the S-corp election timing (§9) and token-tax issues. |
| 10 | **File trademarks** (ECHO word mark + logos) | Trademark attorney | ~$250–350/class gov + counsel | filing now, ~8–12 mo to register | Protect "ECHO," "Echo Message," "Echo Comply," "Echo Passport," and the ripple logo. File in the right classes (software, financial/messaging services). |
| 11 | **Get insurance** (general liability + tech E&O / cyber) | Broker | ~$1–5k/yr to start | 1–2 weeks | Tech E&O / cyber matters a lot for a messaging + identity + payments product. |
| 12 | **Stand up privacy-ops hygiene** | You | low | 1 week | Virtual business address, business phone/VoIP, role-based emails (legal@, privacy@, security@), DMCA agent registration. See §8. |

> ### 📝 Fill-these-in box (set once; they flow into every doc)
> - **Final entity name:** ________________________ (default: *Echo Holdings, LLC*)
> - **Formation state:** Wyoming  ·  **Home/operating state:** ____________ (your no-tax state)
> - **Responsible party (EIN) / control person:** _______________ (you)
> - **Registered agent:** _______________  ·  **Virtual business address:** _______________
> - **Legal contact email:** legal@__________  ·  **Privacy contact:** privacy@__________
> - **Public website:** __________

---

## 6. Per-product compliance packets

Each product lists **Now** (do during/just after formation), **Near-term** (before broad launch), and
**Trigger-based** (only when a condition fires).

### 6.0 Cross-cutting (applies to all three)

- **Now:** entity-named **ToS + Privacy Policy** (fill the existing drafts); **Data Processing
  Addenda (DPAs)** with every processor (cloud, IDV, analytics, APNs); **IP assignment** (§5 step 6);
  **trademark** filings; **App Store** setup (entity legal name + D-U-N-S; app privacy "nutrition
  labels"); **OFAC/sanctions screening** posture (don't serve embargoed jurisdictions/persons);
  **insurance**.
- **Near-term:** **CCPA/CPRA** (California) + other US state privacy laws (Virginia, Colorado, Texas,
  etc.) readiness — privacy rights request flow, "do not sell/share" (you already don't sell);
  **GDPR/UK GDPR** if you take EU/UK users (legal bases, DPA, possibly an EU representative);
  **accessibility** (ADA/WCAG) for the web surfaces; a **vulnerability disclosure / security** policy.
- **Trigger-based:** **SOC 2** once enterprise buyers ask (Comply will drive this).

### 6.1 Echo Message (consumer E2E messaging)

- **Now:**
  - **Age gating / COPPA:** your ToS draft sets 18+; enforce it (under-13 data triggers COPPA, and
    teen-privacy rules are tightening). Keep the floor explicit.
  - **Law-enforcement response policy:** a written process for subpoenas/warrants. Because messages are
    **E2E encrypted and content-blind**, you *can't* produce content — but you can receive and must
    handle legal process for the limited metadata you hold. Document what you have and don't have.
  - **CSAM / NCMEC reporting:** US providers of messaging services have **mandatory reporting duties**
    to NCMEC for known child-sexual-abuse material. E2E encryption doesn't remove the legal duty to
    report what you *do* become aware of — build the reporting path and policy. **This is a hard legal
    obligation, not optional.** Get counsel on how it interacts with your encryption model.
  - **Abuse / acceptable-use policy** + reporting/blocking flow (you already have trust-tier + report
    primitives); **DMCA agent** registration for any user-generated/shared content.
- **Near-term:** transparency-report posture; retention schedule for the minimal metadata you keep;
  jurisdiction/age verification for any region that mandates it.
- **Trigger-based:** if you add **broadcast channels / large groups**, revisit platform-liability and
  content-moderation obligations.

### 6.2 Echo Comply (enterprise compliance / audit / eDiscovery)

> This is your **highest-liability product** and the **first spin-out candidate** (§3.3).

- **Now:**
  - **Enterprise contract templates:** MSA, order form, **DPA**, SLA, mutual NDA, **indemnification**
    and limitation-of-liability terms sized for enterprise risk.
  - Position the audit/evidence backend's guarantees carefully in writing (you store hashes/CIDs and
    aggregate metrics, **never message content** — make that contractual, since it's also your defense).
- **Near-term:**
  - **HIPAA capability:** if any customer handles PHI, you become a **Business Associate** and must be
    able to sign a **BAA** and meet the Security Rule. Decide deliberately whether you're "HIPAA-ready"
    or explicitly out of scope — don't drift into it.
  - **SOC 2 Type II** roadmap (enterprise buyers will require it); vendor security questionnaires.
  - **eDiscovery / litigation-hold / retention** contractual terms and the matching product controls
    (these are the core value and the core obligation).
- **Trigger-based:**
  - **Spin into `Echo Comply, LLC`** at the first enterprise BAA or material enterprise contract.
  - Sector add-ons as you sell into them: **FOIA** workflows (gov), **FINRA/SEC 17a-4** WORM-retention
    (finance), legal-hold standards (law firms).

### 6.3 Echo Passport (credential wallet + selective disclosure + payments)

- **Now:**
  - **Biometric-privacy laws** for selfie/liveness/IDV: **Illinois BIPA** (private right of action —
    real class-action risk), **Texas CUBI**, **Washington**, and a growing list. You need **explicit
    biometric consent**, a retention/destruction schedule, and ideally to keep raw biometrics with the
    **IDV vendor**, not yourself (your design already does this — document it).
  - **IDV vendor contract + DPA**; surface the vendor's role in your Privacy Policy (the draft already
    has the placeholder).
- **Near-term:**
  - **GLBA** considerations if Passport touches financial-account credentials.
  - **Document the non-custodial invariant** as a hard rule (see §7): Echo is the *identity + consent*
    layer, never the holder of funds or raw card data.
- **Trigger-based (get counsel BEFORE building, not after):**
  - **Any real money movement** → **money-transmission analysis** (federal FinCEN MSB + state
    money-transmitter licensing). Native non-custodial P2P and licensed-partner rails are the design
    that *avoids* this — but the line is fact-specific and must be cleared by money-transmission
    counsel **before** the feature ships.
  - **Custody of credentials/keys/funds** → KYC/AML program obligations.
  - Spin into **`Echo Passport, LLC`** when payments turn on.

---

## 7. Token + Foundation track (DEFERRED, pre-architected)

This is the **highest-regulatory-risk** part of Echo. Because your token timing is *undecided*, the
plan is: **architect it, but don't trigger it.** Here's how to stay clean until you're ready.

### 7.1 Securities risk — the thing to respect most

A token that people **stake for "5–15% APY,"** earn as **rewards**, and that pays a **fee-split** has
several hallmarks of an **investment contract** under the **Howey** test (money invested in a common
enterprise with an expectation of profit from others' efforts). If ECHO is deemed a security and you
distributed/sold it without registration or an exemption, the consequences are severe.

**What NOT to do now (until securities counsel clears a path):**
- ❌ **Don't sell, list, or publicly distribute the token.**
- ❌ **Don't advertise "APY," yield, returns, or price appreciation** — drop "5–15% APY" language from
  any public-facing material; describe rewards as **functional/usage-based**, not investment yield.
- ❌ **Don't let rewards be freely transferable/tradeable** before clearance — the closer it looks to a
  tradeable yield-bearing asset, the more it looks like a security.
- ❌ **Don't presell, do a SAFT, or take "investment" for tokens** without securities counsel.

**What you CAN do now:** keep building the *functional* mechanics (rewards accounting, trust tiers)
internally, framed as product features, ideally **non-transferable / closed-loop** until cleared.

### 7.2 Money-transmission / MSB posture (payments)

Your design intent is exactly right and should be locked as a **hard invariant**:

> **Echo is the identity + consent layer, never the money transmitter.** Users self-custody
> (Echo never controls their keys or funds); merchant/fiat flows route through **licensed partners**
> (card-network tokenization / Open-Banking PISP); Echo never touches a raw PAN or holds customer
> funds.

If Echo *did* control or transmit user funds, it would likely be a **FinCEN-registered MSB** and need
**state money-transmitter licenses** (a very expensive, multi-state, bonded undertaking). The
non-custodial + licensed-partner design is what keeps you out of that — but **get a money-transmission
legal opinion before any pay feature ships**, because the analysis is fact-specific.

### 7.3 The Foundation — when and which vehicle

**Form the Foundation only when a trigger fires:** you decide to actually distribute the token, launch
public staking, or hand protocol governance to the community. Until then, leave it as a documented
plan (this section).

| Vehicle | What it is | Pros | Cons / watch-outs |
|---|---|---|---|
| **Wyoming DUNA** *(lead fit; your docs assume it)* | Decentralized Unincorporated Nonprofit Association — Wyoming statute purpose-built for DAOs/token stewards | Legal personhood (can hold assets, contract, sue/be sued); limited liability for members; designed for token governance; can be tax-transparent | Newer law, less precedent; needs careful governance docs; tax treatment needs CPA |
| **501(c)(4) social welfare** | Federal tax-exempt non-profit for social-welfare/advocacy | Tax-exempt; broad advocacy latitude; fits a "public good" mission | Donations **not** tax-deductible; can't primarily benefit private parties; IRS scrutiny |
| **501(c)(3) charity** | Federal charitable non-profit | Donations deductible; strong public-trust signal | Very restrictive; **no private inurement**; poor fit for a token-stewarding org; slow to obtain |
| **Offshore foundation** (e.g. Cayman, Zug, Panama) | Foundation company / Stiftung used widely by crypto protocols | Jurisdiction-neutral; established token-foundation playbook; can isolate token from US operating co | Cost + complexity; cross-border tax (CFC/PFIC) issues; US-person reporting; needs specialized counsel |

**Recommendation:** plan for the **Wyoming DUNA** as primary (it matches your stack and keeps things
onshore/simple), and have **crypto/securities counsel** pressure-test DUNA-vs-offshore at the moment
the token actually moves — the right answer depends on where your users and validators are and how the
token is distributed.

### 7.4 Token tax (flag for the CPA)

Token issuance, treasury holdings, rewards, and staking each have **income-recognition, basis, and
timing** consequences for both the entity and recipients. Don't move tokens — even internally — without
the CPA modeling the tax. (More in §9.)

---

## 8. Anonymity & privacy operations playbook

Goal: **the entity is the public face; your name appears only where the law requires (bank, IRS).**

1. **Commercial registered agent** — their WY address is the public contact, not yours.
2. **Organizer service** — have the agent/attorney act as "organizer" so even the formation filer
   isn't you.
3. **Manager-managed LLC** — members aren't named; if you want extra distance, a separate manager
   entity can be the operational signatory.
4. **Virtual business address** (not your home) on every public surface, app store listing, WHOIS,
   ToS/Privacy contact.
5. **Domain privacy / WHOIS protection** on all domains; register them in the **entity's** name.
6. **Role-based emails** (legal@, privacy@, security@) — never your personal email on public docs.
7. **The entity, not you, signs everything** — contracts, app store, vendor agreements.
8. **EIN responsible party** — this is you, disclosed to the IRS only. Normal and required.
9. **Bank KYC** — you disclose yourself as beneficial owner + control person. Normal and required.
   *Do not attempt to hide this — that's where anonymity stops being legal.*
10. **FinCEN CTA / beneficial-ownership (BOI):** as of mid-2026 a March-2025 interim rule generally
    **exempted US-formed companies** from BOI reporting. **This area has flipped repeatedly in
    litigation — re-verify the current rule with counsel at formation**, and if reporting is required,
    file it (it goes to FinCEN, not the public).
11. **Optional: hold your interest through a revocable living trust** — adds a privacy layer and gives
    you a clean estate path to your son without a nominee (§11).

---

## 9. Tax overview

> Confirm everything here with your CPA — this is orientation, not a return.

- **Default entity tax (single-member LLC):** a **disregarded entity** — profit/loss flows onto your
  personal return (Schedule C). Simple to start.
- **S-corp election (Form 2553) — later trigger:** once the business throws off meaningful profit,
  electing **S-corp** treatment can cut self-employment tax (you pay yourself a "reasonable salary,"
  the rest is a distribution). Your CPA models the breakeven (often around the point where net profit
  comfortably exceeds a reasonable salary). **Don't elect prematurely** — it adds payroll + filing
  overhead.
- **C-corp — only if you raise VC.** If you take institutional venture money, investors usually want a
  **Delaware C-corp** (QSBS, preferred stock, option pools). That's a *conversion* decision for later,
  not now. The WY-LLC-now path doesn't block it.
- **State tax:** your home state has **no income tax**, so there's no personal state income tax on the
  pass-through — a real simplification. Watch for: **foreign-qualification fees/annual reports** if you
  WY-form + home-qualify; sales/use tax if you sell taxable goods/services; any **gross-receipts**
  quirks of your specific state.
- **Wyoming:** no state income tax; an **annual report + license-tax** (small, asset-based) is due
  yearly to keep the LLC in good standing.
- **Token tax (flag, not advice):** tokens received as rewards are generally **income at fair value
  when received**, establishing basis; later disposal is gain/loss; a treasury holding tokens has its
  own accounting. **Model this with the CPA before any token moves.**
- **Son / gifting (only if you pursue it transparently):** granting your son a membership interest is a
  **gift** — track the annual exclusion + lifetime exemption, file a **gift-tax return (Form 709)** if
  required, and get a **valuation**. A trust changes the mechanics. CPA + estate attorney territory.
- **R&D:** software development costs and the **R&D credit** may apply — keep clean records so the CPA
  can capture them.
- **Recordkeeping discipline:** separate bank account, no commingling, contemporaneous books,
  board/member resolutions for big decisions. This is also what *preserves the liability shield.*

---

## 10. Compliance calendar + document register

### 10.1 Recurring obligations

| Item | Frequency | Owner | Notes |
|---|---|---|---|
| Wyoming annual report + license tax | Annual (on formation anniversary) | Registered agent / you | Keeps LLC in good standing |
| Registered-agent renewal | Annual | You | Don't lapse — it's your legal-service address |
| Home-state foreign-qual annual report (if applicable) | Annual | You | Only if you WY-form + home-qualify |
| Federal + state tax filings | Annual / quarterly estimates | CPA | Estimated taxes quarterly |
| Trademark maintenance | Per USPTO deadlines (≈ yrs 5–6, 9–10) | TM attorney | Don't miss declarations of use |
| Insurance renewal | Annual | Broker | Re-scope as products grow |
| Privacy policy / ToS review | Annual or on feature change | Privacy counsel | Update when payments/token go live |
| FinCEN BOI status check | Re-verify periodically | Counsel | Rule has changed repeatedly |
| DPA / vendor review | Annual | You | Especially IDV + cloud |

### 10.2 Document register (keep originals safe + a digital set)

Formation: Articles of Organization · Operating Agreement · EIN letter (CP-575) · registered-agent
agreement · DBA/trade-name registrations.
IP: **IP assignment(s)** · trademark filings/registrations · domain records · open-source license
inventory.
Commercial: ToS · Privacy Policy · DPAs · enterprise MSA/DPA/SLA templates (Comply) · IDV vendor
contract (Passport).
Financial: bank records · bookkeeping · tax returns · cap table / member ledger.
Deferred (when triggered): Foundation governing docs · token legal opinion · money-transmission
opinion · BAA template (Comply) · SOC 2 reports.

---

## 11. Informational ideas (appendix)

A few extra ideas you asked for — optional, but worth knowing:

1. **Separate IP-holding subsidiary.** Once there's value, an `Echo IP, LLC` that *owns* the IP and
   licenses it to the operating entities further isolates the crown jewels from operating liability.
   (Overkill today; revisit at the first subsidiary spin-out.)
2. **Revocable living trust as the LLC member.** Hold your membership interest in a trust: keeps your
   name off the public record, lets the company pass to your son **without probate**, and avoids the
   nominee problems of §4.1. Pair with the Operating Agreement's transfer-on-death terms.
3. **Transparent equity for your son.** If you want him involved, give him a *real, documented*
   minority interest with a vesting schedule, or make him a beneficiary of the trust — legitimate,
   and a far better story than a hidden nominee if anyone ever looks.
4. **Trademark + brand defense early.** File the marks *before* launch noise; secure social handles +
   defensive domains in the entity's name.
5. **DUNA timing.** Stand up the Foundation in the *same window* you finalize token mechanics, so
   governance and treasury exist before the first public distribution — not after.
6. **Onshore vs offshore Foundation.** If your validator/user base becomes heavily non-US, revisit an
   offshore foundation with counsel; if it stays US-centric, the Wyoming DUNA is simpler and cleaner.
7. **Privacy Commons Treasury as a public-goods flywheel.** Funding legal defense/research from
   protocol fees is a strong differentiator *and* a credibility anchor for the "un-ruggable" ownership
   thesis — but it lives in the Foundation, not the for-profit.
8. **Insurance stack as you grow:** add D&O once you have a board/outside money; raise cyber/E&O limits
   as Comply enterprise deals land.
9. **Data-residency / sub-processor map.** Keep a living list of where data sits and who processes it;
   enterprise (Comply) buyers and EU users will both ask.

---

## 12. Who to hire & what to ask

Bring this list to each. Rough budgets are early-stage ballparks.

| Specialist | Why | Ask them | Rough budget |
|---|---|---|---|
| **Wyoming business-formation attorney** | Form the LLC correctly; draft the Operating Agreement; advise WY-vs-home formation + foreign qualification | "WY-form + foreign-qualify in [my state], or form directly? Manager-managed for privacy? Operating Agreement with transfer-on-death to my son/trust? IP assignment from me to the LLC?" | $1.5–4k formation package |
| **Crypto / securities counsel** | The token. Whether ECHO is a security; how/whether to distribute; Foundation vehicle (DUNA vs offshore) | "Is our reward/staking model a security? What can we say publicly? What's a compliant distribution path? DUNA vs offshore foundation for our user base?" | $10k+ for an analysis/opinion |
| **Money-transmission counsel** | Before any payment feature ships | "Does our non-custodial + licensed-partner design avoid MSB/state MTL? What triggers licensing? Review the pay-in-chat flow." | $5–15k for an opinion |
| **Privacy / data counsel** | Finalize ToS + Privacy; CCPA/CPRA/GDPR; biometric (BIPA) for Passport; DPAs | "Finalize our drafts for the entity; biometric-consent flow for IDV; state-privacy + GDPR readiness; processor DPAs." | $3–10k initial |
| **CPA (with crypto experience)** | Entity tax election timing; books; token tax; gift/estate if you involve your son | "Disregarded vs S-corp election timing? Token-reward and treasury accounting? Quarterly estimates? Gift mechanics if I grant my son an interest?" | $50–300/mo + returns |
| **Trademark attorney** | File ECHO marks + logos in the right classes | "File word + logo marks for ECHO and the three products in the correct classes; clearance search first." | ~$1.5–3k for the set |
| *(Later)* **Estate attorney** | The revocable trust + clean transfer to your son | "Set up a revocable living trust to hold my LLC interest and pass it to my son without probate." | $1.5–4k |

---

## 13. Glossary + sources

**Glossary**

- **Beneficial owner** — the real person(s) who own/control/benefit from an entity, regardless of whose
  name is on paper. Disclosed to banks (KYC) and, when required, FinCEN.
- **BOI / CTA** — Beneficial Ownership Information reporting under the Corporate Transparency Act
  (status in flux as of 2026; US-formed companies largely exempted by a March-2025 interim rule).
- **Charging-order protection** — limits a member's personal creditor to a "charging order" against
  distributions rather than seizing the LLC; strong in Wyoming.
- **CIP / KYC** — Customer Identification Program / Know-Your-Customer; banks must verify who you are.
- **DUNA** — Decentralized Unincorporated Nonprofit Association; Wyoming statute for DAO/token stewards.
- **Disregarded entity** — a single-member LLC taxed on the owner's personal return by default.
- **Howey test** — the US test for whether something is an "investment contract" (a security).
- **MSB / money transmitter** — a regulated money-services business; federal FinCEN registration +
  state licenses; triggered by controlling/transmitting customer funds.
- **Nominee / straw owner** — putting someone else's name on ownership to hide the real owner; legally
  risky and, if used to deceive, potentially fraudulent.
- **Private inurement** — a charity benefiting a private individual (e.g. the founder); prohibited for
  501(c)(3).
- **Veil piercing / alter ego** — when courts disregard the LLC's liability shield because it wasn't
  respected as a separate entity.

**Internal sources used (this repo)**

- `docs/NETWORK_STATE_OWNERSHIP_THESIS.md` — token, validators, Foundation, Privacy Commons Treasury.
- `docs/ECHO_PASSPORT_PLAN.md` — custody/non-custodial invariant, payments routing, "Echo stays the
  identity + consent layer, never the money transmitter."
- `docs/CROSS_PRODUCT_GAP_REVIEW.md` — the three products + Passport; Comply enterprise surface.
- `docs/PRD.md` — token rewards / staking "5–15% APY," success metrics, roadmap.
- `docs/legal/TERMS_OF_SERVICE_DRAFT.md`, `docs/legal/PRIVACY_POLICY_DRAFT.md` — existing drafts to
  fill with the entity name.

**External topics to verify with counsel (not legal citations — confirmation prompts):** Wyoming LLC
Act + DUNA statute; FinCEN CTA/BOI current status; FinCEN MSB / state money-transmitter licensing; SEC
Howey/investment-contract guidance; Illinois BIPA / Texas CUBI / Washington biometric laws; HIPAA
Business Associate rules; CCPA/CPRA + state privacy laws; GDPR/UK GDPR; NCMEC/CSAM reporting duties for
messaging providers; USPTO trademark classes.

---

*End of document. This is an informational planning aid, not legal or tax advice. Engage the licensed
professionals in §12 before acting on the token, payments, HIPAA, or cross-border items.*
