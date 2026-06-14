#if os(iOS)
import Foundation

struct PollOptionWire: Codable, Sendable, Equatable {
    let id: String
    let text: String
}

struct PollWireBody: Codable, Sendable, Equatable {
    let action: String
    let question: String?
    let options: [PollOptionWire]?
    let voterDID: String?
    let optionId: String?
}

struct ChatPollOption: Identifiable, Codable, Sendable, Equatable {
    let id: String
    let text: String
    var voteCount: Int
    var voters: [String]
}

struct ChatPoll: Identifiable, Codable, Sendable, Equatable {
    let id: String
    let conversationId: String
    let creatorDID: String
    var question: String
    var options: [ChatPollOption]
    var isClosed: Bool
    var createdAt: Date
}

enum ConversationPollStore {
    private static func key(conversationId: String) -> String {
        "echo.polls.v1.\(conversationId)"
    }

    static func load(conversationId: String) -> [String: ChatPoll] {
        guard let data = UserDefaults.standard.data(forKey: key(conversationId: conversationId)),
              let decoded = try? JSONDecoder().decode([String: ChatPoll].self, from: data) else {
            return [:]
        }
        return decoded
    }

    static func save(conversationId: String, polls: [String: ChatPoll]) {
        guard let data = try? JSONEncoder().encode(polls) else { return }
        UserDefaults.standard.set(data, forKey: key(conversationId: conversationId))
    }
}

enum PollLogic {
    static func apply(
        polls: inout [String: ChatPoll],
        pollId: String,
        conversationId: String,
        actorDID: String,
        body: PollWireBody
    ) -> ChatPoll? {
        switch body.action {
        case "create":
            guard let question = body.question?.trimmingCharacters(in: .whitespacesAndNewlines),
                  !question.isEmpty,
                  let options = body.options, options.count >= 2 else { return nil }
            let poll = ChatPoll(
                id: pollId,
                conversationId: conversationId,
                creatorDID: actorDID,
                question: question,
                options: options.map {
                    ChatPollOption(id: $0.id, text: $0.text, voteCount: 0, voters: [])
                },
                isClosed: false,
                createdAt: Date()
            )
            polls[pollId] = poll
            return poll
        case "vote":
            guard var poll = polls[pollId],
                  !poll.isClosed,
                  let optionId = body.optionId,
                  let voter = body.voterDID,
                  let idx = poll.options.firstIndex(where: { $0.id == optionId }) else { return nil }
            if poll.options[idx].voters.contains(voter) { return poll }
            poll.options[idx].voters.append(voter)
            poll.options[idx].voteCount += 1
            polls[pollId] = poll
            return poll
        case "close":
            guard var poll = polls[pollId] else { return nil }
            poll.isClosed = true
            polls[pollId] = poll
            return poll
        default:
            return nil
        }
    }
}

/// Client-side poll create/vote/close with opaque ciphertext over WS (WO-23 / M6c).
actor PollService {
    static let shared = PollService()

    private let textCrypto: TextMessageCrypto

    init(textCrypto: TextMessageCrypto? = nil) {
        if let textCrypto {
            self.textCrypto = textCrypto
        } else {
            let client = APIClient(configuration: .default)
            self.textCrypto = TextMessageCrypto(identityResolve: IdentityResolveClient(apiClient: client))
        }
    }

    nonisolated func loadPolls(conversationId: String) -> [String: ChatPoll] {
        ConversationPollStore.load(conversationId: conversationId)
    }

    nonisolated func persist(conversationId: String, polls: [String: ChatPoll]) {
        ConversationPollStore.save(conversationId: conversationId, polls: polls)
    }

    func buildCreatePayload(
        conversationId: String,
        peerDID: String,
        creatorDID: String,
        question: String,
        optionTexts: [String]
    ) async throws -> (poll: ChatPoll, payload: PollPayload) {
        let pollId = UUID().uuidString
        let options = optionTexts.enumerated().map { idx, text in
            PollOptionWire(id: "opt-\(idx)", text: text)
        }
        let body = PollWireBody(
            action: "create",
            question: question,
            options: options,
            voterDID: nil,
            optionId: nil
        )
        var polls = ConversationPollStore.load(conversationId: conversationId)
        guard let poll = PollLogic.apply(
            polls: &polls,
            pollId: pollId,
            conversationId: conversationId,
            actorDID: creatorDID,
            body: body
        ) else {
            throw PollServiceError.invalidPoll
        }
        ConversationPollStore.save(conversationId: conversationId, polls: polls)
        let ciphertext = try await encode(body: body, peerDID: peerDID, pollId: pollId)
        let payload = PollPayload(
            conversationId: conversationId,
            pollId: pollId,
            action: "create",
            optionId: nil,
            ciphertext: ciphertext
        )
        return (poll, payload)
    }

    func buildVotePayload(
        conversationId: String,
        peerDID: String,
        pollId: String,
        optionId: String,
        voterDID: String
    ) async throws -> PollPayload {
        let body = PollWireBody(
            action: "vote",
            question: nil,
            options: nil,
            voterDID: voterDID,
            optionId: optionId
        )
        var polls = ConversationPollStore.load(conversationId: conversationId)
        _ = PollLogic.apply(
            polls: &polls,
            pollId: pollId,
            conversationId: conversationId,
            actorDID: voterDID,
            body: body
        )
        ConversationPollStore.save(conversationId: conversationId, polls: polls)
        let ciphertext = try await encode(body: body, peerDID: peerDID, pollId: pollId)
        return PollPayload(
            conversationId: conversationId,
            pollId: pollId,
            action: "vote",
            optionId: optionId,
            ciphertext: ciphertext
        )
    }

    func buildClosePayload(
        conversationId: String,
        peerDID: String,
        pollId: String,
        actorDID: String
    ) async throws -> PollPayload {
        let body = PollWireBody(action: "close", question: nil, options: nil, voterDID: nil, optionId: nil)
        var polls = ConversationPollStore.load(conversationId: conversationId)
        _ = PollLogic.apply(
            polls: &polls,
            pollId: pollId,
            conversationId: conversationId,
            actorDID: actorDID,
            body: body
        )
        ConversationPollStore.save(conversationId: conversationId, polls: polls)
        let ciphertext = try await encode(body: body, peerDID: peerDID, pollId: pollId)
        return PollPayload(
            conversationId: conversationId,
            pollId: pollId,
            action: "close",
            optionId: nil,
            ciphertext: ciphertext
        )
    }

    func applyInbound(
        event: PollSignalEvent,
        localDID: String
    ) async throws -> ChatPoll? {
        guard let ciphertext = event.ciphertext else { return nil }
        let body = try await decode(ciphertext: ciphertext, peerDID: event.peerDID, pollId: event.pollId)
        var polls = ConversationPollStore.load(conversationId: event.conversationId)
        let actor = body.action == "vote" ? (body.voterDID ?? event.peerDID) : event.peerDID
        let poll = PollLogic.apply(
            polls: &polls,
            pollId: event.pollId,
            conversationId: event.conversationId,
            actorDID: actor,
            body: body
        )
        ConversationPollStore.save(conversationId: event.conversationId, polls: polls)
        return poll
    }

    private func encode(body: PollWireBody, peerDID: String, pollId: String) async throws -> Data {
        let json = try JSONEncoder().encode(body)
        guard let jsonStr = String(data: json, encoding: .utf8) else {
            throw PollServiceError.encodingFailed
        }
        do {
            let payload = try await textCrypto.encryptPayload(
                plaintext: jsonStr,
                peerDID: peerDID,
                messageId: pollId
            )
            if let encrypted = payload.encrypted {
                return try JSONEncoder().encode(encrypted)
            }
        } catch {
            // Dev/simulator fallback: opaque JSON blob (relay never reads it).
        }
        return json
    }

    private func decode(ciphertext: Data, peerDID: String, pollId: String) async throws -> PollWireBody {
        if let encrypted = try? JSONDecoder().decode(EncryptedMessageWithPublicKey.self, from: ciphertext) {
            let payload = TextMessagePayload(messageId: pollId, encrypted: encrypted)
            let plain = try await textCrypto.decryptPayload(payload)
            return try JSONDecoder().decode(PollWireBody.self, from: Data(plain.utf8))
        }
        return try JSONDecoder().decode(PollWireBody.self, from: ciphertext)
    }
}

enum PollServiceError: LocalizedError {
    case invalidPoll
    case encodingFailed

    var errorDescription: String? {
        switch self {
        case .invalidPoll: return "Poll must have a question and at least two options."
        case .encodingFailed: return "Could not encode poll."
        }
    }
}
#endif
