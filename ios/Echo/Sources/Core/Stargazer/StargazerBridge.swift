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
}

// MARK: - Mock for Testing

#if DEBUG
final class MockWalletAPIClient: WalletAPIClient {
    var balance = BalanceInfo(total: 1250, available: 750)
    var locks: [TokenLockPosition] = []
    var delegations: [DelegationPosition] = []
    var validators: [ValidatorInfo] = []
    var txHash = "mock_tx_hash"
    var shouldError = false

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
}
#endif
