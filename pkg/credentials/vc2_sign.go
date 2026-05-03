package credentials

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

func (i *Issuer) effectiveIssuerDID() string {
	if i.config.MetagraphConfig.IssuerDID != "" {
		return i.config.MetagraphConfig.IssuerDID
	}
	return i.config.IssuerConfig.IssuerDID
}

func (i *Issuer) verificationMethodForIssuer() string {
	issuerDID := i.effectiveIssuerDID()
	frag := i.config.IssuerConfig.PublicKeyID
	if frag == "" && strings.HasPrefix(issuerDID, "did:key:") {
		frag = strings.TrimPrefix(issuerDID, "did:key:")
	}
	return fmt.Sprintf("%s#%s", issuerDID, frag)
}

func (i *Issuer) loadIssuerPrivateKey() (*ecdsa.PrivateKey, error) {
	i.privMu.Lock()
	defer i.privMu.Unlock()
	if i.privKey != nil {
		return i.privKey, nil
	}
	if i.config.IssuerConfig.PrivateKeyPath == "" {
		return nil, NewCredentialError(ErrCodeInvalidCredential, "issuer private key path is required for VC 2.0 DataIntegrity")
	}
	pemBytes, err := os.ReadFile(i.config.IssuerConfig.PrivateKeyPath)
	if err != nil {
		return nil, NewCredentialErrorWithDetails(ErrCodeInvalidCredential, "read issuer private key", err.Error())
	}
	key, err := didkey.ParseECPrivateKeyPEM(pemBytes)
	if err != nil {
		return nil, NewCredentialErrorWithDetails(ErrCodeInvalidCredential, "parse issuer private key PEM", err.Error())
	}
	if i.config.CredentialConfig.UseW3CVC2 {
		derived, derr := didkey.Derive(&key.PublicKey)
		if derr != nil {
			return nil, NewCredentialErrorWithDetails(ErrCodeInvalidCredential, "derive did:key from issuer key", derr.Error())
		}
		if want := i.effectiveIssuerDID(); want != "" && derived != want {
			return nil, NewCredentialErrorWithDetails(
				ErrCodeInvalidCredential,
				"issuer private key does not match configured issuer DID (expected "+want+", derived "+derived+")",
				"",
			)
		}
	}
	i.privKey = key
	return i.privKey, nil
}

func (i *Issuer) signVC2DataIntegrity(vc *VerifiableCredential) error {
	doc, err := vc2CredentialDocMap(vc, false)
	if err != nil {
		return err
	}
	canon, err := canonicalJSON(doc)
	if err != nil {
		return NewCredentialErrorWithDetails(ErrCodeInvalidProof, "canonical JSON for VC2", err.Error())
	}
	priv, err := i.loadIssuerPrivateKey()
	if err != nil {
		return err
	}
	sig, err := didkey.SignECDSAP256SHA256ASN1(priv, canon)
	if err != nil {
		return NewCredentialErrorWithDetails(ErrCodeInvalidProof, "sign VC2 DataIntegrity", err.Error())
	}
	vc.Proof.ProofValue = base64.RawURLEncoding.EncodeToString(sig)
	return nil
}
