import XCTest
@testable import Echo

#if os(iOS)
final class BackupCryptoTests: XCTestCase {
    private let stubWords = [
        "abandon", "ability", "able", "about", "above", "absent", "absorb", "abstract",
        "absurd", "abuse", "access", "accident", "account", "accuse", "achieve", "acid",
        "acoustic", "acquire", "across", "act", "action", "actor", "actress", "actual"
    ]

    func testEncryptDecryptRoundTrip() throws {
        guard let phrase = RecoveryPhrase(words: stubWords) else {
            XCTFail("stub phrase invalid")
            return
        }
        let bundle = HistorySyncBundle(
            conversations: [
                StoredConversation(id: "dm:1", contactName: "Sam", peerDID: "did:key:sam")
            ],
            threads: [
                "dm:1": [
                    StoredThreadMessage(id: "m1", senderDID: "did:key:sam", content: "hi", timestamp: "Now")
                ]
            ]
        )
        let archive = try BackupCrypto.encrypt(bundle: bundle, phrase: phrase)
        let recovered = try BackupCrypto.decrypt(archiveData: archive, phrase: phrase)
        XCTAssertEqual(recovered, bundle)
    }

    func testDecryptRejectsBadPhrase() throws {
        guard let phrase = RecoveryPhrase(words: stubWords) else {
            XCTFail("stub phrase invalid")
            return
        }
        let bundle = HistorySyncBundle(conversations: [], threads: [:])
        let archive = try BackupCrypto.encrypt(bundle: bundle, phrase: phrase)
        var bad = stubWords
        bad[0] = "zebra"
        guard let wrong = RecoveryPhrase(words: bad) else {
            XCTFail("expected valid word count")
            return
        }
        XCTAssertThrowsError(try BackupCrypto.decrypt(archiveData: archive, phrase: wrong))
    }
}
#endif
