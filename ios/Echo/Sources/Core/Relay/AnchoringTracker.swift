import Foundation

/// Tracks message commitment hashes and updates status when
/// the metagraph confirms anchoring in a finalized snapshot.
@MainActor
final class AnchoringTracker: ObservableObject {

    /// Messages pending on-chain anchoring
    @Published private(set) var pendingAnchors: [String: PendingAnchor] = [:]

    /// Anchored messages with snapshot info
    @Published private(set) var anchoredMessages: [String: AnchorConfirmation] = [:]

    struct PendingAnchor {
        let messageId: String
        let commitment: Data
        let submittedAt: Date
    }

    struct AnchorConfirmation {
        let messageId: String
        let snapshotHash: String
        let snapshotHeight: Int
        let confirmedAt: Date
    }

    /// Track a new message commitment
    func track(messageId: String, commitment: Data) {
        pendingAnchors[messageId] = PendingAnchor(
            messageId: messageId,
            commitment: commitment,
            submittedAt: Date()
        )
    }

    /// Called when WebSocket receives a confirmation from the relay
    /// (type: "confirmation" with snapshotHash and optional Merkle proof)
    func confirmAnchoring(
        messageId: String,
        snapshotHash: String,
        snapshotHeight: Int,
        merkleProof: [Data]?
    ) {
        pendingAnchors.removeValue(forKey: messageId)

        anchoredMessages[messageId] = AnchorConfirmation(
            messageId: messageId,
            snapshotHash: snapshotHash,
            snapshotHeight: snapshotHeight,
            confirmedAt: Date()
        )

        // Phase 3+: Verify Merkle proof locally
        if let proof = merkleProof, !proof.isEmpty {
            // verifyMerkleInclusion(commitment, proof, snapshotHash)
        }

        // Notify UI to update message delivery status to .anchored
        NotificationCenter.default.post(
            name: .messageAnchored,
            object: nil,
            userInfo: [
                "messageId": messageId,
                "snapshotHash": snapshotHash,
                "snapshotHeight": snapshotHeight
            ]
        )
    }

    /// Number of messages awaiting anchoring
    var pendingCount: Int {
        pendingAnchors.count
    }

    /// Check if a message has been anchored
    func isAnchored(_ messageId: String) -> Bool {
        anchoredMessages[messageId] != nil
    }

    /// Remove stale pending anchors older than the given interval
    func pruneStale(olderThan interval: TimeInterval = 3600) {
        let cutoff = Date().addingTimeInterval(-interval)
        pendingAnchors = pendingAnchors.filter { $0.value.submittedAt > cutoff }
    }
}

// MARK: - Notification Name

extension Notification.Name {
    static let messageAnchored = Notification.Name("echo.message.anchored")
    static let messageVerified = Notification.Name("echo.message.verified")
}
