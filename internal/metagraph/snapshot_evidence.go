package metagraph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// SignedSnapshotInfo is what the gateway (and cmd/verifyproof) can read out of
// a Tessellation signed snapshot's JSON body, `{"value": {...}, "proofs": [...]}`,
// without Tessellation's own serialization.
//
// What this does and does not establish:
//   - Signers are the peer ids listed in `proofs`. The secp256k1 signatures
//     are NOT verified here: they are over the snapshot hash, which is computed
//     from Tessellation's canonical serialization of `value`. Reproducing that
//     is out of scope, so signer ids are claims carried by the body.
//   - For a Metagraph L0 snapshot, `value.dataApplication.onChainState` is a
//     JSON array of byte values. For Echo's data application those bytes are
//     UTF-8 JSON of DataLayerOnChainState (see metagraph DataSerializers), whose
//     `merkleRoots` map is keyed by root. DataRoots holds those keys, so root
//     presence in the snapshot body can be checked directly.
//   - A Global L0 snapshot carries metagraph snapshots only as compressed
//     binary under `value.stateChannelSnapshots`; nothing is decoded from it.
type SignedSnapshotInfo struct {
	Ordinal int64
	Signers []string
	// HasDataState is true when a data-application on-chain state was present
	// and decoded as Echo's DataLayerOnChainState JSON.
	HasDataState bool
	DataRoots    map[string]struct{}
}

// ContainsRoot reports whether root is in the decoded on-chain state.
func (i SignedSnapshotInfo) ContainsRoot(root string) bool {
	_, ok := i.DataRoots[root]
	return ok
}

// ParseSignedSnapshot extracts the ordinal, signer ids and (when present) the
// data-application merkle roots from a signed snapshot JSON body.
func ParseSignedSnapshot(body []byte) (SignedSnapshotInfo, error) {
	var doc struct {
		Value *struct {
			Ordinal         int64 `json:"ordinal"`
			DataApplication *struct {
				OnChainState []int `json:"onChainState"`
			} `json:"dataApplication"`
		} `json:"value"`
		Proofs []struct {
			ID string `json:"id"`
		} `json:"proofs"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return SignedSnapshotInfo{}, fmt.Errorf("decode signed snapshot: %w", err)
	}
	if doc.Value == nil {
		return SignedSnapshotInfo{}, fmt.Errorf("signed snapshot has no value")
	}
	info := SignedSnapshotInfo{Ordinal: doc.Value.Ordinal, Signers: []string{}}
	seen := map[string]struct{}{}
	for _, p := range doc.Proofs {
		if p.ID == "" {
			continue
		}
		if _, dup := seen[p.ID]; !dup {
			seen[p.ID] = struct{}{}
			info.Signers = append(info.Signers, p.ID)
		}
	}
	sort.Strings(info.Signers)

	if da := doc.Value.DataApplication; da != nil && len(da.OnChainState) > 0 {
		raw := make([]byte, len(da.OnChainState))
		for i, v := range da.OnChainState {
			raw[i] = byte(v) // circe emits signed bytes; byte() wraps negatives
		}
		var state struct {
			MerkleRoots map[string]json.RawMessage `json:"merkleRoots"`
		}
		// A state that is not Echo's JSON layout is left undecoded rather than
		// treated as an error: the snapshot is still archivable evidence.
		if err := json.Unmarshal(raw, &state); err == nil && state.MerkleRoots != nil {
			info.HasDataState = true
			info.DataRoots = make(map[string]struct{}, len(state.MerkleRoots))
			for k := range state.MerkleRoots {
				info.DataRoots[k] = struct{}{}
			}
		}
	}
	return info, nil
}

// SHA256Hex returns the lowercase hex SHA-256 of data.
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
