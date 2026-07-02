#if os(iOS)
import BackgroundTasks
import Foundation

/// Delivers due scheduled messages via the normal relay path (WO-65).
@MainActor
enum ScheduledMessageDelivery {
    static func deliverDueMessages() async {
        let due = ScheduledMessageStore.pending()
        guard !due.isEmpty else { return }
        guard let signalService = DIContainer.shared.resolveConversationSignalService() else { return }

        let apiClient: APIClient = DIContainer.shared.resolve("networking.apiClient")
            ?? APIClient(configuration: .default)
        let textCrypto = TextMessageCrypto(identityResolve: IdentityResolveClient(apiClient: apiClient))

        for record in due {
            do {
                let payload = try await textCrypto.encryptPayload(
                    plaintext: record.plaintext,
                    peerDID: record.peerDID,
                    messageId: record.id
                )
                try await signalService.sendTextMessage(
                    conversationId: record.conversationId,
                    peerDID: record.peerDID,
                    payload: payload
                )
                ScheduledMessageStore.markDelivered(id: record.id)
            } catch {
                // Leave pending for a later BGTask attempt.
            }
        }
    }
}

/// BGProcessingTask for scheduled message delivery (WO-65 / WO-67).
enum ScheduledMessageBGTask {
    static let taskIdentifier = "com.echo.scheduled.send"

    static func register() {
        BGTaskScheduler.shared.register(forTaskWithIdentifier: taskIdentifier, using: nil) { task in
            guard let processing = task as? BGProcessingTask else {
                task.setTaskCompleted(success: false)
                return
            }
            handle(processing)
        }
    }

    @MainActor
    static func scheduleNext(fireAt: Date? = nil) {
        let request = BGProcessingTaskRequest(identifier: taskIdentifier)
        request.requiresNetworkConnectivity = true
        if let fireAt {
            request.earliestBeginDate = fireAt
        } else if let next = ScheduledMessageStore.all()
            .filter({ $0.status == .pending })
            .map(\.fireAt)
            .min() {
            request.earliestBeginDate = next
        } else {
            request.earliestBeginDate = Date(timeIntervalSinceNow: 3600)
        }
        try? BGTaskScheduler.shared.submit(request)
    }

    @MainActor
    private static func handle(_ task: BGProcessingTask) {
        scheduleNext()
        task.expirationHandler = {}
        Task {
            await ScheduledMessageDelivery.deliverDueMessages()
            task.setTaskCompleted(success: true)
        }
    }
}
#endif
