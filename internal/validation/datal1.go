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
