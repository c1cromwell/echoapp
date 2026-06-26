import XCTest
@testable import Echo

#if os(iOS)
final class MediaMessageWireTests: XCTestCase {
    func testVoiceNoteCodec_normalizesWaveform() {
        let samples: [CGFloat] = [0.1, 0.5, 0.9, 0.3]
        let bars = VoiceNoteCodec.normalizedWaveform(samples, targetCount: 4)
        XCTAssertEqual(bars.count, 4)
        XCTAssertEqual(VoiceNoteCodec.wireMimeType, "audio/mp4")
    }

    func testWireRoundTrip() throws {
        let wire = MediaMessageWire(
            messageId: "m-media-1",
            media: MediaAttachmentRef(
                fileId: "file-abc",
                mimeType: "image/jpeg",
                mediaKind: MediaKind.image.rawValue,
                byteSize: 1024,
                chunkCount: 1,
                caption: "Sunset"
            ),
            caption: "Sunset"
        )
        let json = String(decoding: try JSONEncoder().encode(wire), as: UTF8.self)
        let parsed = MediaMessageService.parseWire(from: json)
        XCTAssertEqual(parsed, wire)
    }

    func testVoiceNoteWireKind() {
        let ref = MediaAttachmentRef(
            fileId: "voice-1",
            mimeType: VoiceNoteCodec.wireMimeType,
            mediaKind: MediaKind.audio.rawValue,
            byteSize: 4096,
            chunkCount: 1,
            caption: "Note",
            waveformBars: [0.2, 0.8, 0.5]
        )
        XCTAssertEqual(ref.mediaKind, "audio")
        XCTAssertEqual(ref.waveformBars?.count, 3)
    }
}
#endif
