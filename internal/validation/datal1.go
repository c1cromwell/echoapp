// Package validation provides server-side pre-validation for Data L1 and reward
// submissions. The metagraph L1 remains the authoritative validator; this layer
// provides a fast-reject tier that reduces unnecessary chain transactions.
//
// # T0–T7 Data Classification (WO-217)
//
// All functions in this package operate exclusively on T5 and T6 data:
//   - DataL1Submission.MerkleRoot — T5: SHA-256 hash commitment of a message batch
//   - TrustCommitmentUpdate.Commitment — T6: H(trust_score|nonce) commitment
//
// T0/T1 secrets, T2 ciphertext, T3 relay blobs, and all PII must never reach
// these functions. Strip sensitive fields before constructing any submission type.
// See docs/data-classification.md for the complete classification guide.
package validation

import (
	"errors"
	"fmt"
	"time"
)

// Errors returned by Data L1 pre-validation (WO-35).
var (
	ErrInvalidMerkleRoot = errors.New("validation: merkle_root must be exactly 32 bytes")
	ErrEmptyBatch        = errors.New("validation: commitment_count must be > 0")
	ErrInvalidTimeRange  = errors.New("validation: time_range: from must be before to")
	ErrUnsupportedSchema = errors.New("validation: schema_version is newer than supported")
)

// TimeRange is an inclusive logical window for a batch commitment.
type TimeRange struct {
	From time.Time
	To   time.Time
}

// DataL1Submission is the payload shape the backend pre-validates before a Data L1 POST (WO-35).
type DataL1Submission struct {
	MerkleRoot      []byte
	CommitmentCount int
	TimeRange       TimeRange
	SchemaVersion   int
}

// ValidateDataL1Submission performs server-side pre-checks; L1 remains authoritative.
func ValidateDataL1Submission(sub DataL1Submission, currentSchemaVersion int) error {
	if len(sub.MerkleRoot) != 32 {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidMerkleRoot, len(sub.MerkleRoot))
	}
	if sub.CommitmentCount <= 0 {
		return fmt.Errorf("%w: got %d", ErrEmptyBatch, sub.CommitmentCount)
	}
	if sub.TimeRange.From.After(sub.TimeRange.To) {
		return ErrInvalidTimeRange
	}
	if sub.SchemaVersion > currentSchemaVersion {
		return fmt.Errorf("%w: %d > %d", ErrUnsupportedSchema, sub.SchemaVersion, currentSchemaVersion)
	}
	return nil
}
