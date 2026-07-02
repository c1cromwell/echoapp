#if os(iOS)
import Foundation

/// ZK proof use-case seam with Midnight deferred (WO-235).
protocol ZKProofUseCase: Sendable {
    func verify(claimType: String, proof: String, nonce: String) async throws -> Bool
}

struct StubZKProofUseCase: ZKProofUseCase {
    func verify(claimType: String, proof: String, nonce: String) async throws -> Bool {
        guard !proof.isEmpty, !nonce.isEmpty else { return false }
        guard let token = try? await KeychainManager.shared.getAuthToken(),
              let url = URL(string: APIConfiguration.default.baseURL.absoluteString + "/v3/zk/verify") else {
            return false
        }
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        struct Body: Encodable {
            let subjectDID: String
            let claimType: String
            let proof: String
            let nonce: String
            enum CodingKeys: String, CodingKey {
                case subjectDID = "subject_did"
                case claimType = "claim_type"
                case proof, nonce
            }
        }
        struct Resp: Decodable { let verified: Bool }
        let did = (try? await KeychainManager.shared.retrieve(key: "echo.did.current", as: String.self)) ?? ""
        request.httpBody = try? JSONEncoder().encode(Body(subjectDID: did, claimType: claimType, proof: proof, nonce: nonce))
        guard let (data, response) = try? await URLSession.shared.data(for: request),
              let http = response as? HTTPURLResponse, http.statusCode == 200,
              let decoded = try? JSONDecoder().decode(Resp.self, from: data) else {
            return false
        }
        return decoded.verified
    }
}
#endif
