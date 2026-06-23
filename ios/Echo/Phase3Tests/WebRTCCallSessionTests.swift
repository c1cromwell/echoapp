import Foundation
import XCTest
@testable import Echo

#if os(iOS)
@MainActor
final class WebRTCCallSessionTests: XCTestCase {
    func testStubEngineRoundTrip() async throws {
        let session = WebRTCCallSession(engine: WebRTCStubCallEngine())
        session.configure(iceServers: [], callType: .voice, onIceCandidate: { _ in }, onConnected: {})
        let offer = try await session.createOffer(callType: .voice)
        let answer = try await session.applyRemoteOffer(offer, callType: .voice)
        try await session.applyAnswer(answer)
        XCTAssertFalse(session.usesLiveWebRTC)
    }
}
#endif
