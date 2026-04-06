import SwiftUI
import Combine

// MARK: - Auth Coordinator (State Machine)

@MainActor
final class AuthCoordinatorVM: ObservableObject {
    @Published private(set) var state: AuthState = .unauthenticated
    @Published var showStepUpSheet = false
    @Published var stepUpAction: StepUpAction?
    @Published var stepUpCompletion: ((String) -> Void)?

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
            Task { await completeOnboarding() }

        case (_, .loginSucceeded(let user)):
            biometricService.saveBiometricState()
            state = .authenticated(user: user)

        case (_, .sessionExpired), (_, .loggedOut):
            tokenManager.clearTokens()
            biometricService.clearBiometricState()
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

    func requestStepUp(for action: StepUpAction, completion: @escaping (String) -> Void) {
        stepUpAction = action
        stepUpCompletion = completion
        showStepUpSheet = true
    }

    func completeStepUp(elevatedToken: String) {
        showStepUpSheet = false
        stepUpCompletion?(elevatedToken)
        stepUpCompletion = nil
        stepUpAction = nil
    }

    // MARK: - Logout

    func logout(allDevices: Bool = false) async {
        do {
            let token = try await tokenManager.getValidAccessToken()
            try await apiClient.logout(token: token, allDevices: allDevices)
        } catch {
            // Still clear local state even if server logout fails
        }
        handle(.loggedOut)
    }

    // MARK: - Private

    private func completeOnboarding() async {
        do {
            let token = try await tokenManager.getValidAccessToken()
            let user = try await apiClient.fetchCurrentUser(token: token)
            biometricService.saveBiometricState()
            state = .authenticated(user: user)
        } catch {
            state = .unauthenticated
        }
    }
}
