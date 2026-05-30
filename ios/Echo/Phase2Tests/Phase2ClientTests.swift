#if os(iOS)
import XCTest
import CryptoKit
@testable import Echo

// MARK: - WO-100 OIDC4VC client

final class OIDC4VCClientTests: XCTestCase {
    func testDefaultSubmissionUsesJWTVcFormat() {
        let submission = OIDC4VCClient.defaultSubmission(credentialType: "KYCLite")
        XCTAssertEqual(submission.definition_id, "pres_def_kyclite")
        XCTAssertEqual(submission.descriptor_map.first?.format, "jwt_vc_json")
    }

    func testFetchVerifierMetadataDecodes() async throws {
        let config = URLSessionConfiguration.ephemeral
        config.protocolClasses = [OIDC4VCStubProtocol.self]
        OIDC4VCStubProtocol.handler = { request in
            XCTAssertTrue(request.url?.path.contains("openid-credential-verifier") == true)
            let body = #"{"verifier_id":"echo","credential_types_supported":["KYCLite"]}"#
            let response = HTTPURLResponse(
                url: request.url!,
                statusCode: 200,
                httpVersion: nil,
                headerFields: ["Content-Type": "application/json"]
            )!
            return (response, Data(body.utf8))
        }
        let session = URLSession(configuration: config)
        let client = OIDC4VCClient(baseURL: URL(string: "https://api.test.local/v1")!, session: session)
        let meta = try await client.fetchVerifierMetadata()
        XCTAssertEqual(meta.verifier_id, "echo")
    }
}

private final class OIDC4VCStubProtocol: URLProtocol {
    nonisolated(unsafe) static var handler: ((URLRequest) -> (HTTPURLResponse, Data))?

    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }

    override func startLoading() {
        guard let handler = Self.handler else {
            client?.urlProtocol(self, didFailWithError: URLError(.badURL))
            return
        }
        let (response, data) = handler(request)
        client?.urlProtocol(self, didReceive: response, cacheStoragePolicy: .notAllowed)
        client?.urlProtocol(self, didLoad: data)
        client?.urlProtocolDidFinishLoading(self)
    }

    override func stopLoading() {}
}

// MARK: - WO-221 PSI discovery

final class ContactDiscoveryServiceTests: XCTestCase {
    func testDiscoverMatchesOPRFKeyInIndex() async throws {
        let phone = "+15551234567"
        let normalized = PhoneNormalizer.normalize(phone)
        let key = mockOPRFKey(for: normalized)

        let oprf = MockOPRFClient()
        let psi = StubPSIEvaluator(index: [key: "did:key:zFriend"])
        let service = ContactDiscoveryService(oprf: oprf, api: psi)

        let matches = try await service.discover(normalized: [
            .init(e164: normalized, label: "Friend")
        ])
        XCTAssertEqual(matches.count, 1)
        XCTAssertEqual(matches[0].did, "did:key:zFriend")
        XCTAssertEqual(matches[0].phoneE164, normalized)
    }
}

private struct StubPSIEvaluator: ContactDiscoveryEvaluating {
    let index: [String: String]

    func evaluate(blinded: [String]) async throws -> ContactDiscoveryAPIClient.PSIResponse {
        ContactDiscoveryAPIClient.PSIResponse(
            evaluated: blinded,
            index: index,
            request_id: "test"
        )
    }
}

private func mockOPRFKey(for phone: String) -> String {
    let digest = SHA256.hash(data: Data("mock-oprf:\(phone)".utf8))
    return digest.map { String(format: "%02x", $0) }.joined()
}
#endif
