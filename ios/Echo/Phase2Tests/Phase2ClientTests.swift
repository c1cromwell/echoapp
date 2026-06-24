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

// MARK: - WO-100 enrollment errors

final class EnrollmentErrorMappingTests: XCTestCase {
    func testOIDC4VCDisabledMessage() {
        let error = EnrollmentError.backendRejected(
            code: "OIDC4VC_DISABLED",
            message: "OIDC4VC verifier is not enabled"
        )
        XCTAssertTrue(error.errorDescription?.contains("OIDC4VC_ENABLED") == true)
    }
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

// MARK: - TestFlight onboarding helpers

final class DidKeyDeriverTests: XCTestCase {
    func testDeriveMatchesGoPkgDidkeyVector() throws {
        let hex = "042a8db0febf8361d5b16c0bd5711625a78d22af9559d0e987666be09ed521459873ec2364e35aa21dbfeb8a63a0b52b61e5c56fbe06fc7ad8cc2143cb1929189a"
        let did = try DidKeyDeriver.deriveFromPublicKeyHex(hex)
        XCTAssertEqual(did, "did:key:zDnaeTJ5Uuw7xjUqbd8Dt6Gdq1sfEFai13hR8PRQP8RDNxWf1")
    }
}

final class EchoAPIBaseURLTests: XCTestCase {
    func testURLAppendsPathToBase() {
        let url = EchoAPIBaseURL.url(path: "/v1/auth/sms-recovery/register")
        XCTAssertTrue(url.path.hasSuffix("/v1/auth/sms-recovery/register"))
    }

    func testDeleteAccountPathUsesV1() {
        XCTAssertEqual(UserEndpoint.deleteAccount.path, "/v1/users/account")
    }
}

// MARK: - WO-39 contact use cases

final class ContactUseCaseTests: XCTestCase {
    func testEchoDeepLink_parsesInviteCode() {
        let url = URL(string: "echo://invite?code=ABC123")!
        guard case .invite(let code)? = EchoDeepLink.parse(url) else {
            return XCTFail("expected invite deep link")
        }
        XCTAssertEqual(code, "ABC123")
    }

    func testContactDiscoverySync_manualSkipsAutomatic() {
        let previous = ContactDiscoverySyncPreferences.cadence
        ContactDiscoverySyncPreferences.cadence = .manual
        XCTAssertFalse(ContactDiscoverySyncPreferences.shouldRunAutomaticSync())
        ContactDiscoverySyncPreferences.cadence = previous
    }

    func testPrivacySettings_encryptionIndicatorPersists() {
        var settings = EnhancedPrivacySettings()
        settings.showEncryptionIndicator = false
        PrivacySettingsStore.save(settings)
        XCTAssertFalse(PrivacySettingsStore.showsEncryptionIndicator)
        settings.showEncryptionIndicator = true
        PrivacySettingsStore.save(settings)
        XCTAssertTrue(PrivacySettingsStore.showsEncryptionIndicator)
    }

    func testQRContactExchange_parsesEchoProfileURL() {
        let useCase = QRContactExchangeUseCase()
        let parsed = useCase.parseScannedPayload("echo://profile?did=did:key:zPeer&u=alice")
        XCTAssertEqual(parsed?.0, "did:key:zPeer")
        XCTAssertEqual(parsed?.1, "alice")
    }

    func testQRContactExchange_parsesRawDID() {
        let useCase = QRContactExchangeUseCase()
        let parsed = useCase.parseScannedPayload("did:key:zRaw")
        XCTAssertEqual(parsed?.0, "did:key:zRaw")
        XCTAssertNil(parsed?.1)
    }

    @MainActor
    func testHiddenFolderSettings_duressPIN() throws {
        let defaults = UserDefaults(suiteName: "echo.hidden.tests")!
        defaults.removePersistentDomain(forName: "echo.hidden.tests")
        let store = HiddenFolderSettingsStore(defaults: defaults)
        try store.setDuressPIN("1234")
        XCTAssertTrue(store.matchesDuressPIN("1234"))
        XCTAssertFalse(store.matchesDuressPIN("9999"))
    }

    @MainActor
    func testHiddenChatsSession_duressHidesVaultContents() {
        let session = HiddenChatsSession.shared
        session.lock()
        session.unlock(duress: true)
        XCTAssertTrue(session.isDuressMode)
        XCTAssertFalse(session.shouldSurfaceNotification(for: "conv-hidden"))
        session.lock()
    }
}
#endif
