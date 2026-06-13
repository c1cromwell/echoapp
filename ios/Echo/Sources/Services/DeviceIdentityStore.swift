#if os(iOS)
import Foundation

/// Stable per-install device id for WO-CA3 sync streams (`target_device_id` / pull `device_id`).
enum DeviceIdentityStore {
    private static let keychainKey = "echo.sync.deviceId"

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
