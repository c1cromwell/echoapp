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
    /// Wave S3 — group Phase 3 signal parity with 1:1 chat.
    var peerIsTyping = false
    var reactionByMessage: [String: [String]] = [:]

    private let keyManager: GroupKeyManager
    private let senderKeys: GroupSenderKeyStore
    private let signalService: ConversationSignalService
    private let keyDistribution: GroupKeyDistributionService
    private var typingClearTask: Task<Void, Never>?

    init(
        groupId: String,
        groupName: String,
        currentUserDID: String,
        keyManager: GroupKeyManager,
        senderKeys: GroupSenderKeyStore,
        signalService: ConversationSignalService,
        keyDistribution: GroupKeyDistributionService
    ) {
        self.groupId = groupId
        self.groupName = groupName
        self.conversationId = "group:\(groupId)"
        self.currentUserDID = currentUserDID
        self.keyManager = keyManager
        self.senderKeys = senderKeys
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

    func onInputChanged(_ text: String) {
        Task {
            // Broadcast typing to group conversation (peerDID unused for group fanout on server).
            try? await signalService.sendTyping(
                conversationId: conversationId,
                peerDID: currentUserDID,
                state: text.isEmpty ? .stop : .start
            )
        }
    }

    func toggleReaction(messageId: String, emoji: String) async {
        var current = reactionByMessage[messageId] ?? []
        if current.contains(emoji) {
            current.removeAll { $0 == emoji }
        } else {
            current.append(emoji)
        }
        reactionByMessage[messageId] = current
        // Group fanout uses conversationId; peerDID required by envelope contract.
        try? await signalService.sendReaction(
            conversationId: conversationId,
            peerDID: currentUserDID,
            messageId: messageId,
            emoji: emoji
        )
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
            let (ciphertext, keyVersion, iteration) = try await encryptForGroup(plaintext)
            let payload = TextMessagePayload(
                messageId: messageId,
                groupCiphertext: ciphertext,
                groupKeyVersion: keyVersion,
                groupSenderKeyIteration: iteration,
                senderDID: currentUserDID
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

    func sendVoiceNote(from recorder: VoiceNoteRecorder, caption: String = "") async {
        guard let data = try? recorder.stopRecording() else { return }
        await sendGroupMedia(data: data, mimeType: VoiceNoteCodec.wireMimeType, mediaKind: .audio)
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
            let (ciphertext, keyVersion, iteration) = try await encryptForGroup(plaintext)
            let payload = TextMessagePayload(
                messageId: messageId,
                groupCiphertext: ciphertext,
                groupKeyVersion: keyVersion,
                groupSenderKeyIteration: iteration,
                senderDID: currentUserDID,
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
        switch event {
        case .typing(let typingEvent):
            guard typingEvent.conversationId == conversationId else { return }
            guard typingEvent.peerDID != currentUserDID else { return }
            peerIsTyping = typingEvent.state == .start
            typingClearTask?.cancel()
            if peerIsTyping {
                typingClearTask = Task { @MainActor in
                    try? await Task.sleep(nanoseconds: 6_000_000_000)
                    peerIsTyping = false
                }
            }
            return
        case .reaction(let reactionEvent):
            guard reactionEvent.conversationId == conversationId else { return }
            var current = reactionByMessage[reactionEvent.messageId] ?? []
            if !current.contains(reactionEvent.emoji) {
                current.append(reactionEvent.emoji)
                reactionByMessage[reactionEvent.messageId] = current
            }
            return
        case .textMessage:
            break // handled below
        default:
            return
        }

        guard case .textMessage(let textEvent) = event else { return }
        guard textEvent.conversationId == conversationId else { return }
        guard textEvent.peerDID != currentUserDID else { return }

        let body: String
        var mediaRef: MediaAttachmentRef?
        if let wire = textEvent.wirePayload,
           let ciphertext = wire.groupCiphertext,
           let version = wire.groupKeyVersion {
            do {
                let plain: Data
                if let iteration = wire.groupSenderKeyIteration {
                    let senderDID = wire.senderDID ?? textEvent.peerDID
                    if !(await senderKeys.hasChain(
                        groupId: groupId,
                        senderDID: senderDID,
                        groupKeyVersion: version
                    )), let rootKey = await keyManager.key(groupId: groupId, version: version)?.key {
                        await senderKeys.seedChain(
                            groupId: groupId,
                            senderDID: senderDID,
                            groupKeyVersion: version,
                            rootKey: rootKey
                        )
                    }
                    plain = try await senderKeys.decrypt(
                        ciphertext: ciphertext,
                        groupId: groupId,
                        senderDID: senderDID,
                        groupKeyVersion: version,
                        iteration: iteration
                    )
                } else {
                    // Compatibility with payloads sent before Wave S2 sender-key rollout.
                    plain = try await keyManager.decryptFromGroup(
                        ciphertext: ciphertext,
                        groupId: groupId,
                        keyVersion: version
                    )
                }
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

    /// Uses a sender chain when the current group root is available. Group-key encryption
    /// remains the compatibility fallback for members that have not received a chain seed.
    private func encryptForGroup(_ plaintext: Data) async throws -> (Data, Int?, UInt32?) {
        guard let keyInfo = await keyManager.latestKey(groupId: groupId) else {
            throw GroupKeyError.noGroupKey
        }
        if !(await senderKeys.hasChain(
            groupId: groupId,
            senderDID: currentUserDID,
            groupKeyVersion: keyInfo.version
        )) {
            await senderKeys.seedChain(
                groupId: groupId,
                senderDID: currentUserDID,
                groupKeyVersion: keyInfo.version,
                rootKey: keyInfo.key
            )
        }
        do {
            let encrypted = try await senderKeys.encrypt(
                plaintext: plaintext,
                groupId: groupId,
                senderDID: currentUserDID,
                groupKeyVersion: keyInfo.version
            )
            return (encrypted.ciphertext, keyInfo.version, encrypted.iteration)
        } catch GroupSenderKeyError.missingChain {
            return (
                try await keyManager.encryptForGroup(plaintext: plaintext, groupId: groupId),
                keyInfo.version,
                nil
            )
        }
    }
}
#endif
