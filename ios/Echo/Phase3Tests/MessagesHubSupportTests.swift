import XCTest
@testable import Echo

final class PinnedConversationsStoreTests: XCTestCase {
    func testMaxPins_isTwenty() {
        XCTAssertEqual(PinnedConversationsStore.maxPins, 20)
    }

    @MainActor
    func testToggle_unpins() {
        let store = PinnedConversationsStore.shared
        store.pin("test-conv-toggle")
        XCTAssertTrue(store.isPinned("test-conv-toggle"))
        store.toggle("test-conv-toggle")
        XCTAssertFalse(store.isPinned("test-conv-toggle"))
        store.unpin("test-conv-toggle")
    }
}

final class MessagesHubSupportTests: XCTestCase {
    func testPinnedItems_preservesOrder() {
        let convA = StoredConversation(id: "a", contactName: "Alice", peerDID: "did:a")
        let convB = StoredConversation(id: "b", contactName: "Bob", peerDID: "did:b")
        let items = MessagesHubSupport.pinnedItems(
            conversations: [convA, convB],
            orderedPinIDs: ["b", "a"]
        )
        XCTAssertEqual(items.map(\.id), ["b", "a"])
        XCTAssertEqual(items.first?.name, "Bob")
    }

    func testFolderUnreadCounts_verifiedOnly() {
        let convs = [
            StoredConversation(id: "1", contactName: "A", peerDID: "d1", unreadCount: 2),
            StoredConversation(id: "2", contactName: "B", peerDID: "d2", unreadCount: 5),
        ]
        let counts = MessagesHubSupport.folderUnreadCounts(conversations: convs) { id in
            id == "1" ? 2 : 0
        }
        XCTAssertEqual(counts[.verified], 2)
        // folderUnreadCounts reports every folder with unread > 0, including unverified.
        XCTAssertEqual(counts[.unverified], 5)
    }
}

#if os(iOS)
final class ContactTrustIndexTests: XCTestCase {
    @MainActor
    func testIngestRemoteContacts_mapsBadge() {
        let index = ContactTrustIndex.shared
        index.ingestRemoteContacts([
            RemoteContact(contactDid: "did:key:bob", ownerDid: nil, addedVia: nil, trustBadge: "verified", blocked: false),
        ])
        XCTAssertEqual(index.tier(peerDID: "did:key:bob"), 2)
    }
}
#endif
