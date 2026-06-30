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
        let body = StakeRequestWire(amount: EchoDatum.toDatum(amount), tier: tier.rawValue)
        let result: TxResultWire = try await apiClient.post(endpoint: WalletEndpoint.stake, body: body)
        cached = nil
        return result.txHash
    }

    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String {
        let wire = try await cachedOrRefresh()
        guard let lock = wire.locks.first(where: { $0.id == stakeId }) else {
            throw StargazerError.transactionFailed("Stake position not found")
        }
        let body = DelegateRequestWire(
            stakeId: stakeId,
            validatorId: validatorId,
            amount: lock.amount
        )
        let result: TxResultWire = try await apiClient.post(endpoint: WalletEndpoint.delegate, body: body)
        cached = nil
        return result.txHash
    }

    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String {
        let body = UnstakeRequestWire(stakeId: stakeId, amount: EchoDatum.toDatum(amount))
        let result: TxResultWire = try await apiClient.post(endpoint: WalletEndpoint.unstake, body: body)
        cached = nil
        return result.txHash
    }

    func submitRewardClaim(rewardTypes: [String]) async throws -> String {
        let result: TxResultWire = try await apiClient.post(
            endpoint: WalletEndpoint.claim,
            body: ClaimRequestWire(types: rewardTypes)
        )
        cached = nil
        return result.txHash
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
                vestingPercent: v.vestingPercent
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
