import CryptoKit
import Foundation

/// Client-side Merkle proof verification (WO-227) — matches Go `parentHash` / `BuildMerkleTree`.
enum MerkleProofLogic {
    static func parentHash(left: String, right: String) -> String {
        let digest = SHA256.hash(data: Data((left + right).utf8))
        return digest.map { String(format: "%02x", $0) }.joined()
    }

    /// Walks sibling hashes from `leafHex` to the root using `leafIndex` for left/right ordering.
    static func verifyMerkleProof(
        leafHex: String,
        siblings: [String],
        onChainRootHex: String,
        leafIndex: Int = 0
    ) -> Bool {
        guard !leafHex.isEmpty, !onChainRootHex.isEmpty else { return false }
        var computed = leafHex
        var idx = leafIndex
        for sibling in siblings {
            if idx % 2 == 0 {
                computed = parentHash(left: computed, right: sibling)
            } else {
                computed = parentHash(left: sibling, right: computed)
            }
            idx /= 2
        }
        return computed.lowercased() == onChainRootHex.lowercased()
    }
}
