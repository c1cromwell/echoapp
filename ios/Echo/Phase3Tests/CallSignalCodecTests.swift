import XCTest
@testable import Echo

#if os(iOS)
final class CallSignalCodecTests: XCTestCase {
    func testOfferEncodeDecode() throws {
        let wire = try CallSignalCodec.encode(
            to: "did:key:bob",
            payload: CallSignalPayload(
                callId: "call-1",
                action: CallSignalAction.offer.rawValue,
                callType: CallType.voice.rawValue,
                sdp: "v=0",
                iceCandidate: nil
            )
        )
        let event = try XCTUnwrap(try CallSignalCodec.decode(from: wire))
        XCTAssertEqual(event.callId, "call-1")
        XCTAssertEqual(event.action, .offer)
        XCTAssertEqual(event.callType, .voice)
        XCTAssertEqual(event.sdp, "v=0")
    }
}
#endif
