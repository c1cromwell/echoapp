#if os(iOS)
import Foundation

/// Registers the dedicated messaging key-agreement public key as an additional device
/// (labelled `MessagingAgreementKey.deviceLabel`) so peers encrypt private messages to
/// it. Without this, peers would resolve the identity signing key — whose private half
/// is locked in the Secure Enclave and cannot drive ECDH on hardware.
///
/// Contract: `POST /identity/devices` (WO-273). The exact request body bytes are signed
/// with the identity key (`echo-identity-signing`); the signature is sent hex-encoded in
/// `X-Identity-Signature`. The backend verifies it against the already-registered signing
/// device (`signing_did`) via ECDSA-P256-SHA256.
struct MessagingKeyRegistrar {
    /// The identity whose key was last registered. Persisted so later calls are a cheap no-op.
    /// It names the identity, not just "done": a new account on the same install has a different
    /// DID and must publish its own key.
    static let registeredDIDKey = "echo.msgKey.registeredDID.v2"

    /// Whether `did` still has to publish its key, given what was registered last.
    static func needsRegistration(did: String, lastRegistered: String?) -> Bool {
        !did.isEmpty && did != lastRegistered
    }

    private let identityKeyId = "echo-identity-signing"
    private let signatureHeader = "X-Identity-Signature"

    /// Best-effort, idempotent. Called at every sign-in (`MessageRelaySession.connect`) and on
    /// chat open: it returns immediately once registered and never throws into the caller.
    ///
    /// It must have run before anyone writes to this account. Until the key is published, a
    /// sender falls back to the identity signing key, and a message encrypted to that cannot be
    /// decrypted on the device.
    func ensureRegistered(did: String) async {
        guard !did.isEmpty else { return }
        guard Self.needsRegistration(did: did, lastRegistered: UserDefaults.standard.string(forKey: Self.registeredDIDKey)) else { return }

        do {
            let pubHex = try await MessagingAgreementKey.publicKeyHex()
            // The bytes we sign MUST equal the bytes we send. Build once, reuse.
            let body = try canonicalBody(subjectDID: did, newPublicKeyHex: pubHex, signingDID: did)
            let signature = try await SecureEnclaveManager.shared.sign(data: body, keyId: identityKeyId)
            let sigHex = signature.map { String(format: "%02x", $0) }.joined()

            var req = URLRequest(url: EchoAPIBaseURL.url(path: "/identity/devices"))
            req.httpMethod = "POST"
            req.setValue("application/json", forHTTPHeaderField: "Content-Type")
            req.setValue(sigHex, forHTTPHeaderField: signatureHeader)
            req.httpBody = body
            req.timeoutInterval = 15

            let (_, response) = try await URLSession.shared.data(for: req)
            guard let http = response as? HTTPURLResponse else { return }
            // 2xx = registered now; 409 = the key is already on file (idempotent success).
            if (200...299).contains(http.statusCode) || http.statusCode == 409 {
                UserDefaults.standard.set(did, forKey: Self.registeredDIDKey)
                await DeviceIdentityStore.pinLocalPublicKey(hex: pubHex)
            }
        } catch {
            // Best-effort: a failure here leaves the flag unset so the next call retries.
        }
    }

    /// Deterministic JSON (sorted keys) so the signed bytes are reproducible.
    private func canonicalBody(subjectDID: String, newPublicKeyHex: String, signingDID: String) throws -> Data {
        let dict: [String: String] = [
            "subject_did": subjectDID,
            "new_public_key_hex": newPublicKeyHex,
            "signing_did": signingDID,
            "device_label": MessagingAgreementKey.deviceLabel,
        ]
        return try JSONSerialization.data(withJSONObject: dict, options: [.sortedKeys])
    }
}
#endif
