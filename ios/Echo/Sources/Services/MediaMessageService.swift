#if os(iOS)
import Foundation

/// M5 media send/receive orchestration: encrypt → upload → WS reference.
actor MediaMessageService {
    private let mediaAPI: MediaAPIClient
    private let mediaCrypto: MediaMessageCrypto
    private let textCrypto: TextMessageCrypto
    private let signalService: ConversationSignalService

    init(
        mediaAPI: MediaAPIClient,
        mediaCrypto: MediaMessageCrypto,
        textCrypto: TextMessageCrypto,
        signalService: ConversationSignalService
    ) {
        self.mediaAPI = mediaAPI
        self.mediaCrypto = mediaCrypto
        self.textCrypto = textCrypto
        self.signalService = signalService
    }

    func sendMedia(
        data: Data,
        mimeType: String,
        mediaKind: MediaKind,
        caption: String?,
        messageId: String,
        peerDID: String,
        conversationId: String,
        trustTier: Int = CurrentUserSession.trustTier(),
        waveformBars: [Float]? = nil
    ) async throws {
        let encryptedFile = try await mediaCrypto.encryptFile(data, peerDID: peerDID)
        let upload = try await mediaAPI.uploadEncrypted(
            data: encryptedFile,
            mimeType: mimeType,
            trustTier: trustTier
        )
        if mediaKind == .image || mediaKind == .video {
            _ = MediaThumbnailCache.thumbnail(for: upload.fileId, from: data)
        }
        let ref = MediaAttachmentRef(
            fileId: upload.fileId,
            mimeType: mimeType,
            mediaKind: mediaKind.rawValue,
            byteSize: data.count,
            chunkCount: upload.chunkCount,
            caption: caption,
            waveformBars: waveformBars
        )
        let wire = MediaMessageWire(messageId: messageId, media: ref, caption: caption)
        let plaintext = String(decoding: try JSONEncoder().encode(wire), as: UTF8.self)
        let payload = try await textCrypto.encryptPayload(
            plaintext: plaintext,
            peerDID: peerDID,
            messageId: messageId
        )
        try await signalService.sendTextMessage(
            conversationId: conversationId,
            peerDID: peerDID,
            payload: payload
        )
    }

    func uploadGroupMedia(
        data: Data,
        mimeType: String,
        mediaKind: MediaKind,
        trustTier: Int = CurrentUserSession.trustTier()
    ) async throws -> MediaAttachmentRef {
        let encryptedFile = try await mediaCrypto.encryptFileSymmetric(data)
        let upload = try await mediaAPI.uploadEncrypted(
            data: encryptedFile,
            mimeType: mimeType,
            trustTier: trustTier
        )
        if mediaKind == .image || mediaKind == .video {
            _ = MediaThumbnailCache.thumbnail(for: upload.fileId, from: data)
        }
        return MediaAttachmentRef(
            fileId: upload.fileId,
            mimeType: mimeType,
            mediaKind: mediaKind.rawValue,
            byteSize: data.count,
            chunkCount: upload.chunkCount,
            caption: nil
        )
    }

    func downloadAndDecrypt(ref: MediaAttachmentRef, peerDID: String) async throws -> Data {
        let encrypted = try await mediaAPI.downloadChunks(fileId: ref.fileId, chunkCount: ref.chunkCount)
        return try await mediaCrypto.decryptFile(encrypted)
    }

    static func parseWire(from decryptedText: String) -> MediaMessageWire? {
        guard let data = decryptedText.data(using: .utf8) else { return nil }
        return try? JSONDecoder().decode(MediaMessageWire.self, from: data)
    }
}
#endif
