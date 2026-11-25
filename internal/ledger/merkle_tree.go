package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

// VerifyMerklePath recomputes the root from a leaf and the sibling path that
// ProofSiblings produced, using the same hashing as BuildMerkleTree.
//
// leafCount is required: when a layer has an odd number of nodes, the last
// node is paired with itself and ProofSiblings emits no sibling for that
// level, so the verifier needs the layer sizes to know where that happened.
func VerifyMerklePath(commitment string, siblings []string, leafIndex, leafCount int, root string) error {
	if leafCount <= 0 {
		return fmt.Errorf("leaf count must be positive, got %d", leafCount)
	}
	if leafIndex < 0 || leafIndex >= leafCount {
		return fmt.Errorf("leaf index %d out of range for %d leaves", leafIndex, leafCount)
	}
	node := commitment
	idx, size, used := leafIndex, leafCount, 0
	for size > 1 {
		switch sib := idx ^ 1; {
		case sib >= size:
			// Last node of an odd layer: duplicated, no sibling in the path.
			node = parentHash(node, node)
		case used >= len(siblings):
			return fmt.Errorf("merkle path too short: need more than %d siblings", len(siblings))
		case idx%2 == 0:
			node = parentHash(node, siblings[used])
			used++
		default:
			node = parentHash(siblings[used], node)
			used++
		}
		idx /= 2
		size = (size + 1) / 2
	}
	if used != len(siblings) {
		return fmt.Errorf("merkle path too long: %d siblings supplied, %d used", len(siblings), used)
	}
	if node != root {
		return fmt.Errorf("merkle path recomputes to %s, not root %s", node, root)
	}
	return nil
}
