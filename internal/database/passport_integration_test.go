//go:build integration

package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thechadcromwell/echoapp/pkg/passport"
)

func TestPassportCredentialRefPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	db := &PostgresDB{pool: pool}

	holder := "did:key:zIntegrationPassportHolder"
	ref := passport.CredentialRef{
		RefID:          "integration-ref-1",
		HolderDID:      holder,
		IssuerDID:      "did:key:zIssuer",
		CredentialType: "ProofOfHumanity",
		CredentialHash: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if err := db.InsertCredentialRef(ctx, ref); err != nil {
		t.Fatalf("insert: %v", err)
	}
	got, err := db.GetCredentialRef(ctx, holder, ref.RefID)
	if err != nil || got == nil {
		t.Fatalf("get: %v", err)
	}
	if got.CredentialHash != ref.CredentialHash {
		t.Fatalf("hash mismatch")
	}
}
