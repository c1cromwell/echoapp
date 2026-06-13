#if os(iOS)
import Foundation
import CryptoKit

/// Stable per-install device id for WO-CA3 sync streams (`target_device_id` / pull `device_id`).
enum DeviceIdentityStore {
    private static let keychainKey = "echo.sync.deviceId"

    /// Deterministic sync stream id from a device's registered P-256 public key hex.
    static func syncDeviceId(fromPublicKeyHex hex: String) -> String {
        var cleaned = hex.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if cleaned.hasPrefix("0x") { cleaned = String(cleaned.dropFirst(2)) }
        let digest = SHA256.hash(data: Data(cleaned.utf8))
        let suffix = digest.prefix(16).map { String(format: "%02x", $0) }.joined()
        return "dev-\(suffix)"
    }

    /// Pins the sync stream id for this install from the linked device's public key.
    static func assignSyncDeviceId(fromPublicKeyHex hex: String) {
        let id = syncDeviceId(fromPublicKeyHex: hex)
        try? KeychainManager.shared.store(value: id, key: keychainKey)
    }

    /// Returns a durable device id, creating and persisting one on first access.
    static func currentDeviceId() -> String {
        if let existing = try? KeychainManager.shared.retrieve(key: keychainKey),
           !existing.isEmpty {
            return existing
        }
        let id = "dev-\(UUID().uuidString.lowercased())"
        try? KeychainManager.shared.store(value: id, key: keychainKey)
        return id
    }

    #if DEBUG
    /// Test-only reset.
    static func resetForTesting() {
        try? KeychainManager.shared.delete(key: keychainKey)
    }
    #endif
}
#endif
