#if os(iOS)
import Foundation
import CryptoKit

/// Stable per-install device id for WO-CA3 sync streams (`target_device_id` / pull `device_id`).
enum DeviceIdentityStore {
    private static let keychainKey = "echo.sync.deviceId"
    private static let localPubKeyKey = "echo.sync.localPubKeyHex"

    /// Deterministic sync stream id from a device's registered P-256 public key hex.
    static func syncDeviceId(fromPublicKeyHex hex: String) -> String {
        var cleaned = hex.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if cleaned.hasPrefix("0x") { cleaned = String(cleaned.dropFirst(2)) }
        let digest = SHA256.hash(data: Data(cleaned.utf8))
        let suffix = digest.prefix(16).map { String(format: "%02x", $0) }.joined()
        return "dev-\(suffix)"
    }

    /// Pins the sync stream id for this install from the linked device's public key.
    static func assignSyncDeviceId(fromPublicKeyHex hex: String) async {
        let id = syncDeviceId(fromPublicKeyHex: hex)
        try? await KeychainManager.shared.store(value: id, key: keychainKey)
    }

    /// Records this install's registered device pubkey and pins the matching sync stream id.
    static func pinLocalPublicKey(hex: String) async {
        var cleaned = hex.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if cleaned.hasPrefix("0x") { cleaned = String(cleaned.dropFirst(2)) }
        guard !cleaned.isEmpty else { return }
        try? await KeychainManager.shared.store(value: cleaned, key: localPubKeyKey)
        await assignSyncDeviceId(fromPublicKeyHex: cleaned)
    }

    static func localPublicKeyHex() async -> String? {
        try? await KeychainManager.shared.retrieve(key: localPubKeyKey)
    }

    /// True when `publicKeyHex` belongs to this install (identity-registered device key).
    static func isLocalDevice(publicKeyHex: String) async -> Bool {
        guard let mine = await localPublicKeyHex(), !mine.isEmpty else { return false }
        var other = publicKeyHex.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if other.hasPrefix("0x") { other = String(other.dropFirst(2)) }
        return mine == other
    }

    /// Returns a durable device id, creating and persisting one on first access.
    static func currentDeviceId() async -> String {
        if let existing = try? await KeychainManager.shared.retrieve(key: keychainKey),
           !existing.isEmpty {
            return existing
        }
        let id = "dev-\(UUID().uuidString.lowercased())"
        try? await KeychainManager.shared.store(value: id, key: keychainKey)
        return id
    }

    #if DEBUG
    /// Test-only reset.
    static func resetForTesting() async {
        try? await KeychainManager.shared.delete(key: keychainKey)
    }
    #endif
}
#endif
