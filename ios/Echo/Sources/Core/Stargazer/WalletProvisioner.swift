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
    private let keyStore: WalletKeyStore

    init(apiClient: APIClient, keychain: KeychainManager = .shared, keyStore: WalletKeyStore = .shared) {
        self.apiClient = apiClient
        self.keychain = keychain
        self.keyStore = keyStore
    }

    /// Returns the user-held DAG address for `did`, generating the wallet key on
    /// first use and linking the address (with proof-of-ownership) on the backend.
    /// In real-funds mode the backend requires the proof; in interim mode it
    /// ignores it (the proof is sent best-effort either way).
    func ensureWalletLinked(did: String) async throws -> String {
        if let existing = try? await keychain.retrieve(key: Self.dagAddressKey, as: String.self),
           !existing.isEmpty {
            return existing
        }

        let account = try await keyStore.ensureWallet()
        let proofHeaders = (try? await buildProofHeaders()) ?? [:]

        struct LinkBody: Encodable { let address: String }
        struct LinkResp: Decodable { let did: String; let address: String }
        let _: LinkResp = try await apiClient.post(
            endpoint: WalletEndpoint.link,
            body: LinkBody(address: account.address),
            headers: proofHeaders
        )
        try await keychain.store(key: Self.dagAddressKey, value: account.address)
        return account.address
    }

    /// Fetches a server challenge, signs it with the wallet key, and returns the
    /// base64(JSON) X-Wallet-Proof header the backend verifier expects.
    private func buildProofHeaders() async throws -> [String: String] {
        struct ChallengeResp: Decodable { let challenge: String; let expiresAt: String }
        let resp: ChallengeResp = try await apiClient.get(endpoint: WalletEndpoint.challenge)
        let signed = try await keyStore.signChallenge(resp.challenge)

        struct Proof: Encodable { let publicKey: String; let challenge: String; let signature: String }
        let proof = Proof(publicKey: signed.publicKey, challenge: resp.challenge, signature: signed.signature)
        let encoded = try JSONEncoder().encode(proof).base64EncodedString()
        return ["X-Wallet-Proof": encoded]
    }

    /// The legacy server-derivable address (DAG = SHA256(did)). Retained only as
    /// the reference value the backend REJECTS in real-funds mode; no longer used
    /// for provisioning. Mirrors Go wallet.ServerDerivableAddress.
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
