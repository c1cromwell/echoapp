import Foundation

/// Tracks message commitment hashes and updates status when
/// the metagraph confirms anchoring in a finalized snapshot (WO-15 / WO-227).
@MainActor
final class AnchoringTracker: ObservableObject {

    enum ConfirmationResult: Equatable {
        case anchored
        case serverConfirmed
        case verificationFailed
    }

    /// Messages pending on-chain anchoring
    @Published private(set) var pendingAnchors: [String: PendingAnchor] = [:]

    /// Anchored messages with snapshot info
    @Published private(set) var anchoredMessages: [String: AnchorConfirmation] = [:]

    struct PendingAnchor {
        let messageId: String
        let commitmentHex: String
        let submittedAt: Date
    }

    struct AnchorConfirmation: Equatable {
        let messageId: String
        let snapshotHash: String
        let snapshotHeight: Int
        let merkleRoot: String?
        let verifiedLocally: Bool
        let confirmedAt: Date
    }

    /// Track a new message commitment (hex SHA-256 from relay payload).
    func track(messageId: String, commitmentHex: String) {
        guard !commitmentHex.isEmpty else { return }
        pendingAnchors[messageId] = PendingAnchor(
            messageId: messageId,
            commitmentHex: commitmentHex,
            submittedAt: Date()
        )
    }

    /// Called when WebSocket receives a confirmation from the relay.
    @discardableResult
    func confirmAnchoring(
        messageId: String,
        snapshotHash: String,
        snapshotHeight: Int,
        merkleProof: [String]?,
        merkleRoot: String?,
        merkleLeafIndex: Int?
    ) -> ConfirmationResult {
        let pending = pendingAnchors.removeValue(forKey: messageId)
        let commitmentHex = pending?.commitmentHex

        var verifiedLocally = false
        var result: ConfirmationResult = .serverConfirmed

        if let siblings = merkleProof, !siblings.isEmpty,
           let root = merkleRoot, !root.isEmpty,
           let leaf = commitmentHex {
            verifiedLocally = MerkleProofLogic.verifyMerkleProof(
                leafHex: leaf,
                siblings: siblings,
                onChainRootHex: root,
                leafIndex: merkleLeafIndex ?? 0
            )
            result = verifiedLocally ? .anchored : .verificationFailed
        } else if merkleRoot == nil {
            result = .serverConfirmed
        }

        anchoredMessages[messageId] = AnchorConfirmation(
            messageId: messageId,
            snapshotHash: snapshotHash,
            snapshotHeight: snapshotHeight,
            merkleRoot: merkleRoot,
            verifiedLocally: verifiedLocally,
            confirmedAt: Date()
        )

        switch result {
        case .anchored, .serverConfirmed:
            NotificationCenter.default.post(
                name: .messageAnchored,
                object: nil,
                userInfo: [
                    "messageId": messageId,
                    "snapshotHash": snapshotHash,
                    "snapshotHeight": snapshotHeight,
                    "verifiedLocally": verifiedLocally,
                ]
            )
        case .verificationFailed:
            NotificationCenter.default.post(
                name: .messageAnchorVerificationFailed,
                object: nil,
                userInfo: [
                    "messageId": messageId,
                    "snapshotHash": snapshotHash,
                ]
            )
        }

        return result
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

// MARK: - Notification Names

extension Notification.Name {
    static let messageAnchored = Notification.Name("echo.message.anchored")
    static let messageVerified = Notification.Name("echo.message.verified")
    static let messageAnchorVerificationFailed = Notification.Name("echo.message.anchorFailed")
}
