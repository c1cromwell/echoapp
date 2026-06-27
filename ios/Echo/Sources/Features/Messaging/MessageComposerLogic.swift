import Foundation

/// Reply, edit window, and forward formatting for chat composer (Wave 0 / Week B).
enum MessageComposerLogic {
    static let editWindowSeconds: TimeInterval = 15 * 60

    static func canEdit(sentAt: Date?, isOwnMessage: Bool) -> Bool {
        guard isOwnMessage, let sentAt else { return false }
        return Date().timeIntervalSince(sentAt) <= editWindowSeconds
    }

    static func replyPreview(authorName: String, content: String) -> String {
        let name = authorName.trimmingCharacters(in: .whitespacesAndNewlines)
        let snippet = content.trimmingCharacters(in: .whitespacesAndNewlines)
        let line = snippet.count > 80 ? String(snippet.prefix(80)) + "…" : snippet
        if name.isEmpty { return line }
        return "\(name): \(line)"
    }

    static func forwardBody(_ content: String) -> String {
        let trimmed = content.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? "" : "↪ \(trimmed)"
    }
}

#if os(iOS)
@MainActor
enum ConversationPinnedMessageStore {
    private static let keyPrefix = "echo.pinnedMessage.v1."

    static func pinnedMessageId(conversationId: String) -> String? {
        guard !conversationId.isEmpty else { return nil }
        return UserDefaults.standard.string(forKey: keyPrefix + conversationId)
    }

    static func setPinnedMessageId(_ messageId: String?, conversationId: String) {
        guard !conversationId.isEmpty else { return }
        let key = keyPrefix + conversationId
        if let messageId, !messageId.isEmpty {
            UserDefaults.standard.set(messageId, forKey: key)
        } else {
            UserDefaults.standard.removeObject(forKey: key)
        }
    }

    static func togglePin(messageId: String, conversationId: String) -> Bool {
        if pinnedMessageId(conversationId: conversationId) == messageId {
            setPinnedMessageId(nil, conversationId: conversationId)
            return false
        }
        setPinnedMessageId(messageId, conversationId: conversationId)
        return true
    }
}

@MainActor
enum MessageForwarder {
    static func forward(
        content: String,
        to conversation: StoredConversation,
        fromMessageId: String? = nil,
        fromConversationId: String? = nil
    ) async {
        let body = MessageComposerLogic.forwardBody(content)
        guard !body.isEmpty,
              let senderDID = await CurrentUserSession.currentDID(),
              !senderDID.isEmpty else { return }

        let message = ChatDetailMessage(
            id: UUID().uuidString,
            senderDID: senderDID,
            currentUserDID: senderDID,
            content: body,
            timestamp: "Now",
            deliveryStatus: .sending,
            sentAt: Date()
        )
        ConversationThreadStore.appendIfNew(conversationId: conversation.id, message: message)
        ConversationStore.shared.appendMessagePreview(conversationId: conversation.id, preview: body)

        guard let service = DIContainer.shared.resolveConversationSignalService(),
              let token = try? await KeychainManager.shared.getAuthToken() else { return }

        do {
            try await service.connect(accessToken: token)
            let client = DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
            let crypto = TextMessageCrypto(identityResolve: IdentityResolveClient(apiClient: client))
            let wirePlain = try ChatMessageEnvelope(
                body: body,
                forwardedFromMessageId: fromMessageId,
                forwardedFromConversationId: fromConversationId
            ).serialized()
            let payload: TextMessagePayload
            do {
                payload = try await crypto.encryptPayload(
                    plaintext: wirePlain,
                    peerDID: conversation.peerDID,
                    messageId: message.id
                )
            } catch {
                payload = TextMessagePayload(messageId: message.id, text: body, encrypted: nil)
            }
            if SealedSenderPolicy.shouldUseSealed(peerDID: conversation.peerDID, conversationId: conversation.id),
               !conversation.id.hasPrefix("group:") {
                let client = DIContainer.shared.resolveAPIClient() ?? APIClient(configuration: .default)
                let sealedService = SealedSenderService(
                    identityResolve: IdentityResolveClient(apiClient: client),
                    apiClient: client
                )
                let sealed = try await sealedService.wrap(
                    payload: payload,
                    recipientDID: conversation.peerDID,
                    senderDID: senderDID
                )
                try await service.sendSealedTextMessage(
                    conversationId: conversation.id,
                    peerDID: conversation.peerDID,
                    payload: sealed
                )
            } else {
                try await service.sendTextMessage(
                    conversationId: conversation.id,
                    peerDID: conversation.peerDID,
                    payload: payload
                )
            }
            if fromMessageId != nil {
                let ops = DIContainer.shared.resolveMessageOpsAPI()
                try? await ops?.putMessageRefs(
                    messageId: message.id,
                    conversationId: conversation.id,
                    replyToMessageId: nil,
                    forwardedFromMessageId: fromMessageId,
                    forwardedFromConversationId: fromConversationId
                )
            }
        } catch {
            // Forward still visible locally if relay fails.
        }
    }
}
#endif
