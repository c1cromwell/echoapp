#if os(iOS)
// Features/Onboarding/Recovery/RecoveryService.swift

import Foundation

enum RecoveryError: Error {
    case deviceAlreadyEnrolled
    case walletDerivationFailed
    case backendRejected(String)
}

@MainActor
final class RecoveryService {
    static let shared = RecoveryService()

    private let secureEnclave = SecureEnclaveManager.shared
    private let passkey = PasskeyManager()

    private init() {}

    /// Restores the user's ECHO identity from a BIP-39 recovery phrase.
    ///
    /// Steps:
    /// 1. Validate this device doesn't already have an enrolled DID.
    /// 2. Restore the Constellation wallet from the phrase via StargazerBridge.
    /// 3. Generate a new Secure Enclave identity key for this device.
    /// 4. Call the backend to re-bind the existing DID to the new device key.
    /// 5. Register a new WebAuthn passkey for this device.
    /// 6. Persist identity artifacts locally and mark first-run complete.
    func restore(from phrase: RecoveryPhrase) async throws -> RestoredIdentity {
        if UserDefaults.standard.string(forKey: "echo.did") != nil {
            throw RecoveryError.deviceAlreadyEnrolled
        }

        let (walletAddress, _) = try await StargazerBridgeForRecovery.shared.restoreWallet(from: phrase)

        let sePublicKey = try await secureEnclave.createIdentityKey()

        let response = try await callRestoreDID(
            walletAddress: walletAddress,
            newDevicePublicKey: sePublicKey
        )

        try await passkey.register(did: response.did, displayName: response.displayName)

        // Persist and mark phrase as exported — the user just proved they have it.
        UserDefaults.standard.set(response.did,         forKey: "echo.did")
        UserDefaults.standard.set(response.displayName, forKey: "echo.displayName")
        UserDefaults.standard.set(true,                 forKey: "echo.hasCompletedFirstRun")
        UserDefaults.standard.set(response.trustTier,   forKey: "echo.trustTier")
        UserDefaults.standard.set(Date(),               forKey: "echo.recoveryPhraseExportedAt")

        RecoveryPromptScheduler.shared.cancelAllPendingReminders()

        return RestoredIdentity(
            did: response.did,
            walletAddress: walletAddress,
            displayName: response.displayName
        )
    }

    // MARK: - Backend call

    private struct RestoreDIDResponse: Decodable {
        let did: String
        let displayName: String
        let trustTier: Int

        enum CodingKeys: String, CodingKey {
            case did
            case displayName = "display_name"
            case trustTier   = "trust_tier"
        }
    }

    private func callRestoreDID(walletAddress: String, newDevicePublicKey: Data) async throws -> RestoreDIDResponse {
        // Stub: in production, first call /v1/auth/restore-challenge to get a nonce,
        // sign it with the wallet key, then POST to /v1/auth/restore-did.
        return RestoreDIDResponse(
            did: "did:key:z6MkfThhVKtpeT52HiPfJT8oLjTJiEwPvY+zx9dp99V4RCHq",
            displayName: UserDefaults.standard.string(forKey: "echo.displayName") ?? "Restored User",
            trustTier: 1
        )
    }
}

// MARK: - PasskeyManager extension for recovery

private extension PasskeyManager {
    func register(did: String, displayName: String) async throws {
        // Production: run WebAuthn registration ceremony bound to the given DID.
    }
}

// MARK: - SecureEnclaveManager extension for recovery

private extension SecureEnclaveManager {
    func createIdentityKey() throws -> Data {
        // Production: create a new P-256 key in Secure Enclave and return the 65-byte public key.
        return Data(count: 65)
    }
}

#endif
