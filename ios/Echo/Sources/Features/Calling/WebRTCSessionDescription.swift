#if os(iOS)
import Foundation

/// JSON SDP envelope used by WebRTCCallSession (M4b). Real WebRTC.framework swaps the generator.
struct WebRTCSessionDescription: Codable, Sendable, Equatable {
    let type: String
    let sdp: String
    let callType: String?

    static func offer(callType: CallType, iceUfrag: String) -> WebRTCSessionDescription {
        WebRTCSessionDescription(
            type: "offer",
            sdp: "v=0\r\no=echo \(UUID().uuidString) 2 IN IP4 0.0.0.0\r\ns=echo-\(callType.rawValue)\r\n" +
                "t=0 0\r\na=ice-ufrag:\(iceUfrag)\r\na=fingerprint:sha-256 echo:phase1\r\n",
            callType: callType.rawValue
        )
    }

    static func answer(for offer: WebRTCSessionDescription) -> WebRTCSessionDescription {
        WebRTCSessionDescription(
            type: "answer",
            sdp: offer.sdp.replacingOccurrences(of: "o=echo", with: "o=echo-answer"),
            callType: offer.callType
        )
    }

    func encoded() throws -> String {
        let data = try JSONEncoder().encode(self)
        guard let text = String(data: data, encoding: .utf8) else {
            throw WebRTCCallSessionError.encodingFailed
        }
        return text
    }

    static func decode(from text: String) throws -> WebRTCSessionDescription {
        guard let data = text.data(using: .utf8) else {
            throw WebRTCCallSessionError.invalidSDP
        }
        return try JSONDecoder().decode(WebRTCSessionDescription.self, from: data)
    }
}

struct ICECandidatePayload: Codable, Sendable, Equatable {
    let candidate: String
    let sdpMid: String?
    let sdpMLineIndex: Int?

    enum CodingKeys: String, CodingKey {
        case candidate
        case sdpMid = "sdp_mid"
        case sdpMLineIndex = "sdp_mline_index"
    }

    func encoded() throws -> Data {
        try JSONEncoder().encode(self)
    }

    static func decode(from data: Data) throws -> ICECandidatePayload {
        try JSONDecoder().decode(ICECandidatePayload.self, from: data)
    }
}

enum WebRTCCallSessionError: LocalizedError {
    case encodingFailed
    case invalidSDP
    case iceUnavailable

    var errorDescription: String? {
        switch self {
        case .encodingFailed: return "Could not encode session description."
        case .invalidSDP: return "Invalid session description."
        case .iceUnavailable: return "ICE servers unavailable."
        }
    }
}
#endif
