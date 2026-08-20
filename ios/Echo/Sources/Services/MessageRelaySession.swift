#if os(iOS)
import Foundation

/// Shared WS reconnect used by the Messages tab and by APNs wake-up.
/// Opening the socket triggers server `flushOffline`, so queued messages land
/// even if the recipient was not in the app when they were sent.
enum MessageRelaySession {
    @MainActor
    static func connect() async {
        guard let service = DIContainer.shared.resolveConversationSignalService() else { return }
        let token: String?
        if let asyncToken = try? await KeychainManager.shared.getAuthToken() {
            token = asyncToken
        } else {
            token = nil
        }
        guard let token, !token.isEmpty else { return }

        service.setInboundTextHandler { event in
            Task { @MainActor in
                await ingestInbound(event)
            }
        }

        try? await service.connect(accessToken: token)
        await IncomingCallPresenter.shared.configureIfNeeded()
        ScreenshotAlertService.shared.configure(signalService: service)
        ScreenshotAlertService.shared.startMonitoring()
        await PushRegistrationService.shared.registerIfNeeded()
    }

    @MainActor
    private static func ingestInbound(_ event: TextMessageSignalEvent) async {
        let resolved = await InboundTextMessageResolver.resolveBody(for: event)
        let store = ConversationStore.shared
        let match = store.conversations.first(where: { $0.id == event.conversationId })
            ?? store.conversations.first(where: { $0.peerDID == event.peerDID })
        let conversation: StoredConversation
        if let match {
            conversation = match
        } else if let created = await ContactThreadHelper.upsertDirectThread(
            peerDID: event.peerDID,
            displayName: event.peerDID
        ) {
            conversation = created
        } else {
            return
        }

        let currentDID = await CurrentUserSession.currentDID() ?? ""
        guard !resolved.senderDID.isEmpty, resolved.senderDID != currentDID else { return }

        let inbound = ChatDetailMessage(
            id: event.messageId,
            senderDID: resolved.senderDID.isEmpty ? event.peerDID : resolved.senderDID,
            currentUserDID: currentDID,
            content: resolved.body,
            timestamp: "Now",
            deliveryStatus: .delivered
        )
        ConversationThreadStore.appendIfNew(conversationId: conversation.id, message: inbound)
        let session = HiddenChatsSession.shared
        guard session.shouldSurfaceNotification(for: conversation.id) else { return }
        let preview = session.redactedPreviewIfNeeded(
            for: conversation.id,
            resolved: resolved.preview
        )
        store.appendMessagePreview(conversationId: conversation.id, preview: preview)
        if ActiveChatRegistry.openConversationId != conversation.id {
            store.incrementUnread(conversationId: conversation.id)
        }
    }
}
#endif
