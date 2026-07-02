#if os(iOS)
import Foundation

/// Cryptographic proof that a message existed on-chain (WO-94).
struct DisappearingMessageProof: Equatable, Sendable {
    let messageId: String
    let conversationId: String
    let senderDID: String
    let recipientDID: String
    let timestamp: Date
    let expiresAt: Date?
    let snapshotHash: String
    let snapshotHeight: Int
    let merkleRoot: String
    let merkleProof: [String]?
    let commitmentHex: String?
    let proofType: ProofType

    enum ProofType: String, Codable, Sendable {
        case beforeDeletion
        case afterDeletion
    }

    var verificationURL: URL? {
        URL(string: "https://dagexplorer.io/snapshot/\(snapshotHash)")
    }

    /// Builds a before-deletion proof from a live message row + anchor confirmation.
    static func beforeDeletion(
        messageId: String,
        conversationId: String,
        senderDID: String,
        recipientDID: String,
        timestamp: Date,
        expiresAt: Date?,
        commitmentHex: String,
        anchor: MessageAnchorProofResponse
    ) -> DisappearingMessageProof {
        DisappearingMessageProof(
            messageId: messageId,
            conversationId: conversationId,
            senderDID: senderDID,
            recipientDID: recipientDID,
            timestamp: timestamp,
            expiresAt: expiresAt,
            snapshotHash: anchor.snapshotHash,
            snapshotHeight: Int(anchor.snapshotHeight ?? 0),
            merkleRoot: anchor.merkleRoot,
            merkleProof: anchor.siblings,
            commitmentHex: commitmentHex,
            proofType: .beforeDeletion
        )
    }

    /// After local deletion — existence proof without content or siblings.
    static func afterDeletion(
        messageId: String,
        conversationId: String,
        senderDID: String,
        recipientDID: String,
        timestamp: Date,
        expiresAt: Date?,
        snapshotHash: String,
        snapshotHeight: Int,
        merkleRoot: String
    ) -> DisappearingMessageProof {
        DisappearingMessageProof(
            messageId: messageId,
            conversationId: conversationId,
            senderDID: senderDID,
            recipientDID: recipientDID,
            timestamp: timestamp,
            expiresAt: expiresAt,
            snapshotHash: snapshotHash,
            snapshotHeight: snapshotHeight,
            merkleRoot: merkleRoot,
            merkleProof: nil,
            commitmentHex: nil,
            proofType: .afterDeletion
        )
    }

    func verifiesAgainstRoot() -> Bool {
        guard proofType == .beforeDeletion,
              let commitmentHex,
              let siblings = merkleProof,
              !siblings.isEmpty else {
            return false
        }
        return MerkleProofLogic.verifyMerkleProof(
            leafHex: commitmentHex,
            siblings: siblings,
            onChainRootHex: merkleRoot,
            leafIndex: 0
        )
    }
}
#endif
