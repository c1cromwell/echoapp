import Foundation
import LocalAuthentication

// MARK: - Biometric Status

enum BiometricStatus {
    case valid
    case enrollmentChanged   // New fingerprint/face added — require re-auth
    case unavailable         // Biometrics disabled or not supported
}

// MARK: - Biometric Integrity Service

final class BiometricIntegrityService {
    private let keychain: DataKeychainProtocol
    private static let stateKey = "echo.biometric.domain_state"

    init(keychain: DataKeychainProtocol) {
        self.keychain = keychain
    }

    /// Check if biometric enrollment has changed since last save
    func checkIntegrity() -> BiometricStatus {
        let context = LAContext()
        guard context.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics, error: nil
        ) else {
            return .unavailable
        }

        guard let currentState = context.evaluatedPolicyDomainState else {
            return .unavailable
        }

        guard let storedState = try? keychain.load(for: Self.stateKey) else {
            // First time — save current state and consider valid
            try? keychain.save(currentState, for: Self.stateKey)
            return .valid
        }

        if storedState != currentState {
            return .enrollmentChanged
        }
        return .valid
    }

    /// Save current biometric state (call after successful auth)
    func saveBiometricState() {
        let context = LAContext()
        guard context.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics, error: nil
        ), let state = context.evaluatedPolicyDomainState else { return }
        try? keychain.save(state, for: Self.stateKey)
    }

    /// Clear saved state (call on logout)
    func clearBiometricState() {
        try? keychain.delete(for: Self.stateKey)
    }
}
