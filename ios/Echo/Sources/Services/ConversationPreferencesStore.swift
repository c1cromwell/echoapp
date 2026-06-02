import Foundation

/// Disappearing-message timer options (spec §5.4 / ux-spec §2.6).
public enum DisappearingTimer: String, Codable, CaseIterable, Sendable {
    case off, s30, m5, h1, h24, d7

    public var label: String {
        switch self {
        case .off: return "Off"
        case .s30: return "30s"
        case .m5:  return "5m"
        case .h1:  return "1h"
        case .h24: return "24h"
        case .d7:  return "7d"
        }
    }
}

/// Per-conversation preferences backing the chat-settings sheet and row badges.
public struct ConversationPreferences: Codable, Equatable, Sendable {
    public var isMuted: Bool
    public var disappearing: DisappearingTimer

    public init(isMuted: Bool = false, disappearing: DisappearingTimer = .off) {
        self.isMuted = isMuted
        self.disappearing = disappearing
    }
}

/// Persistence seam so the store can be unit-tested with an in-memory backing.
public protocol ConversationPreferencesBacking: AnyObject {
    func data(forKey key: String) -> Data?
    func set(_ data: Data?, forKey key: String)
}

extension UserDefaults: ConversationPreferencesBacking {
    public func set(_ data: Data?, forKey key: String) {
        if let data { set(data as Any, forKey: key) } else { removeObject(forKey: key) }
    }
}

/// In-memory backing for tests/previews.
public final class InMemoryPreferencesBacking: ConversationPreferencesBacking {
    private var store: [String: Data] = [:]
    public init() {}
    public func data(forKey key: String) -> Data? { store[key] }
    public func set(_ data: Data?, forKey key: String) {
        if let data { store[key] = data } else { store.removeValue(forKey: key) }
    }
}

/// Lightweight per-conversation preferences store. Mute/silent + disappearing timer.
/// Real cross-device sync (backend TTL policy) is deferred (spec §4.2 P2).
@MainActor
public final class ConversationPreferencesStore {
    static let shared = ConversationPreferencesStore()

    private let backing: ConversationPreferencesBacking
    private let prefix = "echo.convprefs."

    public init(backing: ConversationPreferencesBacking = UserDefaults.standard) {
        self.backing = backing
    }

    public func preferences(for conversationId: String) -> ConversationPreferences {
        guard
            let data = backing.data(forKey: prefix + conversationId),
            let prefs = try? JSONDecoder().decode(ConversationPreferences.self, from: data)
        else { return ConversationPreferences() }
        return prefs
    }

    public func save(_ prefs: ConversationPreferences, for conversationId: String) {
        let data = try? JSONEncoder().encode(prefs)
        backing.set(data, forKey: prefix + conversationId)
    }

    public func setMuted(_ muted: Bool, for conversationId: String) {
        var prefs = preferences(for: conversationId)
        prefs.isMuted = muted
        save(prefs, for: conversationId)
    }

    public func setDisappearing(_ timer: DisappearingTimer, for conversationId: String) {
        var prefs = preferences(for: conversationId)
        prefs.disappearing = timer
        save(prefs, for: conversationId)
    }

    public func isMuted(_ conversationId: String) -> Bool {
        preferences(for: conversationId).isMuted
    }
}
