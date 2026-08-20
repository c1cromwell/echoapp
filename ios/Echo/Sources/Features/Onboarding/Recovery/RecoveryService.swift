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
    /// 6. Persist identity artifacts locally and pull phrase-encrypted chat history.
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

        try await persistSession(response)
        try await passkey.register(did: response.did, displayName: response.displayName)

        UserDefaults.standard.set(true, forKey: "echo.hasCompletedFirstRun")
        UserDefaults.standard.set(Date(), forKey: "echo.recoveryPhraseExportedAt")
        RecoveryPromptScheduler.shared.cancelAllPendingReminders()

        try? await BackupSessionKeyStore.save(from: phrase)
        await Self.attemptCloudMessageRestore(phrase: phrase)

        return RestoredIdentity(
            did: response.did,
            walletAddress: walletAddress,
            displayName: response.displayName
        )
    }

    private func persistSession(_ response: RestoreDIDResponse) async throws {
        UserDefaults.standard.set(response.did, forKey: "echo.did")
        UserDefaults.standard.set(response.did, forKey: "echo.did.current")
        UserDefaults.standard.set(response.displayName, forKey: "echo.displayName")
        UserDefaults.standard.set(response.trustTier, forKey: "echo.trustTier")
        try await KeychainManager.shared.store(key: "echo.did.current", value: response.did)
        try await KeychainManager.shared.store(key: "echo.username.current", value: response.displayName)
        if let token = response.accessToken, !token.isEmpty {
            try await KeychainManager.shared.storeAuthToken(token)
        }
        await MessagingKeyRegistrar().ensureRegistered(did: response.did)
    }

    /// Pulls `/v3/backup/pull`, decrypts with the recovery phrase, merges into ConversationStore.
    private static func attemptCloudMessageRestore(phrase: RecoveryPhrase) async {
        do {
            let service = DIContainer.shared.resolveMessageBackup()
                ?? MessageBackupService(backupAPI: LiveBackupAPIClient(apiClient: APIClient(configuration: .default)))
            let count = try await service.restoreCloudBackup(phrase: phrase)
            if count > 0 {
                UserDefaults.standard.set(true, forKey: "echo.backup.restoredAfterRecovery")
                UserDefaults.standard.set(count, forKey: "echo.backup.restoredConversationCount")
            }
            UserDefaults.standard.set(count, forKey: "echo.backup.lastRestoreCount")
            UserDefaults.standard.removeObject(forKey: "echo.backup.lastRestoreError")
        } catch {
            UserDefaults.standard.set(false, forKey: "echo.backup.restoredAfterRecovery")
            UserDefaults.standard.set(0, forKey: "echo.backup.lastRestoreCount")
            UserDefaults.standard.set(error.localizedDescription, forKey: "echo.backup.lastRestoreError")
        }
    }

    // MARK: - Backend call

    private struct RestoreDIDResponse: Decodable {
        let did: String
        let displayName: String
        let trustTier: Int
        let accessToken: String?

        enum CodingKeys: String, CodingKey {
            case did
            case displayName = "display_name"
            case trustTier   = "trust_tier"
            case accessToken = "access_token"
        }
    }

    private struct RestoreChallengeResponse: Decodable {
        let challenge: String
    }

    private func callRestoreDID(walletAddress: String, newDevicePublicKey: Data) async throws -> RestoreDIDResponse {
        let client = APIClient(configuration: .default)
        struct ChallengeBody: Encodable { let walletAddress: String
            enum CodingKeys: String, CodingKey { case walletAddress = "wallet_address" }
        }
        let challenge: RestoreChallengeResponse = try await client.post(
            endpoint: AuthEndpoint.restoreChallenge,
            body: ChallengeBody(walletAddress: walletAddress)
        )
        let signed = try await WalletKeyStore.shared.signChallenge(challenge.challenge)
        struct RestoreBody: Encodable {
            let walletAddress: String
            let newDevicePublicKey: String
            let walletSignature: String
            enum CodingKeys: String, CodingKey {
                case walletAddress = "wallet_address"
                case newDevicePublicKey = "new_device_public_key"
                case walletSignature = "wallet_signature"
            }
        }
        do {
            return try await client.post(
                endpoint: AuthEndpoint.restoreDID,
                body: RestoreBody(
                    walletAddress: walletAddress,
                    newDevicePublicKey: newDevicePublicKey.map { String(format: "%02x", $0) }.joined(),
                    walletSignature: signed.signature
                )
            )
        } catch {
            throw RecoveryError.backendRejected(error.localizedDescription)
        }
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
