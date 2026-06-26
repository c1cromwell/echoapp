// Features/Safety/FirstContactStore.swift
//
// Tracks which peers we've already opened a thread with, so the safety banner fires only on a
// genuine first contact (not every time you reopen a known chat).

import Foundation

public enum FirstContactStore {
    private static let key = "echo.seenContacts"

    public static func isFirstContact(_ did: String, defaults: UserDefaults = .standard) -> Bool {
        !(defaults.stringArray(forKey: key) ?? []).contains(did)
    }

    public static func markSeen(_ did: String, defaults: UserDefaults = .standard) {
        var seen = Set(defaults.stringArray(forKey: key) ?? [])
        guard !seen.contains(did) else { return }
        seen.insert(did)
        defaults.set(Array(seen), forKey: key)
    }
}
