package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"strings"
	"time"
)

// appleRootCAG3PEM is Apple Root CA - G3, the trust anchor for StoreKit 2
// signed transactions. Source:
// https://www.apple.com/certificateauthority/AppleRootCA-G3.cer
// SHA-256 63343abfb89a6a03ebb57e9b3f5fa7be7c4f5c756f3017b3a8c488c3653e9179
// (asserted in TestAppleRootCAG3Fingerprint). The file is PEM; it carries a
// .crt extension because the repository ignores *.pem.
//
//go:embed certs/AppleRootCA-G3.crt
var appleRootCAG3PEM []byte

// Apple marker extensions that distinguish the App Store receipt-signing chain
// from every other certificate issued under Apple Root CA - G3. Without these
// checks any Apple-issued G3 leaf (for example a third-party Apple Pay
// certificate) could sign a "transaction".
var (
	oidAppleStoreReceiptLeaf = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}
	oidAppleWWDRIntermediate = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 2, 1}
)

// storeKitUnverifiedEnv lets a non-production gateway accept transactions from
// Xcode's local StoreKit test signer (ios/Echo/Configs/EchoVIP.storekit), whose
// chain does not lead to Apple. It is ignored when ENVIRONMENT=production.
const storeKitUnverifiedEnv = "VIP_ALLOW_UNVERIFIED_STOREKIT"

func storeKitUnverifiedAllowed() bool {
	return os.Getenv("ENVIRONMENT") != "production" && os.Getenv(storeKitUnverifiedEnv) == "true"
}

// storeKitSandboxEnv accepts Apple-signed Sandbox transactions. TestFlight
// builds and App Review both purchase in Sandbox against the production
// backend, so this must be "true" while the build is in review or beta, and
// turned off once only App Store builds should earn VIP. The local-only
// unverified bypass implies it (Xcode transactions report environment "Xcode").
const storeKitSandboxEnv = "VIP_ALLOW_SANDBOX_STOREKIT"

func storeKitSandboxAllowed() bool {
	return os.Getenv(storeKitSandboxEnv) == "true" || storeKitUnverifiedAllowed()
}

func appleStoreKitRoots() (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(appleRootCAG3PEM) {
		return nil, errors.New("embedded Apple Root CA G3 is not valid PEM")
	}
	return pool, nil
}

// storeKitRootPool returns the trust anchors for StoreKit transactions: the
// router's injected pool (tests) or the embedded Apple Root CA - G3.
func (rt *Router) storeKitRootPool() (*x509.CertPool, error) {
	if rt.storeKitRoots != nil {
		return rt.storeKitRoots, nil
	}
	return appleStoreKitRoots()
}

func hasExtension(cert *x509.Certificate, oid asn1.ObjectIdentifier) bool {
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(oid) {
			return true
		}
	}
	return false
}

// verifyStoreKitJWS checks a StoreKit 2 signed transaction (JWS compact form):
// the x5c chain must lead leaf -> intermediate -> one of roots, carry Apple's
// receipt-signing marker extensions, and the ES256 signature must verify with
// the leaf key. It returns the raw payload bytes; it does not interpret them.
func verifyStoreKitJWS(jws string, roots *x509.CertPool, now time.Time) ([]byte, error) {
	if roots == nil {
		// x509 would fall back to the system roots; never verify against those.
		return nil, errStoreKit("subscription verification is not available")
	}
	parts := strings.Split(jws, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil, errStoreKit("signed_transaction must be a StoreKit transaction")
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errStoreKit("signed_transaction header is not valid")
	}
	var header struct {
		Alg string   `json:"alg"`
		X5C []string `json:"x5c"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, errStoreKit("signed_transaction header is not valid")
	}
	if header.Alg != "ES256" {
		return nil, errStoreKit("signed_transaction must be signed with ES256")
	}
	// Apple sends leaf, intermediate, root. The root in the header is never
	// trusted; only the pinned pool anchors the chain.
	if len(header.X5C) < 2 || len(header.X5C) > 4 {
		return nil, errStoreKit("signed_transaction certificate chain is missing")
	}
	certs := make([]*x509.Certificate, 0, len(header.X5C))
	for _, b64 := range header.X5C {
		der, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, errStoreKit("signed_transaction certificate chain is not valid")
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, errStoreKit("signed_transaction certificate chain is not valid")
		}
		certs = append(certs, cert)
	}
	leaf := certs[0]
	intermediates := x509.NewCertPool()
	for _, cert := range certs[1:] {
		intermediates.AddCert(cert)
	}
	chains, err := leaf.Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   now,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if err != nil {
		return nil, errStoreKit("signed_transaction is not signed by Apple")
	}
	appleChain := false
	for _, chain := range chains {
		// leaf -> intermediate -> root, exactly.
		if len(chain) == 3 && hasExtension(chain[0], oidAppleStoreReceiptLeaf) && hasExtension(chain[1], oidAppleWWDRIntermediate) {
			appleChain = true
			break
		}
	}
	if !appleChain {
		return nil, errStoreKit("signed_transaction is not signed by an App Store certificate")
	}

	pub, ok := leaf.PublicKey.(*ecdsa.PublicKey)
	if !ok || pub.Curve != elliptic.P256() {
		return nil, errStoreKit("signed_transaction signing key is not valid")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(sig) != 64 {
		return nil, errStoreKit("signed_transaction signature is not valid")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	if !ecdsa.Verify(pub, digest[:], r, s) {
		return nil, errStoreKit("signed_transaction signature is not valid")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errStoreKit("signed_transaction payload is not valid")
	}
	return payload, nil
}

// unverifiedStoreKitPayload decodes the payload without checking who signed
// it. Only reachable behind storeKitUnverifiedAllowed.
func unverifiedStoreKitPayload(jws string) ([]byte, error) {
	parts := strings.Split(jws, ".")
	if len(parts) != 3 || parts[1] == "" {
		return nil, errStoreKit("signed_transaction must be a StoreKit transaction")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errStoreKit("signed_transaction payload is not valid")
	}
	return payload, nil
}
