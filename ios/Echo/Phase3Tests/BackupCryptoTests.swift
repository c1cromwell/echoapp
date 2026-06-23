import XCTest
@testable import Echo

#if os(iOS)
final class BackupCryptoTests: XCTestCase {
    // Valid 24-word BIP-39 mnemonic (all-zero entropy) — passes checksum so `entropy` is non-nil.
    private let stubWords = Array(repeating: "abandon", count: 23) + ["art"]
    // A different valid mnemonic (all-ones entropy) for wrong-key rejection.
    private let otherWords = Array(repeating: "zoo", count: 23) + ["vote"]

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
        // A different valid recovery phrase derives a different key → decrypt must fail.
        guard let wrong = RecoveryPhrase(words: otherWords) else {
            XCTFail("other phrase invalid")
            return
        }
        XCTAssertThrowsError(try BackupCrypto.decrypt(archiveData: archive, phrase: wrong))
    }
}
#endif
