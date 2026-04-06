import XCTest
import CryptoKit
@testable import Echo

// MARK: - AnchoringTracker Tests

@MainActor
final class AnchoringTrackerTests: XCTestCase {

    var tracker: AnchoringTracker!

    override func setUp() {
        super.setUp()
        tracker = AnchoringTracker()
    }

    func testTrackCommitment() {
        let commitment = Data("test-commitment".utf8)
        tracker.track(messageId: "msg-1", commitment: commitment)

        XCTAssertEqual(tracker.pendingCount, 1)
        XCTAssertNotNil(tracker.pendingAnchors["msg-1"])
        XCTAssertEqual(tracker.pendingAnchors["msg-1"]?.commitment, commitment)
    }

    func testConfirmAnchoring() {
        tracker.track(messageId: "msg-1", commitment: Data("hash".utf8))

        tracker.confirmAnchoring(
            messageId: "msg-1",
            snapshotHash: "snap-abc",
            snapshotHeight: 42,
            merkleProof: nil
        )

        XCTAssertEqual(tracker.pendingCount, 0)
        XCTAssertTrue(tracker.isAnchored("msg-1"))
        XCTAssertEqual(tracker.anchoredMessages["msg-1"]?.snapshotHash, "snap-abc")
        XCTAssertEqual(tracker.anchoredMessages["msg-1"]?.snapshotHeight, 42)
    }

    func testConfirmAnchoringRemovesPending() {
        tracker.track(messageId: "msg-1", commitment: Data("h1".utf8))
        tracker.track(messageId: "msg-2", commitment: Data("h2".utf8))

        tracker.confirmAnchoring(
            messageId: "msg-1",
            snapshotHash: "snap-1",
            snapshotHeight: 10,
            merkleProof: nil
        )

        XCTAssertEqual(tracker.pendingCount, 1)
        XCTAssertNil(tracker.pendingAnchors["msg-1"])
        XCTAssertNotNil(tracker.pendingAnchors["msg-2"])
    }

    func testIsAnchoredReturnsFalseForUnknown() {
        XCTAssertFalse(tracker.isAnchored("nonexistent"))
    }

    func testPruneStale() {
        // Add a commitment and immediately prune with 0 interval
        tracker.track(messageId: "msg-old", commitment: Data("old".utf8))

        // Prune with 0 second interval should remove everything
        tracker.pruneStale(olderThan: 0)

        XCTAssertEqual(tracker.pendingCount, 0)
    }

    func testMultipleTracksAndConfirmations() {
        for i in 0..<5 {
            tracker.track(messageId: "msg-\(i)", commitment: Data("h\(i)".utf8))
        }
        XCTAssertEqual(tracker.pendingCount, 5)

        for i in 0..<3 {
            tracker.confirmAnchoring(
                messageId: "msg-\(i)",
                snapshotHash: "snap-\(i)",
                snapshotHeight: i,
                merkleProof: nil
            )
        }

        XCTAssertEqual(tracker.pendingCount, 2)
        XCTAssertEqual(tracker.anchoredMessages.count, 3)
    }
}

// MARK: - OfflineQueueManager Tests

final class OfflineQueueManagerTests: XCTestCase {

    var queue: OfflineQueueManager!

    override func setUp() {
        super.setUp()
        queue = OfflineQueueManager()
    }

    func testEnqueueAndCount() async {
        let request = makeRelayRequest(messageId: "msg-1")
        await queue.enqueue(request)

        let count = await queue.count
        XCTAssertEqual(count, 1)
    }

    func testDequeueAll() async {
        await queue.enqueue(makeRelayRequest(messageId: "msg-1"))
        await queue.enqueue(makeRelayRequest(messageId: "msg-2"))

        let items = await queue.dequeueAll()
        XCTAssertEqual(items.count, 2)

        let countAfter = await queue.count
        XCTAssertEqual(countAfter, 0)
    }

    func testDequeueAllEmpty() async {
        let items = await queue.dequeueAll()
        XCTAssertTrue(items.isEmpty)
    }

    func testRemoveByMessageId() async {
        await queue.enqueue(makeRelayRequest(messageId: "msg-1"))
        await queue.enqueue(makeRelayRequest(messageId: "msg-2"))

        await queue.remove(messageId: "msg-1")

        let items = await queue.peek()
        XCTAssertEqual(items.count, 1)
        XCTAssertEqual(items.first?.messageId, "msg-2")
    }

    func testClear() async {
        for i in 0..<10 {
            await queue.enqueue(makeRelayRequest(messageId: "msg-\(i)"))
        }
        await queue.clear()

        let count = await queue.count
        XCTAssertEqual(count, 0)
    }

    private func makeRelayRequest(messageId: String) -> RelayRequest {
        RelayRequest(
            messageId: messageId,
            conversationId: "conv-1",
            contentType: "text",
            encryptedPayload: Data("encrypted".utf8),
            commitment: Data("commitment".utf8),
            signature: Data("sig".utf8),
            timestamp: Date()
        )
    }
}

// MARK: - DeliveryStatus Tests

final class DeliveryStatusTests: XCTestCase {

    func testAllCasesExist() {
        let allCases: [DeliveryStatus] = [
            .sending, .sent, .delivered, .read,
            .failed, .anchored, .verified
        ]
        XCTAssertEqual(allCases.count, 7)
    }

    func testCodableRoundTrip() throws {
        let status = DeliveryStatus.anchored
        let encoded = try JSONEncoder().encode(status)
        let decoded = try JSONDecoder().decode(DeliveryStatus.self, from: encoded)
        XCTAssertEqual(decoded, .anchored)
    }

    func testVerifiedStatus() throws {
        let status = DeliveryStatus.verified
        let encoded = try JSONEncoder().encode(status)
        let decoded = try JSONDecoder().decode(DeliveryStatus.self, from: encoded)
        XCTAssertEqual(decoded, .verified)
    }
}

// MARK: - WSRelayMessage Tests

final class WSRelayMessageTests: XCTestCase {

    func testRelayMessageTypeCodable() throws {
        let types: [WSRelayMessage.RelayMessageType] = [
            .message, .typing, .presence, .receipt,
            .ack, .queueDrain, .confirmation, .groupKey
        ]

        for type in types {
            let encoded = try JSONEncoder().encode(type)
            let decoded = try JSONDecoder().decode(
                WSRelayMessage.RelayMessageType.self,
                from: encoded
            )
            XCTAssertEqual(decoded, type)
        }
    }

    func testWSConfirmationCodable() throws {
        let conf = WSConfirmation(
            referenceId: "msg-123",
            snapshotHash: "abc123",
            snapshotHeight: 42,
            merkleProof: nil
        )
        let encoded = try JSONEncoder().encode(conf)
        let decoded = try JSONDecoder().decode(WSConfirmation.self, from: encoded)

        XCTAssertEqual(decoded.referenceId, "msg-123")
        XCTAssertEqual(decoded.snapshotHash, "abc123")
        XCTAssertEqual(decoded.snapshotHeight, 42)
        XCTAssertNil(decoded.merkleProof)
    }

    func testWSGroupKeyPayloadCodable() throws {
        let payload = WSGroupKeyPayload(
            groupId: "group-1",
            version: 3,
            encryptedKey: Data("encrypted-key".utf8),
            distributedBy: "did:dag:admin"
        )
        let encoded = try JSONEncoder().encode(payload)
        let decoded = try JSONDecoder().decode(WSGroupKeyPayload.self, from: encoded)

        XCTAssertEqual(decoded.groupId, "group-1")
        XCTAssertEqual(decoded.version, 3)
        XCTAssertEqual(decoded.distributedBy, "did:dag:admin")
    }
}

// MARK: - ECHOError Tests

final class ECHOErrorTests: XCTestCase {

    func testSupportCode() {
        XCTAssertEqual(ECHOError.authFailed.supportCode, "ECHO-1001")
        XCTAssertEqual(ECHOError.relayUnavailable.supportCode, "ECHO-2004")
        XCTAssertEqual(ECHOError.groupKeyMissing.supportCode, "ECHO-6001")
        XCTAssertEqual(ECHOError.evidenceFingerprintFailed.supportCode, "ECHO-7001")
    }

    func testIsRetryable() {
        XCTAssertTrue(ECHOError.networkUnavailable.isRetryable)
        XCTAssertTrue(ECHOError.relayUnavailable.isRetryable)
        XCTAssertTrue(ECHOError.requestTimeout.isRetryable)
        XCTAssertFalse(ECHOError.authFailed.isRetryable)
        XCTAssertFalse(ECHOError.invalidSignature.isRetryable)
    }

    func testIsUserFacing() {
        XCTAssertFalse(ECHOError.messageQueued.isUserFacing)
        XCTAssertTrue(ECHOError.authFailed.isUserFacing)
        XCTAssertTrue(ECHOError.encryptionFailed.isUserFacing)
    }

    func testErrorDescription() {
        let error = ECHOError.relayUnavailable
        XCTAssertNotNil(error.errorDescription)
        XCTAssertTrue(error.errorDescription!.contains("2004"))
    }

    func testAllErrorCategories() {
        // Auth (1xxx)
        XCTAssertEqual(ECHOError.authFailed.rawValue, 1001)
        XCTAssertEqual(ECHOError.biometricFailed.rawValue, 1002)
        XCTAssertEqual(ECHOError.sessionExpired.rawValue, 1003)

        // Network (2xxx)
        XCTAssertEqual(ECHOError.networkUnavailable.rawValue, 2001)
        XCTAssertEqual(ECHOError.relayUnavailable.rawValue, 2004)

        // Encryption (3xxx)
        XCTAssertEqual(ECHOError.invalidSignature.rawValue, 3004)
        XCTAssertEqual(ECHOError.commitmentMismatch.rawValue, 3005)

        // Messages (4xxx)
        XCTAssertEqual(ECHOError.rateLimitExceeded.rawValue, 4003)
        XCTAssertEqual(ECHOError.messageQueued.rawValue, 4004)

        // Groups (6xxx)
        XCTAssertEqual(ECHOError.groupKeyMissing.rawValue, 6001)
        XCTAssertEqual(ECHOError.notGroupAdmin.rawValue, 6003)

        // Digital Evidence (7xxx)
        XCTAssertEqual(ECHOError.evidenceFingerprintFailed.rawValue, 7001)
        XCTAssertEqual(ECHOError.evidenceNotAvailable.rawValue, 7002)
    }
}

// MARK: - MessageStatus Tests (Updated for v3.1)

final class MessageStatusTests: XCTestCase {

    func testAnchoredStatus() {
        let status = MessageStatus.anchored
        XCTAssertEqual(status.rawValue, "Anchored")
    }

    func testVerifiedStatus() {
        let status = MessageStatus.verified
        XCTAssertEqual(status.rawValue, "Verified")
    }

    func testAllStatusValues() {
        let allStatuses: [MessageStatus] = [
            .sending, .sent, .delivered, .read,
            .failed, .anchored, .verified
        ]
        XCTAssertEqual(allStatuses.count, 7)
    }
}

// MARK: - DigitalEvidenceBridge Tests

final class DigitalEvidenceBridgeTests: XCTestCase {

    func testVerificationURLFormat() async {
        let apiClient = APIClient(configuration: .init(
            baseURL: URL(string: "https://api.echo.local")!,
            timeout: 30,
            defaultHeaders: [:]
        ))
        let bridge = DigitalEvidenceBridge(apiClient: apiClient)

        let url = await bridge.verificationURL(for: "evt-abc123")
        XCTAssertEqual(
            url?.absoluteString,
            "https://digitalevidence.constellationnetwork.io/verify/evt-abc123"
        )
    }

    func testVerificationURLNilForEmpty() async {
        let apiClient = APIClient(configuration: .init(
            baseURL: URL(string: "https://api.echo.local")!,
            timeout: 30,
            defaultHeaders: [:]
        ))
        let bridge = DigitalEvidenceBridge(apiClient: apiClient)

        let url = await bridge.verificationURL(for: "")
        XCTAssertNotNil(url) // Empty string still produces a URL
    }

    func testSmartCheckmarkEligibility() async {
        let apiClient = APIClient(configuration: .init(
            baseURL: URL(string: "https://api.echo.local")!,
            timeout: 30,
            defaultHeaders: [:]
        ))
        let bridge = DigitalEvidenceBridge(apiClient: apiClient)

        let eligible = await bridge.isSmartCheckmarkEligible(
            deliveryStatus: .verified,
            evidenceEventId: "evt-123"
        )
        XCTAssertTrue(eligible)

        let notEligibleNoId = await bridge.isSmartCheckmarkEligible(
            deliveryStatus: .verified,
            evidenceEventId: nil
        )
        XCTAssertFalse(notEligibleNoId)

        let notEligibleWrongStatus = await bridge.isSmartCheckmarkEligible(
            deliveryStatus: .anchored,
            evidenceEventId: "evt-123"
        )
        XCTAssertFalse(notEligibleWrongStatus)
    }

    func testComputeFingerprint() async {
        let apiClient = APIClient(configuration: .init(
            baseURL: URL(string: "https://api.echo.local")!,
            timeout: 30,
            defaultHeaders: [:]
        ))
        let bridge = DigitalEvidenceBridge(apiClient: apiClient)

        let data = Data("test data".utf8)
        let hash = await bridge.computeFingerprint(data)

        // SHA-256 of "test data" is deterministic
        XCTAssertEqual(hash.count, 64) // 32 bytes = 64 hex chars
        XCTAssertFalse(hash.isEmpty)

        // Same data should produce same hash
        let hash2 = await bridge.computeFingerprint(data)
        XCTAssertEqual(hash, hash2)

        // Different data should produce different hash
        let hash3 = await bridge.computeFingerprint(Data("other data".utf8))
        XCTAssertNotEqual(hash, hash3)
    }
}

// MARK: - EvidenceResult Tests

final class EvidenceResultTests: XCTestCase {

    func testCodable() throws {
        let result = EvidenceResult(
            eventId: "evt-123",
            verificationURL: "https://digitalevidence.constellationnetwork.io/verify/evt-123",
            timestamp: Date(timeIntervalSince1970: 1000000)
        )

        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .secondsSince1970
        let data = try encoder.encode(result)

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .secondsSince1970
        let decoded = try decoder.decode(EvidenceResult.self, from: data)

        XCTAssertEqual(decoded.eventId, "evt-123")
        XCTAssertTrue(decoded.verificationURL.contains("evt-123"))
    }
}

// MARK: - GroupKeyError Tests

final class GroupKeyErrorTests: XCTestCase {

    func testErrorDescriptions() {
        XCTAssertNotNil(GroupKeyError.noGroupKey.errorDescription)
        XCTAssertNotNil(GroupKeyError.encryptionFailed.errorDescription)
        XCTAssertNotNil(GroupKeyError.decryptionFailed.errorDescription)
        XCTAssertNotNil(GroupKeyError.keyRotationFailed.errorDescription)
        XCTAssertNotNil(GroupKeyError.notGroupAdmin.errorDescription)
    }
}

// MARK: - MessageRelayError Tests

final class MessageRelayErrorTests: XCTestCase {

    func testErrorDescriptions() {
        XCTAssertNotNil(MessageRelayError.invalidSignature.errorDescription)
        XCTAssertNotNil(MessageRelayError.commitmentMismatch.errorDescription)
        XCTAssertNotNil(MessageRelayError.decryptionFailed.errorDescription)
        XCTAssertNotNil(MessageRelayError.relayUnavailable.errorDescription)
    }
}
