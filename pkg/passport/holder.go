package passport

import (
	"time"
)

// CredentialRef is holder-side metadata for a VC the user possesses (T7 refs only — no PII).
type CredentialRef struct {
	RefID             string     `json:"ref_id"`
	HolderDID         string     `json:"holder_did"`
	IssuerDID         string     `json:"issuer_did"`
	CredentialType    string     `json:"credential_type"`
	CredentialHash    string     `json:"credential_hash"` // hex(SHA-256) of canonical VC bytes
	StatusListIndex   *int       `json:"status_list_index,omitempty"`
	StatusListCred    string     `json:"status_list_credential,omitempty"`
	RevocationStatus  string     `json:"revocation_status"` // active | revoked | unknown
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// RegisterRefRequest registers holder metadata when a VC is added to the wallet.
type RegisterRefRequest struct {
	IssuerDID       string `json:"issuer_did"`
	CredentialType  string `json:"credential_type"`
	CredentialHash  string `json:"credential_hash"`
	StatusListIndex *int   `json:"status_list_index,omitempty"`
	StatusListCred  string `json:"status_list_credential,omitempty"`
}
