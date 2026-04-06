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
        maskedPhone = UserDefaults.standard.string(forKey: "echo.display.masked_phone") ?? ""
        username = UserDefaults.standard.string(forKey: "echo.display.username") ?? ""
    }

    func loginWithPasskey() async -> AuthUserProfile? {
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

            // Step 5: Save display info for next login
            if let user = authResponse.user {
                if let name = user.username {
                    UserDefaults.standard.set(name, forKey: "echo.display.username")
                }
            }

            return authResponse.user
        } catch let error as AuthAPIError where error.code == "AUTH_007" {
            errorMessage = "New device detected. Additional verification needed."
            return nil
        } catch let error as AuthAPIError where error.code == "AUTH_009" {
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
