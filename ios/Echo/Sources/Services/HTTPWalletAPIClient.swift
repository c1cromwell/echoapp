#if os(iOS)
import Foundation

// MARK: - Wire models (match internal/wallet/models.go)

private struct WalletStateWire: Codable, Sendable {
    let did: String
    let totalBalance: Int64
    let available: Int64
    let staked: Int64
    let pendingRewards: Int64
    let locks: [TokenLockWire]
    let delegations: [DelegationWire]
    let dailyRewards: AutoScaleWire?
    let vesting: VestingWire?
}

private struct TokenLockWire: Codable, Sendable {
    let id: String
    let amount: Int64
    let tier: String
    let lockedUntil: Date
    let vestingType: String?
    let delegatedTo: String?
}

private struct DelegationWire: Codable, Sendable {
    let id: String
    let stakeId: String
    let validatorId: String
    let amount: Int64
    let since: Date
    let validator: ValidatorWire
}

private struct ValidatorWire: Codable, Sendable {
    let id: String
    let address: String
    let uptimePercent: Double
    let commissionPercent: Double
    let totalDelegated: Int64
    let delegatorCount: Int
    let layer: String
    let estimatedApr: Double
}

private struct AutoScaleWire: Codable, Sendable {
    let currentRate: Int64
    let dailyBudget: Int64
    let effectiveDailyBudget: Int64
    let budgetUsedToday: Int64
    let remainingToday: Int64
    let totalActivityWeight: Double?
    let lastUpdated: String
}

private struct VestingWire: Codable, Sendable {
    let role: String
    let totalAllocated: Int64
    let vested: Int64
    let locked: Int64
    let withdrawable: Int64
    let nextUnlockAmount: Int64
    let nextUnlockDate: Date?
    let cliffDate: Date
    let cliffCompleted: Bool
    let vestingPercent: Double
    let explorerUrl: String?
}

private struct ValidatorsResponse: Codable, Sendable {
    let validators: [ValidatorWire]
}

private struct StakeRequestWire: Encodable, Sendable {
    let amount: Int64
    let tier: String
}

private struct UnstakeRequestWire: Encodable, Sendable {
    let stakeId: String
    let amount: Int64
}

private struct DelegateRequestWire: Encodable, Sendable {
    let stakeId: String
    let validatorId: String
    let amount: Int64
}

private struct ClaimRequestWire: Encodable, Sendable {
    let types: [String]
}

private struct TxResultWire: Decodable, Sendable {
    let txHash: String
}

private struct LinkRequestWire: Encodable, Sendable {
    let address: String
}

// MARK: - Endpoints

enum WalletEndpoint: APIEndpoint {
    case state
    case stake
    case unstake
    case delegate
    case claim
    case validators
    case link
    case challenge
    case txContext(type: String)

    var path: String {
        switch self {
        case .state: return "/v3/wallet"
        case .stake: return "/v3/wallet/stake"
        case .unstake: return "/v3/wallet/unstake"
        case .delegate: return "/v3/wallet/delegate"
        case .claim: return "/v3/wallet/claim"
        case .validators: return "/v3/wallet/validators"
        case .link: return "/v3/wallet/link"
        case .challenge: return "/v3/wallet/challenge"
        case .txContext(let type): return "/v3/wallet/tx-context?type=\(type)"
        }
    }
}

// MARK: - Live client

/// Authenticated wallet API backed by `/v3/wallet/*` (passkey-signed via APIClient).
actor HTTPWalletAPIClient: WalletAPIClient {
    private let apiClient: APIClient
    private let keychain: KeychainManager
    private var cached: WalletStateWire?

    init(apiClient: APIClient, keychain: KeychainManager = .shared) {
        self.apiClient = apiClient
        self.keychain = keychain
    }

    func fetchWalletState() async throws -> WalletState {
        let wire: WalletStateWire = try await apiClient.get(endpoint: WalletEndpoint.state)
        cached = wire
        return mapWalletState(wire)
    }

    func createWallet() async throws -> WalletInfo {
        guard let did = try? await keychain.retrieve(key: "echo.did.current", as: String.self) else {
            throw StargazerError.walletCreationFailed
        }
        let provisioner = WalletProvisioner(apiClient: apiClient, keychain: keychain)
        let address = try await provisioner.ensureWalletLinked(did: did)
        return WalletInfo(address: address, publicKey: "")
    }

    func importWallet(mnemonic: String) async throws -> WalletInfo {
        _ = mnemonic
        throw StargazerError.transactionFailed("Mnemonic import requires Stargazer SDK")
    }

    func getBalance() async throws -> BalanceInfo {
        let wire = try await cachedOrRefresh()
        return BalanceInfo(
            total: EchoDatum.fromDatum(wire.totalBalance),
            available: EchoDatum.fromDatum(wire.available)
        )
    }

    func getTokenLocks() async throws -> [TokenLockPosition] {
        let wire = try await cachedOrRefresh()
        return wire.locks.map(mapLock)
    }

    func getDelegations() async throws -> [DelegationPosition] {
        let wire = try await cachedOrRefresh()
        return wire.delegations.map {
            DelegationPosition(
                id: $0.id,
                stakeId: $0.stakeId,
                validatorId: $0.validatorId,
                amount: EchoDatum.fromDatum($0.amount),
                since: $0.since
            )
        }
    }

    func getValidators() async throws -> [ValidatorInfo] {
        let resp: ValidatorsResponse = try await apiClient.get(endpoint: WalletEndpoint.validators)
        return resp.validators.map(mapValidator)
    }

    func submitTokenLock(amount: Decimal, tier: StakingTier) async throws -> String {
        let datum = EchoDatum.toDatum(amount)
        // Real-funds: sign the TokenLock locally and let the backend relay it.
        // Gated by the real-funds kill-switch; also falls back to the
        // server-originated path when signing isn't possible (no tx-context).
        if WalletFeatureFlags.realFundsSigningEnabled,
           let signed = try? await buildSignedTx(type: "tokenLock", extra: ["amount": datum]) {
            let proof = (try? await buildProofHeader()) ?? [:]
            let result: TxResultWire = try await apiClient.postJSON(
                endpoint: WalletEndpoint.stake,
                json: ["amount": datum, "tier": tier.rawValue, "signed": signed],
                headers: proof)
            cached = nil
            return result.txHash
        }
        let body = StakeRequestWire(amount: datum, tier: tier.rawValue)
        let result: TxResultWire = try await apiClient.post(endpoint: WalletEndpoint.stake, body: body)
        cached = nil
        return result.txHash
    }

    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String {
        let wire = try await cachedOrRefresh()
        guard let lock = wire.locks.first(where: { $0.id == stakeId }) else {
            throw StargazerError.transactionFailed("Stake position not found")
        }
        if WalletFeatureFlags.realFundsSigningEnabled,
           let signed = try? await buildSignedTx(type: "delegatedStake", extra: [
            "nodeId": validatorId, "amount": lock.amount, "fee": 0, "tokenLockRef": stakeId,
        ]) {
            let proof = (try? await buildProofHeader()) ?? [:]
            let result: TxResultWire = try await apiClient.postJSON(
                endpoint: WalletEndpoint.delegate,
                json: ["stakeId": stakeId, "validatorId": validatorId, "amount": lock.amount, "signed": signed],
                headers: proof)
            cached = nil
            return result.txHash
        }
        let body = DelegateRequestWire(stakeId: stakeId, validatorId: validatorId, amount: lock.amount)
        let result: TxResultWire = try await apiClient.post(endpoint: WalletEndpoint.delegate, body: body)
        cached = nil
        return result.txHash
    }

    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String {
        let datum = EchoDatum.toDatum(amount)
        if WalletFeatureFlags.realFundsSigningEnabled,
           let signed = try? await buildSignedTx(type: "withdrawDelegatedStake", extra: ["stakeRef": stakeId]) {
            let proof = (try? await buildProofHeader()) ?? [:]
            let result: TxResultWire = try await apiClient.postJSON(
                endpoint: WalletEndpoint.unstake,
                json: ["stakeId": stakeId, "amount": datum, "signed": signed],
                headers: proof)
            cached = nil
            return result.txHash
        }
        let body = UnstakeRequestWire(stakeId: stakeId, amount: datum)
        let result: TxResultWire = try await apiClient.post(endpoint: WalletEndpoint.unstake, body: body)
        cached = nil
        return result.txHash
    }

    // MARK: - Real-funds signing helpers

    /// Builds a client-signed {value, proofs} for a Currency-L1 tx: fetch
    /// tx-context (source + parent), assemble the body, sign via the embedded
    /// dag4 bundle. `extra` carries the type-specific fields (amount, nodeId, …).
    /// `withdrawDelegatedStake` needs no parent (its body is {source, stakeRef}).
    private func buildSignedTx(type: String, extra: [String: Any]) async throws -> [String: Any] {
        let account = try await WalletKeyStore.shared.ensureWallet()
        let ctx = try await fetchTxContext(type: type)
        var body: [String: Any] = ["source": ctx.source]
        if type != "withdrawDelegatedStake" { body["parent"] = ctx.parent }
        body.merge(extra) { _, new in new }
        let signed = try await StargazerSigner.shared.signTransaction(
            privateKey: account.privateKey, publicKey: account.publicKey, body: body)
        return ["value": signed.value, "proofs": [["id": signed.proofID, "signature": signed.signature]]]
    }

    private struct TxContext { let source: String; let parent: [String: Any] }

    private func fetchTxContext(type: String) async throws -> TxContext {
        struct ParentWire: Decodable { let hash: String; let ordinal: Int }
        struct CtxWire: Decodable { let source: String; let parent: ParentWire }
        let wire: CtxWire = try await apiClient.get(endpoint: WalletEndpoint.txContext(type: type))
        return TxContext(source: wire.source, parent: ["hash": wire.parent.hash, "ordinal": wire.parent.ordinal])
    }

    private func buildProofHeader() async throws -> [String: String] {
        struct ChallengeResp: Decodable { let challenge: String; let expiresAt: String }
        let resp: ChallengeResp = try await apiClient.get(endpoint: WalletEndpoint.challenge)
        let signed = try await WalletKeyStore.shared.signChallenge(resp.challenge)
        let proof: [String: String] = [
            "publicKey": signed.publicKey, "challenge": resp.challenge, "signature": signed.signature,
        ]
        let data = try JSONSerialization.data(withJSONObject: proof)
        return ["X-Wallet-Proof": data.base64EncodedString()]
    }

    func submitRewardClaim(rewardTypes: [String]) async throws -> String {
        let result: TxResultWire = try await apiClient.post(
            endpoint: WalletEndpoint.claim,
            body: ClaimRequestWire(types: rewardTypes)
        )
        cached = nil
        return result.txHash
    }

    func fetchEmissionStatus() async throws -> EmissionStatus {
        let wire: EmissionStatusWire = try await apiClient.get(endpoint: TokenomicsEndpoint.emissionStatus)
        return TokenomicsMapping.mapEmission(wire)
    }

    func fetchFounderVesting() async throws -> FounderVestingProfile? {
        do {
            let wire: FounderVestingWire = try await apiClient.get(endpoint: TokenomicsEndpoint.founderVesting)
            guard wire.isFounder else { return nil }
            let walletWire = try await cachedOrRefresh()
            let lockId = walletWire.locks.first(where: { $0.vestingType == "founder" })?.id
            return TokenomicsMapping.mapFounderVesting(wire, lockId: lockId)
        } catch APIError.notFound {
            return nil
        }
    }

    func linkDAGAddress(_ address: String) async throws {
        struct LinkResp: Decodable { let did: String; let address: String }
        let _: LinkResp = try await apiClient.post(
            endpoint: WalletEndpoint.link,
            body: LinkRequestWire(address: address)
        )
    }

    // MARK: - Private

    private func cachedOrRefresh() async throws -> WalletStateWire {
        if let cached { return cached }
        let wire: WalletStateWire = try await apiClient.get(endpoint: WalletEndpoint.state)
        cached = wire
        return wire
    }

    private func mapWalletState(_ wire: WalletStateWire) -> WalletState {
        let daily: DailyRewardProgress?
        if let d = wire.dailyRewards {
            let zero = RewardEarningEntry(earned: 0, count: 0)
            daily = DailyRewardProgress(
                messaging: zero,
                referrals: zero,
                staking: zero,
                paymentRail: zero,
                autoScaleState: AutoScaleRewardState(
                    currentRate: EchoDatum.fromDatum(d.currentRate),
                    dailyBudget: EchoDatum.fromDatum(d.dailyBudget),
                    effectiveDailyBudget: EchoDatum.fromDatum(d.effectiveDailyBudget),
                    budgetUsedToday: EchoDatum.fromDatum(d.budgetUsedToday),
                    remainingToday: EchoDatum.fromDatum(d.remainingToday),
                    totalActivityToday: d.totalActivityWeight ?? 0,
                    lastUpdated: ISO8601DateFormatter().date(from: d.lastUpdated) ?? Date()
                )
            )
        } else {
            daily = nil
        }

        var vesting: VestingState?
        if let v = wire.vesting {
            vesting = VestingState(
                role: v.role,
                totalAllocated: EchoDatum.fromDatum(v.totalAllocated),
                vested: EchoDatum.fromDatum(v.vested),
                locked: EchoDatum.fromDatum(v.locked),
                withdrawable: EchoDatum.fromDatum(v.withdrawable),
                nextUnlockAmount: EchoDatum.fromDatum(v.nextUnlockAmount),
                nextUnlockDate: v.nextUnlockDate,
                cliffDate: v.cliffDate,
                cliffCompleted: v.cliffCompleted,
                vestingPercent: v.vestingPercent,
                explorerURL: v.explorerUrl.flatMap { URL(string: $0) }
            )
        }

        return WalletState(
            totalBalance: EchoDatum.fromDatum(wire.totalBalance),
            available: EchoDatum.fromDatum(wire.available),
            staked: EchoDatum.fromDatum(wire.staked),
            pendingRewards: EchoDatum.fromDatum(wire.pendingRewards),
            locks: wire.locks.map(mapLock),
            delegations: wire.delegations.map {
                DelegationPosition(
                    id: $0.id,
                    stakeId: $0.stakeId,
                    validatorId: $0.validatorId,
                    amount: EchoDatum.fromDatum($0.amount),
                    since: $0.since
                )
            },
            dailyRewards: daily,
            vesting: vesting
        )
    }

    private func mapLock(_ lock: TokenLockWire) -> TokenLockPosition {
        TokenLockPosition(
            id: lock.id,
            amount: EchoDatum.fromDatum(lock.amount),
            tier: lock.tier,
            lockedUntil: lock.lockedUntil,
            vestingType: lock.vestingType,
            originalAmount: EchoDatum.fromDatum(lock.amount),
            cliffDate: nil,
            cliffCompleted: false,
            vestedAmount: 0,
            withdrawableAmount: 0,
            nextUnlockDate: nil,
            nextUnlockAmount: 0,
            delegatedTo: lock.delegatedTo
        )
    }

    private func mapValidator(_ v: ValidatorWire) -> ValidatorInfo {
        ValidatorInfo(
            id: v.id,
            address: v.address,
            uptimePercent: v.uptimePercent,
            commissionPercent: v.commissionPercent,
            totalDelegated: EchoDatum.fromDatum(v.totalDelegated),
            delegatorCount: v.delegatorCount,
            layer: v.layer,
            estimatedAPR: v.estimatedApr
        )
    }
}
#endif
