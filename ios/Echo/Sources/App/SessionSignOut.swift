#if os(iOS)
import Foundation

/// Clears local auth session so Settings → Sign Out returns to the login gate.
@MainActor
enum SessionSignOut {
    private static let skipAutoUnlockKey = "echo.skipLoginAutoUnlock"

    /// After sign-out, login screen waits for explicit Face ID tap (not auto-trigger).
    static var shouldSkipLoginAutoUnlock: Bool {
        UserDefaults.standard.bool(forKey: skipAutoUnlockKey)
    }

    static func consumeSkipLoginAutoUnlock() {
        UserDefaults.standard.removeObject(forKey: skipAutoUnlockKey)
    }

    static func perform() async {
        await bestEffortServerLogout()
        await clearLocalCredentials()
        await disconnectRealtime()
        ActiveChatRegistry.openConversationId = nil
        UserDefaults.standard.set(true, forKey: skipAutoUnlockKey)
    }

    /// Wipes first-run flag so "Create account" can run full onboarding (dev / second account).
    static func performIncludingOnboardingReset() async {
        await perform()
        UserDefaults.standard.set(false, forKey: "echo.hasCompletedFirstRun")
    }

    private static func bestEffortServerLogout() async {
        guard let client = DIContainer.shared.resolveAPIClient(),
              (try? await KeychainManager.shared.getAuthToken()) != nil else {
            return
        }
        struct Empty: Encodable {}
        struct RevokeResponse: Decodable { let revoked: Bool? }
        let _: RevokeResponse? = try? await client.post(endpoint: AuthEndpoint.logout, body: Empty())
    }

    private static func clearLocalCredentials() async {
        let tokenManager = TokenManager(keychain: KeychainAdapter(), apiClient: AuthAPIClient())
        tokenManager.clearTokens()
        BiometricIntegrityService(keychain: KeychainAdapter()).clearBiometricState()

        try? await KeychainManager.shared.clearAuthCredentials()
        for service in ["com.echo.auth", "com.echo.tokens", "com.echo.credentials"] {
            try? await KeychainManager.shared.clearAll(service: service)
        }
        // Keep echo.username.current / echo.did.current so returning-user login still works.
    }

    private static func disconnectRealtime() async {
        guard let service = DIContainer.shared.resolveConversationSignalService() else { return }
        await service.disconnect()
    }
}
#endif
