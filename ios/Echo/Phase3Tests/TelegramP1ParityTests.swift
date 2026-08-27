import XCTest
@testable import Echo

final class TelegramP1ParityTests: XCTestCase {
    func testComposerDraftRoundTrip() {
        let id = "dm:test-draft-\(UUID().uuidString)"
        ComposerDraftStore.clear(conversationId: id)
        XCTAssertEqual(ComposerDraftStore.load(conversationId: id), "")
        ComposerDraftStore.save(conversationId: id, text: "hello draft")
        XCTAssertEqual(ComposerDraftStore.load(conversationId: id), "hello draft")
        ComposerDraftStore.clear(conversationId: id)
        XCTAssertEqual(ComposerDraftStore.load(conversationId: id), "")
    }

    func testSavedMessagesConversationIdIsSelfPair() {
        let did = "did:key:zTestSaved"
        let id = SavedMessagesStore.conversationId(localDID: did)
        // Self-pair mirrors ConversationID.direct(localDID:peerDID:) sorted-pair algorithm.
        XCTAssertEqual(id, "dm:\(did):\(did)")
        let conv = StoredConversation(id: id, contactName: "Saved Messages", peerDID: did)
        XCTAssertTrue(SavedMessagesStore.isSavedMessages(conv, localDID: did))
        XCTAssertFalse(SavedMessagesStore.isSavedMessages(conv, localDID: "did:key:other"))
    }

    func testLinkPreviewExtractsURLAndDomain() {
        let text = "Check https://www.example.com/path?q=1 please"
        let url = MessageLinkPreview.firstURL(in: text)
        XCTAssertEqual(url?.host, "www.example.com")
        XCTAssertEqual(MessageLinkPreview.domain(for: url!), "example.com")
        let stub = MessageLinkPreview.previewStub(from: text)
        XCTAssertEqual(stub?.domain, "example.com")
        XCTAssertNil(MessageLinkPreview.firstURL(in: "no links here"))
    }

    func testLinkPreviewTitleExtraction() {
        let html = "<html><head><title>Echo Notes</title></head><body></body></html>"
        XCTAssertEqual(MessageLinkPreview.extractTitle(from: html), "Echo Notes")
    }

    func testReplyPreviewAndForwardBody() {
        let preview = MessageComposerLogic.replyPreview(authorName: "Alice", content: "Hello world")
        XCTAssertTrue(preview.contains("Alice"))
        XCTAssertTrue(preview.contains("Hello"))
        XCTAssertEqual(MessageComposerLogic.forwardBody("  hi  "), "↪ hi")
        XCTAssertEqual(MessageComposerLogic.forwardBody("   "), "")
    }

    func testScheduledRemoteCodecKeysAndEnvelope() throws {
        let fireAt = Date(timeIntervalSince1970: 1_788_220_800) // 2026-09-01T00:00:00Z
        let request = ScheduledMessageRemoteCodec.createRequest(
            conversationId: "dm:a:b",
            ciphertext: "sealed-b64",
            fireAt: fireAt,
            timezone: "UTC",
            silent: true
        )
        let data = try JSONEncoder().encode(request)
        let json = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
        XCTAssertEqual(json["conversation_id"] as? String, "dm:a:b")
        XCTAssertEqual(json["content"] as? String, "sealed-b64")
        XCTAssertEqual(json["content_type"] as? String, "text")
        XCTAssertEqual(json["scheduled_at"] as? String, "2026-09-01T00:00:00Z")
        XCTAssertEqual(json["timezone"] as? String, "UTC")
        XCTAssertEqual(json["silent"] as? Bool, true)
        XCTAssertEqual(ScheduledMessageRemoteCodec.collectionPath, "/v3/messages/schedule")
        XCTAssertEqual(ScheduledMessageRemoteCodec.itemPath(id: "abc"), "/v3/messages/schedule/abc")

        let envelope = try ScheduledMessageRemoteCodec.envelopeJSON(peerDID: "did:key:zPeer", plaintext: "hi")
        let parsed = try ScheduledMessageRemoteCodec.parseEnvelope(envelope)
        XCTAssertEqual(parsed.peerDID, "did:key:zPeer")
        XCTAssertEqual(parsed.body, "hi")
    }
}
