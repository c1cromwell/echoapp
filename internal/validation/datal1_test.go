package validation

import (
	"errors"
	"testing"
	"time"
)

func TestValidateDataL1Submission(t *testing.T) {
	const schema = 3
	baseTime := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
	goodRoot := make([]byte, 32)
	for i := range goodRoot {
		goodRoot[i] = byte(i)
	}

	t.Run("ok", func(t *testing.T) {
		err := ValidateDataL1Submission(DataL1Submission{
			MerkleRoot:      goodRoot,
			CommitmentCount: 10,
			TimeRange:       TimeRange{From: baseTime, To: baseTime.Add(time.Hour)},
			SchemaVersion:   2,
		}, schema)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("equal time endpoints ok", func(t *testing.T) {
		err := ValidateDataL1Submission(DataL1Submission{
			MerkleRoot:      goodRoot,
			CommitmentCount: 1,
			TimeRange:       TimeRange{From: baseTime, To: baseTime},
			SchemaVersion:   1,
		}, schema)
		if err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("bad merkle length", func(t *testing.T) {
		err := ValidateDataL1Submission(DataL1Submission{
			MerkleRoot:      []byte{1, 2, 3},
			CommitmentCount: 1,
			TimeRange:       TimeRange{From: baseTime, To: baseTime.Add(time.Minute)},
			SchemaVersion:   1,
		}, schema)
		if !errors.Is(err, ErrInvalidMerkleRoot) {
			t.Fatalf("want ErrInvalidMerkleRoot, got %v", err)
		}
	})

	t.Run("zero commitments", func(t *testing.T) {
		err := ValidateDataL1Submission(DataL1Submission{
			MerkleRoot:      goodRoot,
			CommitmentCount: 0,
			TimeRange:       TimeRange{From: baseTime, To: baseTime.Add(time.Minute)},
			SchemaVersion:   1,
		}, schema)
		if !errors.Is(err, ErrEmptyBatch) {
			t.Fatalf("want ErrEmptyBatch, got %v", err)
		}
	})

	t.Run("inverted time range", func(t *testing.T) {
		err := ValidateDataL1Submission(DataL1Submission{
			MerkleRoot:      goodRoot,
			CommitmentCount: 1,
			TimeRange:       TimeRange{From: baseTime.Add(time.Hour), To: baseTime},
			SchemaVersion:   1,
		}, schema)
		if !errors.Is(err, ErrInvalidTimeRange) {
			t.Fatalf("want ErrInvalidTimeRange, got %v", err)
		}
	})

	t.Run("future schema", func(t *testing.T) {
		err := ValidateDataL1Submission(DataL1Submission{
			MerkleRoot:      goodRoot,
			CommitmentCount: 1,
			TimeRange:       TimeRange{From: baseTime, To: baseTime.Add(time.Minute)},
			SchemaVersion:   schema + 1,
		}, schema)
		if !errors.Is(err, ErrUnsupportedSchema) {
			t.Fatalf("want ErrUnsupportedSchema, got %v", err)
		}
	})
}
