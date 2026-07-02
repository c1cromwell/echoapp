#if os(iOS)
import CryptoKit
import Foundation

/// Per-folder encryption key manager for hidden conversations (WO-7/18).
/// In production, keys are derived from the Secure Enclave and cached per session.
actor HiddenFolderKeyManager {
    static let shared = HiddenFolderKeyManager()

    private var cache: [String: SymmetricKey] = [:]

    private init() {}

    func keyForFolder(_ folderId: String) async throws -> SymmetricKey {
        if let cached = cache[folderId] { return cached }
        let key = deriveKey(folderId: folderId)
        cache[folderId] = key
        return key
    }

    func purgeCache() {
        cache.removeAll()
    }

    private func deriveKey(folderId: String) -> SymmetricKey {
        let salt = Data(folderId.utf8)
        let ikm = SymmetricKey(size: .bits256)
        return HKDF<SHA256>.deriveKey(
            inputKeyMaterial: ikm,
            salt: salt,
            info: Data("echo.hidden-folder.v1".utf8),
            outputByteCount: 32
        )
    }
}
#endif
