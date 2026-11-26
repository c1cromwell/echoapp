#if os(iOS)
// Core/Stargazer/WalletProvisioner.swift
// Creates the account recovery key and links its address to the account.

import CryptoKit
import Foundation

/// Creates and persists the account recovery key's address.
actor WalletProvisioner {
    static let dagAddressKey = WalletKeychain.dagAddressKey

    private let apiClient: APIClient
    private let keychain: KeychainManager
    private let keyStore: WalletKeyStore

    init(apiClient: APIClient, keychain: KeychainManager = .shared, keyStore: WalletKeyStore = .shared) {
        self.apiClient = apiClient
        self.keychain = keychain
        self.keyStore = keyStore
    }

    /// Returns the recovery key's address for `did`, creating the key on first use, and tells the
    /// server which account the address belongs to. Restoring the account on a new device depends
    /// on the server knowing that.
    ///
    /// Telling the server is best-effort here so that first run is never held up by it; a link
    /// that did not get through is sent again after sign-in (`RecoveryLinkSync`).
    func ensureWalletLinked(did: String) async throws -> String {
        let account = try await keyStore.ensureWallet()
        try await keychain.store(key: Self.dagAddressKey, value: account.address)
        await RecoveryLinkSync.syncIfNeeded(apiClient: apiClient, did: did)
        return account.address
    }

    /// The placeholder address the server hands out before the real one is linked
    /// (DAG + SHA256(did)). No key corresponds to it. Mirrors Go wallet.ServerDerivableAddress.
    static func deterministicDAGAddress(did: String) -> String {
        let digest = SHA256.hash(data: Data(did.utf8))
        let hex = digest.map { String(format: "%02x", $0) }.joined()
        return "DAG" + String(hex.prefix(36))
    }
}

/// Stargazer provisioning adapter for silent first-run wallet step.
final class RealProvisionStargazer: ProvisionStargazerProtocol, @unchecked Sendable {
    func createWallet() async throws -> String {
        guard let did = try? await KeychainManager.shared.retrieve(key: "echo.did.current", as: String.self) else {
            throw StargazerError.walletCreationFailed
        }
        let client = await MainActor.run {
            DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
        }
        let provisioner = WalletProvisioner(apiClient: client)
        return try await provisioner.ensureWalletLinked(did: did)
    }
}
#endif
