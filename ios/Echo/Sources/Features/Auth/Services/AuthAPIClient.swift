import Foundation

// MARK: - Protocol

protocol AuthAPIClientProtocol {
    func registerPhone(phone: String, countryCode: String, deviceInfo: AuthDeviceInfo)
        async throws -> PhoneRegistrationResponse
    func verifyOTP(verificationId: String, code: String, deviceInfo: AuthDeviceInfo)
        async throws -> AuthTokenResponse
    func registerPasskey(attestation: PasskeyRegistrationResult, deviceInfo: AuthDeviceInfo)
        async throws -> AuthTokenResponse
    func getLoginChallenge() async throws -> ChallengeResponse
    func login(assertion: PasskeyAssertionResult, deviceInfo: AuthDeviceInfo)
        async throws -> AuthTokenResponse
    func refreshToken(_ token: String) async throws -> AuthTokenResponse
    func logout(token: String, allDevices: Bool) async throws
    func fetchCurrentUser(token: String) async throws -> AuthUserProfile
    func listDevices(token: String) async throws -> [DeviceSession]
    func revokeDevice(id: String, elevatedToken: String) async throws
    func getAuditLog(token: String) async throws -> [AuthAuditLogEntry]
    func requestStepUp(action: String, assertion: PasskeyAssertionResult,
                       token: String) async throws -> StepUpTokenResponse
}

// MARK: - Step-Up Response

struct StepUpTokenResponse: Decodable {
    let elevatedToken: String
    let expiresAt: Date
    let action: String

    enum CodingKeys: String, CodingKey {
        case elevatedToken = "elevated_token"
        case expiresAt = "expires_at"
        case action
    }
}

// MARK: - Implementation

final class AuthAPIClient: AuthAPIClientProtocol {
    private let baseURL: URL
    private let session: URLSession
    private let deviceService: DeviceFingerprintService
    private let decoder: JSONDecoder

    init(
        baseURL: URL = URL(string: "https://api.echo.app/v1")!,
        certificatePinner: CertificatePinner = CertificatePinner()
    ) {
        self.baseURL = baseURL
        self.deviceService = DeviceFingerprintService()
        self.session = URLSession(
            configuration: .default,
            delegate: certificatePinner,
            delegateQueue: nil
        )
        self.decoder = JSONDecoder()
        self.decoder.dateDecodingStrategy = .iso8601
    }

    // MARK: - Generic Request Builder

    private func makeRequest(
        path: String,
        method: String = "POST",
        body: Encodable? = nil,
        token: String? = nil
    ) async throws -> Data {
        var request = URLRequest(url: baseURL.appendingPathComponent(path))
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(
            deviceService.deviceInfoHeader(),
            forHTTPHeaderField: "X-Device-Info"
        )
        if let token {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        if let body {
            let encoder = JSONEncoder()
            encoder.keyEncodingStrategy = .convertToSnakeCase
            request.httpBody = try encoder.encode(body)
        }

        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw AuthAPIError.networkError
        }

        if !(200...299).contains(httpResponse.statusCode) {
            let errorResponse = try? decoder.decode(AuthErrorResponse.self, from: data)
            throw AuthAPIError(
                code: errorResponse?.error.code ?? "UNKNOWN",
                userMessage: errorResponse?.error.message ?? "Something went wrong.",
                httpStatus: httpResponse.statusCode
            )
        }

        return data
    }

    // MARK: - Phone Registration

    func registerPhone(
        phone: String,
        countryCode: String,
        deviceInfo: AuthDeviceInfo
    ) async throws -> PhoneRegistrationResponse {
        struct Body: Encodable {
            let phoneNumber: String
            let countryCode: String
            let deviceInfo: AuthDeviceInfo
        }
        let data = try await makeRequest(
            path: "auth/phone/register",
            body: Body(phoneNumber: phone, countryCode: countryCode, deviceInfo: deviceInfo)
        )
        return try decoder.decode(PhoneRegistrationResponse.self, from: data)
    }

    // MARK: - OTP Verification

    func verifyOTP(
        verificationId: String,
        code: String,
        deviceInfo: AuthDeviceInfo
    ) async throws -> AuthTokenResponse {
        struct Body: Encodable {
            let verificationId: String
            let code: String
            let deviceInfo: AuthDeviceInfo
        }
        let data = try await makeRequest(
            path: "auth/phone/verify",
            body: Body(verificationId: verificationId, code: code, deviceInfo: deviceInfo)
        )
        return try decoder.decode(AuthTokenResponse.self, from: data)
    }

    // MARK: - Passkey Registration

    func registerPasskey(
        attestation: PasskeyRegistrationResult,
        deviceInfo: AuthDeviceInfo
    ) async throws -> AuthTokenResponse {
        struct CredentialData: Encodable {
            let id: String
            let rawId: String
            let clientDataJSON: String
            let attestationObject: String
        }
        struct Body: Encodable {
            let attestationResponse: CredentialData
            let deviceInfo: AuthDeviceInfo
        }
        let body = Body(
            attestationResponse: CredentialData(
                id: attestation.credentialId.base64EncodedString(),
                rawId: attestation.rawId.base64EncodedString(),
                clientDataJSON: attestation.clientDataJSON.base64EncodedString(),
                attestationObject: attestation.attestationObject.base64EncodedString()
            ),
            deviceInfo: deviceInfo
        )
        let data = try await makeRequest(path: "auth/passkey/register", body: body)
        return try decoder.decode(AuthTokenResponse.self, from: data)
    }

    // MARK: - Login Challenge

    func getLoginChallenge() async throws -> ChallengeResponse {
        let data = try await makeRequest(path: "auth/login/challenge", method: "POST")
        return try decoder.decode(ChallengeResponse.self, from: data)
    }

    // MARK: - Login

    func login(
        assertion: PasskeyAssertionResult,
        deviceInfo: AuthDeviceInfo
    ) async throws -> AuthTokenResponse {
        struct CredentialData: Encodable {
            let id: String
            let rawId: String
            let clientDataJSON: String
            let authenticatorData: String
            let signature: String
        }
        struct Body: Encodable {
            let authType: String
            let credential: CredentialData
            let deviceInfo: AuthDeviceInfo
        }
        let body = Body(
            authType: "passkey",
            credential: CredentialData(
                id: assertion.credentialId.base64EncodedString(),
                rawId: assertion.rawId.base64EncodedString(),
                clientDataJSON: assertion.clientDataJSON.base64EncodedString(),
                authenticatorData: assertion.authenticatorData.base64EncodedString(),
                signature: assertion.signature.base64EncodedString()
            ),
            deviceInfo: deviceInfo
        )
        let data = try await makeRequest(path: "auth/login", body: body)
        return try decoder.decode(AuthTokenResponse.self, from: data)
    }

    // MARK: - Refresh

    func refreshToken(_ token: String) async throws -> AuthTokenResponse {
        struct Body: Encodable {
            let refreshToken: String
        }
        let data = try await makeRequest(
            path: "auth/token/refresh",
            body: Body(refreshToken: token)
        )
        return try decoder.decode(AuthTokenResponse.self, from: data)
    }

    // MARK: - Logout

    func logout(token: String, allDevices: Bool) async throws {
        struct Body: Encodable {
            let allDevices: Bool
        }
        _ = try await makeRequest(
            path: "auth/logout",
            body: Body(allDevices: allDevices),
            token: token
        )
    }

    // MARK: - Fetch Current User

    func fetchCurrentUser(token: String) async throws -> AuthUserProfile {
        let data = try await makeRequest(path: "auth/me", method: "GET", token: token)
        return try decoder.decode(AuthUserProfile.self, from: data)
    }

    // MARK: - Device Management

    func listDevices(token: String) async throws -> [DeviceSession] {
        let data = try await makeRequest(path: "auth/devices", method: "GET", token: token)
        return try decoder.decode([DeviceSession].self, from: data)
    }

    func revokeDevice(id: String, elevatedToken: String) async throws {
        _ = try await makeRequest(
            path: "auth/devices/\(id)/revoke",
            token: elevatedToken
        )
    }

    // MARK: - Audit Log

    func getAuditLog(token: String) async throws -> [AuthAuditLogEntry] {
        let data = try await makeRequest(path: "auth/audit-log", method: "GET", token: token)
        return try decoder.decode([AuthAuditLogEntry].self, from: data)
    }

    // MARK: - Step-Up

    func requestStepUp(
        action: String,
        assertion: PasskeyAssertionResult,
        token: String
    ) async throws -> StepUpTokenResponse {
        struct CredentialData: Encodable {
            let id: String
            let rawId: String
            let clientDataJSON: String
            let authenticatorData: String
            let signature: String
        }
        struct Body: Encodable {
            let method: String
            let action: String
            let credential: CredentialData
        }
        let body = Body(
            method: "passkey",
            action: action,
            credential: CredentialData(
                id: assertion.credentialId.base64EncodedString(),
                rawId: assertion.rawId.base64EncodedString(),
                clientDataJSON: assertion.clientDataJSON.base64EncodedString(),
                authenticatorData: assertion.authenticatorData.base64EncodedString(),
                signature: assertion.signature.base64EncodedString()
            )
        )
        let data = try await makeRequest(path: "auth/step-up", body: body, token: token)
        return try decoder.decode(StepUpTokenResponse.self, from: data)
    }
}

// MARK: - Mock for Testing

#if DEBUG
final class MockAuthAPIClient: AuthAPIClientProtocol {
    var phoneResponse: PhoneRegistrationResponse?
    var tokenResponse: AuthTokenResponse?
    var challengeResponse: ChallengeResponse?
    var userProfile: AuthUserProfile?
    var devices: [DeviceSession] = []
    var auditLog: [AuthAuditLogEntry] = []
    var stepUpResponse: StepUpTokenResponse?
    var errorToThrow: Error?

    func registerPhone(phone: String, countryCode: String, deviceInfo: AuthDeviceInfo) async throws -> PhoneRegistrationResponse {
        if let error = errorToThrow { throw error }
        return phoneResponse ?? PhoneRegistrationResponse(
            verificationId: "mock-vid",
            expiresAt: Date().addingTimeInterval(600),
            retryAfter: 60
        )
    }

    func verifyOTP(verificationId: String, code: String, deviceInfo: AuthDeviceInfo) async throws -> AuthTokenResponse {
        if let error = errorToThrow { throw error }
        return tokenResponse ?? mockTokenResponse()
    }

    func registerPasskey(attestation: PasskeyRegistrationResult, deviceInfo: AuthDeviceInfo) async throws -> AuthTokenResponse {
        if let error = errorToThrow { throw error }
        return tokenResponse ?? mockTokenResponse()
    }

    func getLoginChallenge() async throws -> ChallengeResponse {
        if let error = errorToThrow { throw error }
        return challengeResponse ?? ChallengeResponse(
            challenge: Data("mock-challenge".utf8).base64EncodedString(),
            timeout: 60,
            rpId: "echo.app"
        )
    }

    func login(assertion: PasskeyAssertionResult, deviceInfo: AuthDeviceInfo) async throws -> AuthTokenResponse {
        if let error = errorToThrow { throw error }
        return tokenResponse ?? mockTokenResponse()
    }

    func refreshToken(_ token: String) async throws -> AuthTokenResponse {
        if let error = errorToThrow { throw error }
        return tokenResponse ?? mockTokenResponse()
    }

    func logout(token: String, allDevices: Bool) async throws {
        if let error = errorToThrow { throw error }
    }

    func fetchCurrentUser(token: String) async throws -> AuthUserProfile {
        if let error = errorToThrow { throw error }
        return userProfile ?? AuthUserProfile(
            id: "user-1", did: "did:dag:mock", displayName: "Test User",
            username: "testuser", trustScore: 50, trustTier: 2
        )
    }

    func listDevices(token: String) async throws -> [DeviceSession] {
        if let error = errorToThrow { throw error }
        return devices
    }

    func revokeDevice(id: String, elevatedToken: String) async throws {
        if let error = errorToThrow { throw error }
        devices.removeAll { $0.id == id }
    }

    func getAuditLog(token: String) async throws -> [AuthAuditLogEntry] {
        if let error = errorToThrow { throw error }
        return auditLog
    }

    func requestStepUp(action: String, assertion: PasskeyAssertionResult, token: String) async throws -> StepUpTokenResponse {
        if let error = errorToThrow { throw error }
        return stepUpResponse ?? StepUpTokenResponse(
            elevatedToken: "mock-elevated-token",
            expiresAt: Date().addingTimeInterval(300),
            action: action
        )
    }

    private func mockTokenResponse() -> AuthTokenResponse {
        AuthTokenResponse(
            accessToken: "mock-access-token",
            refreshToken: "mock-refresh-token",
            expiresAt: Date().addingTimeInterval(900),
            user: AuthUserProfile(
                id: "user-1", did: "did:dag:mock", displayName: "Test User",
                username: "testuser", trustScore: 50, trustTier: 2
            ),
            passkeyChallenge: nil
        )
    }
}
#endif
