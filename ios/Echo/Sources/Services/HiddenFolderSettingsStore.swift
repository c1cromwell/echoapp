#if os(iOS)
import Foundation
import CryptoKit

/// Device-local hidden-folder preferences (WO-7). No server sync.
public enum HiddenNotificationMode: String, Codable, CaseIterable, Sendable {
    case suppressed
    case redacted

    public var label: String {
        switch self {
        case .suppressed: return "Suppressed"
        case .redacted: return "Redacted"
        }
    }
}

@MainActor
public final class HiddenFolderSettingsStore {
    public static let shared = HiddenFolderSettingsStore()

    private let defaults: UserDefaults
    private enum Key {
        static let autoLock = "echo.hidden.autolockSeconds"
        static let lockOnScreenshot = "echo.hidden.lockOnScreenshot"
        static let notificationMode = "echo.hidden.notificationMode"
        static let duressPINHash = "echo.hidden.duressPINHash"
    }

    public init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    public var autoLockSeconds: TimeInterval {
        get {
            let stored = defaults.double(forKey: Key.autoLock)
            return stored > 0 ? stored : 120
        }
        set { defaults.set(newValue, forKey: Key.autoLock) }
    }

    public var lockOnScreenshot: Bool {
        get {
            if defaults.object(forKey: Key.lockOnScreenshot) == nil { return true }
            return defaults.bool(forKey: Key.lockOnScreenshot)
        }
        set { defaults.set(newValue, forKey: Key.lockOnScreenshot) }
    }

    public var notificationMode: HiddenNotificationMode {
        get {
            guard let raw = defaults.string(forKey: Key.notificationMode),
                  let mode = HiddenNotificationMode(rawValue: raw) else {
                return .suppressed
            }
            return mode
        }
        set { defaults.set(newValue.rawValue, forKey: Key.notificationMode) }
    }

    public var hasDuressPIN: Bool {
        defaults.string(forKey: Key.duressPINHash) != nil
    }

    public func setDuressPIN(_ pin: String) throws {
        let trimmed = pin.trimmingCharacters(in: .whitespacesAndNewlines)
        guard trimmed.count >= 4, trimmed.count <= 8, trimmed.allSatisfy(\.isNumber) else {
            throw HiddenFolderSettingsError.invalidPIN
        }
        defaults.set(Self.hashPIN(trimmed), forKey: Key.duressPINHash)
    }

    public func clearDuressPIN() {
        defaults.removeObject(forKey: Key.duressPINHash)
    }

    public func matchesDuressPIN(_ pin: String) -> Bool {
        guard let stored = defaults.string(forKey: Key.duressPINHash) else { return false }
        let trimmed = pin.trimmingCharacters(in: .whitespacesAndNewlines)
        return stored == Self.hashPIN(trimmed)
    }

    private static func hashPIN(_ pin: String) -> String {
        let digest = SHA256.hash(data: Data(pin.utf8))
        return digest.map { String(format: "%02x", $0) }.joined()
    }
}

public enum HiddenFolderSettingsError: LocalizedError {
    case invalidPIN

    public var errorDescription: String? {
        switch self {
        case .invalidPIN: return "Enter a 4–8 digit PIN."
        }
    }
}
#endif
