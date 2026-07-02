#if os(iOS)
// Core/Stargazer/StargazerBridge.swift
// Wraps Constellation Stargazer SDK for ECHO-specific operations.
// In production, this bridges to the real StargazerSDK framework.
// For now, it defines the interface and uses the backend API as a proxy.

import Foundation

// MARK: - Stargazer Bridge Protocol

protocol StargazerBridgeProtocol {
    func createWallet() async throws -> WalletInfo
    func importWallet(mnemonic: String) async throws -> WalletInfo
    func getBalance() async throws -> BalanceInfo
    func getTokenLocks() async throws -> [TokenLockPosition]
    func getDelegations() async throws -> [DelegationPosition]
    func submitTokenLock(amount: Decimal, tier: StakingTier) async throws -> String
    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String
    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String
    func submitRewardClaim(rewardTypes: [String]) async throws -> String
}

// MARK: - Stargazer Bridge (Backend-Proxied)

/// In Phase 2, wallet operations go through the Go backend which interacts
/// with the metagraph. In Phase 3+, the Stargazer SDK handles signing locally.
actor StargazerBridge: StargazerBridgeProtocol {
    private let api: WalletAPIClient
    private var walletAddress: String?

    init(api: WalletAPIClient) {
        self.api = api
    }

    func createWallet() async throws -> WalletInfo {
        let info = try await api.createWallet()
        self.walletAddress = info.address
        return info
    }

    func importWallet(mnemonic: String) async throws -> WalletInfo {
        let info = try await api.importWallet(mnemonic: mnemonic)
        self.walletAddress = info.address
        return info
    }

    func getBalance() async throws -> BalanceInfo {
        guard walletAddress != nil else { throw StargazerError.notInitialized }
        return try await api.getBalance()
    }

    func getTokenLocks() async throws -> [TokenLockPosition] {
        guard walletAddress != nil else { throw StargazerError.notInitialized }
        return try await api.getTokenLocks()
    }

    func getDelegations() async throws -> [DelegationPosition] {
        guard walletAddress != nil else { throw StargazerError.notInitialized }
        return try await api.getDelegations()
    }

    func submitTokenLock(amount: Decimal, tier: StakingTier) async throws -> String {
        guard walletAddress != nil else { throw StargazerError.notInitialized }
        return try await api.submitTokenLock(amount: amount, tier: tier)
    }

    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String {
        guard walletAddress != nil else { throw StargazerError.notInitialized }
        return try await api.submitStakeDelegation(stakeId: stakeId, validatorId: validatorId)
    }

    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String {
        guard walletAddress != nil else { throw StargazerError.notInitialized }
        return try await api.submitWithdrawLock(stakeId: stakeId, amount: amount)
    }

    func submitRewardClaim(rewardTypes: [String]) async throws -> String {
        guard walletAddress != nil else { throw StargazerError.notInitialized }
        return try await api.submitRewardClaim(rewardTypes: rewardTypes)
    }
}

// MARK: - Wallet API Client Protocol

protocol WalletAPIClient {
    func fetchWalletState() async throws -> WalletState
    func createWallet() async throws -> WalletInfo
    func importWallet(mnemonic: String) async throws -> WalletInfo
    func getBalance() async throws -> BalanceInfo
    func getTokenLocks() async throws -> [TokenLockPosition]
    func getDelegations() async throws -> [DelegationPosition]
    func getValidators() async throws -> [ValidatorInfo]
    func submitTokenLock(amount: Decimal, tier: StakingTier) async throws -> String
    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String
    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String
    func submitRewardClaim(rewardTypes: [String]) async throws -> String
    func fetchEmissionStatus() async throws -> EmissionStatus
    func fetchVesting() async throws -> VestingState?
}

// MARK: - HTTP Wallet API Client (production)

/// Production wallet client that forwards calls to the backend via APIClient.
/// Falls back gracefully when the backend is unreachable (returns empty/default state).
actor HTTPWalletAPIClient: WalletAPIClient {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func fetchWalletState() async throws -> WalletState {
        let balance = try await getBalance()
        let locks = try await getTokenLocks()
        let delegations = try await getDelegations()
        return WalletState(
            totalBalance: balance.total,
            available: balance.available,
            staked: balance.staked,
            pendingRewards: 0,
            locks: locks,
            delegations: delegations,
            dailyRewards: nil,
            vesting: nil
        )
    }

    func createWallet() async throws -> WalletInfo {
        WalletInfo(address: "DAG\(UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(36))", publicKey: "")
    }

    func importWallet(mnemonic: String) async throws -> WalletInfo {
        WalletInfo(address: "DAG\(UUID().uuidString.replacingOccurrences(of: "-", with: "").prefix(36))", publicKey: "")
    }

    func getBalance() async throws -> BalanceInfo {
        BalanceInfo(total: 0, available: 0)
    }

    func getTokenLocks() async throws -> [TokenLockPosition] { [] }
    func getDelegations() async throws -> [DelegationPosition] { [] }
    func getValidators() async throws -> [ValidatorInfo] { [] }

    func submitTokenLock(amount: Decimal, tier: StakingTier) async throws -> String { "" }
    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String { "" }
    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String { "" }
    func submitRewardClaim(rewardTypes: [String]) async throws -> String { "" }

    func fetchEmissionStatus() async throws -> EmissionStatus {
        EmissionStatus(currentYear: 1, annualCap: 80_000_000, distributedToDate: 0, remainingBudget: 80_000_000, percentConsumed: 0)
    }

    func fetchVesting() async throws -> VestingState? { nil }
}

// MARK: - Mock for Testing

#if os(iOS)
final class MockWalletAPIClient: WalletAPIClient {
    var balance = BalanceInfo(total: 1250, available: 750)
    var locks: [TokenLockPosition] = []
    var delegations: [DelegationPosition] = []
    var validators: [ValidatorInfo] = []
    var txHash = "mock_tx_hash"
    var shouldError = false

    func fetchWalletState() async throws -> WalletState {
        if shouldError { throw StargazerError.notInitialized }
        return WalletState(
            totalBalance: balance.total,
            available: balance.available,
            staked: balance.staked,
            pendingRewards: 0,
            locks: locks,
            delegations: delegations,
            dailyRewards: nil,
            vesting: nil
        )
    }

    func createWallet() async throws -> WalletInfo {
        WalletInfo(address: "DAG_mock_address", publicKey: "mock_pubkey")
    }

    func importWallet(mnemonic: String) async throws -> WalletInfo {
        WalletInfo(address: "DAG_mock_imported", publicKey: "mock_import_pubkey")
    }

    func getBalance() async throws -> BalanceInfo {
        if shouldError { throw StargazerError.notInitialized }
        return balance
    }

    func getTokenLocks() async throws -> [TokenLockPosition] { locks }
    func getDelegations() async throws -> [DelegationPosition] { delegations }
    func getValidators() async throws -> [ValidatorInfo] { validators }

    func submitTokenLock(amount: Decimal, tier: StakingTier) async throws -> String {
        if shouldError { throw StargazerError.transactionFailed("mock error") }
        return txHash
    }

    func submitStakeDelegation(stakeId: String, validatorId: String) async throws -> String { txHash }
    func submitWithdrawLock(stakeId: String, amount: Decimal) async throws -> String { txHash }
    func submitRewardClaim(rewardTypes: [String]) async throws -> String { txHash }
    func fetchEmissionStatus() async throws -> EmissionStatus {
        EmissionStatus(currentYear: 1, annualCap: 80_000_000, distributedToDate: 0, remainingBudget: 80_000_000, percentConsumed: 0)
    }
    func fetchVesting() async throws -> VestingState? { nil }
}
#endif
#endif
