import Foundation

// MARK: - Auth Response (from server)

struct AuthTokenResponse: Decodable {
    let accessToken: String
    let refreshToken: String?
    let expiresAt: Date
    let user: AuthUserProfile?
    let passkeyChallenge: String?

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case expiresAt = "expires_at"
        case user
        case passkeyChallenge = "passkey_challenge"
    }
}

// MARK: - Phone Registration Response

struct PhoneRegistrationResponse: Decodable {
    let verificationId: String
    let expiresAt: Date
    let retryAfter: Int

    enum CodingKeys: String, CodingKey {
        case verificationId = "verification_id"
        case expiresAt = "expires_at"
        case retryAfter = "retry_after"
    }
}

// MARK: - Challenge Response

struct ChallengeResponse: Decodable {
    let challenge: String
    let timeout: Int
    let rpId: String

    enum CodingKeys: String, CodingKey {
        case challenge
        case timeout
        case rpId = "rp_id"
    }

    var challengeData: Data {
        Data(base64Encoded: challenge) ?? Data()
    }
}

// MARK: - Device Session

struct DeviceSession: Identifiable, Decodable {
    let id: String
    let friendlyName: String
    let platform: String
    let osVersion: String
    let lastIP: String
    let lastLocation: String?
    let lastActiveAt: Date
    let isCurrentDevice: Bool

    enum CodingKeys: String, CodingKey {
        case id
        case friendlyName = "friendly_name"
        case platform
        case osVersion = "os_version"
        case lastIP = "last_ip"
        case lastLocation = "last_location"
        case lastActiveAt = "last_active_at"
        case isCurrentDevice = "is_current_device"
    }
}

// MARK: - Auth Audit Log Entry

struct AuthAuditLogEntry: Identifiable, Decodable {
    let id: String
    let eventType: String
    let result: String
    let ipAddress: String
    let deviceName: String?
    let createdAt: Date

    enum CodingKeys: String, CodingKey {
        case id
        case eventType = "event_type"
        case result
        case ipAddress = "ip_address"
        case deviceName = "device_name"
        case createdAt = "created_at"
    }

    var isSuccess: Bool { result == "success" }
    var isFailed: Bool { result == "failed" }
    var isBlocked: Bool { result == "blocked" }
}
