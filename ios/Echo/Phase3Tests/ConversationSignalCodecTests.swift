import CryptoKit
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
final class RecoveryLinkSyncTests: XCTestCase {
    func testWhenTheAddressNeedsSending() {
        XCTAssertFalse(RecoveryLinkSync.needsSync(address: nil, lastSent: nil))
        XCTAssertFalse(RecoveryLinkSync.needsSync(address: "", lastSent: nil))
        XCTAssertTrue(RecoveryLinkSync.needsSync(address: "DAG4abc", lastSent: nil))
        XCTAssertFalse(RecoveryLinkSync.needsSync(address: "DAG4abc", lastSent: "DAG4abc"))
        XCTAssertTrue(RecoveryLinkSync.needsSync(address: "DAG4new", lastSent: "DAG4abc"), "a new recovery key must be sent")
    }

    func testEndpoint() {
        XCTAssertEqual(RecoveryLinkSync.LinkEndpoint().path, "/v1/identity/link-wallet")
    }
}
@MainActor
final class DIContainerTests: XCTestCase {
    // Every screen that talks to the server asks the container for its client. If the container
    // cannot build one, Contacts, Rewards, search and key registration all silently do nothing.
    func testFactoryBackedServicesResolve() {
        XCTAssertNotNil(DIContainer.shared.resolveAPIClient(), "the API client did not resolve")
        XCTAssertNotNil(DIContainer.shared.resolveGamificationAPI(), "the gamification API did not resolve")
        XCTAssertNotNil(DIContainer.shared.resolveConversationSignalService(), "the signal service did not resolve")
        XCTAssertNotNil(DIContainer.shared.resolveKeychain())
    }

    func testAResolvedServiceIsReused() {
        let first = DIContainer.shared.resolveAPIClient()
        let second = DIContainer.shared.resolveAPIClient()
        XCTAssertTrue(first === second, "resolving twice built two clients")
    }
}
final class RequestSigningTests: XCTestCase {
    // What the app signs must be the very bytes the server rebuilds and hashes once. If the app
    // signs a hash of them instead, no signed request is ever accepted.
    func testSigningInputIsTheCanonicalStringItself() {
        let input = PasskeySigningInterceptor.signingInput(
            method: "POST", path: "/v1/auth/session", timestamp: "1795737600", body: Data("{}".utf8)
        )
        XCTAssertEqual(
            String(decoding: input, as: UTF8.self),
            "POST\n/v1/auth/session\n1795737600\n44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
        )
        XCTAssertNotEqual(input.count, 32, "the signing input is a 32-byte hash; it must be the message")
    }

    func testEmptyBodyHashesTheEmptyString() {
        let input = PasskeySigningInterceptor.signingInput(method: "GET", path: "/v3/prefs", timestamp: "1", body: Data())
        XCTAssertTrue(String(decoding: input, as: UTF8.self)
            .hasSuffix("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))
    }

    // End to end with a software key: a signature made the way the app makes one verifies against
    // the message hashed once, which is what the server does.
    func testSignatureVerifiesTheWayTheServerChecksIt() throws {
        let key = P256.Signing.PrivateKey()
        let input = PasskeySigningInterceptor.signingInput(method: "PUT", path: "/v3/messages/delivery-key", timestamp: "42", body: Data("{\"a\":1}".utf8))
        let signature = try key.signature(for: input) // hashes with SHA-256, as SecureEnclaveManager.sign does
        XCTAssertTrue(key.publicKey.isValidSignature(signature, for: SHA256.hash(data: input)))
    }
}
final class IdentityEndpointTests: XCTestCase {
    // Must match the server route (internal/api/router.go: "/identity/resolve/").
    func testPeerKeysAreResolvedAtTheRouteTheServerServes() {
        XCTAssertEqual(IdentityEndpoint.resolveDID(did: "did:key:zAbc").path, "/identity/resolve/did:key:zAbc")
    }
}

final class SessionTokenOwnershipTests: XCTestCase {
    private func jwt(sub: String) -> String {
        func b64(_ object: [String: Any]) -> String {
            let data = try! JSONSerialization.data(withJSONObject: object)
            return data.base64EncodedString().replacingOccurrences(of: "+", with: "-")
                .replacingOccurrences(of: "/", with: "_").replacingOccurrences(of: "=", with: "")
        }
        return "\(b64(["alg": "ES256"])).\(b64(["sub": sub, "exp": 4_102_444_800])).sig"
    }

    // The Keychain outlives the app. A token left by a previous account must not be used once
    // this device is a different identity.
    func testATokenForAnotherIdentityIsNotThisDevices() {
        let old = jwt(sub: "did:key:zOldAccount")
        XCTAssertEqual(SessionTokenBootstrap.subject(of: old), "did:key:zOldAccount")
        XCTAssertTrue(SessionTokenBootstrap.belongsToSomeoneElse(old, currentDID: "did:key:zNewAccount"))
        XCTAssertFalse(SessionTokenBootstrap.belongsToSomeoneElse(old, currentDID: "did:key:zOldAccount"))
    }

    func testAnUnreadableTokenOrUnknownIdentityIsLeftAlone() {
        XCTAssertFalse(SessionTokenBootstrap.belongsToSomeoneElse("not-a-jwt", currentDID: "did:key:zNew"))
        XCTAssertFalse(SessionTokenBootstrap.belongsToSomeoneElse(jwt(sub: "did:key:zOld"), currentDID: nil))
    }
}

final class MessagingKeyRegistrationTests: XCTestCase {
    // A second account on the same install has its own identity and must publish its own key.
    func testANewIdentityStillHasToRegister() {
        XCTAssertTrue(MessagingKeyRegistrar.needsRegistration(did: "did:key:zNew", lastRegistered: nil))
        XCTAssertTrue(MessagingKeyRegistrar.needsRegistration(did: "did:key:zNew", lastRegistered: "did:key:zOld"))
        XCTAssertFalse(MessagingKeyRegistrar.needsRegistration(did: "did:key:zNew", lastRegistered: "did:key:zNew"))
        XCTAssertFalse(MessagingKeyRegistrar.needsRegistration(did: "", lastRegistered: nil))
    }
}

final class FirstRunIdentityTests: XCTestCase {
    // First run must end with one identity, not two: the provisioning step reuses the key the
    // Face ID step made when that key is the one the recorded DID was derived from.
    func testTheKeyFromTheFaceIDStepIsReused() throws {
        let key = P256.Signing.PrivateKey().publicKey.x963Representation
        let hex = key.map { String(format: "%02x", $0) }.joined()
        let did = try DidKeyDeriver.deriveFromPublicKeyHex(hex)
        XCTAssertTrue(RealProvisionSecureEnclave.shouldReuseExistingKey(recordedDID: did, existingPublicKey: key))
    }

    func testANewKeyIsMadeWhenThereIsNothingToReuse() throws {
        let key = P256.Signing.PrivateKey().publicKey.x963Representation
        let other = P256.Signing.PrivateKey().publicKey.x963Representation
        let otherDID = try DidKeyDeriver.deriveFromPublicKeyHex(other.map { String(format: "%02x", $0) }.joined())
        XCTAssertFalse(RealProvisionSecureEnclave.shouldReuseExistingKey(recordedDID: nil, existingPublicKey: key))
        XCTAssertFalse(RealProvisionSecureEnclave.shouldReuseExistingKey(recordedDID: "did:key:zX", existingPublicKey: nil))
        XCTAssertFalse(RealProvisionSecureEnclave.shouldReuseExistingKey(recordedDID: otherDID, existingPublicKey: key),
                       "a key that does not belong to the recorded identity must not be reused")
    }
}

final class UsernameRegistrationTests: XCTestCase {
    func testWhenAUsernameStillHasToBeRegistered() {
        let did = "did:key:zNew"
        XCTAssertTrue(UsernameRegistration.needsSync(did: did, username: "alice", lastRegistered: nil))
        let done = UsernameRegistration.marker(did: did, username: "alice")
        XCTAssertFalse(UsernameRegistration.needsSync(did: did, username: "alice", lastRegistered: done))
        XCTAssertFalse(UsernameRegistration.needsSync(did: did, username: "Alice", lastRegistered: done), "case alone is not a new name")
        XCTAssertTrue(UsernameRegistration.needsSync(did: "did:key:zOther", username: "alice", lastRegistered: done), "a new account registers its own name")
        XCTAssertTrue(UsernameRegistration.needsSync(did: did, username: "alice2", lastRegistered: done))
        XCTAssertFalse(UsernameRegistration.needsSync(did: nil, username: "alice", lastRegistered: nil))
        XCTAssertFalse(UsernameRegistration.needsSync(did: did, username: nil, lastRegistered: nil))
    }

    // Must match the server's rule, and must never claim the placeholder name.
    func testOnlyRealUsernamesAreRegistered() {
        for good in ["bob", "Alice_1", "a23456789012345678901234567890"] {
            XCTAssertTrue(UsernameRegistration.isRegistrable(good), good)
        }
        for bad in ["", "ab", "You", "has space", "émile", "semi;colon", "a234567890123456789012345678901"] {
            XCTAssertFalse(UsernameRegistration.isRegistrable(bad), bad)
        }
    }

    func testEndpoint() {
        XCTAssertEqual(UsernameRegistration.Endpoint().path, "/v3/auth/register")
        XCTAssertEqual(UsernameRegistration.AvailabilityEndpoint(username: "alice").path, "/v1/users/check-username?username=alice")
    }

    // The server refuses for two reasons. Only "someone else holds this name" asks the person
    // to pick another; "this account already has a different name" leaves nothing to choose.
    func testARefusalAsksForAnotherNameOnlyWhenTheNameIsTaken() {
        XCTAssertEqual(UsernameRegistration.refusalOutcome(nameIsAvailable: false), .taken)
        XCTAssertEqual(UsernameRegistration.refusalOutcome(nameIsAvailable: true), .accountAlreadyNamed)
        XCTAssertEqual(UsernameRegistration.refusalOutcome(nameIsAvailable: nil), .taken, "unknown: ask rather than leave the account unfindable")
    }

    func testPromptOnlySubmitsANameThatMayBeFree() {
        typealias Prompt = UsernameConflictPrompt
        XCTAssertTrue(Prompt.canSubmit(check: .available, isSaving: false))
        XCTAssertTrue(Prompt.canSubmit(check: .unreachable, isSaving: false), "the server decides when it could not be asked first")
        for blocked in [Prompt.Check.idle, .checking, .unavailable, .invalid] {
            XCTAssertFalse(Prompt.canSubmit(check: blocked, isSaving: false))
        }
        XCTAssertFalse(Prompt.canSubmit(check: .available, isSaving: true))
        XCTAssertNil(Prompt.hint(for: .idle))
        XCTAssertEqual(Prompt.hint(for: .unavailable), "That username is taken.")
    }
}

final class DemoDataTests: XCTestCase {
    // A Debug build starts empty unless demo data is asked for.
    func testDemoDataIsOffUnlessAskedFor() {
        XCTAssertFalse(DemoData.isRequested(arguments: ["Echo"], environment: [:]))
        XCTAssertFalse(DemoData.isRequested(arguments: ["Echo", "-uiTestMode"], environment: ["ECHO_DEMO_DATA": "0"]))
        XCTAssertTrue(DemoData.isRequested(arguments: ["Echo", "-echoDemoData"], environment: [:]))
        XCTAssertTrue(DemoData.isRequested(arguments: ["Echo"], environment: ["ECHO_DEMO_DATA": "1"]))
    }
}

final class MessageOrderingTests: XCTestCase {
    // Work added to the queue runs in the order it was added, even when an earlier item takes
    // longer than a later one. Without this, two messages that arrive together can swap.
    func testSerialQueueKeepsArrivalOrder() async {
        let queue = SerialAsyncQueue()
        let recorder = OrderRecorder()
        for index in 0..<20 {
            queue.enqueue {
                // Earlier items sleep longer: unordered execution would finish them last.
                try? await Task.sleep(nanoseconds: UInt64(20 - index) * 2_000_000)
                await recorder.append(index)
            }
        }
        await queue.drain()
        let order = await recorder.values
        XCTAssertEqual(order, Array(0..<20))
    }

    // Read state for a 1:1 chat is not sent to the server: it can no longer match a message by
    // id, and the call would tell it which message this person read.
    func testOnlyGroupsUseServerReceipts() {
        XCTAssertFalse(ChatDetailViewModel.usesServerReceipts(conversationId: "dm:did:key:zA:did:key:zB"))
        XCTAssertTrue(ChatDetailViewModel.usesServerReceipts(conversationId: "group:abc"))
    }

    func testAnUndecryptableMessageIsCountedAndAnnounced() {
        let defaults = UserDefaults(suiteName: "undecryptable-\(UUID().uuidString)")!
        let announced = expectation(forNotification: .echoRelayError, object: nil) { note in
            (note.userInfo?["code"] as? String) == "UNDECRYPTABLE"
        }
        UndecryptableMessages.record(defaults: defaults)
        wait(for: [announced], timeout: 2)
        XCTAssertEqual(defaults.integer(forKey: UndecryptableMessages.countKey), 1)
    }
}

private actor OrderRecorder {
    private(set) var values: [Int] = []
    func append(_ value: Int) { values.append(value) }
}

final class RequestURLTests: XCTestCase {
    private let base = URL(string: "http://localhost:8000")!

    // A query string must reach the server as a query string. It used to be folded into the path
    // with its "?" encoded, and the server answered 404 to every search.
    func testQueryIsNotFoldedIntoThePath() {
        let url = APIClient.requestURL(base: base, path: "/v3/contacts/search?handle=bob")
        XCTAssertEqual(url.absoluteString, "http://localhost:8000/v3/contacts/search?handle=bob")
        XCTAssertEqual(url.path, "/v3/contacts/search")
        XCTAssertEqual(url.query, "handle=bob")
    }

    func testPathWithoutAQueryIsUnchanged() {
        XCTAssertEqual(APIClient.requestURL(base: base, path: "/v3/prefs").absoluteString, "http://localhost:8000/v3/prefs")
    }

    func testAnAlreadyEncodedValueIsNotEncodedTwice() {
        let url = APIClient.requestURL(base: base, path: "/v3/profile/view?did=did%3Akey%3AzAbc")
        XCTAssertEqual(url.absoluteString, "http://localhost:8000/v3/profile/view?did=did%3Akey%3AzAbc")
    }

    // A DID encoded by the caller must reach the server encoded once, not twice.
    func testAPreEncodedIdentifierInThePathIsNotEncodedTwice() {
        let did = "did:key:zDnaeAbc"
        let preEncoded = did.addingPercentEncoding(withAllowedCharacters: .alphanumerics)!
        XCTAssertEqual(preEncoded, "did%3Akey%3AzDnaeAbc")
        let fromEncoded = APIClient.requestURL(base: base, path: "/identity/resolve/\(preEncoded)")
        let fromRaw = APIClient.requestURL(base: base, path: "/identity/resolve/\(did)")
        XCTAssertEqual(fromEncoded.path, "/identity/resolve/did:key:zDnaeAbc", "the server would see a different identifier")
        XCTAssertEqual(fromEncoded.path, fromRaw.path)
        XCTAssertFalse(fromEncoded.absoluteString.contains("%25"), "the percent sign was encoded again")
    }

    func testEveryEndpointThatCarriesAQueryKeepsIt() {
        let leaderboard = APIClient.requestURL(base: base, path: GamificationEndpoint.leaderboard(window: "weekly").path)
        XCTAssertEqual(leaderboard.path, "/v1/gamification/leaderboard")
        XCTAssertEqual(leaderboard.query, "window=weekly")
    }
}
#endif
