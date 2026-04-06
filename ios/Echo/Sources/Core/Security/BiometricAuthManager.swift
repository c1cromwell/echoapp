import Foundation
import LocalAuthentication

/// Manages biometric authentication (Face ID / Touch ID)
class BiometricAuthManager: NSObject {
    
    // MARK: - Properties
    
    private let context = LAContext()
    private let policy = LAPolicy.deviceOwnerAuthenticationWithBiometrics
    
    // MARK: - Biometric State
    
    var isBiometricAvailable: Bool {
        var error: NSError?
        return context.canEvaluatePolicy(policy, error: &error)
    }
    
    var biometricType: BiometricType {
        guard isBiometricAvailable else { return .none }
        
        if #available(iOS 16.0, *) {
            switch context.biometryType {
            case .faceID:
                return .faceID
            case .touchID:
                return .touchID
            case .opticID:
                return .opticID
            @unknown default:
                return .unknown
            }
        } else {
            // Fallback for iOS 15
            if context.canEvaluatePolicy(policy, error: nil) {
                return .touchID
            }
            return .none
        }
    }
    
    // MARK: - Authentication
    
    /// Authenticate user with biometrics
    /// - Parameters:
    ///   - reason: The reason to display to the user
    ///   - fallbackTitle: Optional title for fallback button
    /// - Returns: Boolean indicating success or failure
    func authenticate(reason: String, fallbackTitle: String? = nil) async -> Bool {
        guard isBiometricAvailable else {
            return false
        }
        
        // Create a new context for this authentication attempt
        let authContext = LAContext()
        
        if let fallbackTitle = fallbackTitle {
            authContext.localizedFallbackTitle = fallbackTitle
        }
        
        do {
            let success = try await authContext.evaluatePolicy(
                policy,
                localizedReason: reason
            )
            return success
        } catch {
            // Handle cancellation and other errors silently
            return false
        }
    }
    
    /// Authenticate with optional fallback to passcode
    /// - Parameters:
    ///   - reason: The reason to display to the user
    ///   - allowFallback: Whether to allow fallback to device passcode
    /// - Returns: Tuple of (success, usedFallback)
    func authenticateWithFallback(
        reason: String,
        allowFallback: Bool = true
    ) async -> (success: Bool, usedFallback: Bool) {
        guard isBiometricAvailable else {
            return (false, false)
        }
        
        let authContext = LAContext()
        
        // Configure fallback behavior
        if allowFallback {
            authContext.localizedFallbackTitle = "Use Passcode"
        } else {
            authContext.localizedFallbackTitle = ""
        }
        
        do {
            let success = try await authContext.evaluatePolicy(
                policy,
                localizedReason: reason
            )
            return (success, authContext.biometryType == .none)
        } catch {
            return (false, false)
        }
    }
    
    /// Check if device has passcode set
    var hasDevicePasscode: Bool {
        let context = LAContext()
        return context.canEvaluatePolicy(.deviceOwnerAuthentication, error: nil)
    }
    
    // MARK: - Invalidation
    
    /// Invalidate the current authentication context
    /// Call this after sensitive operations to require re-authentication
    func invalidate() {
        context.invalidate()
    }
}

// MARK: - Biometric Type

enum BiometricType: Equatable {
    case faceID
    case touchID
    case opticID
    case unknown
    case none
    
    var displayName: String {
        switch self {
        case .faceID:
            return "Face ID"
        case .touchID:
            return "Touch ID"
        case .opticID:
            return "Optic ID"
        case .unknown:
            return "Biometric Authentication"
        case .none:
            return "Not Available"
        }
    }
}

// MARK: - Authentication Errors

enum BiometricAuthError: LocalizedError {
    case notAvailable
    case userCancelled
    case userFallback
    case invalidContext
    case locked
    case unknown(String)
    
    var errorDescription: String? {
        switch self {
        case .notAvailable:
            return "Biometric authentication is not available on this device"
        case .userCancelled:
            return "Authentication was cancelled by the user"
        case .userFallback:
            return "User selected fallback option"
        case .invalidContext:
            return "Authentication context is invalid"
        case .locked:
            return "Biometric authentication is locked due to too many failed attempts"
        case .unknown(let message):
            return "Authentication failed: \(message)"
        }
    }
}
