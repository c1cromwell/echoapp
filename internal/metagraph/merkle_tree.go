package metagraph

import (
	"crypto/sha256"
	"encoding/hex"
)

// MerkleTree holds layered hashes for proof generation (WO-15).
type MerkleTree struct {
	Root   string
	Layers [][]string
}

// parentHash combines two sibling hex hashes (SHA-256 of concatenation).
func parentHash(left, right string) string {
	sum := sha256.Sum256([]byte(left + right))
	return hex.EncodeToString(sum[:])
}

// BuildMerkleTree constructs a Merkle tree from leaf hex hashes and returns
// the root plus per-leaf sibling paths for proof verification (WO-227).
func BuildMerkleTree(leafHashes []string) *MerkleTree {
	if len(leafHashes) == 0 {
		return &MerkleTree{}
	}
	layers := [][]string{append([]string(nil), leafHashes...)}
	current := append([]string(nil), leafHashes...)
	for len(current) > 1 {
		if len(current)%2 != 0 {
			current = append(current, current[len(current)-1])
		}
		next := make([]string, 0, len(current)/2)
		for i := 0; i < len(current); i += 2 {
			next = append(next, parentHash(current[i], current[i+1]))
		}
		layers = append(layers, next)
		current = next
	}
	return &MerkleTree{Root: current[0], Layers: layers}
}

// ProofSiblings returns the sibling hashes from leaf index to the root.
func (t *MerkleTree) ProofSiblings(leafIndex int) []string {
	if t == nil || len(t.Layers) == 0 || leafIndex < 0 || leafIndex >= len(t.Layers[0]) {
		return nil
	}
	idx := leafIndex
	var siblings []string
	for level := 0; level < len(t.Layers)-1; level++ {
		layer := t.Layers[level]
		sibIdx := idx ^ 1
		if sibIdx < len(layer) {
			siblings = append(siblings, layer[sibIdx])
		}
		idx /= 2
	}
	return siblings
}
