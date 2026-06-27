#if os(iOS)
// Core/Stargazer/WalletProvisioner.swift
// Links a deterministic DAG address to the authenticated DID until Stargazer SPM ships.

import CryptoKit
import Foundation

/// Provisions and persists the user's Constellation wallet address (T7 public).
actor WalletProvisioner {
    static let dagAddressKey = WalletKeychain.dagAddressKey

    private let apiClient: APIClient
    private let keychain: KeychainManager

    init(apiClient: APIClient, keychain: KeychainManager = .shared) {
        self.apiClient = apiClient
        self.keychain = keychain
    }

    /// Returns a stable DAG address for `did`, linking it on the backend when needed.
    func ensureWalletLinked(did: String) async throws -> String {
        if let existing = try? await keychain.retrieve(key: Self.dagAddressKey, as: String.self),
           !existing.isEmpty {
            return existing
        }

        let address = Self.deterministicDAGAddress(did: did)
        struct LinkBody: Encodable { let address: String }
        struct LinkResp: Decodable { let did: String; let address: String }
        let _: LinkResp = try await apiClient.post(
            endpoint: WalletEndpoint.link,
            body: LinkBody(address: address)
        )
        try await keychain.store(key: Self.dagAddressKey, value: address)
        return address
    }

    /// Mirrors Go `deterministicDAGAddress` (enrollment_handlers.go).
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
