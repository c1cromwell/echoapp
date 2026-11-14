#if os(iOS)
import Foundation
import CryptoKit

// MARK: - Certificate Pinner

/// Pins the API host's TLS key when real pins are configured.
///
/// Pins are read from the Info.plist array `EchoPinnedSPKIHashes`, each entry
/// `sha256/<base64>` of the certificate's SubjectPublicKeyInfo — the value
/// produced by:
///
///     openssl s_client -connect api.echo.app:443 -servername api.echo.app </dev/null \
///       | openssl x509 -pubkey -noout | openssl pkey -pubin -outform der \
///       | openssl dgst -sha256 -binary | base64
///
/// Ship at least two (current + backup key). With NO pins configured the
/// pinner does not block anything: connections fall back to the system's
/// standard TLS validation. The previous behaviour — placeholder hashes that
/// could never match, and cancelling every challenge in Release — made every
/// request from a Release build fail.
final class CertificatePinner: NSObject, URLSessionDelegate {

    /// `sha256/<base64>` hashes of trusted SubjectPublicKeyInfo structures.
    private let pinnedKeyHashes: Set<String>

    /// Hosts that require certificate pinning
    private let pinnedHosts: Set<String>

    /// True when at least one real pin is configured.
    var isPinningActive: Bool { !pinnedKeyHashes.isEmpty && !pinnedHosts.isEmpty }

    init(
        pinnedKeyHashes: Set<String> = CertificatePinner.configuredPins(),
        pinnedHosts: Set<String> = CertificatePinner.configuredHosts()
    ) {
        self.pinnedKeyHashes = pinnedKeyHashes
        self.pinnedHosts = pinnedHosts
    }

    /// Pins from Info.plist `EchoPinnedSPKIHashes`; malformed entries are dropped.
    static func configuredPins(bundle: Bundle = .main) -> Set<String> {
        let raw = bundle.object(forInfoDictionaryKey: "EchoPinnedSPKIHashes") as? [String] ?? []
        return Set(raw.filter(isWellFormedPin))
    }

    /// A pin is `sha256/` + base64 of exactly 32 bytes.
    static func isWellFormedPin(_ pin: String) -> Bool {
        guard pin.hasPrefix("sha256/"),
              let data = Data(base64Encoded: String(pin.dropFirst("sha256/".count)))
        else { return false }
        return data.count == 32
    }

    /// The host of the configured API base URL.
    static func configuredHosts() -> Set<String> {
        guard let host = EchoAPIBaseURL.resolved.host, !host.isEmpty else { return [] }
        return [host]
    }

    /// Decides what to do with a server-trust challenge. Pure, so it is unit-tested.
    enum Decision: Equatable { case defaultHandling, evaluatePin }

    func decision(forHost host: String, authenticationMethod: String) -> Decision {
        guard isPinningActive,
              authenticationMethod == NSURLAuthenticationMethodServerTrust,
              pinnedHosts.contains(host)
        else { return .defaultHandling }
        return .evaluatePin
    }

    func urlSession(
        _ session: URLSession,
        didReceive challenge: URLAuthenticationChallenge
    ) async -> (URLSession.AuthChallengeDisposition, URLCredential?) {

        #if DEBUG
        // Allow any certificate in debug builds for testing with proxies
        if let serverTrust = challenge.protectionSpace.serverTrust {
            return (.useCredential, URLCredential(trust: serverTrust))
        }
        return (.performDefaultHandling, nil)
        #else
        let space = challenge.protectionSpace
        guard decision(forHost: space.host, authenticationMethod: space.authenticationMethod) == .evaluatePin,
              let serverTrust = space.serverTrust
        else {
            // Not a pinned host / no pins configured: normal system TLS validation.
            return (.performDefaultHandling, nil)
        }

        // Evaluate the trust chain
        var error: CFError?
        guard SecTrustEvaluateWithError(serverTrust, &error) else {
            return (.cancelAuthenticationChallenge, nil)
        }

        // Extract server certificate's public key
        guard let certChain = SecTrustCopyCertificateChain(serverTrust)
                as? [SecCertificate],
              let serverCert = certChain.first,
              let serverKey = SecCertificateCopyKey(serverCert),
              let serverKeyHash = Self.spkiHash(of: serverKey)
        else {
            return (.cancelAuthenticationChallenge, nil)
        }

        if pinnedKeyHashes.contains(serverKeyHash) {
            return (.useCredential, URLCredential(trust: serverTrust))
        }

        // Pin mismatch — reject connection
        return (.cancelAuthenticationChallenge, nil)
        #endif
    }

    /// ASN.1 SubjectPublicKeyInfo headers. `SecKeyCopyExternalRepresentation`
    /// returns only the raw key, but a pin is the hash of the whole SPKI, so the
    /// algorithm header has to be put back before hashing.
    private static let spkiHeaders: [String: [UInt8]] = [
        "EC-256": [0x30, 0x59, 0x30, 0x13, 0x06, 0x07, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02, 0x01,
                   0x06, 0x08, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x03, 0x01, 0x07, 0x03, 0x42, 0x00],
        "EC-384": [0x30, 0x76, 0x30, 0x10, 0x06, 0x07, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02, 0x01,
                   0x06, 0x05, 0x2b, 0x81, 0x04, 0x00, 0x22, 0x03, 0x62, 0x00],
        "RSA-2048": [0x30, 0x82, 0x01, 0x22, 0x30, 0x0d, 0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7,
                     0x0d, 0x01, 0x01, 0x01, 0x05, 0x00, 0x03, 0x82, 0x01, 0x0f, 0x00],
        "RSA-4096": [0x30, 0x82, 0x02, 0x22, 0x30, 0x0d, 0x06, 0x09, 0x2a, 0x86, 0x48, 0x86, 0xf7,
                     0x0d, 0x01, 0x01, 0x01, 0x05, 0x00, 0x03, 0x82, 0x02, 0x0f, 0x00]
    ]

    /// `sha256/<base64>` of the key's SubjectPublicKeyInfo, or nil for a key type
    /// we cannot rebuild the SPKI for (the caller then rejects the connection).
    static func spkiHash(of key: SecKey) -> String? {
        guard let attributes = SecKeyCopyAttributes(key) as? [CFString: Any],
              let type = attributes[kSecAttrKeyType] as? String,
              let bits = attributes[kSecAttrKeySizeInBits] as? Int,
              let keyData = SecKeyCopyExternalRepresentation(key, nil) as Data?
        else { return nil }
        let family = type == (kSecAttrKeyTypeRSA as String) ? "RSA" : "EC"
        guard let header = spkiHeaders["\(family)-\(bits)"] else { return nil }
        return spkiHash(header: header, rawKey: keyData)
    }

    static func spkiHash(header: [UInt8], rawKey: Data) -> String {
        var spki = Data(header)
        spki.append(rawKey)
        return "sha256/" + Data(SHA256.hash(data: spki)).base64EncodedString()
    }
}
#endif
