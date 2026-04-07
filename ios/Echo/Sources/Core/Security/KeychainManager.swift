import Foundation
import Security

/// Secure wrapper around iOS Keychain for credential storage
actor KeychainManager {
    
    // MARK: - Singleton
    static let shared = KeychainManager()
    
    // MARK: - Service Identifiers
    
    private enum KeychainService: String {
        case authentication = "com.echo.auth"
        case encryption = "com.echo.encryption"
        case tokens = "com.echo.tokens"
        case credentials = "com.echo.credentials"
        case privateKey = "com.echo.privateKey"
        case publicKey = "com.echo.publicKey"
    }
    
    // MARK: - Initialization
    
    private init() {}
    
    // MARK: - Store Operations
    
    /// Store a string value in Keychain
    func store(
        value: String,
        key: String,
        service: String = KeychainService.authentication.rawValue
    ) throws {
        guard let data = value.data(using: .utf8) else {
            throw KeychainError.encodingError
        }
        
        try store(data: data, key: key, service: service)
    }
    
    /// Store data in Keychain
    func store(
        data: Data,
        key: String,
        service: String = KeychainService.authentication.rawValue
    ) throws {
        // First, try to delete existing value
        try? delete(key: key, service: service)
        
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        
        let status = SecItemAdd(query as CFDictionary, nil)
        
        guard status == errSecSuccess else {
            throw KeychainError.storageError(status)
        }
    }
    
    // MARK: - Retrieve Operations
    
    /// Retrieve a string value from Keychain
    func retrieve(
        key: String,
        service: String = KeychainService.authentication.rawValue
    ) throws -> String? {
        guard let data = try retrieveData(key: key, service: service) else {
            return nil
        }
        
        return String(data: data, encoding: .utf8)
    }
    
    /// Retrieve data from Keychain
    func retrieveData(
        key: String,
        service: String = KeychainService.authentication.rawValue
    ) throws -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true
        ]
        
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        
        switch status {
        case errSecSuccess:
            return result as? Data
        case errSecItemNotFound:
            return nil
        default:
            throw KeychainError.retrievalError(status)
        }
    }
    
    // MARK: - Delete Operations
    
    /// Delete a value from Keychain
    func delete(
        key: String,
        service: String = KeychainService.authentication.rawValue
    ) throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key
        ]
        
        let status = SecItemDelete(query as CFDictionary)
        
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw KeychainError.deletionError(status)
        }
    }
    
    /// Clear all stored credentials
    func clearAll(service: String = KeychainService.authentication.rawValue) throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service
        ]
        
        let status = SecItemDelete(query as CFDictionary)
        
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw KeychainError.deletionError(status)
        }
    }
    
    // MARK: - Credential Management
    
    /// Store authentication token
    func storeAuthToken(_ token: String) throws {
        try store(value: token, key: "authToken", service: KeychainService.authentication.rawValue)
    }
    
    /// Retrieve authentication token
    func getAuthToken() throws -> String? {
        try retrieve(key: "authToken", service: KeychainService.authentication.rawValue)
    }
    
    /// Store refresh token
    func storeRefreshToken(_ token: String) throws {
        try store(value: token, key: "refreshToken", service: KeychainService.authentication.rawValue)
    }
    
    /// Retrieve refresh token
    func getRefreshToken() throws -> String? {
        try retrieve(key: "refreshToken", service: KeychainService.authentication.rawValue)
    }
    
    /// Store private key
    func storePrivateKey(_ key: Data) throws {
        try store(data: key, key: "privateKey", service: KeychainService.privateKey.rawValue)
    }
    
    /// Retrieve private key
    func getPrivateKey() throws -> Data? {
        try retrieveData(key: "privateKey", service: KeychainService.privateKey.rawValue)
    }
    
    /// Store public key
    func storePublicKey(_ key: Data) throws {
        try store(data: key, key: "publicKey", service: KeychainService.publicKey.rawValue)
    }
    
    /// Retrieve public key
    func getPublicKey() throws -> Data? {
        try retrieveData(key: "publicKey", service: KeychainService.publicKey.rawValue)
    }
    
    /// Clear authentication credentials
    func clearAuthCredentials() throws {
        try delete(key: "authToken", service: KeychainService.authentication.rawValue)
        try delete(key: "refreshToken", service: KeychainService.authentication.rawValue)
    }
    
    // MARK: - Generic Codable Operations
    
    /// Store a Codable value in Keychain
    func store<T: Codable>(key: String, value: T) async throws {
        let data = try JSONEncoder().encode(value)
        
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]
        
        // Delete existing
        SecItemDelete(query as CFDictionary)
        
        // Add new
        let status = SecItemAdd(query as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw KeychainError.storageError(status)
        }
    }
    
    /// Retrieve a Codable value from Keychain
    func retrieve<T: Codable>(key: String, as type: T.Type) async throws -> T? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true
        ]
        
        var result: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        
        guard status == errSecSuccess else {
            return nil
        }
        
        guard let data = result as? Data else {
            return nil
        }
        
        return try JSONDecoder().decode(T.self, from: data)
    }
    
    // MARK: - Verification
    
    /// Check if a key exists in Keychain
    func exists(key: String, service: String = KeychainService.authentication.rawValue) -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key,
            kSecReturnData as String: false
        ]
        
        let status = SecItemCopyMatching(query as CFDictionary, nil)
        return status == errSecSuccess
    }
}

// MARK: - Keychain Errors

enum KeychainError: LocalizedError {
    case encodingError
    case storageError(OSStatus)
    case retrievalError(OSStatus)
    case deletionError(OSStatus)
    case invalidFormat
    case keyNotFound
    
    var errorDescription: String? {
        switch self {
        case .encodingError:
            return "Failed to encode value for Keychain storage"
        case .storageError(let status):
            return "Failed to store value in Keychain (status: \(status))"
        case .retrievalError(let status):
            return "Failed to retrieve value from Keychain (status: \(status))"
        case .deletionError(let status):
            return "Failed to delete value from Keychain (status: \(status))"
        case .invalidFormat:
            return "Invalid data format in Keychain"
        case .keyNotFound:
            return "Key not found in Keychain"
        }
    }
}
