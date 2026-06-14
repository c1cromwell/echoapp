import XCTest
@testable import Echo

#if os(iOS)
final class MessageSearchTests: XCTestCase {
    func testTokenizer_splitsWords() {
        let tokens = MessageSearchTokenizer.tokenize("Hello World testing")
        XCTAssertTrue(tokens.contains("hello"))
        XCTAssertTrue(tokens.contains("world"))
        XCTAssertTrue(tokens.contains("testing"))
    }

    func testParseQuery_phrase() {
        let terms = MessageSearchTokenizer.parseQuery("\"exact phrase\" hello")
        XCTAssertEqual(terms, ["exact phrase", "hello"])
    }

    func testKeywordSearch_findsMatch() async {
        await LocalMessageIndexer.shared.rebuildFromLocalThreads()
        await LocalMessageIndexer.shared.indexMessage(
            conversationId: "conv-search",
            messageId: "msg-1",
            senderDID: "did:key:alice",
            body: "meet me at the lighthouse tonight",
            sentAt: Date(),
            contentType: "text"
        )
        let hits = await KeywordSearchEngine.shared.search(query: "lighthouse")
        XCTAssertFalse(hits.isEmpty)
        XCTAssertEqual(hits.first?.messageId, "msg-1")
    }

    func testKeywordSearch_excludesArchivedByDefault() async {
        ConversationArchiveStore.setArchived(true, conversationId: "conv-archived")
        await LocalMessageIndexer.shared.indexMessage(
            conversationId: "conv-archived",
            messageId: "msg-arch",
            senderDID: "did:key:alice",
            body: "secret archived keyword",
            sentAt: Date(),
            contentType: "text"
        )
        await LocalMessageIndexer.shared.indexMessage(
            conversationId: "conv-active",
            messageId: "msg-active",
            senderDID: "did:key:alice",
            body: "secret archived keyword",
            sentAt: Date(),
            contentType: "text"
        )
        let hits = await KeywordSearchEngine.shared.search(query: "archived keyword")
        XCTAssertEqual(hits.count, 1)
        XCTAssertEqual(hits.first?.conversationId, "conv-active")
        ConversationArchiveStore.setArchived(false, conversationId: "conv-archived")
    }
}
#endif
