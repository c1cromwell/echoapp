# Membership Interest Issuance Guide — ECHO Networks LLC

**Entity:** ECHO Networks LLC (a Wyoming limited liability company)
**Document:** How to "Issue Shares" in an LLC — Units, Ledger, and Adding Owners · **DRAFT v0.1**
**Date:** 2026-07-21

> ## ⚠️ Read this first — important disclaimer
>
> **This is a template + explainer, not legal, tax, or securities advice, and creates no professional
> relationship.** Two cautions before you issue anything to anyone other than yourself: (1) issuing
> interests to investors can implicate **securities law** — coordinate with counsel; (2) issuing/gifting
> interests has **tax and valuation** consequences — coordinate with your CPA (and, for a gift to your
> son, the gift-tax rules in `CORPORATE_STRUCTURE_AND_COMPLIANCE.md` §9/§11). Adding a second member also
> changes the Company's tax classification and requires amending the Operating Agreement.

---

## 1. "Shares" vs. LLC units — the key concept
An LLC does **not** have stock or shares. Ownership is expressed as **membership interests**, which this
Company represents as **Membership Units** ("Units"). As the sole Member, **you already own 100%** — you
don't need to "issue yourself shares"; your Units were issued at formation and recorded in the ledger
(see the organizational consent, `ORGANIZATIONAL_CONSENT_DRAFT.md` §6).

- **To bring in another owner, you either:** (a) **issue new Units** to them (which dilutes everyone
  proportionally), or (b) **transfer some of your existing Units** to them. Issuing new Units in exchange
  for a capital contribution is the usual way to raise money; transferring your Units is the usual way to
  gift/sell part of what you own.
- **Percentage vs. unit count.** You can express ownership as a raw percentage, but using a fixed unit
  count (e.g., 10,000,000 Units = 100%) makes future issuances and cap-table math cleaner. Pick one and
  be consistent.

## 2. What controls issuance
- The **Operating Agreement** (`OPERATING_AGREEMENT_DRAFT.md`, Article 7) governs transfer restrictions
  and admission of new Members.
- New Units/Members require the **Manager's action + the Member's written consent**, recorded by a
  **written consent** and reflected on the **membership ledger**.

## 3. Step-by-step: add an owner (or investor)
1. **Decide the deal.** How much of the Company (units/percent), for what (cash, services, IP), at what
   valuation, and with what vesting.
2. **Securities check.** If the person is paying to invest, confirm with counsel how to do it compliantly
   (e.g., an exemption) before taking money.
3. **Tax/valuation check.** Have the CPA confirm treatment (capital contribution, profits interest, or
   gift) and any Section 83(b)/valuation steps.
4. **Amend the Operating Agreement.** Moving from single-member to multi-member converts the tax default
   to **partnership** (or a corporate election) and needs multi-member governance terms — adopt an
   amended/restated Operating Agreement.
5. **Sign an Admission Joinder** (template below): the new Member agrees to be bound by the Operating
   Agreement.
6. **Issue the Units / record the transfer** and, optionally, deliver a **Unit Certificate** (template
   below).
7. **Update the membership ledger** (template below) — this is the Company's cap table of record.
8. **Document by written consent** and keep it in the minute book.
9. **Capital contribution / grant mechanics.** Record cash received, or the profits-interest grant, and
   update capital accounts with the CPA.

## 4. Special case — gifting an interest to your son
- Do it **transparently** (not as a hidden nominee — `CORPORATE_STRUCTURE_AND_COMPLIANCE.md` §4.1): grant
  a documented minority interest (optionally with vesting), or hold your interest in a **revocable living
  trust** naming him as beneficiary for a clean, probate-free transfer (§11).
- A gift of an interest may require a **valuation** and a **gift-tax return (Form 709)**; track the annual
  exclusion and lifetime exemption with the CPA/estate attorney.

## 5. Employee/contractor equity — "profits interests" (the LLC analog of options)
An LLC can't issue stock options, but it can grant a **profits interest** — a right to a share of
**future** appreciation/profits above the value on the grant date, commonly used to reward team members.
- Grant by a **profits-interest award agreement** with a vesting schedule; set a "threshold/hurdle" equal
  to the Company's value at grant so only future value is shared.
- Consider an **83(b) election** at grant (talk to the CPA — timing is strict).
- Keep a pool concept (e.g., reserve up to `[FILL IN: %]` for team) so dilution is planned.

## 6. If you later raise venture capital — convert to a C-corp
The LLC-now path does not block this. When you raise a priced institutional round, investors will almost
always want a **Delaware C-corp** for preferred stock, an option pool, and **QSBS** eligibility.
- **Trigger:** a term sheet for priced equity, or a lead investor requiring a corp.
- **Path:** a statutory conversion or "LLC-to-C-corp" reorganization, exchanging Units for stock; counsel
  and the CPA handle the mechanics and tax. Do it **before** the round closes.
- Until then, membership Units + profits interests cover founder, team, and small/angel investors.

---

## Template A — Membership (Unit) Ledger / Cap Table

| Member | Units held | % ownership | Contribution (cash/IP/services) | Date issued | Certificate # | Notes |
|---|---|---|---|---|---|---|
| `[FILL IN: your legal name]` | `[FILL IN: e.g., 10,000,000]` | 100% | IP Assignment + `[FILL IN: $]` | `[FILL IN]` | 1 | Founder / sole Member |
| | | | | | | |
| **Total** | `[FILL IN]` | 100% | | | | |

*Update this table on every issuance/transfer; it is the Company's cap table of record.*

## Template B — Unit Certificate

```
ECHO NETWORKS LLC — MEMBERSHIP UNIT CERTIFICATE
Certificate No.: ____    Date: ____
This certifies that ____________________ ("Member") is the holder of ______ Membership Units of
ECHO Networks LLC, a Wyoming limited liability company, representing ____% of the Company, subject
to the Company's Operating Agreement (including transfer restrictions).
These Units have not been registered under securities laws and may not be transferred except as
permitted by the Operating Agreement and applicable law.

ECHO Networks LLC
By: ____________________  (Manager)
```

## Template C — Admission of Member (Joinder)

```
JOINDER TO OPERATING AGREEMENT — ECHO NETWORKS LLC
The undersigned, in consideration of the issuance/transfer of ______ Membership Units, hereby
becomes a Member of ECHO Networks LLC and agrees to be bound by the Company's Operating Agreement,
as amended, as fully as if an original party. Contribution: ____________. Vesting (if any): ______.
New Member: ____________________  Date: ______
Accepted — ECHO Networks LLC, By: ____________________ (Manager)  Date: ______
```

## Template D — Profits-Interest Grant (outline)

```
PROFITS INTEREST AWARD — ECHO NETWORKS LLC
Grantee: ____   Grant date: ____   Profits units: ____   Threshold/hurdle value: $____
Vesting: ____ (e.g., 4-year, 1-year cliff)   83(b) filed by grantee: [ ] yes  (within 30 days)
This award entitles the Grantee to future appreciation/profits above the threshold, per the
Operating Agreement and this award. Subject to CPA/counsel review before grant.
```

---

*Cross-references: governed by `OPERATING_AGREEMENT_DRAFT.md` Article 7; initial issuance recorded in
`ORGANIZATIONAL_CONSENT_DRAFT.md` §6; tax/gift mechanics in `CORPORATE_STRUCTURE_AND_COMPLIANCE.md`
§9/§11.*
