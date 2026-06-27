# ECHO Wallet, Balance & Staking Launch Plan

**Status:** In progress (2026-05-29)  
**Scope:** Real Currency L1 balances, staking positions, reward claims, iOS wallet tab, Stargazer signing path  
**Phases:** Backend → iOS → Stargazer SDK bridge → E2E

---

## Launch principles (crypto · privacy · engineering · UX)

| Lens | Requirement |
|------|-------------|
| **Custody** | TokenLock keeps funds in the user's wallet; backend never holds private keys |
| **Privacy (T0–T7)** | On-chain payloads are tier names + amounts only — no PII, no message content |
| **Auth** | All `/v3/wallet/*` mutations require JWT; DID from token must match request body |
| **Integrity** | Stake/unstake/claim submit to Currency L1 when `CURRENCY_L1_URL` is set; PG cache mirrors chain |
| **UX** | Show datum as human ECHO (÷1e8); empty → genesis credit in dev; errors are actionable |
| **Degrade** | No L1 URL → ledger + cache still work for TestFlight; chain submit is best-effort logged |

---

## Architecture

```text
iOS WalletTab / StakingDetailView
        │  JWT
        ▼
GET/POST /v3/wallet/*  (Go API :8000)
        │
        ├── wallet.Store (PostgreSQL: wallet_balance_cache, staking_positions)
        ├── rewards.Service (pending + auto-scale)
        └── MetagraphClient.SubmitCurrencyL1 (TokenLock / WithdrawLock / AtomicAction)
                    │
                    ▼
            Echo Currency L1 (TokenLockUpdate validators)
```

**Balance model (datum = 1e8 per ECHO):**

- `total` — allocated ECHO for DID (genesis credit + claimed rewards)
- `staked` — sum of active `staking_positions.amount`
- `available` — `total - staked` (simplified; withdraw cooldown tracked per lock)
- On-chain locks indexed by `senderDid` in metagraph state; PG is query cache + source when L1 offline

**Tier mapping (iOS ↔ API ↔ L1):**

| iOS / API | Scala L1 | Lock days | APR |
|-----------|----------|-----------|-----|
| bronze | Tier 1 | 30 | 5% |
| silver | Tier 2 | 90 | 8% |
| gold | Tier 3 | 180 | 10% |
| platinum | Tier 5 | 365 | 15% |

---

## Step 1 — Backend ✅ (this commit)

- `internal/wallet/pg_store.go` — PostgreSQL store (migration 003 tables + 024 `wallet_accounts`)
- `internal/wallet/metagraph_querier.go` — `MetagraphQuerier` implementation
- `internal/wallet/rewards_adapter.go` — `RewardsQuerier` → `rewards.Service`
- `internal/api/wallet_handlers.go` — `/v3/wallet` routes
- Wire in `main.go` + `V3Handlers`
- Tests: `wallet_handlers_test.go`, existing `wallet/service_test.go`

**Env:**

| Variable | Purpose |
|----------|---------|
| `CURRENCY_L1_URL` | Currency L1 base URL for submissions |
| `ECHO_WALLET_GENESIS_ECHO` | Dev genesis credit (default 1000 ECHO) when cache empty |
| `ECHO_WALLET_GENESIS_AUTO` | `1` to auto-credit on first wallet read (dev/TestFlight) |

---

## Step 2 — iOS

- `HTTPWalletAPIClient` implementing `WalletAPIClient`
- `DIContainer.resolveWalletAPI()`
- Inject into `WalletTab`, `StakingDetailView`, `ValidatorBrowserView`
- Datum ↔ Decimal formatting helper
- Wallet empty state + staking confirmation copy

---

## Step 3 — Stargazer SDK bridge

- `WalletProvisioner` — link DAG address via `POST /v3/wallet/link` after enrollment
- `StargazerBridge` initialized with live API client
- Keychain: `echo.wallet.dagAddress`
- **Interim:** deterministic DAG address (same as enrollment proxy) until Constellation SPM ships
- **Production:** replace with Stargazer SDK `createWallet` / `submitTokenLock` local signing

---

## Step 4 — E2E

- `scripts/validate-wallet.sh` — JWT → GET wallet → stake → verify position → claim rewards path
- Document in `docs/E2E_QUICK_START.md` § wallet
- Manual: Wallet tab → Staking Details → stake bronze tier on dev cluster

---

## Out of scope (post-launch)

- Cross-chain bridges, PacaSwap LP, liquid staking
- Full L0 calculated-state combiner for currency (use PG mirror until L0 query API lands)
- Pay-in-chat (WO-299/300)
