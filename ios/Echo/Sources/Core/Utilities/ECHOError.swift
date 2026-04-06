import Foundation

/// User-facing error codes (no sensitive details exposed).
/// Format: ECHO-{category}{number}
enum ECHOError: Int, LocalizedError {
    // Authentication (1xxx)
    case authFailed = 1001
    case biometricFailed = 1002
    case sessionExpired = 1003

    // Network (2xxx)
    case networkUnavailable = 2001
    case requestTimeout = 2002
    case serverError = 2003
    case relayUnavailable = 2004       // WebSocket relay disconnected

    // Encryption (3xxx)
    case encryptionFailed = 3001
    case decryptionFailed = 3002
    case keyNotFound = 3003
    case invalidSignature = 3004      // Sender signature verification failed
    case commitmentMismatch = 3005    // Commitment hash doesn't match content

    // Messages (4xxx)
    case messageSendFailed = 4001
    case messageNotFound = 4002
    case rateLimitExceeded = 4003     // Relay rate limit hit
    case messageQueued = 4004         // Not an error; message queued for offline send

    // Identity (5xxx)
    case didCreationFailed = 5001
    case verificationFailed = 5002

    // Groups (6xxx)
    case groupKeyMissing = 6001
    case groupKeyRotationFailed = 6002
    case notGroupAdmin = 6003

    // Digital Evidence (7xxx)
    case evidenceFingerprintFailed = 7001
    case evidenceNotAvailable = 7002   // User tier does not support Digital Evidence
    case evidenceVerificationFailed = 7003

    var errorDescription: String? {
        "Error \(rawValue). Please try again or contact support."
    }

    var supportCode: String {
        "ECHO-\(rawValue)"
    }

    /// Whether this error should be shown to the user
    var isUserFacing: Bool {
        switch self {
        case .messageQueued:
            return false // Informational, not an error
        default:
            return true
        }
    }

    /// Whether an automatic retry is appropriate
    var isRetryable: Bool {
        switch self {
        case .networkUnavailable, .requestTimeout, .serverError,
             .relayUnavailable, .messageSendFailed:
            return true
        default:
            return false
        }
    }
}
