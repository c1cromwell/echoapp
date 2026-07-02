package metagraph

import "testing"

func TestBuildMerkleTree_ProofVerifiesRoot(t *testing.T) {
	leaves := []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	tree := BuildMerkleTree(leaves)
	if tree.Root == "" {
		t.Fatal("empty root")
	}
	siblings := tree.ProofSiblings(0)
	if len(siblings) != 1 || siblings[0] != leaves[1] {
		t.Fatalf("unexpected siblings: %v", siblings)
	}
	recomputed := parentHash(leaves[0], siblings[0])
	if recomputed != tree.Root {
		t.Fatalf("proof path root %s != tree root %s", recomputed, tree.Root)
	}
}
