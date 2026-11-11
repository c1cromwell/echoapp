# Company Policies — ECHO Networks LLC

**Entity:** ECHO Networks LLC (a Wyoming limited liability company)
**Document:** Core Internal Policies Bundle · **DRAFT v0.1** · **Date:** 2026-07-21

> ## ⚠️ Read this first — important disclaimer
>
> **This is a template draft, not legal advice, and it creates no attorney–client relationship.** These
> are lightweight internal hygiene policies — the everyday practices that keep your liability shield
> intact. They are **not** the regulated, product-specific compliance packets (COPPA, HIPAA, BIPA,
> money-transmission, securities); those live in `CORPORATE_STRUCTURE_AND_COMPLIANCE.md` §6–§7 and are
> routed to specialist counsel. Adopt these by the organizational consent, follow them, and revisit as
> the Company grows.

---

## Policy 1 — Recordkeeping & No-Commingling (the shield-preserving policy)

**Why this matters most:** for a single-member LLC, the biggest risk is a court "piercing the veil" and
reaching you personally on the theory that the Company wasn't a real separate entity. Commingling funds is
the #1 way that happens. These rules are the defense.

- **Separate accounts.** The Company keeps its own bank and card accounts. Personal and Company money
  never mix — in either direction.
- **Company pays Company costs; you pay personal costs.** If you pay a Company expense personally, record
  it and reimburse through the expense policy; never pay a personal expense from a Company account —
  instead take a documented **distribution** to yourself, then pay it personally.
- **Contract and hold in the Company's name.** Sign as "ECHO Networks LLC, by [name], Manager." Hold IP,
  domains, and property in the Company's name.
- **Books from day one.** Maintain contemporaneous bookkeeping (accounting software + a CPA), a general
  ledger, and a document register (`CORPORATE_STRUCTURE_AND_COMPLIANCE.md` §10.2).
- **Minute book.** Keep signed originals of the Articles, Operating Agreement, EIN letter (CP-575),
  IP Assignment, organizational consent, and later written consents together, physically and digitally.
- **Material decisions in writing.** Document significant actions (new accounts, big contracts, new
  members, elections) by written consent.
- **Retention.** Keep formation and tax records permanently; keep operational records at least
  `[FILL IN: e.g., 7]` years or as counsel/CPA advise.

## Policy 2 — Expense & Reimbursement

- **Business purpose required.** Only ordinary, necessary business expenses are Company expenses.
- **Documentation.** Keep receipts and note the business purpose; use the Company card where possible so
  the record is automatic.
- **Reimbursements.** Founder/contractor out-of-pocket Company expenses are reimbursed on a documented
  request; reimbursements are not compensation.
- **No personal use.** Company accounts and cards are not for personal spending.

## Policy 3 — Conflict of Interest

- **Disclose.** The Manager/Founder discloses any transaction between the Company and the Founder (or a
  related party or Founder-controlled entity), including a manager-entity arrangement or the Founder
  Services Agreement.
- **Arm's-length terms.** Related-party transactions are on fair, documented terms and recorded by written
  consent.
- **Corporate opportunity.** Business opportunities in the Company's line are offered to the Company
  first.

## Policy 4 — Data Handling & Security (stub)

> This is a short internal stub. The user-facing commitments live in the Privacy Policy
> (`PRIVACY_POLICY_DRAFT.md`) and Terms (`TERMS_OF_SERVICE_DRAFT.md`); product-specific privacy law
> (CCPA/CPRA, GDPR, BIPA, COPPA, HIPAA) is handled in `CORPORATE_STRUCTURE_AND_COMPLIANCE.md` §6 with
> privacy counsel.

- **Data minimization.** Collect and retain the least data necessary; Echo's E2E-encrypted, content-blind
  design is a core control — preserve it.
- **Access control.** Limit access to systems, keys, and secrets; use strong auth; no shared credentials.
- **Vendor/processor discipline.** Sign a **DPA** with every processor (cloud, IDV, analytics, APNs);
  keep a sub-processor/data-residency map.
- **Incident response.** Maintain a basic process to detect, contain, assess, and (where legally required)
  notify on a security incident; preserve the mandatory reporting paths (e.g., NCMEC/CSAM for messaging —
  `CORPORATE_STRUCTURE_AND_COMPLIANCE.md` §6.1).
- **Secrets & keys.** Never commit secrets to the repo; store keys in a secrets manager / secure enclave
  per the product's security design.

## Policy 5 — Acceptable Signing Authority (quick reference)

- The **Manager** signs for the Company. The Member, acting as Member, does not bind the Company.
- Anything creating material liability (enterprise contracts, BAAs, financing, token/payment features)
  gets the review routed in `CORPORATE_STRUCTURE_AND_COMPLIANCE.md` §6–§7 & §12 **before** signing.

---

**Adopted by:** the organizational written consent (`ORGANIZATIONAL_CONSENT_DRAFT.md`).

*Cross-references: preserves the shield asserted in `OPERATING_AGREEMENT_DRAFT.md` Article 8; expense and
compensation tie to `FOUNDER_SERVICES_AGREEMENT_DRAFT.md`. Master plan:
`CORPORATE_STRUCTURE_AND_COMPLIANCE.md` §8–§10.*
