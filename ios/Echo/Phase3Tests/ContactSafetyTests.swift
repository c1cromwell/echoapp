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

// MARK: - Safety reports (POST /v1/safety/reports)

final class SafetyReportRequestTests: XCTestCase {
    private func json(_ request: SafetyReportRequest) throws -> [String: Any] {
        let data = try JSONEncoder().encode(request)
        return try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: Any])
    }

    func testReasonsMatchTheServerAllowList() {
        XCTAssertEqual(
            Set(SafetyReportReason.allCases.map(\.rawValue)),
            ["spam", "harassment", "impersonation", "other"]
        )
    }

    func testContactReportSendsOnlyTheOriginalFields() throws {
        let body = try json(SafetyReportTarget.contact(did: "did:key:bob").request(reason: .spam, details: "  "))
        XCTAssertEqual(body["subject_did"] as? String, "did:key:bob")
        XCTAssertEqual(body["reason"] as? String, "spam")
        XCTAssertEqual(body["context"] as? String, "contact")
        XCTAssertEqual(Set(body.keys), ["subject_did", "reason", "context"])
    }

    func testMessageReportCarriesMessageAndConversation() throws {
        let target = SafetyReportTarget.message(id: "m-1", senderDID: "did:key:bob", conversationId: "conv-1")
        let body = try json(target.request(reason: .harassment, details: " rude "))
        XCTAssertEqual(body["context"] as? String, "message")
        XCTAssertEqual(body["subject_did"] as? String, "did:key:bob")
        XCTAssertEqual(body["message_id"] as? String, "m-1")
        XCTAssertEqual(body["conversation_id"] as? String, "conv-1")
        XCTAssertEqual(body["details"] as? String, "rude")
        XCTAssertNil(body["channel_id"])
    }

    func testChannelAndGroupReportsUseTheChannelContext() throws {
        let channel = try json(SafetyReportTarget.channel(id: "ch-9", ownerDID: "did:key:owner").request(reason: .other, details: nil))
        XCTAssertEqual(channel["context"] as? String, "channel")
        XCTAssertEqual(channel["channel_id"] as? String, "ch-9")
        XCTAssertNil(channel["conversation_id"])

        let group = try json(SafetyReportTarget.group(id: "g-3", ownerDID: "did:key:owner").request(reason: .spam, details: nil))
        XCTAssertEqual(group["context"] as? String, "channel")
        XCTAssertEqual(group["conversation_id"] as? String, "group:g-3")
        XCTAssertNil(group["channel_id"])
    }

    func testDetailsAreCapped() {
        let long = String(repeating: "x", count: SafetyReportRequest.maxDetailsLength + 50)
        let request = SafetyReportRequest(subjectDID: "did:key:bob", reason: .other, context: .contact, details: long)
        XCTAssertEqual(request.details?.count, SafetyReportRequest.maxDetailsLength)
    }
}

@MainActor
final class SafetyReportViewModelTests: XCTestCase {
    private struct Boom: LocalizedError { var errorDescription: String? { "The network is down." } }

    func testNothingIsSentUntilAReasonIsChosen() async {
        var sent: [SafetyReportRequest] = []
        let vm = SafetyReportViewModel(target: .contact(did: "did:key:bob"), submit: { sent.append($0) })
        XCTAssertFalse(vm.canSubmit)
        await vm.submitReport()
        XCTAssertTrue(sent.isEmpty)
        XCTAssertEqual(vm.phase, .editing)
    }

    func testSubmitSendsTheReportAndSucceeds() async {
        var sent: [SafetyReportRequest] = []
        var blocked: [String] = []
        let vm = SafetyReportViewModel(
            target: .message(id: "m-1", senderDID: "did:key:bob", conversationId: "conv-1"),
            submit: { sent.append($0) },
            block: { blocked.append($0) }
        )
        vm.selectedReason = .spam
        await vm.submitReport()
        XCTAssertEqual(vm.phase, .succeeded)
        XCTAssertEqual(sent.count, 1)
        XCTAssertEqual(sent.first?.messageId, "m-1")
        XCTAssertTrue(blocked.isEmpty, "a plain report must not block")
        XCTAssertFalse(vm.canSubmit, "a sent report cannot be sent twice")
    }

    func testFailureIsShownAndCanBeRetried() async {
        var attempts = 0
        let vm = SafetyReportViewModel(target: .contact(did: "did:key:bob"), submit: { _ in
            attempts += 1
            if attempts == 1 { throw Boom() }
        })
        vm.selectedReason = .other
        await vm.submitReport()
        XCTAssertEqual(vm.errorMessage, "The network is down.")
        XCTAssertTrue(vm.canSubmit)
        await vm.submitReport()
        XCTAssertEqual(vm.phase, .succeeded)
        XCTAssertEqual(attempts, 2)
    }

    func testBlockAndReportBlocksThenReports() async {
        var order: [String] = []
        let vm = SafetyReportViewModel(
            target: .contact(did: "did:key:bob", conversationId: "conv-1", alsoBlocks: true),
            submit: { _ in order.append("report") },
            block: { order.append("block:\($0)") }
        )
        vm.selectedReason = .harassment
        await vm.submitReport()
        XCTAssertEqual(order, ["block:did:key:bob", "report"])
        XCTAssertEqual(vm.phase, .succeeded)
    }

    func testReportFailureAfterBlockSaysSoAndRetryDoesNotBlockAgain() async {
        var blocks = 0
        var reports = 0
        let vm = SafetyReportViewModel(
            target: .contact(did: "did:key:bob", alsoBlocks: true),
            submit: { _ in
                reports += 1
                if reports == 1 { throw Boom() }
            },
            block: { _ in blocks += 1 }
        )
        vm.selectedReason = .spam
        await vm.submitReport()
        XCTAssertEqual(vm.errorMessage, "They're blocked, but the report wasn't sent. The network is down.")
        await vm.submitReport()
        XCTAssertEqual(vm.phase, .succeeded)
        XCTAssertEqual(blocks, 1)
        XCTAssertEqual(reports, 2)
    }

    func testBlockFailureSendsNoReport() async {
        var reports = 0
        let vm = SafetyReportViewModel(
            target: .contact(did: "did:key:bob", alsoBlocks: true),
            submit: { _ in reports += 1 },
            block: { _ in throw Boom() }
        )
        vm.selectedReason = .spam
        await vm.submitReport()
        XCTAssertEqual(vm.errorMessage, "The network is down.")
        XCTAssertEqual(reports, 0)
    }
}
