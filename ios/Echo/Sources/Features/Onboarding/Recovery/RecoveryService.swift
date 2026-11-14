// Features/Onboarding/Recovery/RecoveryService.swift

import Foundation

enum RecoveryError: Error, Equatable {
    case deviceAlreadyEnrolled
    case walletDerivationFailed
    case backendRejected(String)
    /// The device key that was just generated is not the one stored under the
    /// identity key id, so later requests could not be signed with it.
    case deviceKeyUnavailable
}

// MARK: - Restored-device identity key (platform-neutral so it is unit-tested headlessly)

/// The one key id every signer uses: `PasskeySigningInterceptor`, login,
/// step-up and `MessagingKeyRegistrar`. A restored device must end up with
/// its key under this id, or every signed request throws `keyNotFound`.
enum IdentitySigningKeyID {
    static let current = "echo-identity-signing"
}

/// The slice of `SecureEnclaveManager` that restore needs. Injected in tests.
protocol RecoveryDeviceKeyStore: Sendable {
    /// Creates a device key under `id` and returns its base64 public key.
    /// Any key already stored under `id` is replaced.
    func generateKey(id: String) async throws -> String
    /// The public key currently stored under `id`.
    func publicKey(id: String) async throws -> Data
}

extension SecureEnclaveManager: RecoveryDeviceKeyStore {
    func generateKey(id: String) async throws -> String {
        try await generateBiometricProtectedKey(id: id)
    }

    func publicKey(id: String) async throws -> Data {
        try await getPublicKey(id: id)
    }
}

/// Where restore writes the Keychain values the returning-user login reads.
protocol RecoveryCredentialStore: Sendable {
    func storeString(_ value: String, key: String) async throws
}

struct RecoveryIdentityProvisioner {
    let keyStore: any RecoveryDeviceKeyStore

    /// Creates this device's identity signing key and returns the public key
    /// to send as `new_device_public_key` in restore-did.
    ///
    /// The key is created under `IdentitySigningKeyID.current`, replacing a
    /// stale key left by an earlier install or an unfinished first run (that
    /// key belongs to no identity this device can still use). The stored key
    /// is then read back: the bytes returned here are the bytes the server
    /// binds, so they must be the key later requests are signed with.
    func provisionDeviceKey() async throws -> Data {
        let encoded = try await keyStore.generateKey(id: IdentitySigningKeyID.current)
        guard let generated = Data(base64Encoded: encoded),
              generated.contains(where: { $0 != 0 }) else {
            throw RecoveryError.deviceKeyUnavailable
        }
        let stored = try await keyStore.publicKey(id: IdentitySigningKeyID.current)
        guard stored == generated else {
            throw RecoveryError.deviceKeyUnavailable
        }
        return generated
    }

    /// Writes the same local state first run writes, so the next launch routes
    /// to the returning-user login (`echo.hasCompletedFirstRun`) and that
    /// screen finds the account (`echo.username.current` in the Keychain).
    static func persistRestoredAccount(
        did: String,
        displayName: String,
        trustTier: Int,
        defaults: UserDefaults,
        credentials: any RecoveryCredentialStore
    ) async throws {
        let username = displayName.trimmingCharacters(in: .whitespacesAndNewlines)
        defaults.set(did, forKey: "echo.did")
        defaults.set(did, forKey: "echo.did.current")
        defaults.set(username, forKey: "echo.displayName")
        defaults.set(trustTier, forKey: "echo.trustTier")
        try await credentials.storeString(did, key: "echo.did.current")
        if !username.isEmpty {
            try await credentials.storeString(username, key: "echo.username.current")
        }
        defaults.set(true, forKey: "echo.hasCompletedFirstRun")
    }
}

#if os(iOS)
import Security

private struct KeychainRecoveryCredentialStore: RecoveryCredentialStore {
    func storeString(_ value: String, key: String) async throws {
        try await KeychainManager.shared.store(key: key, value: value)
    }
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

        let sePublicKey = try await RecoveryIdentityProvisioner(keyStore: secureEnclave)
            .provisionDeviceKey()

        let response = try await callRestoreDID(
            walletAddress: walletAddress,
            newDevicePublicKey: sePublicKey
        )

        try await persistSession(response)
        // The device key above is what later requests are signed with. The
        // WebAuthn ceremony is attempted on a real phone and does not undo a
        // restore that already bound that key.
        try? await passkey.register(did: response.did, displayName: response.displayName)

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
        try await RecoveryIdentityProvisioner.persistRestoredAccount(
            did: response.did,
            displayName: response.displayName,
            trustTier: response.trustTier,
            defaults: .standard,
            credentials: KeychainRecoveryCredentialStore()
        )
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
            let walletPublicKey: String
            let newDevicePublicKey: String
            let walletSignature: String
            enum CodingKeys: String, CodingKey {
                case walletAddress = "wallet_address"
                case walletPublicKey = "wallet_public_key"
                case newDevicePublicKey = "new_device_public_key"
                case walletSignature = "wallet_signature"
            }
        }
        do {
            return try await client.post(
                endpoint: AuthEndpoint.restoreDID,
                body: RestoreBody(
                    walletAddress: walletAddress,
                    // The server verifies the signature against this key and
                    // checks that it derives `walletAddress`.
                    walletPublicKey: signed.publicKey,
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
        #if targetEnvironment(simulator)
        UserDefaults.standard.set(did, forKey: "echo.passkey.registeredDID")
        #else
        var challenge = Data(count: 32)
        let status = challenge.withUnsafeMutableBytes { SecRandomCopyBytes(kSecRandomDefault, 32, $0.baseAddress!) }
        guard status == errSecSuccess else {
            throw RecoveryError.backendRejected("could not start passkey registration")
        }
        let name = displayName.isEmpty ? did : displayName
        _ = try await createPasskey(challenge: challenge, userId: Data(did.utf8), userName: name)
        #endif
    }
}

#endif
