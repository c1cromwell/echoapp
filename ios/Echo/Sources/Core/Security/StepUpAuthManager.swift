#if os(iOS)
// Core/Security/StepUpAuthManager.swift
// Centralises step-up authentication used by:
//   - PersonaGateView  (hidden personas, folders, chats)
//   - StepUpSheetView  (sensitive account actions: revoke device, delete account, …)
//   - GlacialLoginScreen passcode fallback
//
// Three methods — all local, no network round-trip:
//   .faceID         — SE key sign (triggers Face ID prompt)
//   .passkey        — same SE key sign (same hardware, different label for UX clarity)
//   .devicePasscode — LAContext .deviceOwnerAuthentication (device PIN / password)

import Foundation
import LocalAuthentication

// MARK: - Method enum

public enum StepUpMethod: String, CaseIterable, Codable {
    case faceID          = "face_id"
    case passkey         = "passkey"
    case devicePasscode  = "device_passcode"

    public var displayName: String {
        switch self {
        case .faceID:         return "Face ID"
        case .passkey:        return "Passkey"
        case .devicePasscode: return "Device Passcode"
        }
    }

    public var systemIcon: String {
        switch self {
        case .faceID:         return "faceid"
        case .passkey:        return "key.horizontal.fill"
        case .devicePasscode: return "lock.fill"
        }
    }

    /// Whether this method is available on the current device.
    public var isAvailable: Bool {
        switch self {
        case .faceID, .passkey:
            let ctx = LAContext()
            return ctx.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil)
        case .devicePasscode:
            let ctx = LAContext()
            return ctx.canEvaluatePolicy(.deviceOwnerAuthentication, error: nil)
        }
    }
}

// MARK: - Manager

@MainActor
public final class StepUpAuthManager {
    public static let shared = StepUpAuthManager()
    private static let preferenceKey = "echo.stepup.preferred_method"

    private init() {}

    /// The user's chosen step-up method (defaults to Face ID).
    public var preferredMethod: StepUpMethod {
        get {
            guard let raw = UserDefaults.standard.string(forKey: Self.preferenceKey),
                  let method = StepUpMethod(rawValue: raw) else { return .faceID }
            return method
        }
        set {
            UserDefaults.standard.set(newValue.rawValue, forKey: Self.preferenceKey)
        }
    }

    /// Authenticate using the preferred method, or an explicit override.
    /// Throws on failure or user cancellation.
    ///
    /// - Parameters:
    ///   - reason: Localised string shown in the system auth prompt.
    ///   - override: Pass a specific method to bypass the stored preference.
    public func authenticate(reason: String, override method: StepUpMethod? = nil) async throws {
        switch method ?? preferredMethod {
        case .faceID, .passkey:
            try await authenticateWithSecureEnclave()
        case .devicePasscode:
            try await authenticateWithDevicePasscode(reason)
        }
    }

    // MARK: - Implementations

    private func authenticateWithSecureEnclave() async throws {
        let nonce = Data("echo-stepup-\(UUID().uuidString)".utf8)
        _ = try await SecureEnclaveManager.shared.sign(
            data: nonce,
            keyId: "echo-identity-signing"
        )
    }

    private func authenticateWithDevicePasscode(_ reason: String) async throws {
        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, Error>) in
            let ctx = LAContext()
            ctx.evaluatePolicy(.deviceOwnerAuthentication, localizedReason: reason) { ok, err in
                if ok {
                    cont.resume()
                } else {
                    cont.resume(throwing: err ?? LAError(.authenticationFailed))
                }
            }
        }
    }

    // MARK: - Passcode-only helper (login fallback, exported for GlacialLoginScreen)

    public func authenticateWithPasscodeOnly(reason: String) async throws {
        try await authenticateWithDevicePasscode(reason)
    }
}
#endif
