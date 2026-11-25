import XCTest
@testable import Echo

final class ConversationSignalCodecTests: XCTestCase {

    func testTypingEnvelope_usesSnakeCaseKeys() throws {
        let json = try ConversationSignalCodec.encodeTyping(
            to: "did:key:peer",
            conversationId: "conv-1",
            state: .start
        )
        XCTAssertTrue(json.contains("\"conversation_id\""))
        XCTAssertTrue(json.contains("\"type\":\"typing\""))
        XCTAssertTrue(json.contains("\"to\":\"did:key:peer\""))
        XCTAssertTrue(json.contains("\"state\":\"start\""))
    }

    func testReadReceiptEnvelope_roundTrip() throws {
        let json = try ConversationSignalCodec.encodeReadReceipt(
            to: "did:key:alice",
            conversationId: "conv-2",
            messageIds: ["m1", "m2"],
            readAt: "2026-05-25T12:00:00Z"
        )
        let event = try XCTUnwrap(ConversationSignalCodec.decodeEvent(from: json))
        guard case .readReceipt(let e) = event else {
            return XCTFail("expected read receipt")
        }
        XCTAssertEqual(e.conversationId, "conv-2")
        XCTAssertEqual(e.messageIds, ["m1", "m2"])
        XCTAssertEqual(e.readAt, "2026-05-25T12:00:00Z")
    }

    func testReactionEnvelope_roundTrip() throws {
        let json = try ConversationSignalCodec.encodeReaction(
            to: "did:key:bob",
            conversationId: "conv-3",
            messageId: "msg-9",
            emoji: "👍"
        )
        let event = try XCTUnwrap(ConversationSignalCodec.decodeEvent(from: json))
        guard case .reaction(let e) = event else {
            return XCTFail("expected reaction")
        }
        XCTAssertEqual(e.messageId, "msg-9")
        XCTAssertEqual(e.emoji, "👍")
    }

    func testTextMessageEnvelope_roundTrip() throws {
        let json = try ConversationSignalCodec.encodeTextMessage(
            to: "did:key:bob",
            conversationId: "conv-1",
            payload: TextMessagePayload(messageId: "m1", text: "Hello", encrypted: nil)
        )
        let event = try XCTUnwrap(ConversationSignalCodec.decodeEvent(from: json))
        guard case .textMessage(let e) = event else {
            return XCTFail("expected text")
        }
        XCTAssertEqual(e.conversationId, "conv-1")
        XCTAssertEqual(e.messageId, "m1")
        XCTAssertEqual(e.text, "Hello")
    }

    func testTextMessageEnvelope_silentFlag() throws {
        let json = try ConversationSignalCodec.encodeTextMessage(
            to: "did:key:bob",
            conversationId: "conv-1",
            payload: TextMessagePayload(messageId: "m1", text: "Quiet", encrypted: nil),
            silent: true
        )
        let data = try XCTUnwrap(json.data(using: .utf8))
        let envelope = try JSONDecoder().decode(WSEnvelope<TextMessagePayload>.self, from: data)
        XCTAssertEqual(envelope.silent, true)
    }

    func testDecodeEvent_unknownTypeReturnsNil() throws {
        let json = """
        {"type":"presence","to":"did:key:x","payload":{}}
        """
        XCTAssertNil(try ConversationSignalCodec.decodeEvent(from: json))
    }

    func testTypingDecode_fromFieldBecomesPeerDID() throws {
        let json = """
        {"type":"typing","from":"did:key:peer","to":"did:key:me","conversation_id":"c1","payload":{"conversation_id":"c1","state":"stop"}}
        """
        let event = try XCTUnwrap(ConversationSignalCodec.decodeEvent(from: json))
        guard case .typing(let e) = event else {
            return XCTFail("expected typing")
        }
        XCTAssertEqual(e.peerDID, "did:key:peer")
        XCTAssertEqual(e.state, .stop)
    }
}

#if os(iOS)
final class SealedDeliveryTests: XCTestCase {
    private let alice = "did:key:zAlice"
    private let bob = "did:key:zBob"
    private let keyHex = String(repeating: "ab", count: 32)

    // The request that carries a sealed message must contain nothing that says who sent it.
    func testDeliveryRequestCarriesNoIdentity() throws {
        let payload = SealedTextPayload(deliveryToken: nil, ciphertext: Data([9, 9, 9]))
        let request = try SealedDelivery.request(
            baseURL: URL(string: "https://relay.example")!,
            deliveryKey: keyHex,
            payload: payload,
            messageId: "m-1",
            commitmentHash: String(repeating: "cd", count: 32),
            silent: false
        )
        XCTAssertEqual(request.url?.absoluteString, "https://relay.example/v1/sealed/deliver")
        XCTAssertEqual(request.httpMethod, "POST")
        XCTAssertEqual(request.allHTTPHeaderFields ?? [:], ["Content-Type": "application/json"])
        XCTAssertFalse(request.httpShouldHandleCookies)
        let body = try XCTUnwrap(String(data: try XCTUnwrap(request.httpBody), encoding: .utf8))
        XCTAssertFalse(body.contains("did:"), "the body names a DID: \(body)")
        XCTAssertFalse(body.contains("delivery_token"), "the body carries a server-issued token: \(body)")
        XCTAssertFalse(body.contains("conversation"), "the body names a conversation: \(body)")
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: try XCTUnwrap(request.httpBody)) as? [String: Any])
        XCTAssertEqual(Set(object.keys), ["delivery_key", "payload", "message_id", "commitment_hash"])
    }

    // On the first-contact path the socket envelope carries a constant, never the pair.
    func testSealedWireConversationIdNamesNoOne() throws {
        let wire = try ConversationSignalCodec.encodeSealedTextMessage(
            to: bob,
            conversationId: ConversationID.sealedWire,
            payload: SealedTextPayload(deliveryToken: "tok", ciphertext: Data([1]))
        )
        XCTAssertFalse(wire.contains(alice))
        XCTAssertTrue(wire.contains("\"conversation_id\":\"dm:sealed\""))
        XCTAssertFalse(ConversationID.sealedWire.contains("did:"))
    }

    // What the server sends for a sealed delivery decodes, with no sender and no token.
    func testServerSealedEnvelopeDecodes() throws {
        let frame = #"{"type":"sealed_text","to":"did:key:zBob","conversation_id":"dm:sealed","payload":{"ciphertext":"AQID"},"timestamp":"2026-11-24T00:00:00Z"}"#
        let event = try XCTUnwrap(try ConversationSignalCodec.decodeEvent(from: frame))
        guard case .textMessage(let text) = event else { return XCTFail("expected a message") }
        XCTAssertEqual(text.sealedPayload, SealedTextPayload(deliveryToken: nil, ciphertext: Data([1, 2, 3])))
        XCTAssertTrue(text.peerDID.isEmpty)
        XCTAssertEqual(text.conversationId, "dm:sealed")
    }

    // The receiving device works out the thread from the sender inside the encryption, and both
    // devices arrive at the same local thread name.
    func testRecipientDerivesTheThreadLocally() throws {
        let inner = TextMessagePayload(messageId: "m-42", text: "hi")
        let opened = OpenedSealedMessage(senderDID: alice, payload: inner, senderDeliveryKey: keyHex)
        let event = try XCTUnwrap(SealedInbound.event(from: opened, localDID: bob, fallbackMessageId: "x"))
        XCTAssertEqual(event.peerDID, alice)
        XCTAssertEqual(event.messageId, "m-42")
        XCTAssertNil(event.sealedPayload)
        XCTAssertEqual(event.conversationId, ConversationID.direct(localDID: alice, peerDID: bob))
        XCTAssertNil(SealedInbound.event(from: opened, localDID: alice, fallbackMessageId: "x"), "a message claiming to be from ourselves")
        XCTAssertNil(SealedInbound.event(from: opened, localDID: "", fallbackMessageId: "x"))
    }

    func testRouteChoice() {
        XCTAssertEqual(DirectMessageSender.route(sealedAllowed: true, isGroup: false, peerKey: keyHex), .sealedDelivery(key: keyHex))
        XCTAssertEqual(DirectMessageSender.route(sealedAllowed: true, isGroup: false, peerKey: nil), .sealedOverSocket)
        XCTAssertEqual(DirectMessageSender.route(sealedAllowed: false, isGroup: false, peerKey: keyHex), .plain)
        XCTAssertEqual(DirectMessageSender.route(sealedAllowed: true, isGroup: true, peerKey: keyHex), .plain)
    }

    func testOwnKeyIsStableUntilRotatedAndRegistrationTracksIt() throws {
        let keys = DeliveryKeys(storage: MemoryDeliveryKeyStorage())
        let first = keys.ownKey()
        XCTAssertTrue(DeliveryKeys.isValid(first))
        XCTAssertEqual(keys.ownKey(), first)
        XCTAssertTrue(keys.needsRegistration)
        keys.markRegistered()
        XCTAssertFalse(keys.needsRegistration)
        let second = keys.rotateOwnKey()
        XCTAssertNotEqual(second, first)
        XCTAssertTrue(keys.needsRegistration, "a replaced key must be registered again")
    }

    // The value the server is given must be the SHA-256 the server computes from the key.
    func testKeyHashMatchesTheServer() {
        XCTAssertEqual(
            DeliveryKeys.hash(ofKeyHex: String(repeating: "00", count: 32)),
            "66687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f2925"
        )
        XCTAssertNil(DeliveryKeys.hash(ofKeyHex: "zz"))
    }

    // A key that arrives in a message is not used until that message is authenticated.
    func testPeerKeyIsUsedOnlyAfterTheMessageIsAuthenticated() {
        let keys = DeliveryKeys(storage: MemoryDeliveryKeyStorage())
        keys.notePending(peerDID: alice, key: keyHex)
        XCTAssertNil(keys.peerKey(for: alice))
        keys.confirm(peerDID: bob)
        XCTAssertNil(keys.peerKey(for: alice), "confirming someone else must not release Alice's key")
        keys.confirm(peerDID: alice)
        XCTAssertEqual(keys.peerKey(for: alice), keyHex)
        keys.notePending(peerDID: alice, key: "not-a-key")
        keys.confirm(peerDID: alice)
        XCTAssertEqual(keys.peerKey(for: alice), keyHex, "a malformed key must not replace a good one")
        keys.forget(peerDID: alice)
        XCTAssertNil(keys.peerKey(for: alice))
    }

    func testBlockedSendersList() {
        let did = "did:key:zBlockedForTest"
        BlockedSenders.set(did, blocked: true)
        XCTAssertTrue(BlockedSenders.contains(did))
        BlockedSenders.set(did, blocked: false)
        XCTAssertFalse(BlockedSenders.contains(did))
    }
}

final class MembershipMarkTests: XCTestCase {
    func testServerValues() {
        XCTAssertEqual(MembershipMark(serverValue: "vip"), .vip)
        XCTAssertEqual(MembershipMark(serverValue: "organization"), .organization)
        XCTAssertEqual(MembershipMark(serverValue: "official"), .official)
        XCTAssertNil(MembershipMark(serverValue: ""))
        XCTAssertNil(MembershipMark(serverValue: nil))
        XCTAssertNil(MembershipMark(serverValue: "verified"))
    }

    // A paid mark must say it is paid, and must not use the word that describes earned trust.
    func testPaidMarkDoesNotClaimVerification() {
        let text = MembershipMark.vip.accessibilityText
        XCTAssertTrue(text.contains("paid"))
        XCTAssertTrue(text.contains("not identity verification"))
    }

    func testEveryMarkHasItsOwnColour() {
        let colours = MembershipMark.allCases.map { $0.color.description }
        XCTAssertEqual(Set(colours).count, MembershipMark.allCases.count)
    }
}

final class LinkPreviewPolicyTests: XCTestCase {
    // A link someone else sent must not make this device contact the linked site on its own.
    func testReceivedLinksAreNotFetchedUnlessTheUserOptsIn() {
        XCTAssertFalse(MessageLinkPreview.shouldFetch(isOwnMessage: false, receivedOptIn: false))
        XCTAssertTrue(MessageLinkPreview.shouldFetch(isOwnMessage: false, receivedOptIn: true))
        XCTAssertTrue(MessageLinkPreview.shouldFetch(isOwnMessage: true, receivedOptIn: false))
    }

    func testReceivedOptInIsOffByDefault() {
        UserDefaults.standard.removeObject(forKey: "echo.link_previews.received")
        XCTAssertFalse(MessageLinkPreview.fetchForReceivedMessages)
        XCTAssertFalse(MessageLinkPreview.shouldFetch(isOwnMessage: false))
    }
}
#endif
