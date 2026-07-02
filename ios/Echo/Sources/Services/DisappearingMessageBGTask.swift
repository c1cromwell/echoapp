#if os(iOS)
import BackgroundTasks
import Foundation

/// Registers background purge for expired disappearing messages (WO-105).
enum DisappearingMessageBGTask {
    static let taskIdentifier = "com.echo.disappearing.purge"

    static func register() {
        BGTaskScheduler.shared.register(forTaskWithIdentifier: taskIdentifier, using: nil) { task in
            guard let refresh = task as? BGAppRefreshTask else {
                task.setTaskCompleted(success: false)
                return
            }
            handle(refresh)
        }
    }

    static func scheduleNext() {
        let request = BGAppRefreshTaskRequest(identifier: taskIdentifier)
        request.earliestBeginDate = Date(timeIntervalSinceNow: 15 * 60)
        try? BGTaskScheduler.shared.submit(request)
    }

    @MainActor
    private static func handle(_ task: BGAppRefreshTask) {
        scheduleNext()
        task.expirationHandler = {}

        let conversationIds = ConversationThreadStore.allStoredConversationIds()
        for conversationId in conversationIds {
            let ttl = ConversationPreferencesStore.shared.preferences(for: conversationId).disappearing.seconds
            _ = DisappearingMessageEnforcer.purgeExpired(conversationId: conversationId, ttlSeconds: ttl)
        }
        task.setTaskCompleted(success: true)
    }
}
#endif
