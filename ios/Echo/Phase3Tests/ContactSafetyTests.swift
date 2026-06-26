import XCTest
@testable import Echo

final class LookalikeDetectorTests: XCTestCase {
    func testPureHomoglyphSwapMatches() {
        // "Ηello" uses a Greek capital Eta (U+0397) in place of Latin H.
        XCTAssertTrue(LookalikeDetector.isLookalike("Ηello", of: "Hello"))
        // Cyrillic 'а'/'о' in "Аlіcо" mimicking "Alice"/"Alico".
        XCTAssertTrue(LookalikeDetector.isLookalike("Аlice", of: "Alice"))
    }

    func testSmallEditDistanceMatches() {
        XCTAssertTrue(LookalikeDetector.isLookalike("Alicia", of: "Alicia "))
        XCTAssertTrue(LookalikeDetector.isLookalike("Jon Smith", of: "John Smith"))
    }

    func testDifferentNamesDoNotMatch() {
        XCTAssertFalse(LookalikeDetector.isLookalike("Bob", of: "Alice"))
        XCTAssertFalse(LookalikeDetector.isLookalike("Charlie", of: "Dave"))
    }

    func testCaseAndDiacriticsAreFolded() {
        XCTAssertTrue(LookalikeDetector.isLookalike("josé", of: "JOSE"))
    }
}

final class ContactSafetyEvaluatorTests: XCTestCase {
    private let eval = ContactSafetyEvaluator(trustedTierThreshold: 3)
    private let trustedAlice = KnownContact(did: "did:key:alice", displayName: "Alice", tier: 4)

    func testVerifiedKnownContactIsOk() {
        let a = eval.evaluate(peerDID: "did:key:bob", peerName: "Bob", peerTier: 4,
                              isFirstContact: false, knownContacts: [trustedAlice])
        XCTAssertEqual(a.level, .ok)
        XCTAssertTrue(a.reasons.isEmpty)
    }

    func testUnverifiedFirstContactIsCaution() {
        let a = eval.evaluate(peerDID: "did:key:x", peerName: "Stranger", peerTier: 0,
                              isFirstContact: true, knownContacts: [])
        XCTAssertEqual(a.level, .caution)
        XCTAssertTrue(a.reasons.contains(.firstContact))
        XCTAssertTrue(a.reasons.contains(.unverifiedSender))
    }

    func testImpersonationOfTrustedContactIsWarning() {
        // Low-trust peer named like trusted Alice (homoglyph), different DID.
        let a = eval.evaluate(peerDID: "did:key:imposter", peerName: "Аlice", peerTier: 0,
                              isFirstContact: true, knownContacts: [trustedAlice])
        XCTAssertEqual(a.level, .warning)
        XCTAssertTrue(a.reasons.contains(.possibleImpersonation(of: "Alice")))
    }

    func testSameNameSameDIDIsNotImpersonation() {
        // The real Alice reconnecting — same DID, no impersonation flag.
        let a = eval.evaluate(peerDID: "did:key:alice", peerName: "Alice", peerTier: 4,
                              isFirstContact: false, knownContacts: [trustedAlice])
        XCTAssertFalse(a.reasons.contains(where: {
            if case .possibleImpersonation = $0 { return true } else { return false }
        }))
    }

    func testPayGuardRequiresConfirmForLowTrust() {
        XCTAssertTrue(eval.requiresExtraConfirmationToPay(peerTier: 0))
        XCTAssertTrue(eval.requiresExtraConfirmationToPay(peerTier: 1))
        XCTAssertFalse(eval.requiresExtraConfirmationToPay(peerTier: 3))
    }
}

final class FirstContactStoreTests: XCTestCase {
    func testMarksAndDetects() {
        let d = UserDefaults(suiteName: "safety.firstcontact.test")!
        defer { d.removePersistentDomain(forName: "safety.firstcontact.test") }
        XCTAssertTrue(FirstContactStore.isFirstContact("did:key:a", defaults: d))
        FirstContactStore.markSeen("did:key:a", defaults: d)
        XCTAssertFalse(FirstContactStore.isFirstContact("did:key:a", defaults: d))
        XCTAssertTrue(FirstContactStore.isFirstContact("did:key:b", defaults: d))
    }
}
