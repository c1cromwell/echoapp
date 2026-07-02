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

    /// Midnight envelope: commitment-bound proof string for `/v3/zk/verify`.
    static func midnightEnvelope(
        subjectDID: String,
        claimType: String,
        nonce: String,
        circuitProof: String = "echo-circuit-v1"
    ) -> String {
        struct Envelope: Encodable {
            let commitment: String
            let publicSignals: [String]
            let proof: String
            enum CodingKeys: String, CodingKey {
                case commitment, proof
                case publicSignals = "public_signals"
            }
        }
        let env = Envelope(
            commitment: build(subjectDID: subjectDID, claimType: claimType, nonce: nonce),
            publicSignals: [],
            proof: circuitProof
        )
        guard let json = try? JSONEncoder().encode(env) else { return "midnight:" }
        return "midnight:" + json.base64EncodedString()
    }
}
#endif
