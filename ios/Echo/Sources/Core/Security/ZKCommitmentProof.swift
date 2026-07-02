#if os(iOS)
import CryptoKit
import Foundation

enum ZKCommitmentProof {
    /// Client-side commitment H(subject||claim||nonce) matching Go verifier (WO-236).
    static func build(subjectDID: String, claimType: String, nonce: String) -> String {
        var data = Data()
        data.append(Data(subjectDID.utf8))
        data.append(0)
        data.append(Data(claimType.utf8))
        data.append(0)
        data.append(Data(nonce.utf8))
        let digest = SHA256.hash(data: data)
        return digest.map { String(format: "%02x", $0) }.joined()
    }
}
#endif
