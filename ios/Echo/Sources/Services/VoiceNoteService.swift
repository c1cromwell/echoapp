#if os(iOS)
import Foundation

/// Voice-note send orchestration (WO-SX5). Wraps media upload + encrypted chat reference.
struct VoiceNoteService: Sendable {
    private let mediaService: MediaMessageService

    init(mediaService: MediaMessageService) {
        self.mediaService = mediaService
    }

    func send(
        audioData: Data,
        waveformBars: [Float],
        caption: String,
        messageId: String,
        peerDID: String,
        conversationId: String
    ) async throws {
        try await mediaService.sendMedia(
            data: audioData,
            mimeType: VoiceNoteCodec.wireMimeType,
            mediaKind: .audio,
            caption: caption.isEmpty ? nil : caption,
            messageId: messageId,
            peerDID: peerDID,
            conversationId: conversationId,
            waveformBars: waveformBars.isEmpty ? nil : waveformBars
        )
    }
}
#endif
