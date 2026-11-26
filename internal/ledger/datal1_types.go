package ledger

import (
	"crypto/sha256"
	"encoding/hex"
)

// DataL1TrustCommitmentUpdate matches shared_data TrustCommitmentUpdate (T6 commitment only).
type DataL1TrustCommitmentUpdate struct {
	Commitment string `json:"commitment"`
	Epoch      int64  `json:"epoch"`
}

// DataL1MerkleRootUpdate is one Merkle root to anchor, with the number of leaves under it.
type DataL1MerkleRootUpdate struct {
	Root      string `json:"root"`
	LeafCount int    `json:"leafCount"`
}

// ComputeMerkleRoot computes a SHA-256 Merkle root from a set of commitment hashes (hex strings,
// hashed pairwise as text). Used for media content roots; message batches use BuildMerkleTree.
func ComputeMerkleRoot(commitments []string) string {
	if len(commitments) == 0 {
		return ""
	}
	if len(commitments) == 1 {
		return commitments[0]
	}
	if len(commitments)%2 != 0 {
		commitments = append(commitments, commitments[len(commitments)-1])
	}
	var nextLevel []string
	for i := 0; i < len(commitments); i += 2 {
		hash := sha256.Sum256([]byte(commitments[i] + commitments[i+1]))
		nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
	}
	return ComputeMerkleRoot(nextLevel)
}
