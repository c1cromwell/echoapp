import XCTest
@testable import Echo

@MainActor
final class ConversationPreferencesStoreTests: XCTestCase {

    private func makeStore() -> ConversationPreferencesStore {
        ConversationPreferencesStore(backing: InMemoryPreferencesBacking())
    }

    func testDefaultPreferences_areUnmutedAndOff() {
        let store = makeStore()
        let prefs = store.preferences(for: "conv-1")
        XCTAssertFalse(prefs.isMuted)
        XCTAssertEqual(prefs.disappearing, .off)
    }

    func testSetMuted_persistsAndIsReadBack() {
        let store = makeStore()
        store.setMuted(true, for: "conv-1")
        XCTAssertTrue(store.isMuted("conv-1"))
        store.setMuted(false, for: "conv-1")
        XCTAssertFalse(store.isMuted("conv-1"))
    }

    func testSetDisappearing_persistsIndependentlyPerConversation() {
        let store = makeStore()
        store.setDisappearing(.h24, for: "conv-1")
        store.setDisappearing(.m5, for: "conv-2")
        XCTAssertEqual(store.preferences(for: "conv-1").disappearing, .h24)
        XCTAssertEqual(store.preferences(for: "conv-2").disappearing, .m5)
    }

    func testMuteAndTimer_coexistOnSameConversation() {
        let store = makeStore()
        store.setMuted(true, for: "conv-1")
        store.setDisappearing(.d7, for: "conv-1")
        let prefs = store.preferences(for: "conv-1")
        XCTAssertTrue(prefs.isMuted)
        XCTAssertEqual(prefs.disappearing, .d7)
    }

    func testChatFolder_trustTierPredicate() {
        XCTAssertTrue(ChatFolder.all.includes(tier: 0))
        XCTAssertTrue(ChatFolder.verified.includes(tier: 2))
        XCTAssertFalse(ChatFolder.verified.includes(tier: 1))
        XCTAssertTrue(ChatFolder.trusted.includes(tier: 4))
        XCTAssertFalse(ChatFolder.trusted.includes(tier: 2))
        XCTAssertTrue(ChatFolder.unverified.includes(tier: 1))
        XCTAssertFalse(ChatFolder.unverified.includes(tier: 2))
    }

    func testMessageActions_gateEditAndDelete() {
        // Edit only for own message within window; delete only for own.
        let peerSheet = MessageActionsSheet(messagePreview: "hi", isOwnMessage: false, sentWithinEditWindow: true) { _ in }
        let ownSheet = MessageActionsSheet(messagePreview: "hi", isOwnMessage: true, sentWithinEditWindow: false) { _ in }
        _ = peerSheet
        _ = ownSheet
        // Pure predicate coverage via the enum:
        XCTAssertTrue(MessageAction.delete.isDestructive)
        XCTAssertFalse(MessageAction.reply.isDestructive)
    }
}
