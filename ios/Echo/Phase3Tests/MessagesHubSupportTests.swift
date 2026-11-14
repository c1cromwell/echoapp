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
            RemoteContact(contactDid: "did:key:bob", ownerDid: nil, addedVia: nil, trustBadge: "verified", blocked: false, barAllowed: nil),
        ])
        XCTAssertEqual(index.tier(peerDID: "did:key:bob"), 2)
    }
}
#endif

#if os(iOS)
final class MessagesHubEmptyStateTests: XCTestCase {
    func testChatsText_nilWhenRowsAreVisible() {
        XCTAssertNil(MessagesHubEmptyState.chatsText(visibleCount: 2, pinnedCount: 0, searchText: "", folderSelected: false))
    }

    func testChatsText_searchWithNoMatchesNamesTheQuery() {
        XCTAssertEqual(
            MessagesHubEmptyState.chatsText(visibleCount: 0, pinnedCount: 3, searchText: " aria ", folderSelected: false),
            "No chats match \u{201C}aria\u{201D}."
        )
    }

    func testChatsText_emptyFolder() {
        XCTAssertEqual(
            MessagesHubEmptyState.chatsText(visibleCount: 0, pinnedCount: 0, searchText: "", folderSelected: true),
            "No chats in this folder yet."
        )
    }

    func testChatsText_firstRunHintOnlyWhenNothingIsPinned() {
        XCTAssertNotNil(MessagesHubEmptyState.chatsText(visibleCount: 0, pinnedCount: 0, searchText: "", folderSelected: false))
        XCTAssertNil(MessagesHubEmptyState.chatsText(visibleCount: 0, pinnedCount: 1, searchText: "", folderSelected: false))
    }
}

final class RewardsHubLoadStateTests: XCTestCase {
    func testResolve_failedOnlyWhenEverythingFailed() {
        XCTAssertEqual(RewardsHubLoadState.resolve(statusLoaded: false, streakLoaded: false, leaderboardLoaded: false), .failed)
        XCTAssertEqual(RewardsHubLoadState.resolve(statusLoaded: false, streakLoaded: true, leaderboardLoaded: false), .loaded)
        XCTAssertEqual(RewardsHubLoadState.resolve(statusLoaded: true, streakLoaded: true, leaderboardLoaded: true), .loaded)
    }
}

final class APIErrorUserTextTests: XCTestCase {
    func testDescriptions_haveNoHTTPJargon() {
        let errors: [APIError] = [
            .invalidResponse, .badRequest, .unauthorized, .forbidden, .notFound,
            .rateLimited(retryAfter: 3), .serverError(503), .unexpectedStatusCode(418),
            .networkError(URLError(.notConnectedToInternet)),
        ]
        for error in errors {
            let text = error.localizedDescription
            XCTAssertNil(text.range(of: #"\d{3}"#, options: .regularExpression), "status code leaked: \(text)")
            for jargon in ["Unauthorized", "Forbidden", "Bad request", "decode", "status code"] {
                XCTAssertFalse(text.contains(jargon), "jargon leaked: \(text)")
            }
        }
    }

    func testDescriptions_sayWhatToDo() {
        XCTAssertEqual(APIError.rateLimited(retryAfter: 3).localizedDescription, "Echo is busy. Try again in 3 seconds.")
        XCTAssertEqual(APIError.rateLimited(retryAfter: 1).localizedDescription, "Echo is busy. Try again in 1 second.")
        XCTAssertEqual(
            APIError.networkError(URLError(.notConnectedToInternet)).localizedDescription,
            "You're offline. Check your connection and try again."
        )
        XCTAssertEqual(APIError.serverError(503).diagnosticDescription, "HTTP 503")
    }
}
#endif
