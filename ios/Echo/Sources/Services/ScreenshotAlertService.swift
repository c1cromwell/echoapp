#if os(iOS)
import Foundation
import UIKit

/// Sends ephemeral screenshot alerts to conversation peers when enabled (M6).
@MainActor
final class ScreenshotAlertService {
    static let shared = ScreenshotAlertService()

    private var observer: NSObjectProtocol?
    private var activeConversationId: String?
    private var activePeerDID: String?
    private var signalService: ConversationSignalService?

    private init() {}

    func configure(signalService: ConversationSignalService) {
        self.signalService = signalService
    }

    func setActiveChat(conversationId: String?, peerDID: String?) {
        activeConversationId = conversationId
        activePeerDID = peerDID
    }

    func startMonitoring() {
        guard observer == nil else { return }
        observer = NotificationCenter.default.addObserver(
            forName: UIApplication.userDidTakeScreenshotNotification,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            Task { @MainActor in await self?.handleScreenshot() }
        }
    }

    func stopMonitoring() {
        if let observer {
            NotificationCenter.default.removeObserver(observer)
            self.observer = nil
        }
    }

    private func handleScreenshot() async {
        guard PrivacySettingsStore.load().screenshotNotifications,
              let conversationId = activeConversationId,
              !ConversationPreferencesStore.shared.isHidden(conversationId),
              let peerDID = activePeerDID,
              let signalService else { return }
        do {
            let wire = try ConversationSignalCodec.encodeScreenshotAlert(
                to: peerDID,
                conversationId: conversationId
            )
            try await signalService.sendRaw(wire: wire)
        } catch {
            // Best-effort alert; never block the UI.
        }
    }
}
#endif
