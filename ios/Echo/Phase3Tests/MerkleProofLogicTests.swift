import XCTest
@testable import Echo

final class MerkleProofLogicTests: XCTestCase {
    func testProofVerifiesRoot_leafZero() {
        let leaves = [
            String(repeating: "a", count: 64),
            String(repeating: "b", count: 64),
        ]
        let root = MerkleProofLogic.parentHash(left: leaves[0], right: leaves[1])
        XCTAssertTrue(
            MerkleProofLogic.verifyMerkleProof(
                leafHex: leaves[0],
                siblings: [leaves[1]],
                onChainRootHex: root,
                leafIndex: 0
            )
        )
    }

    func testProofFailsOnTamperedSibling() {
        let leaves = [
            String(repeating: "a", count: 64),
            String(repeating: "b", count: 64),
        ]
        let root = MerkleProofLogic.parentHash(left: leaves[0], right: leaves[1])
        XCTAssertFalse(
            MerkleProofLogic.verifyMerkleProof(
                leafHex: leaves[0],
                siblings: [String(repeating: "c", count: 64)],
                onChainRootHex: root,
                leafIndex: 0
            )
        )
    }
}

#if os(iOS)
@MainActor
final class AnchoringTrackerVerificationTests: XCTestCase {
    func testConfirmAnchoring_verifiesMerkleProof() {
        let tracker = AnchoringTracker()
        let leaves = [
            String(repeating: "a", count: 64),
            String(repeating: "b", count: 64),
        ]
        let root = MerkleProofLogic.parentHash(left: leaves[0], right: leaves[1])
        tracker.track(messageId: "msg-1", commitmentHex: leaves[0])

        let result = tracker.confirmAnchoring(
            messageId: "msg-1",
            snapshotHash: "tx-1",
            snapshotHeight: 1,
            merkleProof: [leaves[1]],
            merkleRoot: root,
            merkleLeafIndex: 0
        )

        XCTAssertEqual(result, .anchored)
        XCTAssertTrue(tracker.anchoredMessages["msg-1"]?.verifiedLocally == true)
    }

    func testConfirmAnchoring_detectsTamperedProof() {
        let tracker = AnchoringTracker()
        let leaf = String(repeating: "a", count: 64)
        tracker.track(messageId: "msg-2", commitmentHex: leaf)

        let result = tracker.confirmAnchoring(
            messageId: "msg-2",
            snapshotHash: "tx-2",
            snapshotHeight: 2,
            merkleProof: [String(repeating: "f", count: 64)],
            merkleRoot: String(repeating: "0", count: 64),
            merkleLeafIndex: 0
        )

        XCTAssertEqual(result, .verificationFailed)
    }
}

final class AnchorConfirmationCodecTests: XCTestCase {
    func testDecodeConfirmationEnvelope() {
        let json = """
        {"type":"confirmation","to":"did:key:me","payload":{"type":"confirmation","messageId":"m-1","snapshotHash":"snap","snapshotHeight":3,"merkleProof":["bb"],"merkleRoot":"root","merkleLeafIndex":0}}
        """
        let payload = AnchorConfirmationCodec.decodeConfirmation(from: json)
        XCTAssertEqual(payload?.messageId, "m-1")
        XCTAssertEqual(payload?.snapshotHash, "snap")
        XCTAssertEqual(payload?.merkleProof?.count, 1)
    }
}
#endif
