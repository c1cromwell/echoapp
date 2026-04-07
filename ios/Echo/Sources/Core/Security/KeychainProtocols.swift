import Foundation

// MARK: - Data Keychain Protocol

protocol DataKeychainProtocol {
    func save(_ data: Data, for key: String) throws
    func load(for key: String) throws -> Data?
    func delete(for key: String) throws
}

// MARK: - Token Keychain Protocol

protocol TokenKeychainProtocol {
    func saveToken(_ token: String) throws
    func retrieveToken() -> String?
    func clearAll()
}