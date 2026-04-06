# ECHO Tokenomics & Founder Allocation

## Document Info

| Field | Value |
|-------|-------|
| Version | 2.0 |
| Date | March 26, 2026 |
| Status | Implementation-ready for Phase 2 genesis |
| Dependencies | Tessellation v3 (TokenLock, StakeDelegation), Currency L1 Scala validation |
| Changes from v1.0 | Revised founder allocation to 18% (CEO 10%, co-founders 2% each). Treasury reduced to 22%. Added trust-tier weighted governance model. Added Midnight blockchain evaluation (Phase 3+). Single-token governance decision documented. |

---

## 1. Total Supply & Allocation

**Total ECHO Supply: 1,000,000,000 (1 billion)**

Fixed supply minted at genesis (Phase 2 mainnet launch). No additional minting after genesis. Deflationary via Phase 5 ECHO burn program.

| Allocation | % | Tokens | Unlock Schedule | Purpose |
|-----------|---|--------|----------------|---------|
| **Community Rewards** | 40% | 400,000,000 | Emitted over 10 years via declining emission curve | Messaging rewards, referrals, staking APY, governance incentives |
| **Treasury** | 22% | 220,000,000 | Controlled by governance (Phase 4+) / founders multi-sig (Phase 1–3) | PacaSwap liquidity, DAG staking collateral, Digital Evidence subscriptions, operations, Phase 5–6 |
| **Founders** | 18% | 180,000,000 | 4-year vesting with 1-year cliff (see Section 2) | Founding team compensation |
| **Future Team & Advisors** | 10% | 100,000,000 | Reserved; same vesting terms when allocated | Future hires, Scala developers, advisors, legal counsel |
| **Ecosystem & Partnerships** | 10% | 100,000,000 | Governance-approved disbursements | PacaSwap LP incentives, DAG delegator rewards, Constellation grants, exchange listing reserves |

### 1.1 Why These Numbers

**40% Community Rewards** — Sacred and untouched. This is the engine of "all users are owners." Over 10 years, 400M ECHO flows to users through messaging, staking, referrals, and governance participation. The declining emission curve means early adopters earn more (incentivizing growth) while long-term inflation stays controlled. At 100K daily active users averaging 50 messages/day at 0.1 ECHO/message with daily caps, approximately 15–25M ECHO is distributed in Year 1.

**22% Treasury (reduced from 25%)** — The 3% reduction funds the expanded founder allocation. At 220M ECHO, the treasury still covers PacaSwap liquidity seeding (80M for ECHO/DAG + ECHO/USDC pools), operational reserve (50M in stablecoins via bridge), DAG staking collateral (purchased from treasury funds), and a 90M locked reserve for Phase 5–6 operations. This is ample — many projects operate with 15% or less in treasury.

**18% Founders (increased from 15%)** — Reflects the CEO/Visionary's outsized contribution to date (full architecture stack, all decisions, all documentation) while providing equal and competitive offers to the 4 co-founders being recruited. The insider total (founders + future team) is 28%, still below the industry average of 35–45% (which typically includes VC allocation). Community + ecosystem retains 50% — majority ownership.

**10% Future Team** — Reserved pool for recruiting. Equal 2% co-founder allocations mean the Future Team pool is now entirely available for non-founding hires, advisors, and performance grants. Unallocated tokens after 3 years revert to treasury via governance vote.

**10% Ecosystem** — PacaSwap liquidity mining rewards, DAG delegator incentives, Constellation ecosystem grants, and exchange listing reserves.

### 1.2 Emission Curve (Community Rewards)

The 400M community reward tokens are emitted over 10 years via a declining curve:

| Year | Emission (% of 400M) | Tokens | Cumulative |
|------|---------------------|--------|------------|
| 1 | 20% | 80,000,000 | 80M |
| 2 | 16% | 64,000,000 | 144M |
| 3 | 13% | 52,000,000 | 196M |
| 4 | 11% | 44,000,000 | 240M |
| 5 | 9% | 36,000,000 | 276M |
| 6 | 7% | 28,000,000 | 304M |
| 7 | 6% | 24,000,000 | 328M |
| 8 | 6% | 24,000,000 | 352M |
| 9 | 6% | 24,000,000 | 376M |
| 10 | 6% | 24,000,000 | 400M |

This front-loads rewards for early adopters (Year 1 gets 20% of all community tokens) while ensuring rewards continue for a full decade. After Year 10, no new ECHO is minted. Staking APY after Year 10 comes from transaction fees and AllowSpend fees, not new emission.

The emission schedule is enforced in Currency L1 Scala validation logic. Validators reject reward claims that exceed the current year's emission cap.

---

## 2. Founder Allocation (18% = 180,000,000 ECHO)

### 2.1 Founder Roles & Split

| Founder | Role | % of Total Supply | ECHO Tokens | Recruiting Status |
|---------|------|-------------------|-------------|-------------------|
| **Founder 1 (CEO)** | CEO / Visionary / Product | **10.0%** | **100,000,000** | Active — project originator |
| **Founder 2** | CTO / Lead iOS Engineer | **2.0%** | **20,000,000** | Recruiting |
| **Founder 3** | Scala / Blockchain Lead | **2.0%** | **20,000,000** | Recruiting |
| **Founder 4** | Head of Growth / Community | **2.0%** | **20,000,000** | Recruiting |
| **Founder 5** | Head of Design / UX | **2.0%** | **20,000,000** | Recruiting |

### 2.2 Rationale

**CEO at 10% of total supply** is justified because:

- Conceived the entire project — vision, architecture, product direction
- Produced all foundational documentation before any co-founder joined: PRD (5 versions), Data Layer Architecture (5 versions), Backend Architecture (2 versions), iOS Frontend Architecture (4 versions), OpenAPI spec (2 versions), Hidden Folders spec, Constellation deployment strategy, Constellation ecosystem gap analysis, Network State governance model, tokenomics framework, Glacial Interface design system
- Made all critical architectural decisions: relay over P2P, Constellation over private chain, Cardano for identity, Midnight evaluation, Digital Evidence integration, Stargazer SDK wallet, single-token governance model
- Will continue as CEO driving fundraising, Constellation partnership, and product strategy
- Bears the most risk — other founders join a project with a comprehensive architecture and clear direction
- 10% is high for a typical startup but defensible for a solo visionary who built the entire foundation. At $100M FDV this is $10M; at $1B it's $100M — both reasonable for a successful CEO/founder

**Co-founders at 2% each of total supply** provides:

- Clean, equal offer to all co-founders — simplifies recruiting ("every co-founder gets 2% with 4-year vest")
- Avoids internal politics of justifying differential allocations
- Competitive at this stage: 2% of a project with comprehensive architecture and clear direction is a strong offer
- The Future Team pool (10%) has room for performance grants if any co-founder proves disproportionately impactful

### 2.3 Vesting Schedule

All founder tokens are held in on-chain TokenLock positions with time-gated withdrawal conditions. There is no way to circumvent vesting — it is enforced by Currency L1 Scala validation code, not by a legal agreement alone.

| Parameter | Value | Enforcement |
|-----------|-------|-------------|
| **Total vesting period** | 4 years from mainnet genesis | TokenLock expiry field |
| **Cliff** | 1 year (12 months) | TokenLock rejects all WithdrawLock before cliff date |
| **Post-cliff vesting** | Monthly (1/36th per month for 36 months) | Monthly WithdrawLock allowance calculated by L1 validation |
| **Early departure (before cliff)** | 0% — all tokens returned to Future Team pool | TokenLock revocation by multi-sig (3-of-5 founders) |
| **Early departure (after cliff)** | Keep vested portion; unvested returned to Future Team pool | Partial TokenLock release + remainder revocation |
| **Acceleration — DAO transition** | 50% of unvested accelerates when governance transitions to full DAO (Phase 5–6) | Governance vote triggers acceleration in L1 code |
| **Acceleration — acquisition** | 100% of unvested accelerates if ECHO entity is acquired (unlikely given DAO structure, but protected) | Governance vote |
| **Lock-up post-vest** | None — vested tokens are freely transferable | Standard L0 token behavior |

### 2.4 Vesting Example (Founder 1 — 100,000,000 ECHO)

| Month | Event | Tokens Vested | Cumulative Vested | Locked |
|-------|-------|--------------|-------------------|--------|
| 0 | Genesis — all tokens in TokenLock | 0 | 0 | 100,000,000 |
| 1–11 | Cliff period — no vesting | 0 | 0 | 100,000,000 |
| 12 | Cliff completes — 25% vests | 25,000,000 | 25,000,000 | 75,000,000 |
| 13 | Monthly vesting begins | 2,083,333 | 27,083,333 | 72,916,667 |
| 24 | End of Year 2 | 2,083,333 | 50,000,000 | 50,000,000 |
| 36 | End of Year 3 | 2,083,333 | 75,000,000 | 25,000,000 |
| 48 | Fully vested | 2,083,333 | 100,000,000 | 0 |

### 2.5 On-Chain Transparency

Every founder's TokenLock position is publicly visible on DAG Explorer. Any ECHO user can verify:

- Total allocated tokens per founder DID
- Cliff date and status (completed or pending)
- Monthly vesting amount
- Total vested vs. locked
- Any WithdrawLock transactions (when founders actually withdraw vested tokens)

This is not optional — it is built into the L1 validation logic. Founder vesting transparency is a core feature, not a nice-to-have. The ECHO wallet UI renders this information for the founder's own view, and DAG Explorer makes it visible to everyone.

### 2.6 Co-Founder Offer Template

When recruiting Founders 2–5, the offer should include:

- ECHO token allocation (from the table above)
- 4-year vesting with 1-year cliff (non-negotiable)
- Role description and expectations
- Explicit statement that allocation is on-chain and publicly visible
- Explicit statement that pre-cliff departure = 0 tokens
- Mutual 30-day notice period (good-leaver provisions for post-cliff departure)
- Right to participate in governance as a founder (permanent board seat per PRD v2.2 governance structure)

---

## 3. Token Genesis Mechanics

At Phase 2 mainnet launch, the Currency L1 Scala validation logic executes token genesis:

```
Genesis Block (Snapshot #1)
├── Mint 1,000,000,000 ECHO total supply
│
├── Community Rewards Pool (400,000,000 ECHO)
│   └── Held in protocol-controlled emission account
│       (releases per emission curve, enforced by L1 validation)
│
├── Treasury (220,000,000 ECHO)
│   ├── 80,000,000 → PacaSwap liquidity seeding (ECHO/DAG + ECHO/USDC pools)
│   ├── 50,000,000 → Operational reserve (stablecoins via bridge)
│   └── 90,000,000 → Locked in treasury multi-sig (3-of-5 founders until DAO transition)
│
├── Founders (180,000,000 ECHO)
│   ├── Founder 1 DID → TokenLock(100,000,000, cliff=12mo, vest=48mo)
│   ├── Founder 2 DID → TokenLock(20,000,000, cliff=12mo, vest=48mo)
│   ├── Founder 3 DID → TokenLock(20,000,000, cliff=12mo, vest=48mo)
│   ├── Founder 4 DID → TokenLock(20,000,000, cliff=12mo, vest=48mo)
│   └── Founder 5 DID → TokenLock(20,000,000, cliff=12mo, vest=48mo)
│
├── Future Team Pool (100,000,000 ECHO)
│   └── Held in protocol-controlled pool
│       (released via multi-sig when new team members are allocated)
│
└── Ecosystem Pool (100,000,000 ECHO)
    └── Held in protocol-controlled pool
        (released via governance vote for LP incentives, grants, listings)
```

---

## 4. Governance Model: Single Token + Trust-Tier Weighting

### 4.1 Decision: Single Token (ECHO for Everything)

ECHO uses one token for utility (messaging rewards, staking, payments) and governance (voting on protocol changes, treasury allocation, board elections). A separate governance token was evaluated and rejected.

**Why not a separate governance token:**

- **Complexity kills adoption.** One token to understand, earn, stake, and vote with. Users don't need to navigate two token economies.
- **Liquidity concentration.** All trading volume and PacaSwap liquidity focuses on ECHO. Splitting across two tokens halves the depth of each.
- **Historical failures.** Many dual-token projects have struggled (LUNA/UST, AXS/SLP, GMT/GST). The utility token often becomes worthless or the governance token captures all value.
- **Staking model breaks.** If governance is a separate token, what does staking ECHO actually do? The dual-purpose (earn APY + earn governance weight) is cleaner with one token.
- **"All users are owners" is clearest with one token.** One token = one community = one governance system.

**How plutocracy is prevented without a separate token:** Trust-tier weighted governance.

### 4.2 Trust-Tier Weighted Governance

Governance votes are weighted by both token stake (via TokenLock) and trust tier. This prevents whales from buying governance power while rewarding verified, active community members.

**Governance Weight Formula:**

```
Governance Weight = StakedECHO × TrustTierMultiplier

Where TrustTierMultiplier:
  Tier 1 (Unverified):  0.0  — no governance participation
  Tier 2 (Newcomer):    0.5
  Tier 3 (Member):      1.0
  Tier 4 (Verified):    1.5
  Tier 5 (Trusted):     2.0
```

**Requirements to vote:**
- Must have staked ECHO (TokenLock position — cannot vote with unstaked tokens)
- Must be Trust Tier 2+ (Tier 1 / Unverified cannot vote)
- One vote per DID per proposal (no splitting votes across multiple positions)
- Governance weight = sum of all staked positions × trust tier multiplier

**Why this works:**

| Scenario | ECHO Staked | Trust Tier | Multiplier | Effective Weight | Outcome |
|----------|-----------|------------|------------|-----------------|---------|
| CEO (Founder 1) | 100,000,000 (vested, locked) | Tier 5 | 2.0 | 200,000,000 | Strong influence, but not majority |
| Whale who buys 50M | 50,000,000 | Tier 1 (unverified) | 0.0 | **0** | Zero governance power |
| Whale who buys 50M + verifies | 50,000,000 | Tier 4 | 1.5 | 75,000,000 | Significant but not controlling |
| 10,000 Tier 5 users × 10K ECHO each | 100,000,000 total | Tier 5 | 2.0 | 200,000,000 | Equal to CEO — community balance |
| 50,000 Tier 3 users × 2K ECHO each | 100,000,000 total | Tier 3 | 1.0 | 100,000,000 | Meaningful collective power |

**Critical property:** The CEO's 10% allocation gives economic alignment but does NOT give 10% governance control. With trust-tier weighting, the community can always outvote any individual founder once the user base reaches scale. A whale who doesn't verify their identity gets zero governance power regardless of how many tokens they buy.

**Staked tokens can vote:** TokenLock positions are eligible for governance voting. This solves the "stakers can't vote" problem and aligns economic commitment with governance participation. In fact, staking is required to vote — unstaked tokens have no governance weight.

**Founder TokenLock positions can vote:** Founder vesting tokens that are still locked (but within the TokenLock) are eligible for governance voting. This gives founders governance participation from day one, which is appropriate during the permissioned phases (1-3). As the community grows and more users stake, founder governance influence dilutes naturally.

### 4.3 Governance Weight Enforcement

Governance weight is calculated by the Data L1 Scala validation logic when processing governance vote transactions. The L1 validator:

1. Queries the voter's DID
2. Sums all TokenLock positions for that DID (including founder vesting locks)
3. Queries the voter's trust tier from the Cardano credential cache
4. Computes: `weight = totalStaked × tierMultiplier`
5. Records the weighted vote on-chain

This is enforced at the consensus layer — the Go backend cannot inflate governance weights.

---

## 5. Midnight Blockchain Evaluation

### 5.1 Decision: Cardano Now, Midnight Phase 3+

ECHO uses Cardano for DIDs and Verifiable Credentials in Phases 1-2. Midnight is evaluated for privacy-credential verification starting Phase 3, after it has proven mainnet stability.

**Rationale:** Midnight launched mainnet in late March 2026 and has zero production track record. Cardano has 8+ years of mainnet operation with mature Plutus tooling. ECHO ships on the proven chain, then adopts Midnight when it's stable.

### 5.2 Midnight Integration Roadmap

| Phase | Identity Layer | Midnight Role |
|-------|---------------|---------------|
| Phase 1-2 | **Cardano only** — DID registration, credential issuance, trust tier UTXO datums | None. Evaluate from distance. |
| Phase 3 | **Cardano + evaluate Midnight** — assess stability, tooling, developer ecosystem after 6+ months of mainnet | Proof-of-concept: ZK trust tier verification ("Prove I'm Tier 3+ without revealing my credential") |
| Phase 3-4 | **Cardano + Midnight for ZK credential verification** | Trust tier verification on Midnight. Keep DID registration + credential issuance on Cardano. Use native Cardano ↔ Midnight bridge. |
| Phase 4+ | **Cardano + Midnight for enterprise privacy** | Org-tier features: private group membership proofs, compliance verification without data exposure, regulatory audit with selective disclosure |

### 5.3 What Stays on Cardano (Always)

- DID Document registration (public by design — contains public keys)
- Credential schema definitions (public)
- Credential issuance and revocation (Plutus bit vector)
- Trust tier UTXO datums (current system works, backward compatible)

### 5.4 What Moves to Midnight (Phase 3+)

- **Trust tier verification:** "Prove I am Trust Tier 3+ and my credential is valid" — ZK proof on Midnight without revealing score or credential details
- **KYC compliance proofs:** "Prove I passed KYC without revealing my passport data" — for Organization tier clients
- **Private group membership:** "Prove I am a member of Group X" without revealing which groups you belong to
- **Age/eligibility verification:** "Prove I am 18+ without revealing my birthdate"

### 5.5 Technical Notes

Midnight uses Compact (TypeScript-based DSL) for smart contracts — does NOT require Scala. An existing web developer can handle Midnight contracts. The Scala/JVM requirement is only for Constellation metagraph L1 validation logic.

Midnight has a dual-token model: NIGHT (governance/staking) and DUST (renewable, non-tradable, pays for private transactions). ECHO does not need to hold NIGHT — Midnight ZK verification can be called as a service by the Go backend.

---

## 6. ECHO Wallet Architecture (Stargazer SDK)

### 4.1 Decision: Build on Stargazer SDK

**Resolved (was Open Question #13).** ECHO builds a native wallet experience inside the iOS app using the Constellation Stargazer Wallet SDK. This replaces the "rewards page" concept with a true decentralized wallet.

**Rationale:**

- "Rewards page" implies gamification points. "Wallet" implies real ownership. For a project built on "all users are owners," the framing matters.
- Stargazer SDK handles key management, transaction signing, L0 token display, delegation, and bridging — ECHO doesn't reinvent wallet infrastructure.
- Users manage their ECHO tokens in the same app where they message, creating a unified experience.
- Founder vesting, staking, delegation, and rewards all live in one place tied to the user's DID.
- External wallet compatibility remains — users can also view/manage ECHO in standalone Stargazer or D'Cent hardware wallet.

### 4.2 Wallet Tab Architecture

The ECHO iOS app adds a "Wallet" tab alongside Messaging and Profile:

```
┌──────────────────────────────────────────────────┐
│  Tab Bar:  💬 Messages  |  👛 Wallet  |  👤 Me  │
└──────────────────────────────────────────────────┘
```

**Wallet Tab — All Users:**

```
┌─────────────────────────────────────────────┐
│  👛 ECHO Wallet                      ⚙️    │
│                                              │
│  ┌─────────────────────────────────────┐    │
│  │  Total Balance                      │    │
│  │  24,830.00 ECHO                     │    │
│  │  ≈ $2,483.00 USD                   │    │
│  │  ▲ 3.2% (24h)                      │    │
│  └─────────────────────────────────────┘    │
│                                              │
│  Available     12,450.00                     │
│  Staked         8,000.00  🔒 Gold Tier       │
│  Delegated to  Validator #7  ↗              │
│  Pending        4,380.00  [Claim All]        │
│                                              │
│  ┌────────┬────────┬────────┬────────┐      │
│  │ Stake  │Delegate│  Swap  │ Bridge │      │
│  └────────┴────────┴────────┴────────┘      │
│                                              │
│  ── Today's Rewards ─────────────────       │
│  💬 Messaging     4.2 / 50.0 ECHO          │
│  🤝 Referrals     0.0 / 50.0 ECHO          │
│  📊 Staking      12.8 ECHO (auto)          │
│  ░░░░░░░░░█████░░░░░░ 34% of daily cap     │
│                                              │
│  ── Recent Activity ─────────────────       │
│  ↓ +2.1 ECHO  Messaging reward  2m ago     │
│  ↓ +12.8 ECHO Staking reward    6h ago     │
│  ↑ -500 ECHO  Staked (Gold)     2d ago     │
│  ↓ +50 ECHO   Referral bonus    5d ago     │
│                                              │
└─────────────────────────────────────────────┘
```

**Wallet Tab — Founder View (additional section):**

```
│  ── Founder Allocation ──────────────────   │
│  Role           CEO / Visionary              │
│  Allocated      60,000,000 ECHO              │
│  Vested         16,250,000 ECHO  (27.1%)     │
│  Locked         43,750,000 ECHO              │
│  Next unlock    1,250,000 ECHO               │
│  Unlock date    April 7, 2027                │
│  Cliff          ✓ Completed Mar 7, 2027     │
│                                              │
│  [████████░░░░░░░░░░░░░░░░░░░░░] 27.1%     │
│                                              │
│  Withdrawable   1,250,000 ECHO  [Withdraw]   │
│                                              │
│  🔍 View on DAG Explorer →                  │
│  ⓘ Founder vesting is on-chain and          │
│    publicly verifiable by all ECHO users.    │
```

### 4.3 Wallet Components (SwiftUI + Stargazer SDK)

```swift
// EchoWallet/WalletTab.swift

import SwiftUI
import StargazerSDK  // Constellation Stargazer Wallet SDK

struct WalletTab: View {
    @StateObject private var viewModel = WalletViewModel()
    
    var body: some View {
        NavigationStack {
            ScrollView {
                BalanceCard(balance: viewModel.totalBalance, 
                           usdValue: viewModel.usdValue)
                
                BalanceBreakdown(
                    available: viewModel.available,
                    staked: viewModel.staked,
                    delegatedTo: viewModel.delegatedValidator,
                    pending: viewModel.pendingRewards
                )
                
                ActionButtons(
                    onStake: { viewModel.showStaking = true },
                    onDelegate: { viewModel.showDelegation = true },
                    onSwap: { viewModel.showSwap = true },
                    onBridge: { viewModel.showBridge = true }
                )
                
                DailyRewardsSection(rewards: viewModel.dailyRewards)
                
                // Founder section — only visible if user's DID has a founder TokenLock
                if let vesting = viewModel.founderVesting {
                    FounderVestingSection(vesting: vesting)
                }
                
                RecentActivityList(activity: viewModel.recentActivity)
            }
            .navigationTitle("ECHO Wallet")
        }
    }
}

// WalletViewModel.swift

@MainActor
class WalletViewModel: ObservableObject {
    private let stargazer: StargazerClient  // Stargazer SDK client
    private let backendAPI: BackendAPIClient
    private let metagraphQuery: MetagraphQueryClient
    
    @Published var totalBalance: Decimal = 0
    @Published var available: Decimal = 0
    @Published var staked: Decimal = 0
    @Published var pendingRewards: Decimal = 0
    @Published var delegatedValidator: ValidatorInfo?
    @Published var founderVesting: VestingInfo?  // nil for non-founders
    @Published var dailyRewards: DailyRewards = .empty
    @Published var recentActivity: [WalletActivity] = []
    
    func loadWallet() async {
        // 1. Query balance from Stargazer SDK (reads metagraph state)
        let balance = try? await stargazer.getBalance(token: .echo)
        self.totalBalance = balance?.total ?? 0
        self.available = balance?.available ?? 0
        
        // 2. Query TokenLock positions (staking)
        let locks = try? await stargazer.getTokenLocks(token: .echo)
        self.staked = locks?.reduce(0) { $0 + $1.amount } ?? 0
        
        // 3. Query StakeDelegation positions
        let delegations = try? await stargazer.getDelegations(token: .echo)
        self.delegatedValidator = delegations?.first?.validator
        
        // 4. Query pending rewards from backend cache
        let rewards = try? await backendAPI.getPendingRewards()
        self.pendingRewards = rewards?.total ?? 0
        self.dailyRewards = rewards?.daily ?? .empty
        
        // 5. Check for founder vesting TokenLock (special type with cliff/vest metadata)
        if let founderLock = locks?.first(where: { $0.isFounderVesting }) {
            self.founderVesting = VestingInfo(
                totalAllocated: founderLock.originalAmount,
                vested: founderLock.vestedAmount,
                locked: founderLock.lockedAmount,
                nextUnlockAmount: founderLock.nextUnlockAmount,
                nextUnlockDate: founderLock.nextUnlockDate,
                cliffCompleted: founderLock.cliffCompleted,
                cliffDate: founderLock.cliffDate,
                withdrawable: founderLock.withdrawableAmount
            )
        }
    }
    
    // Claim rewards via AtomicAction (verify tier + claim + update cap)
    func claimRewards() async throws {
        try await stargazer.submitAtomicAction([
            .verifyTrustTier(did: currentDID),
            .claimRewards(did: currentDID, types: dailyRewards.claimableTypes),
            .updateDailyCap(did: currentDID)
        ])
        await loadWallet()  // Refresh
    }
    
    // Stake ECHO via TokenLock
    func stakeEcho(amount: Decimal, tier: StakingTier) async throws {
        try await stargazer.submitTokenLock(TokenLockRequest(
            token: .echo,
            amount: amount,
            tier: tier.rawValue,
            duration: tier.durationDays
        ))
        await loadWallet()
    }
    
    // Delegate staked ECHO via StakeDelegation
    func delegateToValidator(_ validatorId: String, stakeId: String) async throws {
        try await stargazer.submitStakeDelegation(StakeDelegationRequest(
            stakeId: stakeId,
            validatorId: validatorId
        ))
        await loadWallet()
    }
    
    // Withdraw vested founder tokens via WithdrawLock
    func withdrawVestedTokens(amount: Decimal) async throws {
        guard let vesting = founderVesting, amount <= vesting.withdrawable else {
            throw WalletError.insufficientVestedBalance
        }
        try await stargazer.submitWithdrawLock(WithdrawLockRequest(
            amount: amount
            // 14-day cooldown enforced by L1 validation
        ))
        await loadWallet()
    }
}
```

### 4.4 Staking Flow

```
User taps [Stake] →
  ├── Select amount (slider + input)
  ├── Select tier:
  │   ├── Bronze (30 days, 5% APR)
  │   ├── Silver (90 days, 8% APR)
  │   ├── Gold (180 days, 12% APR)
  │   └── Platinum (365 days, 15% APR)
  ├── Review: "Lock 8,000 ECHO for 180 days at 12% APR"
  ├── Biometric confirmation (Secure Enclave signs transaction)
  └── Stargazer SDK → TokenLock transaction → Currency L1
      ├── Success: balance updates, staking position appears
      └── Failure: error message, tokens unchanged
```

### 4.5 Delegation Flow

```
User taps [Delegate] →
  ├── Validator Browser:
  │   ├── List of active L1 validators
  │   ├── Per validator: uptime %, commission %, total delegated, APR estimate
  │   ├── Sort by: APR, uptime, commission, total delegated
  │   └── Filter: Currency L1, Data L1, both
  ├── Select validator → "Delegate 8,000 staked ECHO to Validator #7"
  ├── Biometric confirmation
  └── Stargazer SDK → StakeDelegation transaction → Currency L1
```

### 4.6 Swap Flow (Phase 3+ — PacaSwap Integration)

```
User taps [Swap] →
  ├── Select pair: ECHO/DAG or ECHO/USDC
  ├── Enter amount
  ├── See: exchange rate, price impact, estimated output
  ├── Confirm → AtomicAction (atomic swap via PacaSwap)
  └── Tokens appear in wallet
```

### 4.7 Bridge Flow (Phase 3+ — Base/Ink)

```
User taps [Bridge] →
  ├── Select destination: Base or Ink
  ├── Enter ECHO amount
  ├── See: bridge fee, estimated time, destination address
  ├── Confirm → Bridge transaction
  └── Status: "Bridging... (~1 minute)" → "Complete"
```

---

## 5. Security Considerations

### 5.1 Wallet Key Management

The Stargazer SDK manages wallet keys. ECHO's integration:

| Concern | Approach |
|---------|----------|
| Key generation | Stargazer SDK generates Constellation keypair; private key stored in iOS Keychain with Secure Enclave protection |
| Transaction signing | All transactions (TokenLock, StakeDelegation, etc.) signed by Stargazer SDK using the local private key; requires biometric authentication |
| Key backup | Stargazer's recovery phrase mechanism; ECHO does not add a separate backup system for wallet keys |
| DID linkage | User's Constellation wallet address is linked to their Cardano DID during registration; both identities map to the same profile |
| Multi-wallet | Users can import an existing Stargazer wallet or create a new one during ECHO setup |

### 5.2 Founder Token Security

| Risk | Mitigation |
|------|-----------|
| Founder private key compromised | TokenLock prevents spending locked tokens even with key access; attacker can only spend vested + withdrawn tokens. Emergency: multi-sig can freeze founder TokenLock pending recovery. |
| Founder coerced to withdraw | WithdrawLock has 14-day cooldown; community can detect unusual founder withdrawal patterns on DAG Explorer and governance can intervene |
| Rogue founder sells vested tokens and dumps price | Vested tokens release monthly (1/36th) — limits sell pressure. Community visibility on DAG Explorer creates social accountability. |
| Multi-sig compromised (treasury) | 3-of-5 threshold; transition to DAO governance in Phase 4–5 removes multi-sig dependency |

---

## 6. Implementation Priority

| Priority | Component | Effort | Phase |
|----------|-----------|--------|-------|
| P0 | Token genesis in Currency L1 Scala code (mint + allocate to pools) | 1 week | Phase 2 |
| P0 | Founder TokenLock positions with cliff/vest logic in L1 validation | 2 weeks | Phase 2 |
| P0 | Stargazer SDK integration in iOS app | 2 weeks | Phase 2 |
| P0 | Wallet tab UI (balance, staking, pending rewards) | 2 weeks | Phase 2 |
| P0 | Claim rewards via AtomicAction | 1 week | Phase 2 |
| P1 | Delegation flow (validator browser, StakeDelegation) | 1 week | Phase 2 |
| P1 | Founder vesting display UI | 3 days | Phase 2 |
| P1 | DAG Explorer visibility (verify founder locks, supply distribution) | 1 week | Phase 2 |
| P2 | PacaSwap swap integration in wallet | 2 weeks | Phase 3 |
| P2 | Bridge integration (Base) | 1 week | Phase 3 |
| P3 | Bridge integration (Ink) | 1 week | Phase 4 |

---

*ECHO Tokenomics & Founder Allocation v1.0*
*March 7, 2026*
*Status: Implementation-ready. Requires co-founder recruiting to finalize Founders 2–5 DID assignments before genesis.*
