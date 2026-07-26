import UserNotifications

/// Notification Service Extension skeleton (Signal Parity Wave S4).
/// Add this file to an `EchoNotificationService` app-extension target in Xcode.
/// Production: decrypt Echo push envelopes using shared Keychain key material;
/// fall back to a generic title when keys are unavailable (device locked).
final class NotificationService: UNNotificationServiceExtension {
    private var contentHandler: ((UNNotificationContent) -> Void)?
    private var bestAttempt: UNMutableNotificationContent?

    override func didReceive(
        _ request: UNNotificationRequest,
        withContentHandler contentHandler: @escaping (UNNotificationContent) -> Void
    ) {
        self.contentHandler = contentHandler
        bestAttempt = (request.content.mutableCopy() as? UNMutableNotificationContent)

        guard let bestAttempt else {
            contentHandler(request.content)
            return
        }

        // Placeholder enrichment — replace with Echo envelope decrypt when Keychain
        // access group + ratchet/Kinnami stores are available to the extension.
        if bestAttempt.title.isEmpty {
            bestAttempt.title = "Echo"
        }
        if bestAttempt.body.isEmpty {
            bestAttempt.body = "New message"
        }
        contentHandler(bestAttempt)
    }

    override func serviceExtensionTimeWillExpire() {
        if let contentHandler, let bestAttempt {
            contentHandler(bestAttempt)
        }
    }
}
