import Foundation

// MARK: - Auth Error

enum AuthError: LocalizedError {
    case passkeyCreationFailed
    case passkeyAssertionFailed
    case noRefreshToken
    case tokenStorageFailed
    case sessionRevoked
    case biometricChanged
    case deviceIntegrityFailed
    case networkUnavailable
    case invalidPhoneNumber
    case otpExpired
    case accountLocked(retryAfter: Date?)

    var errorDescription: String? {
        switch self {
        case .passkeyCreationFailed: return "Could not create passkey."
        case .passkeyAssertionFailed: return "Sign in failed. Please try again."
        case .noRefreshToken: return "Please sign in again."
        case .tokenStorageFailed: return "Could not save credentials."
        case .sessionRevoked: return "Your session was ended. Please sign in."
        case .biometricChanged: return "Biometric data changed. Please verify."
        case .deviceIntegrityFailed: return "Device verification failed."
        case .networkUnavailable: return "No internet connection."
        case .invalidPhoneNumber: return "Please enter a valid phone number."
        case .otpExpired: return "Verification code expired. Please request a new one."
        case .accountLocked: return "Account temporarily locked. Please try again later."
        }
    }

    var isRetryable: Bool {
        switch self {
        case .networkUnavailable, .passkeyAssertionFailed:
            return true
        default:
            return false
        }
    }
}

// MARK: - Auth API Error

struct AuthAPIError: Error {
    let code: String
    let userMessage: String
    let httpStatus: Int

    static let networkError = AuthAPIError(
        code: "NETWORK",
        userMessage: "No internet connection.",
        httpStatus: 0
    )
}

// MARK: - Auth Error Response

struct AuthErrorResponse: Decodable {
    let error: ErrorDetail

    struct ErrorDetail: Decodable {
        let code: String
        let message: String
    }
}
