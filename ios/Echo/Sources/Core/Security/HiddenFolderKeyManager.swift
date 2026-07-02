#if os(iOS)
import CryptoKit
import Foundation
import LocalAuthentication
import Security

/// Biometric-derived per-folder AES keys (WO-18).
actor HiddenFolderKeyManager {
    static let shared = HiddenFolderKeyManager()

    private let masterKeyId = "echo-hidden-folder-master"
    private let keychainService = "com.echo.hidden-folders"
    private let domainStateKey = "echo.hidden.biometric.domain"
    private let secureEnclave = SecureEnclaveManager.shared

    private var cachedKeys: [String: SymmetricKey] = [:]

    func keyForFolder(_ folderId: String) async throws -> SymmetricKey {
        if let cached = cachedKeys[folderId] { return cached }
        try await ensureMasterKey()
        try await rotateIfBiometricChanged()

        let account = "hidden_folder_key_\(folderId)"
        if let stored = try await loadBiometricProtected(account: account) {
            let key = SymmetricKey(data: stored)
            cachedKeys[folderId] = key
            return key
        }

        let signature = try await secureEnclave.sign(
            data: Data(folderId.utf8),
            keyId: masterKeyId
        )
        let derived = HKDF<SHA256>.deriveKey(
            inputKeyMaterial: SymmetricKey(data: signature),
            salt: Data(folderId.utf8),
            info: Data("echo.hidden_folder.\(folderId)".utf8),
            outputByteCount: 32
        )
        let raw = derived.withUnsafeBytes { Data($0) }
        try await storeBiometricProtected(account: account, value: raw)
        cachedKeys[folderId] = derived
        return derived
    }

    func destroyKey(_ folderId: String) async {
        cachedKeys.removeValue(forKey: folderId)
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: "hidden_folder_key_\(folderId)",
        ]
        SecItemDelete(query as CFDictionary)
    }

    func purgeCache() {
        cachedKeys.removeAll()
    }

    private func ensureMasterKey() async throws {
        #if targetEnvironment(simulator)
        _ = try await secureEnclave.generateBiometricProtectedKey(id: masterKeyId)
        #else
        let query: [String: Any] = [
            kSecClass as String: kSecClassKey,
            kSecAttrLabel as String: masterKeyId,
            kSecReturnRef as String: true,
        ]
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        if status == errSecItemNotFound {
            _ = try await secureEnclave.generateBiometricProtectedKey(id: masterKeyId)
        } else if status != errSecSuccess {
            throw HiddenFolderKeyError.keyUnavailable
        }
        #endif
    }

    private func rotateIfBiometricChanged() async throws {
        let context = LAContext()
        guard let domain = context.evaluatedPolicyDomainState else { return }
        let encoded = domain.base64EncodedString()
        let prior = UserDefaults.standard.string(forKey: domainStateKey)
        guard let prior, prior != encoded else {
            UserDefaults.standard.set(encoded, forKey: domainStateKey)
            return
        }
        purgeCache()
        let deleteQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
        ]
        SecItemDelete(deleteQuery as CFDictionary)
        UserDefaults.standard.set(encoded, forKey: domainStateKey)
    }

    private func storeBiometricProtected(account: String, value: Data) async throws {
        var error: Unmanaged<CFError>?
        guard let access = SecAccessControlCreateWithFlags(
            kCFAllocatorDefault,
            kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            [.biometryCurrentSet],
            &error
        ) else {
            throw HiddenFolderKeyError.keyUnavailable
        }
        let deleteQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: account,
        ]
        SecItemDelete(deleteQuery as CFDictionary)
        let addQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: account,
            kSecValueData as String: value,
            kSecAttrAccessControl as String: access,
            kSecAttrSynchronizable as String: false,
        ]
        let status = SecItemAdd(addQuery as CFDictionary, nil)
        guard status == errSecSuccess else { throw HiddenFolderKeyError.keyUnavailable }
    }

    private func loadBiometricProtected(account: String) async throws -> Data? {
        let context = LAContext()
        context.localizedReason = "Unlock hidden folder encryption"
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: account,
            kSecReturnData as String: true,
            kSecUseAuthenticationContext as String: context,
        ]
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        switch status {
        case errSecSuccess:
            return item as? Data
        case errSecItemNotFound:
            return nil
        default:
            throw HiddenFolderKeyError.keyUnavailable
        }
    }
}

enum HiddenFolderKeyError: LocalizedError {
    case keyUnavailable

    var errorDescription: String? {
        switch self {
        case .keyUnavailable: return "Hidden folder encryption key is unavailable."
        }
    }
}
#endif
