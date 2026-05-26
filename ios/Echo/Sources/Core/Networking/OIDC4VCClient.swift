#if os(iOS)
import Foundation

/// OIDC4VC verifier client (WO-100). Talks to `/verification/*` on the API gateway.
struct OIDC4VCClient: Sendable {
    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init(baseURL: URL? = nil, session: URLSession = .shared) {
        if let baseURL {
            self.baseURL = baseURL
        } else if let raw = ProcessInfo.processInfo.environment["ECHO_API_URL"],
                  let url = URL(string: raw) {
            self.baseURL = url
        } else {
            self.baseURL = APIConfiguration.default.baseURL
        }
        self.session = session
        self.decoder = JSONDecoder()
        self.encoder = JSONEncoder()
    }

    struct VerifierMetadata: Decodable, Sendable {
        let verifier_id: String
        let credential_types_supported: [String]?
    }

    struct PresentationRequest: Decodable, Sendable {
        let client_id: String
        let redirect_uri: String
        let state: String
        let nonce: String?
        let presentation_definition: PresentationDefinition?
    }

    struct PresentationDefinition: Decodable, Sendable {
        let id: String
        let input_descriptors: [InputDescriptor]?
    }

    struct InputDescriptor: Decodable, Sendable {
        let id: String
        let name: String?
    }

    struct PresentationSubmission: Encodable, Sendable {
        let id: String
        let definition_id: String
        let descriptor_map: [DescriptorMap]
    }

    struct DescriptorMap: Encodable, Sendable {
        let id: String
        let format: String
        let path: String
    }

    struct SubmitResponse: Decodable, Sendable {
        let presentationId: String?
        let status: String?
        let verificationResult: VerificationResult?
    }

    struct VerificationResult: Decodable, Sendable {
        let isValid: Bool?
        let holderDid: String?
    }

    func fetchVerifierMetadata() async throws -> VerifierMetadata {
        try await get(path: "/.well-known/openid-credential-verifier")
    }

    func createPresentationRequest(
        credentialType: String,
        redirectURI: String = "echo-enroll://callback"
    ) async throws -> PresentationRequest {
        var components = URLComponents(
            url: baseURL.appendingPathComponent("verification/request"),
            resolvingAgainstBaseURL: false
        )!
        components.queryItems = [
            URLQueryItem(name: "credential_type", value: credentialType),
            URLQueryItem(name: "redirect_uri", value: redirectURI),
        ]
        guard let url = components.url else { throw URLError(.badURL) }
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        return try await decode(OIDC4VCClient.PresentationRequest.self, from: request)
    }

    func submitPresentation(
        vpToken: String,
        state: String,
        submission: PresentationSubmission
    ) async throws -> SubmitResponse {
        struct Body: Encodable {
            let vp_token: String
            let state: String
            let presentation_submission: PresentationSubmission
        }
        return try await post(
            path: "/verification/submit",
            body: Body(vp_token: vpToken, state: state, presentation_submission: submission)
        )
    }

    static func defaultSubmission(credentialType: String) -> PresentationSubmission {
        PresentationSubmission(
            id: UUID().uuidString,
            definition_id: "pres_def_\(credentialType.lowercased())",
            descriptor_map: [
                DescriptorMap(id: "credential_0", format: "jwt_vc_json", path: "$.vp.verifiableCredential[0]")
            ]
        )
    }

    // MARK: - Transport

    private func get<R: Decodable>(path: String) async throws -> R {
        var request = URLRequest(url: baseURL.appendingPathComponent(path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))))
        request.httpMethod = "GET"
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        return try await decode(R.self, from: request)
    }

    private func post<B: Encodable, R: Decodable>(path: String, body: B) async throws -> R {
        var request = URLRequest(url: baseURL.appendingPathComponent(path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try encoder.encode(body)
        return try await decode(R.self, from: request)
    }

    private func decode<R: Decodable>(_ type: R.Type, from request: URLRequest) async throws -> R {
        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse, (200 ... 299).contains(http.statusCode) else {
            throw EnrollmentError.backendRejected(
                code: "HTTP_ERROR",
                message: "OIDC4VC request failed"
            )
        }
        return try decoder.decode(R.self, from: data)
    }
}
#endif
