#if os(iOS)
import Foundation

/// Shared credential-enrollment tail: Secure Enclave key → did → wallet → passkey (WO-100).
@MainActor
enum CredentialEnrollmentTailService {
    struct Result: Sendable {
        let did: String
        let walletAddress: String
        let publicKeyHex: String
        let trustTier: Int
    }

    enum Step: Sendable {
        case creatingIdentity
        case creatingWallet
        case registeringPasskey
    }

    static func run(
        bundle: VerifiedIdentityBundle,
        api: EnrollmentAPIClient = .shared,
        passkeyManager: PasskeyManagerProtocol = PasskeyManager(),
        onStep: ((Step) -> Void)? = nil
    ) async throws -> Result {
        onStep?(.creatingIdentity)
        let publicKeyBase64 = try await SecureEnclaveManager.shared
            .generateBiometricProtectedKey(id: identityKeyId)
        guard let pubData = Data(base64Encoded: publicKeyBase64) else {
            throw EnrollmentError.transportFailed(underlying: "Invalid Secure Enclave public key encoding")
        }
        let publicKeyHex = pubData.map { String(format: "%02x", $0) }.joined()

        let challenge = Data("echo-enrollment-verify-\(UUID().uuidString)".utf8)
        _ = try await SecureEnclaveManager.shared.sign(data: challenge, keyId: identityKeyId)

        let did = try await api.registerDIDFromBundle(bundle, publicKeyHex: publicKeyHex)

        onStep?(.creatingWallet)
        let walletAddress = try await api.createWalletForDID(did: did)

        onStep?(.registeringPasskey)
        let passkey = try await passkeyManager.createPasskey(
            challenge: challenge,
            userId: Data(did.utf8),
            userName: displayName(from: bundle)
        )
        try await api.registerPasskey(
            did: did,
            publicKeyHex: publicKeyHex,
            attestation: passkey
        )

        try await persistLocalSession(did: did, bundle: bundle, publicKeyHex: publicKeyHex)

        return Result(
            did: did,
            walletAddress: walletAddress,
            publicKeyHex: publicKeyHex,
            trustTier: bundle.assuranceLevel.trustTier
        )
    }

    private static let identityKeyId = "echo-identity-signing"

    private static func displayName(from bundle: VerifiedIdentityBundle) -> String {
        if let given = bundle.disclosedClaims["givenName"], !given.isEmpty {
            if let family = bundle.disclosedClaims["familyName"], !family.isEmpty {
                return "\(given) \(family)"
            }
            return given
        }
        return "ECHO User"
    }

    private static func persistLocalSession(
        did: String,
        bundle: VerifiedIdentityBundle,
        publicKeyHex: String
    ) async throws {
        let tier = bundle.assuranceLevel.trustTier
        UserDefaults.standard.set(did, forKey: "echo.did")
        UserDefaults.standard.set(tier, forKey: "echo.trustTier")
        UserDefaults.standard.set(true, forKey: "echo.hasCompletedFirstRun")

        try await KeychainManager.shared.store(key: "echo.did.current", value: did)
        if let name = displayName(from: bundle).nonEmpty {
            try await KeychainManager.shared.store(key: "echo.username.current", value: name)
            UserDefaults.standard.set(name, forKey: "echo.displayName")
        }

        _ = publicKeyHex
    }
}

private extension String {
    var nonEmpty: String? {
        let trimmed = trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }
}
#endif
