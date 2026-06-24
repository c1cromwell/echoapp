#if os(iOS)
import Foundation

enum CallSignalType {
    static let callSignal = "call_signal"
}

enum CallSignalAction: String, Codable, Sendable {
    case offer
    case answer
    case ice
    case hangup
    case reject
    case ring
    case screenShareStart = "screen_share_start"
    case screenShareStop = "screen_share_stop"
}

struct CallSignalPayload: Codable, Sendable, Equatable {
    let callId: String
    let action: String
    let callType: String?
    let sdp: String?
    let iceCandidate: Data?

    enum CodingKeys: String, CodingKey {
        case callId = "call_id"
        case action
        case callType = "call_type"
        case sdp
        case iceCandidate = "ice_candidate"
    }

    var actionEnum: CallSignalAction? { CallSignalAction(rawValue: action) }
}

struct CallSignalEvent: Sendable, Equatable {
    let callId: String
    let peerDID: String
    let action: CallSignalAction
    let callType: CallType?
    let sdp: String?
    let iceCandidate: Data?
}

enum CallSignalCodec {
    private static let encoder = JSONEncoder()
    private static let decoder = JSONDecoder()

    static func encode(
        to peerDID: String,
        payload: CallSignalPayload,
        conversationId: String? = nil
    ) throws -> String {
        let envelope = WSEnvelope(
            type: CallSignalType.callSignal,
            to: peerDID,
            from: nil,
            conversationId: conversationId,
            payload: payload,
            timestamp: ISO8601DateFormatter().string(from: Date())
        )
        let data = try encoder.encode(envelope)
        guard let text = String(data: data, encoding: .utf8) else {
            throw ConversationSignalError.encodingFailed
        }
        return text
    }

    static func decode(from text: String) throws -> CallSignalEvent? {
        let data = Data(text.utf8)
        let header = try decoder.decode(WSEnvelopeHeader.self, from: data)
        guard header.type == CallSignalType.callSignal else { return nil }
        let envelope = try decoder.decode(WSEnvelope<CallSignalPayload>.self, from: data)
        guard let action = envelope.payload.actionEnum else { return nil }
        let callType = envelope.payload.callType.flatMap { CallType(rawValue: $0) }
        return CallSignalEvent(
            callId: envelope.payload.callId,
            peerDID: header.from ?? "",
            action: action,
            callType: callType,
            sdp: envelope.payload.sdp,
            iceCandidate: envelope.payload.iceCandidate
        )
    }
}
#endif
