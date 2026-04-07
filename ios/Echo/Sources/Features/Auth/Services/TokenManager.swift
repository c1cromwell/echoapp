import Foundation
import Combine

// MARK: - Token Manager

final class TokenManager: ObservableObject {
    @Published private(set) var isAuthenticated = false

    // Access token lives in memory only — never persisted
    private var accessToken: String?
    private var accessTokenExpiry: Date?

    private let keychain: DataKeychainProtocol
    private let apiClient: AuthAPIClientProtocol
    private var refreshTask: Task<String, Error>?

    static let refreshTokenKey = "echo.auth.refresh_token"
    static let biometricStateKey = "echo.auth.biometric_state"

    init(keychain: DataKeychainProtocol, apiClient: AuthAPIClientProtocol) {
        self.keychain = keychain
        self.apiClient = apiClient
        self.isAuthenticated = hasStoredRefreshToken()
    }

    // MARK: - Store Tokens from Auth Response

    func storeTokens(_ response: AuthTokenResponse) throws {
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

    // MARK: - Current Access Token (no refresh)

    var currentAccessToken: String? { accessToken }

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
        } catch let error as AuthAPIError where error.code == "AUTH_006" {
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

// MARK: - Keychain Adapter

final class KeychainAdapter: DataKeychainProtocol {
    
    func save(_ data: Data, for key: String) throws {
        let semaphore = DispatchSemaphore(value: 0)
        var thrownError: Error?
        Task {
            do {
                let manager = KeychainManager.shared
                try await manager.store(data: data, key: key)
            } catch {
                thrownError = error
            }
            semaphore.signal()
        }
        semaphore.wait()
        if let error = thrownError { throw error }
    }

    func load(for key: String) throws -> Data? {
        let semaphore = DispatchSemaphore(value: 0)
        var result: Data?
        var thrownError: Error?
        Task {
            do {
                let manager = KeychainManager.shared
                result = try await manager.retrieveData(key: key)
            } catch {
                thrownError = error
            }
            semaphore.signal()
        }
        semaphore.wait()
        if let error = thrownError { throw error }
        return result
    }

    func delete(for key: String) throws {
        let semaphore = DispatchSemaphore(value: 0)
        var thrownError: Error?
        Task {
            do {
                let manager = KeychainManager.shared
                try await manager.delete(key: key)
            } catch {
                thrownError = error
            }
            semaphore.signal()
        }
        semaphore.wait()
        if let error = thrownError { throw error }
    }
}

// MARK: - Mock for Testing

#if DEBUG
final class MockDataKeychainManager: DataKeychainProtocol {
    var storage: [String: Data] = [:]

    func save(_ data: Data, for key: String) throws {
        storage[key] = data
    }

    func load(for key: String) throws -> Data? {
        storage[key]
    }

    func delete(for key: String) throws {
        storage.removeValue(forKey: key)
    }
}
#endif
