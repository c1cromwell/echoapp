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
    /// Persisted once registration succeeds, so this is a cheap no-op on later calls.
    static let registeredFlag = "echo.msgKey.registered.v1"

    private let identityKeyId = "echo-identity-signing"
    private let signatureHeader = "X-Identity-Signature"

    /// Best-effort, idempotent. Safe to call on every session start / chat open: it
    /// returns immediately once registered and never throws into the caller.
    func ensureRegistered(did: String) async {
        guard !did.isEmpty else { return }
        guard !UserDefaults.standard.bool(forKey: Self.registeredFlag) else { return }

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
                UserDefaults.standard.set(true, forKey: Self.registeredFlag)
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
