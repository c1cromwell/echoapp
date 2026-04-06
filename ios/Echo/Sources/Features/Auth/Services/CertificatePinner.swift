import Foundation
import CryptoKit

// MARK: - Certificate Pinner

final class CertificatePinner: NSObject, URLSessionDelegate {

    /// SHA-256 hashes of the SPKI of trusted certificates.
    /// Include both current and backup certificate hashes.
    private let pinnedKeyHashes: Set<String>

    /// Hosts that require certificate pinning
    private let pinnedHosts: Set<String>

    init(
        pinnedKeyHashes: Set<String> = [
            // Production certificate (current) — replace with real hash
            "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
            // Backup certificate (rotate in advance)
            "sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="
        ],
        pinnedHosts: Set<String> = ["api.echo.app"]
    ) {
        self.pinnedKeyHashes = pinnedKeyHashes
        self.pinnedHosts = pinnedHosts
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
        guard challenge.protectionSpace.authenticationMethod
                == NSURLAuthenticationMethodServerTrust,
              pinnedHosts.contains(challenge.protectionSpace.host),
              let serverTrust = challenge.protectionSpace.serverTrust
        else {
            return (.cancelAuthenticationChallenge, nil)
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
              let serverKey = SecCertificateCopyKey(serverCert)
        else {
            return (.cancelAuthenticationChallenge, nil)
        }

        // Hash the public key and compare
        let serverKeyHash = hashPublicKey(serverKey)
        if pinnedKeyHashes.contains(serverKeyHash) {
            return (.useCredential, URLCredential(trust: serverTrust))
        }

        // Pin mismatch — reject connection
        return (.cancelAuthenticationChallenge, nil)
        #endif
    }

    private func hashPublicKey(_ key: SecKey) -> String {
        guard let keyData = SecKeyCopyExternalRepresentation(key, nil) as Data? else {
            return ""
        }
        let hash = SHA256.hash(data: keyData)
        return "sha256/" + Data(hash).base64EncodedString()
    }
}
