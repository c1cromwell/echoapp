#if os(iOS)
import Foundation
import Observation
import UIKit

/// Biometric-gated hidden-chats vault session (WO-7/18).
@MainActor
@Observable
public final class HiddenChatsSession {
    public static let shared = HiddenChatsSession()

    private(set) public var isUnlocked = false
    /// True when the user entered the duress PIN — show an empty decoy vault.
    private(set) public var isDuressMode = false

    private var unlockedAt: Date?
    private var screenshotObserver: NSObjectProtocol?

    private init() {}

    public func unlock(biometric: Bool = true, duress: Bool = false) {
        isUnlocked = true
        isDuressMode = duress
        unlockedAt = Date()
        HiddenFolderAuditLog.record(duress ? .duressEntry : .unlock)
        if screenshotObserver == nil {
            screenshotObserver = NotificationCenter.default.addObserver(
                forName: UIApplication.userDidTakeScreenshotNotification,
                object: nil,
                queue: .main
            ) { [weak self] _ in
                Task { @MainActor in
                    guard HiddenFolderSettingsStore.shared.lockOnScreenshot else { return }
                    self?.lock()
                }
            }
        }
    }

    public func lock() {
        if isUnlocked { HiddenFolderAuditLog.record(.lock) }
        isUnlocked = false
        isDuressMode = false
        unlockedAt = nil
    }

    /// Lock immediately when the app leaves the foreground.
    public func noteDidEnterBackground() {
        if isUnlocked { lock() }
    }

    /// Re-lock if the user stayed away longer than the configured timeout.
    public func refreshForegroundLockIfNeeded() {
        guard isUnlocked, let unlockedAt else { return }
        let timeout = HiddenFolderSettingsStore.shared.autoLockSeconds
        if Date().timeIntervalSince(unlockedAt) > timeout {
            lock()
        }
    }

    /// Whether inbox previews and badge counts should update for a hidden chat.
    public func shouldSurfaceNotification(for conversationId: String) -> Bool {
        guard ConversationPreferencesStore.shared.isHidden(conversationId) else { return true }
        switch HiddenFolderSettingsStore.shared.notificationMode {
        case .suppressed:
            return false
        case .redacted:
            return isUnlocked && !isDuressMode
        }
    }

    public func redactedPreviewIfNeeded(for conversationId: String, resolved: String) -> String {
        guard ConversationPreferencesStore.shared.isHidden(conversationId) else { return resolved }
        switch HiddenFolderSettingsStore.shared.notificationMode {
        case .suppressed:
            return resolved
        case .redacted:
            return "New message"
        }
    }
}
#endif
