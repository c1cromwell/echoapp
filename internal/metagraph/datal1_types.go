package metagraph

// DataL1TrustCommitmentUpdate matches shared_data TrustCommitmentUpdate (T6 commitment only).
type DataL1TrustCommitmentUpdate struct {
	Commitment string `json:"commitment"`
	Epoch      int64  `json:"epoch"`
}
