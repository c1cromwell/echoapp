#if os(iOS)
import XCTest
@testable import Echo

/// Two-device restore proof: back up device-A history, encrypt with a recovery
/// phrase, then on a fresh device-B decrypt + apply and assert the conversation
/// and messages are reproduced. Exercises the real backup→restore path
/// (BackupCrypto + HistorySyncBundleMerger) end to end, no network.
@MainActor
final class BackupRestoreTests: XCTestCase {
    // Valid 24-word BIP-39 mnemonic (all-zero entropy) — passes checksum.
    private let phraseWords = Array(repeating: "abandon", count: 23) + ["art"]
    private let otherWords = Array(repeating: "zoo", count: 23) + ["vote"]
    private let convId = "restore-test:dm1"

    override func tearDown() {
        UserDefaults.standard.removeObject(forKey: "echo.thread.v1.\(convId)")
        UserDefaults.standard.removeObject(forKey: "echo.polls.v1.\(convId)")
        super.tearDown()
    }

    func testTwoDeviceBackupRestore() throws {
        guard let phrase = RecoveryPhrase(words: phraseWords) else {
            return XCTFail("phrase invalid")
        }

        // Device A: a conversation with two messages.
        let conversation = StoredConversation(id: convId, contactName: "Sam", peerDID: "did:key:sam")
        let messages = [
            StoredThreadMessage(id: "m1", senderDID: "did:key:sam", content: "first", timestamp: "Now"),
            StoredThreadMessage(id: "m2", senderDID: "did:key:me", content: "second", timestamp: "Now"),
        ]
        let bundle = HistorySyncBundle(conversations: [conversation], threads: [convId: messages])

        // Back up (encrypt).
        let archive = try BackupCrypto.encrypt(bundle: bundle, phrase: phrase)

        // Device B: fresh — no local history for this conversation.
        UserDefaults.standard.removeObject(forKey: "echo.thread.v1.\(convId)")
        XCTAssertTrue(ConversationThreadStore.exportMessages(conversationId: convId).isEmpty,
                      "device B should start empty")

        // Restore: decrypt + apply.
        let recovered = try BackupCrypto.decrypt(archiveData: archive, phrase: phrase)
        XCTAssertEqual(recovered, bundle, "decrypted bundle must match the backup")
        HistorySyncBundleMerger.apply(recovered)

        // Device B now has the messages and the conversation.
        let restored = ConversationThreadStore.exportMessages(conversationId: convId)
        XCTAssertEqual(restored.map(\.id), ["m1", "m2"], "messages restored in order")
        XCTAssertEqual(restored.map(\.content), ["first", "second"])
        XCTAssertNotNil(ConversationStore.shared.conversation(id: convId), "conversation restored")

        // Re-applying is idempotent (no duplicates).
        HistorySyncBundleMerger.apply(recovered)
        XCTAssertEqual(ConversationThreadStore.exportMessages(conversationId: convId).count, 2,
                       "restore is idempotent")
    }

    func testWrongPhraseCannotRestore() throws {
        guard let phrase = RecoveryPhrase(words: phraseWords),
              let wrong = RecoveryPhrase(words: otherWords) else {
            return XCTFail("phrase invalid")
        }
        let bundle = HistorySyncBundle(
            conversations: [StoredConversation(id: convId, contactName: "Sam", peerDID: "did:key:sam")],
            threads: [convId: [StoredThreadMessage(id: "m1", senderDID: "did:key:sam", content: "secret", timestamp: "Now")]]
        )
        let archive = try BackupCrypto.encrypt(bundle: bundle, phrase: phrase)
        XCTAssertThrowsError(try BackupCrypto.decrypt(archiveData: archive, phrase: wrong),
                             "a different recovery phrase must not decrypt the backup")
    }
}
#endif
