// Core/Mesh/MeshPanicWipe.swift
//
// Emergency "panic" wipe (P4): a triple-tap clears all local data instantly. The data-clearing
// half is pure (UserDefaults) and unit-testable; the full wipe also stops the mesh and locks
// secure storage on device.

import Foundation

public enum MeshPanicWipe {
    private static let wipePrefixes = ["echo.thread.v1.", "echo.mesh."]

    /// Clear locally-persisted message + mesh state. Returns the number of keys removed.
    @discardableResult
    public static func clearPersistedData(_ defaults: UserDefaults = .standard) -> Int {
        var removed = 0
        for key in defaults.dictionaryRepresentation().keys
        where wipePrefixes.contains(where: key.hasPrefix) {
            defaults.removeObject(forKey: key)
            removed += 1
        }
        return removed
    }
}

#if os(iOS)
extension MeshPanicWipe {
    /// Triple-tap emergency wipe: stop the mesh, clear local data, lock secure storage.
    static func wipe(stack: MeshMessagingStack? = nil) {
        stack?.stop()
        clearPersistedData()
        Task { await SecureEnclaveManager.shared.lockStorage() }
    }
}
#endif
