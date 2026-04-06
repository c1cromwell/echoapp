# Echo Authentication — iOS Swift Implementation Spec

> **Version:** 2.0 · **Date:** March 2026 · **Status:** Ready for Development
> **iOS Target:** 16.0+ · **Swift:** 5.9+ · **Architecture:** MVVM + Clean

---

## Table of Contents

1. [Overview](#1-overview)
2. [Module Structure](#2-module-structure)
3. [Auth State Machine](#3-auth-state-machine)
4. [Services](#4-services)
5. [ViewModels](#5-viewmodels)
6. [Views](#6-views)
7. [Navigation Updates](#7-navigation-updates)
8. [API Client (Auth-Specific)](#8-api-client-auth-specific)
9. [Security Layer](#9-security-layer)
10. [Error Handling](#10-error-handling)
11. [Accessibility](#11-accessibility)
12. [Testing](#12-testing)

---

## 1. Overview

The iOS Swift authentication module handles the complete client-side auth experience: phone entry, OTP verification, WebAuthn passkey creation/assertion via Secure Enclave, secure token storage, device fingerprinting, and biometric change detection. It integrates with the existing MVVM + Clean Architecture, theme system, and navigation router.

### Key Frameworks

| Framework | Purpose |
|---|---|
| `AuthenticationServices` | WebAuthn / passkey creation and assertion (iOS 16+) |
| `LocalAuthentication` | Biometric availability check, domain state tracking |
| `Security` | Keychain for refresh token storage |
| `CryptoKit` | SHA-256 hashing, key generation |

### Dependencies (Add to SPM)

```swift
// No new external dependencies required.
// All auth functionality uses Apple's built-in frameworks.
// Existing dependencies (Alamofire, KeychainSwift) are sufficient.
```

---

## 2. Module Structure

All new files belong under `Features/Auth/`. This extends the existing directory structure from the iOS Swift Spec.

```
Features/Auth/
├── Views/
│   ├── WelcomeView.swift                  # UPDATE existing
│   ├── PhoneEntryView.swift               # NEW
│   ├── OTPVerificationView.swift          # NEW
│   ├── PasskeySetupView.swift             # NEW
│   ├── ProfileSetupView.swift             # NEW (already spec'd in onboarding)
│   ├── TrustIntroView.swift               # NEW
│   ├── LoginView.swift                    # NEW - returning user passkey login
│   ├── DeviceManagementView.swift         # NEW
│   ├── DeviceRowView.swift                # NEW - device list item component
│   ├── StepUpSheetView.swift              # NEW - bottom sheet modal
│   ├── RecoveryView.swift                 # NEW
│   └── AccountLockedView.swift            # NEW
│
├── ViewModels/
│   ├── AuthCoordinator.swift              # NEW - top-level auth state coordinator
│   ├── PhoneEntryViewModel.swift          # NEW
│   ├── OTPViewModel.swift                 # NEW
│   ├── LoginViewModel.swift               # NEW
│   ├── DeviceManagementViewModel.swift    # NEW
│   └── RecoveryViewModel.swift            # NEW
│
├── Models/
│   ├── AuthState.swift                    # NEW - state enum + events
│   ├── AuthToken.swift                    # NEW - token model
│   ├── AuthError.swift                    # NEW - typed error enum
│   ├── DeviceSession.swift                # NEW - device list model
│   └── AuditLogEntry.swift               # NEW - login history model
│
└── Services/
    ├── PasskeyManager.swift               # NEW - WebAuthn wrapper
    ├── TokenManager.swift                 # NEW - secure token lifecycle
    ├── DeviceFingerprintService.swift     # NEW - device info collection
    ├── BiometricIntegrityService.swift    # NEW - enrollment change detection
    ├── CertificatePinner.swift            # NEW - TLS pinning
    └── AuthAPIClient.swift                # NEW - auth-specific API layer
```

---

## 3. Auth State Machine

### 3.1 State Enum

```swift
// Features/Auth/Models/AuthState.swift

import Foundation

enum AuthState: Equatable {
    case unauthenticated
    case phoneEntry
    case otpVerification(phone: String, verificationId: String)
    case passkeySetup(tempToken: String)
    case profileSetup
    case trustIntro
    case authenticated(user: UserProfile)
    case locked(reason: LockReason, retryAfter: Date?)
    case recovery(method: RecoveryMethod)
}

enum LockReason: Equatable {
    case tooManyAttempts
    case suspiciousActivity
    case accountSuspended
}

enum RecoveryMethod: String, CaseIterable {
    case recoveryPhrase = "recovery_phrase"
    case trustedContacts = "trusted_contacts"
    case phoneReverification = "phone"
}

enum AuthEvent {
    case phoneSubmitted(phone: String, verificationId: String)
    case otpVerified(tempToken: String)
    case passkeyCreated
    case profileCompleted
    case trustIntroDismissed
    case loginSucceeded(UserProfile)
    case sessionExpired
    case loggedOut
    case accountLocked(LockReason, retryAfter: Date?)
    case recoveryInitiated(RecoveryMethod)
    case recoveryCompleted
}

enum StepUpAction: String, Hashable {
    case revokeDevice = "revoke_device"
    case changePhone = "change_phone"
    case sendLargePayment = "send_large_payment"
    case exportRecoveryPhrase = "export_recovery_phrase"
    case deleteAccount = "delete_account"
}
```

### 3.2 Auth Coordinator

```swift
// Features/Auth/ViewModels/AuthCoordinator.swift

import SwiftUI
import Combine

@MainActor
final class AuthCoordinator: ObservableObject {
    @Published private(set) var state: AuthState = .unauthenticated
    @Published var showStepUpSheet = false
    @Published var stepUpAction: StepUpAction?
    
    private let tokenManager: TokenManager
    private let passkeyManager: PasskeyManagerProtocol
    private let biometricService: BiometricIntegrityService
    private let apiClient: AuthAPIClientProtocol
    
    init(
        tokenManager: TokenManager,
        passkeyManager: PasskeyManagerProtocol,
        biometricService: BiometricIntegrityService,
        apiClient: AuthAPIClientProtocol
    ) {
        self.tokenManager = tokenManager
        self.passkeyManager = passkeyManager
        self.biometricService = biometricService
        self.apiClient = apiClient
    }
    
    // MARK: - State Transitions
    
    func handle(_ event: AuthEvent) {
        switch (state, event) {
        case (.unauthenticated, .phoneSubmitted(let phone, let vId)):
            state = .otpVerification(phone: phone, verificationId: vId)
            
        case (.otpVerification, .otpVerified(let tempToken)):
            state = .passkeySetup(tempToken: tempToken)
            
        case (.passkeySetup, .passkeyCreated):
            state = .profileSetup
            
        case (.profileSetup, .profileCompleted):
            state = .trustIntro
            
        case (.trustIntro, .trustIntroDismissed):
            // Fetch user profile and transition to authenticated
            Task { await completeOnboarding() }
            
        case (_, .loginSucceeded(let user)):
            state = .authenticated(user: user)
            
        case (_, .sessionExpired), (_, .loggedOut):
            tokenManager.clearTokens()
            state = .unauthenticated
            
        case (_, .accountLocked(let reason, let retryAfter)):
            state = .locked(reason: reason, retryAfter: retryAfter)
            
        case (_, .recoveryInitiated(let method)):
            state = .recovery(method: method)
            
        case (.recovery, .recoveryCompleted):
            state = .passkeySetup(tempToken: "")
            
        default:
            break // Invalid transition — ignore
        }
    }
    
    // MARK: - Session Restore
    
    func checkExistingSession() async {
        // Check biometric integrity first
        let biometricStatus = biometricService.checkIntegrity()
        if biometricStatus == .enrollmentChanged {
            // Biometric enrollment changed — require re-auth
            tokenManager.clearTokens()
            state = .unauthenticated
            return
        }
        
        // Try to restore session with stored refresh token
        do {
            let token = try await tokenManager.getValidAccessToken()
            let user = try await apiClient.fetchCurrentUser(token: token)
            state = .authenticated(user: user)
        } catch {
            state = .unauthenticated
        }
    }
    
    // MARK: - Step-Up
    
    func requestStepUp(for action: StepUpAction) {
        stepUpAction = action
        showStepUpSheet = true
    }
    
    private func completeOnboarding() async {
        do {
            let token = try await tokenManager.getValidAccessToken()
            let user = try await apiClient.fetchCurrentUser(token: token)
            state = .authenticated(user: user)
        } catch {
            state = .unauthenticated
        }
    }
}
```

---

## 4. Services

### 4.1 PasskeyManager

```swift
// Features/Auth/Services/PasskeyManager.swift

import AuthenticationServices
import CryptoKit

protocol PasskeyManagerProtocol {
    func createPasskey(
        challenge: Data,
        userId: Data,
        userName: String
    ) async throws -> PasskeyRegistrationResult
    
    func authenticateWithPasskey(
        challenge: Data
    ) async throws -> PasskeyAssertionResult
}

struct PasskeyRegistrationResult {
    let credentialId: Data
    let rawId: Data
    let clientDataJSON: Data
    let attestationObject: Data
}

struct PasskeyAssertionResult {
    let credentialId: Data
    let rawId: Data
    let clientDataJSON: Data
    let authenticatorData: Data
    let signature: Data
    let userHandle: Data
}

final class PasskeyManager: NSObject, PasskeyManagerProtocol,
    ASAuthorizationControllerDelegate,
    ASAuthorizationControllerPresentationContextProviding
{
    private let relyingPartyIdentifier = "echo.app"
    private var authContinuation: CheckedContinuation<ASAuthorization, Error>?
    
    // MARK: - Registration
    
    func createPasskey(
        challenge: Data,
        userId: Data,
        userName: String
    ) async throws -> PasskeyRegistrationResult {
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: relyingPartyIdentifier
        )
        
        let request = provider.createCredentialRegistrationRequest(
            challenge: challenge,
            name: userName,
            userID: userId
        )
        // Request direct attestation for server-side validation
        request.attestationPreference = .direct
        
        let authorization = try await performAuthorizationRequest(request)
        
        guard let credential = authorization.credential
            as? ASAuthorizationPlatformPublicKeyCredentialRegistration
        else {
            throw AuthError.passkeyCreationFailed
        }
        
        return PasskeyRegistrationResult(
            credentialId: credential.credentialID,
            rawId: credential.credentialID,
            clientDataJSON: credential.rawClientDataJSON,
            attestationObject: credential.rawAttestationObject ?? Data()
        )
    }
    
    // MARK: - Assertion (Login)
    
    func authenticateWithPasskey(
        challenge: Data
    ) async throws -> PasskeyAssertionResult {
        let provider = ASAuthorizationPlatformPublicKeyCredentialProvider(
            relyingPartyIdentifier: relyingPartyIdentifier
        )
        
        let request = provider.createCredentialAssertionRequest(
            challenge: challenge
        )
        // Allow the system to auto-discover available passkeys
        
        let authorization = try await performAuthorizationRequest(request)
        
        guard let credential = authorization.credential
            as? ASAuthorizationPlatformPublicKeyCredentialAssertion
        else {
            throw AuthError.passkeyAssertionFailed
        }
        
        return PasskeyAssertionResult(
            credentialId: credential.credentialID,
            rawId: credential.credentialID,
            clientDataJSON: credential.rawClientDataJSON,
            authenticatorData: credential.rawAuthenticatorData,
            signature: credential.signature,
            userHandle: credential.userID
        )
    }
    
    // MARK: - Authorization Execution
    
    private func performAuthorizationRequest(
        _ request: ASAuthorizationRequest
    ) async throws -> ASAuthorization {
        try await withCheckedThrowingContinuation { continuation in
            self.authContinuation = continuation
            let controller = ASAuthorizationController(authorizationRequests: [request])
            controller.delegate = self
            controller.presentationContextProvider = self
            controller.performRequests()
        }
    }
    
    // MARK: - ASAuthorizationControllerDelegate
    
    func authorizationController(
        controller: ASAuthorizationController,
        didCompleteWithAuthorization authorization: ASAuthorization
    ) {
        authContinuation?.resume(returning: authorization)
        authContinuation = nil
    }
    
    func authorizationController(
        controller: ASAuthorizationController,
        didCompleteWithError error: Error
    ) {
        authContinuation?.resume(throwing: error)
        authContinuation = nil
    }
    
    // MARK: - Presentation Context
    
    func presentationAnchor(
        for controller: ASAuthorizationController
    ) -> ASPresentationAnchor {
        guard let scene = UIApplication.shared.connectedScenes.first as? UIWindowScene,
              let window = scene.windows.first
        else {
            fatalError("No window available for passkey presentation")
        }
        return window
    }
}
```

### 4.2 TokenManager

```swift
// Features/Auth/Services/TokenManager.swift

import Foundation
import Combine

final class TokenManager: ObservableObject {
    @Published private(set) var isAuthenticated = false
    
    // Access token lives in memory only — never persisted
    private var accessToken: String?
    private var accessTokenExpiry: Date?
    
    private let keychain: KeychainManagerProtocol
    private let apiClient: AuthAPIClientProtocol
    private var refreshTask: Task<String, Error>?
    
    private static let refreshTokenKey = "echo.auth.refresh_token"
    private static let biometricStateKey = "echo.auth.biometric_state"
    
    init(keychain: KeychainManagerProtocol, apiClient: AuthAPIClientProtocol) {
        self.keychain = keychain
        self.apiClient = apiClient
        self.isAuthenticated = hasStoredRefreshToken()
    }
    
    // MARK: - Store Tokens from Auth Response
    
    func storeTokens(_ response: AuthResponse) throws {
        // Access token — memory only
        self.accessToken = response.accessToken
        self.accessTokenExpiry = response.expiresAt
        
        // Refresh token — Keychain (device-only, not backed up)
        if let refreshToken = response.refreshToken {
            guard let data = refreshToken.data(using: .utf8) else {
                throw AuthError.tokenStorageFailed
            }
            try keychain.save(data, for: Self.refreshTokenKey)
        }
        
        isAuthenticated = true
    }
    
    // MARK: - Get Valid Access Token (auto-refresh)
    
    func getValidAccessToken() async throws -> String {
        // Return current token if still valid (with 30s buffer)
        if let token = accessToken, let expiry = accessTokenExpiry,
           expiry > Date().addingTimeInterval(30) {
            return token
        }
        
        // Coalesce concurrent refresh requests into one network call
        if let existingTask = refreshTask {
            return try await existingTask.value
        }
        
        let task = Task<String, Error> {
            defer { refreshTask = nil }
            return try await performRefresh()
        }
        refreshTask = task
        return try await task.value
    }
    
    // MARK: - Refresh
    
    private func performRefresh() async throws -> String {
        guard let refreshData = try keychain.load(for: Self.refreshTokenKey),
              let refreshToken = String(data: refreshData, encoding: .utf8)
        else {
            throw AuthError.noRefreshToken
        }
        
        do {
            let response = try await apiClient.refreshToken(refreshToken)
            try storeTokens(response)
            return response.accessToken
        } catch let error as APIError where error.code == "AUTH_006" {
            // Refresh token was reused (replay detected) — server revoked all sessions
            clearTokens()
            throw AuthError.sessionRevoked
        }
    }
    
    // MARK: - Clear
    
    func clearTokens() {
        accessToken = nil
        accessTokenExpiry = nil
        try? keychain.delete(for: Self.refreshTokenKey)
        isAuthenticated = false
    }
    
    // MARK: - Helpers
    
    private func hasStoredRefreshToken() -> Bool {
        (try? keychain.load(for: Self.refreshTokenKey)) != nil
    }
}
```

### 4.3 DeviceFingerprintService

```swift
// Features/Auth/Services/DeviceFingerprintService.swift

import UIKit
import LocalAuthentication
import CryptoKit

struct DeviceInfo: Codable {
    let deviceId: String
    let platform: String
    let osVersion: String
    let appVersion: String
    let model: String
    let locale: String
    let timezone: String
    let secureEnclave: Bool
    let biometricType: String
    let jailbreakStatus: Bool
    
    enum CodingKeys: String, CodingKey {
        case deviceId = "device_id"
        case platform
        case osVersion = "os_version"
        case appVersion = "app_version"
        case model
        case locale
        case timezone
        case secureEnclave = "secure_enclave"
        case biometricType = "biometric_type"
        case jailbreakStatus = "jailbreak_status"
    }
}

final class DeviceFingerprintService {
    
    func collectDeviceInfo() -> DeviceInfo {
        let device = UIDevice.current
        let context = LAContext()
        _ = context.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics, error: nil
        )
        
        return DeviceInfo(
            deviceId: computeDeviceId(),
            platform: "ios",
            osVersion: device.systemVersion,
            appVersion: Bundle.main.infoDictionary?["CFBundleShortVersionString"]
                as? String ?? "unknown",
            model: deviceModelIdentifier(),
            locale: Locale.current.identifier,
            timezone: TimeZone.current.identifier,
            secureEnclave: hasSecureEnclave(),
            biometricType: biometricTypeString(context.biometryType),
            jailbreakStatus: isJailbroken()
        )
    }
    
    /// JSON-encoded device info for the X-Device-Info header
    func deviceInfoHeader() -> String {
        let info = collectDeviceInfo()
        guard let data = try? JSONEncoder().encode(info),
              let string = String(data: data, encoding: .utf8)
        else { return "{}" }
        return string
    }
    
    // MARK: - Private
    
    private func computeDeviceId() -> String {
        // Use identifierForVendor + model as stable device identifier
        let vendorId = UIDevice.current.identifierForVendor?.uuidString ?? "unknown"
        let model = deviceModelIdentifier()
        let data = "\(vendorId):\(model)".data(using: .utf8)!
        let hash = SHA256.hash(data: data)
        return hash.compactMap { String(format: "%02x", $0) }.joined()
    }
    
    private func deviceModelIdentifier() -> String {
        var systemInfo = utsname()
        uname(&systemInfo)
        return withUnsafePointer(to: &systemInfo.machine) {
            $0.withMemoryRebound(to: CChar.self, capacity: 1) {
                String(cString: $0)
            }
        }
    }
    
    private func hasSecureEnclave() -> Bool {
        // All devices with Face ID or Touch ID have Secure Enclave
        let context = LAContext()
        return context.canEvaluatePolicy(
            .deviceOwnerAuthenticationWithBiometrics, error: nil
        )
    }
    
    private func biometricTypeString(_ type: LABiometryType) -> String {
        switch type {
        case .faceID: return "face_id"
        case .touchID: return "touch_id"
        case .opticID: return "optic_id"
        default: return "none"
        }
    }
    
    private func isJailbroken() -> Bool {
        #if targetEnvironment(simulator)
        return false
        #else
        // Check for common jailbreak indicators
        let suspiciousPaths = [
            "/Applications/Cydia.app",
            "/Library/MobileSubstrate/MobileSubstrate.dylib",
            "/bin/bash",
            "/usr/sbin/sshd",
            "/etc/apt",
            "/private/var/lib/apt/"
        ]
        for path in suspiciousPaths {
            if FileManager.default.fileExists(atPath: path) {
                return true
            }
        }
        // Check if app can write outside sandbox
        let testPath = "/private/jailbreak_test_\(UUID().uuidString)"
        do {
            try "test".write(toFile: testPath, atomically: true, encoding: .utf8)
            try FileManager.default.removeItem(atPath: testPath)
            return true
        } catch {
            return false
        }
        #endif
    }
}
```

### 4.4 BiometricIntegrityService

```swift
// Features/Auth/Services/BiometricIntegrityService.swift

import LocalAuthentication

enum BiometricStatus {
    case valid
    case enrollmentChanged   // New fingerprint/face added — require re-auth
    case unavailable         // Biometrics disabled or not supported
}

final class BiometricIntegrityService {
    private let keychain: KeychainManagerProtocol
    private static let stateKey = "echo.biometric.domain_state"
    
    init(keychain: KeychainManagerProtocol) {
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
```

### 4.5 CertificatePinner

```swift
// Features/Auth/Services/CertificatePinner.swift

import Foundation
import CryptoKit

final class CertificatePinner: NSObject, URLSessionDelegate {
    
    /// SHA-256 hashes of the Subject Public Key Info (SPKI) of trusted certificates.
    /// Include both current and backup certificate hashes.
    private let pinnedKeyHashes: Set<String> = [
        // Production certificate (current)
        "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
        // Backup certificate (rotate in advance)
        "sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="
    ]
    
    func urlSession(
        _ session: URLSession,
        didReceive challenge: URLAuthenticationChallenge
    ) async -> (URLSession.AuthChallengeDisposition, URLCredential?) {
        
        guard challenge.protectionSpace.authenticationMethod
                == NSURLAuthenticationMethodServerTrust,
              challenge.protectionSpace.host == "api.echo.app",
              let serverTrust = challenge.protectionSpace.serverTrust
        else {
            return (.cancelAuthenticationChallenge, nil)
        }
        
        // Evaluate the trust chain
        var error: CFError?
        guard SecTrustEvaluateWithError(serverTrust, &error) else {
            return (.cancelAuthenticationChallenge, nil)
        }
        
        // Extract server certificate's public key
        guard let certChain = SecTrustCopyCertificateChain(serverTrust)
                as? [SecCertificate],
              let serverCert = certChain.first,
              let serverKey = SecCertificateCopyKey(serverCert)
        else {
            return (.cancelAuthenticationChallenge, nil)
        }
        
        // Hash the public key and compare
        let serverKeyHash = hashPublicKey(serverKey)
        if pinnedKeyHashes.contains(serverKeyHash) {
            return (.useCredential, URLCredential(trust: serverTrust))
        }
        
        // Pin mismatch — reject connection
        return (.cancelAuthenticationChallenge, nil)
    }
    
    private func hashPublicKey(_ key: SecKey) -> String {
        guard let keyData = SecKeyCopyExternalRepresentation(key, nil) as Data? else {
            return ""
        }
        let hash = SHA256.hash(data: keyData)
        return "sha256/" + Data(hash).base64EncodedString()
    }
}
```

---

## 5. ViewModels

### 5.1 PhoneEntryViewModel

```swift
// Features/Auth/ViewModels/PhoneEntryViewModel.swift

import SwiftUI

@MainActor
final class PhoneEntryViewModel: ObservableObject {
    @Published var countryCode = "+1"
    @Published var phoneNumber = ""
    @Published var isLoading = false
    @Published var errorMessage: String?
    
    var isValid: Bool {
        let digits = phoneNumber.filter(\.isNumber)
        return digits.count >= 10 && digits.count <= 15
    }
    
    var formattedDisplay: String {
        // Format as (555) 123-4567 for US numbers
        let digits = phoneNumber.filter(\.isNumber)
        guard digits.count == 10, countryCode == "+1" else { return phoneNumber }
        let area = digits.prefix(3)
        let mid = digits.dropFirst(3).prefix(3)
        let last = digits.suffix(4)
        return "(\(area)) \(mid)-\(last)"
    }
    
    private let apiClient: AuthAPIClientProtocol
    private let deviceService: DeviceFingerprintService
    
    init(apiClient: AuthAPIClientProtocol, deviceService: DeviceFingerprintService) {
        self.apiClient = apiClient
        self.deviceService = deviceService
    }
    
    func sendOTP() async -> (phone: String, verificationId: String)? {
        guard isValid else { return nil }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        
        do {
            let digits = phoneNumber.filter(\.isNumber)
            let response = try await apiClient.registerPhone(
                phone: digits,
                countryCode: countryCode,
                deviceInfo: deviceService.collectDeviceInfo()
            )
            return (phone: "\(countryCode)\(digits)", verificationId: response.verificationId)
        } catch let error as APIError {
            errorMessage = error.userMessage
            return nil
        } catch {
            errorMessage = "Something went wrong. Please try again."
            return nil
        }
    }
}
```

### 5.2 OTPViewModel

```swift
// Features/Auth/ViewModels/OTPViewModel.swift

import SwiftUI
import Combine

@MainActor
final class OTPViewModel: ObservableObject {
    @Published var code: [String] = Array(repeating: "", count: 6)
    @Published var focusedIndex = 0
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published var resendCountdown = 60
    @Published var canResend = false
    
    let phoneNumber: String
    let verificationId: String
    
    private let apiClient: AuthAPIClientProtocol
    private let deviceService: DeviceFingerprintService
    private var timer: AnyCancellable?
    
    var fullCode: String { code.joined() }
    var isComplete: Bool { fullCode.count == 6 && fullCode.allSatisfy(\.isNumber) }
    
    init(
        phoneNumber: String,
        verificationId: String,
        apiClient: AuthAPIClientProtocol,
        deviceService: DeviceFingerprintService
    ) {
        self.phoneNumber = phoneNumber
        self.verificationId = verificationId
        self.apiClient = apiClient
        self.deviceService = deviceService
        startResendTimer()
    }
    
    func verifyOTP() async -> String? {
        guard isComplete else { return nil }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        
        do {
            let response = try await apiClient.verifyOTP(
                verificationId: verificationId,
                code: fullCode,
                deviceInfo: deviceService.collectDeviceInfo()
            )
            return response.accessToken  // Temp token for passkey registration
        } catch let error as APIError {
            errorMessage = error.userMessage
            // Shake animation triggered by errorMessage change
            code = Array(repeating: "", count: 6)
            focusedIndex = 0
            return nil
        } catch {
            errorMessage = "Verification failed. Please try again."
            return nil
        }
    }
    
    func resendCode() async {
        guard canResend else { return }
        canResend = false
        resendCountdown = 60
        startResendTimer()
        // Re-send OTP via same phone number
        _ = try? await apiClient.registerPhone(
            phone: phoneNumber.filter(\.isNumber),
            countryCode: String(phoneNumber.prefix(2)),
            deviceInfo: deviceService.collectDeviceInfo()
        )
    }
    
    private func startResendTimer() {
        timer = Timer.publish(every: 1, on: .main, in: .common)
            .autoconnect()
            .sink { [weak self] _ in
                guard let self else { return }
                if self.resendCountdown > 0 {
                    self.resendCountdown -= 1
                } else {
                    self.canResend = true
                    self.timer?.cancel()
                }
            }
    }
}
```

### 5.3 LoginViewModel

```swift
// Features/Auth/ViewModels/LoginViewModel.swift

import SwiftUI

@MainActor
final class LoginViewModel: ObservableObject {
    @Published var isAuthenticating = false
    @Published var errorMessage: String?
    @Published var maskedPhone: String = ""
    @Published var username: String = ""
    
    private let passkeyManager: PasskeyManagerProtocol
    private let tokenManager: TokenManager
    private let apiClient: AuthAPIClientProtocol
    private let deviceService: DeviceFingerprintService
    
    init(
        passkeyManager: PasskeyManagerProtocol,
        tokenManager: TokenManager,
        apiClient: AuthAPIClientProtocol,
        deviceService: DeviceFingerprintService
    ) {
        self.passkeyManager = passkeyManager
        self.tokenManager = tokenManager
        self.apiClient = apiClient
        self.deviceService = deviceService
    }
    
    func loadStoredAccountInfo() {
        // Load masked phone and username from Keychain/UserDefaults
        // These are non-sensitive display values saved during registration
        maskedPhone = UserDefaults.standard.string(forKey: "echo.display.masked_phone") ?? ""
        username = UserDefaults.standard.string(forKey: "echo.display.username") ?? ""
    }
    
    func loginWithPasskey() async -> UserProfile? {
        isAuthenticating = true
        errorMessage = nil
        defer { isAuthenticating = false }
        
        do {
            // Step 1: Get challenge from server
            let challenge = try await apiClient.getLoginChallenge()
            
            // Step 2: Perform passkey assertion (triggers Face ID / Touch ID)
            let assertion = try await passkeyManager.authenticateWithPasskey(
                challenge: challenge.challengeData
            )
            
            // Step 3: Send assertion to server for verification
            let authResponse = try await apiClient.login(
                assertion: assertion,
                deviceInfo: deviceService.collectDeviceInfo()
            )
            
            // Step 4: Store tokens
            try tokenManager.storeTokens(authResponse)
            
            return authResponse.user
        } catch let error as APIError where error.code == "AUTH_007" {
            errorMessage = "New device detected. Additional verification needed."
            return nil
        } catch let error as APIError where error.code == "AUTH_009" {
            errorMessage = "Account temporarily locked. Please try again later."
            return nil
        } catch is CancellationError {
            // User cancelled Face ID — not an error
            return nil
        } catch {
            errorMessage = "Sign in failed. Please try again."
            return nil
        }
    }
}
```

### 5.4 DeviceManagementViewModel

```swift
// Features/Auth/ViewModels/DeviceManagementViewModel.swift

import SwiftUI

@MainActor
final class DeviceManagementViewModel: ObservableObject {
    @Published var currentDevice: DeviceSession?
    @Published var otherDevices: [DeviceSession] = []
    @Published var isLoading = true
    @Published var errorMessage: String?
    
    private let apiClient: AuthAPIClientProtocol
    private let tokenManager: TokenManager
    
    init(apiClient: AuthAPIClientProtocol, tokenManager: TokenManager) {
        self.apiClient = apiClient
        self.tokenManager = tokenManager
    }
    
    func loadDevices() async {
        isLoading = true
        defer { isLoading = false }
        
        do {
            let token = try await tokenManager.getValidAccessToken()
            let devices = try await apiClient.listDevices(token: token)
            currentDevice = devices.first(where: \.isCurrentDevice)
            otherDevices = devices.filter { !$0.isCurrentDevice }
        } catch {
            errorMessage = "Could not load devices."
        }
    }
    
    func revokeDevice(id: String, elevatedToken: String) async -> Bool {
        do {
            try await apiClient.revokeDevice(
                id: id, elevatedToken: elevatedToken
            )
            otherDevices.removeAll { $0.id == id }
            return true
        } catch {
            errorMessage = "Could not remove device."
            return false
        }
    }
    
    func logoutAllDevices() async {
        do {
            let token = try await tokenManager.getValidAccessToken()
            try await apiClient.logout(token: token, allDevices: true)
            tokenManager.clearTokens()
        } catch {
            errorMessage = "Could not log out devices."
        }
    }
}
```

---

## 6. Views

### 6.1 PhoneEntryView

```swift
// Features/Auth/Views/PhoneEntryView.swift

import SwiftUI

struct PhoneEntryView: View {
    @Environment(\.theme) var theme
    @StateObject var viewModel: PhoneEntryViewModel
    let onOTPSent: (String, String) -> Void  // (phone, verificationId)
    
    var body: some View {
        VStack(spacing: 0) {
            // Header
            VStack(alignment: .leading, spacing: 8) {
                Text("Enter your phone number")
                    .font(.system(size: 28, weight: .bold))
                    .foregroundColor(theme.colors.textPrimary)
                
                Text("We'll send a verification code to confirm it's you.")
                    .font(.system(size: 16))
                    .foregroundColor(theme.colors.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.top, 24)
            
            Spacer().frame(height: 32)
            
            // Phone Input
            HStack(spacing: 12) {
                // Country Code Picker
                Menu {
                    Button("+1 US") { viewModel.countryCode = "+1" }
                    Button("+44 UK") { viewModel.countryCode = "+44" }
                    Button("+91 IN") { viewModel.countryCode = "+91" }
                } label: {
                    HStack {
                        Text(viewModel.countryCode)
                            .font(.system(size: 16, weight: .medium))
                        Image(systemName: "chevron.down")
                            .font(.system(size: 12))
                    }
                    .foregroundColor(theme.colors.textPrimary)
                    .frame(width: 80, height: 56)
                    .background(theme.colors.surface)
                    .cornerRadius(12)
                }
                
                // Phone Number Field
                TextField("(555) 123-4567", text: $viewModel.phoneNumber)
                    .keyboardType(.phonePad)
                    .font(.system(size: 16))
                    .frame(height: 56)
                    .padding(.horizontal, 16)
                    .background(theme.colors.surface)
                    .cornerRadius(12)
            }
            
            // Error message
            if let error = viewModel.errorMessage {
                Text(error)
                    .font(.system(size: 14))
                    .foregroundColor(theme.colors.error)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.top, 8)
            }
            
            Spacer()
            
            // Send Code Button
            Button {
                Task {
                    if let result = await viewModel.sendOTP() {
                        onOTPSent(result.phone, result.verificationId)
                    }
                }
            } label: {
                Group {
                    if viewModel.isLoading {
                        ProgressView()
                            .tint(theme.colors.textOnPrimary)
                    } else {
                        Text("Send Code")
                    }
                }
                .font(.system(size: 16, weight: .semibold))
                .foregroundColor(theme.colors.textOnPrimary)
                .frame(maxWidth: .infinity)
                .frame(height: 56)
                .background(
                    viewModel.isValid
                        ? theme.colors.primaryGradient
                        : LinearGradient(
                            colors: [theme.colors.border, theme.colors.border],
                            startPoint: .leading, endPoint: .trailing
                          )
                )
                .cornerRadius(12)
            }
            .disabled(!viewModel.isValid || viewModel.isLoading)
            
            // Legal text
            Text("By continuing, you agree to our Terms and Privacy Policy")
                .font(.system(size: 12))
                .foregroundColor(theme.colors.textTertiary)
                .multilineTextAlignment(.center)
                .padding(.top, 16)
                .padding(.bottom, 8)
        }
        .padding(.horizontal, 24)
    }
}
```

### 6.2 OTPVerificationView

```swift
// Features/Auth/Views/OTPVerificationView.swift

import SwiftUI

struct OTPVerificationView: View {
    @Environment(\.theme) var theme
    @StateObject var viewModel: OTPViewModel
    let onVerified: (String) -> Void  // tempToken
    
    @FocusState private var focusedField: Int?
    @State private var shake = false
    
    var body: some View {
        VStack(spacing: 0) {
            VStack(alignment: .leading, spacing: 8) {
                Text("Enter verification code")
                    .font(.system(size: 28, weight: .bold))
                    .foregroundColor(theme.colors.textPrimary)
                
                Text("Sent to \(viewModel.phoneNumber)")
                    .font(.system(size: 16))
                    .foregroundColor(theme.colors.textSecondary)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.top, 24)
            
            Spacer().frame(height: 40)
            
            // OTP Input Boxes
            HStack(spacing: 8) {
                ForEach(0..<6, id: \.self) { index in
                    TextField("", text: $viewModel.code[index])
                        .keyboardType(.numberPad)
                        .multilineTextAlignment(.center)
                        .font(.system(size: 24, weight: .semibold))
                        .frame(width: 48, height: 56)
                        .background(
                            viewModel.code[index].isEmpty
                                ? theme.colors.surface
                                : theme.colors.primarySubtle
                        )
                        .cornerRadius(12)
                        .overlay(
                            RoundedRectangle(cornerRadius: 12)
                                .stroke(
                                    focusedField == index
                                        ? theme.colors.primary
                                        : theme.colors.border,
                                    lineWidth: focusedField == index ? 2 : 1
                                )
                        )
                        .focused($focusedField, equals: index)
                        .onChange(of: viewModel.code[index]) { newValue in
                            // Auto-advance on digit entry
                            if newValue.count == 1 && index < 5 {
                                focusedField = index + 1
                            }
                            // Auto-submit when complete
                            if viewModel.isComplete {
                                Task {
                                    if let token = await viewModel.verifyOTP() {
                                        onVerified(token)
                                    } else {
                                        shake = true
                                    }
                                }
                            }
                        }
                }
            }
            .modifier(ShakeEffect(shakes: shake ? 2 : 0))
            .animation(.default, value: shake)
            .onChange(of: shake) { _ in
                DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
                    shake = false
                }
            }
            
            // Error & Resend
            VStack(spacing: 16) {
                if let error = viewModel.errorMessage {
                    Text(error)
                        .font(.system(size: 14))
                        .foregroundColor(theme.colors.error)
                }
                
                Button {
                    Task { await viewModel.resendCode() }
                } label: {
                    Text(viewModel.canResend
                         ? "Resend code"
                         : "Resend code (0:\(String(format: "%02d", viewModel.resendCountdown)))")
                        .font(.system(size: 16))
                        .foregroundColor(
                            viewModel.canResend
                                ? theme.colors.primary
                                : theme.colors.textTertiary
                        )
                }
                .disabled(!viewModel.canResend)
            }
            .padding(.top, 24)
            
            Spacer()
        }
        .padding(.horizontal, 24)
        .onAppear { focusedField = 0 }
    }
}

// MARK: - Shake Effect

struct ShakeEffect: GeometryEffect {
    var amount: CGFloat = 6
    var shakes: Int
    var animatableData: CGFloat {
        get { CGFloat(shakes) }
        set { shakes = Int(newValue) }
    }
    func effectValue(size: CGSize) -> ProjectionTransform {
        ProjectionTransform(CGAffineTransform(
            translationX: amount * sin(animatableData * .pi * 2),
            y: 0
        ))
    }
}
```

### 6.3 LoginView (Returning User)

```swift
// Features/Auth/Views/LoginView.swift

import SwiftUI

struct LoginView: View {
    @Environment(\.theme) var theme
    @StateObject var viewModel: LoginViewModel
    let onLoggedIn: (UserProfile) -> Void
    let onSwitchAccount: () -> Void
    let onRecover: () -> Void
    
    var body: some View {
        VStack(spacing: 0) {
            Spacer()
            
            // Logo (smaller than Welcome)
            Image("echo-logo")
                .resizable()
                .frame(width: 64, height: 64)
                .cornerRadius(16)
                .padding(.bottom, 24)
            
            Text("Welcome back")
                .font(.system(size: 28, weight: .bold))
                .foregroundColor(theme.colors.textPrimary)
                .padding(.bottom, 32)
            
            // Sign In Button (biometric)
            Button {
                Task {
                    if let user = await viewModel.loginWithPasskey() {
                        onLoggedIn(user)
                    }
                }
            } label: {
                HStack(spacing: 12) {
                    Image(systemName: "faceid")
                        .font(.system(size: 20))
                    
                    if viewModel.isAuthenticating {
                        ProgressView()
                            .tint(theme.colors.textOnPrimary)
                    } else {
                        Text("Sign In")
                            .font(.system(size: 16, weight: .semibold))
                    }
                }
                .foregroundColor(theme.colors.textOnPrimary)
                .frame(maxWidth: .infinity)
                .frame(height: 56)
                .background(theme.colors.primaryGradient)
                .cornerRadius(12)
            }
            .disabled(viewModel.isAuthenticating)
            
            // Account info
            VStack(spacing: 4) {
                if !viewModel.username.isEmpty {
                    Text("Signed in as @\(viewModel.username)")
                        .font(.system(size: 14))
                        .foregroundColor(theme.colors.textSecondary)
                }
                if !viewModel.maskedPhone.isEmpty {
                    Text(viewModel.maskedPhone)
                        .font(.system(size: 14))
                        .foregroundColor(theme.colors.textTertiary)
                }
            }
            .padding(.top, 16)
            
            // Error
            if let error = viewModel.errorMessage {
                Text(error)
                    .font(.system(size: 14))
                    .foregroundColor(theme.colors.error)
                    .multilineTextAlignment(.center)
                    .padding(.top, 12)
            }
            
            Spacer()
            
            // Switch account
            Button("Not you? Switch account", action: onSwitchAccount)
                .font(.system(size: 14, weight: .medium))
                .foregroundColor(theme.colors.primary)
            
            Divider()
                .padding(.vertical, 16)
            
            // Recovery
            Button("Lost access? Recover", action: onRecover)
                .font(.system(size: 14))
                .foregroundColor(theme.colors.textSecondary)
                .padding(.bottom, 32)
        }
        .padding(.horizontal, 24)
        .onAppear {
            viewModel.loadStoredAccountInfo()
        }
    }
}
```

### 6.4 StepUpSheetView

```swift
// Features/Auth/Views/StepUpSheetView.swift

import SwiftUI

struct StepUpSheetView: View {
    @Environment(\.theme) var theme
    @Environment(\.dismiss) var dismiss
    
    let action: StepUpAction
    let onVerified: (String) -> Void  // elevated token
    
    @State private var isAuthenticating = false
    @State private var showOTPFallback = false
    
    var body: some View {
        VStack(spacing: 24) {
            // Drag indicator
            Capsule()
                .fill(theme.colors.border)
                .frame(width: 36, height: 5)
                .padding(.top, 8)
            
            VStack(spacing: 8) {
                Text("Verify it's you")
                    .font(.system(size: 24, weight: .bold))
                    .foregroundColor(theme.colors.textPrimary)
                
                Text("This action requires additional verification.")
                    .font(.system(size: 16))
                    .foregroundColor(theme.colors.textSecondary)
                    .multilineTextAlignment(.center)
            }
            
            // Biometric button
            Button {
                isAuthenticating = true
                // Trigger passkey assertion for step-up
            } label: {
                HStack(spacing: 12) {
                    Image(systemName: "faceid")
                        .font(.system(size: 20))
                    Text("Verify now")
                        .font(.system(size: 16, weight: .semibold))
                }
                .foregroundColor(theme.colors.textOnPrimary)
                .frame(maxWidth: .infinity)
                .frame(height: 56)
                .background(theme.colors.primaryGradient)
                .cornerRadius(12)
            }
            
            // OTP fallback
            Button("Use verification code") {
                showOTPFallback = true
            }
            .font(.system(size: 14, weight: .medium))
            .foregroundColor(theme.colors.primary)
            
            Spacer()
        }
        .padding(.horizontal, 24)
        .presentationDetents([.fraction(0.4)])
        .presentationDragIndicator(.hidden)
    }
}
```

---

## 7. Navigation Updates

### 7.1 Updated Route Enum

```swift
// App/Navigation/AppRouter.swift — updated

enum Route: Hashable {
    // Existing routes
    case welcome
    case phoneEntry
    case otpVerification(phoneNumber: String)
    case passkeySetup
    case personaCreation
    case home
    case chat(conversationId: String)
    case contactProfile(contactId: String)
    case settings
    case trustDashboard
    
    // NEW auth routes
    case login                                  // Returning user passkey login
    case trustIntro                             // Post-registration trust overview
    case deviceManagement                       // View/revoke device sessions
    case accountRecovery                        // Lost access flow
    case stepUpVerification(action: StepUpAction)   // Elevated auth
    case auditLog                               // Login history
}
```

### 7.2 Root View (Auth Gating)

```swift
// App/EchoApp.swift

@main
struct EchoApp: App {
    @StateObject private var authCoordinator = AuthCoordinator(...)
    @StateObject private var themeManager = ThemeManager()
    @StateObject private var router = AppRouter()
    
    var body: some Scene {
        WindowGroup {
            Group {
                switch authCoordinator.state {
                case .unauthenticated:
                    WelcomeView(
                        onPhoneRegistration: { authCoordinator.handle(.phoneSubmitted(...)) },
                        onLogin: { router.navigate(to: .login) }
                    )
                case .phoneEntry:
                    PhoneEntryView(...)
                case .otpVerification(let phone, let vId):
                    OTPVerificationView(...)
                case .passkeySetup:
                    PasskeySetupView(...)
                case .profileSetup:
                    ProfileSetupView(...)
                case .trustIntro:
                    TrustIntroView(...)
                case .authenticated:
                    MainTabView()
                        .environmentObject(router)
                case .locked(let reason, let retryAfter):
                    AccountLockedView(reason: reason, retryAfter: retryAfter)
                case .recovery(let method):
                    RecoveryView(method: method)
                }
            }
            .themed(themeManager)
            .environmentObject(authCoordinator)
            .task { await authCoordinator.checkExistingSession() }
        }
    }
}
```

---

## 8. API Client (Auth-Specific)

```swift
// Features/Auth/Services/AuthAPIClient.swift

import Foundation

protocol AuthAPIClientProtocol {
    func registerPhone(phone: String, countryCode: String, deviceInfo: DeviceInfo)
        async throws -> PhoneRegistrationResponse
    func verifyOTP(verificationId: String, code: String, deviceInfo: DeviceInfo)
        async throws -> AuthResponse
    func registerPasskey(attestation: PasskeyRegistrationResult, deviceInfo: DeviceInfo)
        async throws -> AuthResponse
    func getLoginChallenge() async throws -> ChallengeResponse
    func login(assertion: PasskeyAssertionResult, deviceInfo: DeviceInfo)
        async throws -> AuthResponse
    func refreshToken(_ token: String) async throws -> AuthResponse
    func logout(token: String, allDevices: Bool) async throws
    func fetchCurrentUser(token: String) async throws -> UserProfile
    func listDevices(token: String) async throws -> [DeviceSession]
    func revokeDevice(id: String, elevatedToken: String) async throws
    func getAuditLog(token: String) async throws -> [AuditLogEntry]
}

struct PhoneRegistrationResponse: Decodable {
    let verificationId: String
    let expiresAt: Date
    let retryAfter: Int
}

struct ChallengeResponse: Decodable {
    let challenge: String  // base64
    let timeout: Int
    let rpId: String
    
    var challengeData: Data {
        Data(base64Encoded: challenge) ?? Data()
    }
}

struct AuthResponse: Decodable {
    let accessToken: String
    let refreshToken: String?
    let expiresAt: Date
    let user: UserProfile
}

struct DeviceSession: Identifiable, Decodable {
    let id: String
    let friendlyName: String
    let platform: String
    let osVersion: String
    let lastIP: String
    let lastLocation: String
    let lastActiveAt: Date
    let isCurrentDevice: Bool
}

struct AuditLogEntry: Identifiable, Decodable {
    let id: String
    let eventType: String
    let result: String
    let ipAddress: String
    let deviceName: String?
    let createdAt: Date
}

// MARK: - Implementation

final class AuthAPIClient: AuthAPIClientProtocol {
    private let baseURL: URL
    private let session: URLSession
    private let deviceService: DeviceFingerprintService
    
    init(
        baseURL: URL = URL(string: "https://api.echo.app/v1")!,
        certificatePinner: CertificatePinner
    ) {
        self.baseURL = baseURL
        self.deviceService = DeviceFingerprintService()
        // Configure session with certificate pinning
        self.session = URLSession(
            configuration: .default,
            delegate: certificatePinner,
            delegateQueue: nil
        )
    }
    
    // All requests automatically include X-Device-Info header
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
            request.httpBody = try JSONEncoder().encode(body)
        }
        
        let (data, response) = try await session.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIError.networkError
        }
        
        if !(200...299).contains(httpResponse.statusCode) {
            let errorResponse = try? JSONDecoder().decode(
                ErrorResponse.self, from: data
            )
            throw APIError(
                code: errorResponse?.error.code ?? "UNKNOWN",
                userMessage: errorResponse?.error.message ?? "Something went wrong.",
                httpStatus: httpResponse.statusCode
            )
        }
        
        return data
    }
    
    // Example implementation for login
    func login(
        assertion: PasskeyAssertionResult,
        deviceInfo: DeviceInfo
    ) async throws -> AuthResponse {
        struct LoginBody: Encodable {
            let authType = "passkey"
            let credential: CredentialData
            
            struct CredentialData: Encodable {
                let id: String
                let rawId: String
                let clientDataJSON: String
                let authenticatorData: String
                let signature: String
            }
        }
        
        let body = LoginBody(credential: .init(
            id: assertion.credentialId.base64EncodedString(),
            rawId: assertion.rawId.base64EncodedString(),
            clientDataJSON: assertion.clientDataJSON.base64EncodedString(),
            authenticatorData: assertion.authenticatorData.base64EncodedString(),
            signature: assertion.signature.base64EncodedString()
        ))
        
        let data = try await makeRequest(path: "auth/login", body: body)
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        return try decoder.decode(AuthResponse.self, from: data)
    }
    
    // ... other endpoint implementations follow same pattern
}
```

---

## 9. Security Layer

### 9.1 Keychain Configuration

All sensitive auth data uses the most restrictive Keychain access level:

```swift
// kSecAttrAccessibleWhenUnlockedThisDeviceOnly
// - Only accessible when device is unlocked
// - NOT included in iCloud Keychain backup
// - NOT migrated to new device (forces re-auth on new device)
```

| Key | Data | Access Level |
|---|---|---|
| `echo.auth.refresh_token` | Refresh token string | `whenUnlockedThisDeviceOnly` |
| `echo.biometric.domain_state` | LAContext domain state | `whenUnlockedThisDeviceOnly` |
| `echo.display.masked_phone` | Masked phone for UI | `afterFirstUnlock` |

### 9.2 Memory Protection

```swift
// Access tokens are NEVER:
// - Written to UserDefaults
// - Written to Keychain
// - Written to disk in any form
// - Logged to console
// - Included in crash reports

// Access tokens ARE:
// - Held as a private var in TokenManager (in-memory only)
// - Cleared on app termination
// - Cleared on background (optional, configurable)
```

### 9.3 Debug vs Release

```swift
#if DEBUG
    // Allow none attestation format for simulator testing
    // Disable certificate pinning for Charles/Proxyman debugging
    // Log auth state transitions to console
#else
    // Require Apple App Attestation
    // Enforce certificate pinning
    // No auth logging
    // Jailbreak detection enabled
#endif
```

---

## 10. Error Handling

```swift
// Features/Auth/Models/AuthError.swift

enum AuthError: LocalizedError {
    case passkeyCreationFailed
    case passkeyAssertionFailed
    case noRefreshToken
    case tokenStorageFailed
    case sessionRevoked
    case biometricChanged
    case deviceIntegrityFailed
    case networkUnavailable
    
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
        }
    }
}

struct APIError: Error {
    let code: String       // e.g. "AUTH_003"
    let userMessage: String
    let httpStatus: Int
}

struct ErrorResponse: Decodable {
    let error: ErrorDetail
    struct ErrorDetail: Decodable {
        let code: String
        let message: String
    }
}
```

---

## 11. Accessibility

All auth views must meet these requirements:

| Requirement | Implementation |
|---|---|
| Tap targets | Minimum 44x44pt on all interactive elements |
| VoiceOver | Accessibility labels on all buttons, inputs, and status indicators |
| Dynamic Type | All text scales with system font size preference |
| Reduce Motion | Disable shake animation, OTP auto-advance, and screen transitions |
| Color contrast | WCAG AA minimum (4.5:1 for text, 3:1 for large text) |
| Keyboard navigation | Full tab order through all auth forms |

```swift
// Example: OTP box accessibility
TextField("", text: $viewModel.code[index])
    .accessibilityLabel("Verification code digit \(index + 1) of 6")
    .accessibilityHint(
        viewModel.code[index].isEmpty
            ? "Enter digit"
            : "Digit entered: \(viewModel.code[index])"
    )
```

---

## 12. Testing

### 12.1 Unit Tests

| Class | Key Test Cases |
|---|---|
| `AuthCoordinator` | All valid state transitions, invalid transition rejection, session restore |
| `TokenManager` | Token storage/retrieval, auto-refresh, concurrent refresh coalescing, replay handling |
| `PhoneEntryViewModel` | Phone validation, formatting, OTP send success/failure |
| `OTPViewModel` | Code entry, auto-advance, auto-submit, resend timer, error clearing |
| `LoginViewModel` | Passkey login, new device detection, account locked handling |
| `DeviceFingerprintService` | Device hash stability, jailbreak detection, model identifier |
| `BiometricIntegrityService` | State save/load, enrollment change detection |

### 12.2 UI Tests

| Flow | Steps |
|---|---|
| Registration | Welcome → Phone → OTP → Passkey → Profile → Trust Intro → Home |
| Login | Welcome → Login → Face ID → Home |
| Session Restore | Kill app → Relaunch → Auto-restore → Home |
| Step-Up | Home → Settings → Devices → Remove → Step-Up Sheet → Confirm |
| Error States | Invalid phone, wrong OTP, expired OTP, locked account |

### 12.3 Mock Protocols

All services use protocol abstractions for testability:

```swift
// Test doubles
final class MockPasskeyManager: PasskeyManagerProtocol { ... }
final class MockTokenManager: TokenManager { ... }
final class MockAuthAPIClient: AuthAPIClientProtocol { ... }
final class MockKeychainManager: KeychainManagerProtocol { ... }
```

---

*Last Updated: March 2026*
*iOS Target: 16.0+*
*Swift Version: 5.9+*
