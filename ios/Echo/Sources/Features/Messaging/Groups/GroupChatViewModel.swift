#if os(iOS)
import Foundation

@MainActor
@Observable
final class GroupChatViewModel {
    struct GroupMessage: Identifiable, Equatable {
        let id: String
        let senderDID: String
        let text: String
        let isOutgoing: Bool
        var mediaRef: MediaAttachmentRef?
    }

    let groupId: String
    let conversationId: String
    let currentUserDID: String
    var groupName: String

    var messages: [GroupMessage] = []
    var composerText = ""
    var errorMessage: String?

    private let keyManager: GroupKeyManager
    private let signalService: ConversationSignalService
    private let keyDistribution: GroupKeyDistributionService

    init(
        groupId: String,
        groupName: String,
        currentUserDID: String,
        keyManager: GroupKeyManager,
        signalService: ConversationSignalService,
        keyDistribution: GroupKeyDistributionService
    ) {
        self.groupId = groupId
        self.groupName = groupName
        self.conversationId = "group:\(groupId)"
        self.currentUserDID = currentUserDID
        self.keyManager = keyManager
        self.signalService = signalService
        self.keyDistribution = keyDistribution
    }

    func configure(accessToken: String) async {
        signalService.setGroupKeyHandler { [weak self] event in
            guard event.groupId == self?.groupId else { return }
            Task { @MainActor in
                try? await self?.keyDistribution.acceptInboundKey(
                    groupId: event.groupId,
                    version: event.version,
                    encryptedPackage: event.encryptedKey
                )
            }
        }

        signalService.setConversationHandler(conversationId: conversationId) { [weak self] event in
            guard let self else { return }
            Task { @MainActor in
                await self.handleSignal(event)
            }
        }

        try? await signalService.connect(accessToken: accessToken)
    }

    func sendMessage() async {
        let text = composerText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }
        guard await keyManager.hasKey(groupId: groupId) else {
            errorMessage = "Waiting for group encryption key…"
            return
        }
        composerText = ""

        let messageId = UUID().uuidString
        let optimistic = GroupMessage(
            id: messageId,
            senderDID: currentUserDID,
            text: text,
            isOutgoing: true
        )
        messages.append(optimistic)

        do {
            let plaintext = Data(text.utf8)
            let ciphertext = try await keyManager.encryptForGroup(plaintext: plaintext, groupId: groupId)
            let keyVersion = await keyManager.latestKeyVersion(groupId: groupId)
            let payload = TextMessagePayload(
                messageId: messageId,
                groupCiphertext: ciphertext,
                groupKeyVersion: keyVersion
            )
            try await signalService.sendGroupTextMessage(
                conversationId: conversationId,
                payload: payload
            )
        } catch {
            errorMessage = error.localizedDescription
            messages.removeAll { $0.id == messageId }
        }
    }

    func sendGroupMedia(data: Data, mimeType: String, mediaKind: MediaKind) async {
        guard !data.isEmpty else { return }
        guard await keyManager.hasKey(groupId: groupId) else {
            errorMessage = "Waiting for group encryption key…"
            return
        }

        let messageId = UUID().uuidString
        let placeholder = TextMessagePayload.mediaPlaceholder(for: MediaAttachmentRef(
            fileId: "pending", mimeType: mimeType, mediaKind: mediaKind.rawValue,
            byteSize: data.count, chunkCount: 0, caption: nil
        ))
        let optimistic = GroupMessage(
            id: messageId, senderDID: currentUserDID, text: placeholder, isOutgoing: true, mediaRef: nil
        )
        messages.append(optimistic)

        do {
            guard let mediaService = DIContainer.shared.resolveMediaMessage() else {
                throw ECHOError.messageSendFailed
            }
            let ref = try await mediaService.uploadGroupMedia(
                data: data, mimeType: mimeType, mediaKind: mediaKind
            )
            if let idx = messages.firstIndex(where: { $0.id == messageId }) {
                messages[idx] = GroupMessage(
                    id: messageId,
                    senderDID: currentUserDID,
                    text: TextMessagePayload.mediaPlaceholder(for: ref),
                    isOutgoing: true,
                    mediaRef: ref
                )
            }
            let wire = MediaMessageWire(messageId: messageId, media: ref, caption: nil)
            let plaintext = try JSONEncoder().encode(wire)
            let ciphertext = try await keyManager.encryptForGroup(plaintext: plaintext, groupId: groupId)
            let keyVersion = await keyManager.latestKeyVersion(groupId: groupId)
            let payload = TextMessagePayload(
                messageId: messageId,
                groupCiphertext: ciphertext,
                groupKeyVersion: keyVersion,
                media: ref
            )
            try await signalService.sendGroupTextMessage(
                conversationId: conversationId,
                payload: payload
            )
        } catch {
            errorMessage = error.localizedDescription
            messages.removeAll { $0.id == messageId }
        }
    }

    private func handleSignal(_ event: ConversationSignalEvent) async {
        guard case .textMessage(let textEvent) = event else { return }
        guard textEvent.conversationId == conversationId else { return }
        guard textEvent.peerDID != currentUserDID else { return }

        let body: String
        var mediaRef: MediaAttachmentRef?
        if let wire = textEvent.wirePayload,
           let ciphertext = wire.groupCiphertext,
           let version = wire.groupKeyVersion {
            do {
                let plain = try await keyManager.decryptFromGroup(
                    ciphertext: ciphertext,
                    groupId: groupId,
                    keyVersion: version
                )
                if let parsed = MediaMessageService.parseWire(from: String(data: plain, encoding: .utf8) ?? "") {
                    body = parsed.caption ?? TextMessagePayload.mediaPlaceholder(for: parsed.media)
                    mediaRef = parsed.media
                } else {
                    body = String(data: plain, encoding: .utf8) ?? textEvent.text
                }
            } catch {
                body = textEvent.text
            }
        } else if let wire = textEvent.wirePayload, let media = wire.media {
            body = TextMessagePayload.mediaPlaceholder(for: media)
            mediaRef = media
        } else {
            body = textEvent.text
        }

        guard !messages.contains(where: { $0.id == textEvent.messageId }) else { return }
        messages.append(GroupMessage(
            id: textEvent.messageId,
            senderDID: textEvent.peerDID,
            text: body,
            isOutgoing: false,
            mediaRef: mediaRef
        ))
    }
}
#endif
