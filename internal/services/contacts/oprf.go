package contacts

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/cloudflare/circl/oprf"
)

// oprfSuite is the RFC 9497 OPRF ciphersuite used for private contact discovery.
var oprfSuite = oprf.SuiteRistretto255

// OPRFService wraps a server-side OPRF (RFC 9497, ristretto255-SHA512). It powers
// private contact discovery: clients blind their contacts' phone numbers, the
// server evaluates them under its secret key without learning the inputs, and
// clients match the results locally. The server stores only OPRF outputs, never
// raw phone numbers or reversible hashes.
//
// The key MUST be stable across restarts/instances — rotating it changes every
// OPRF output and invalidates the discovery index.
type OPRFService struct {
	server oprf.Server
}

// NewOPRFService loads the server key from CONTACT_OPRF_KEY (hex of the marshaled
// private key). When unset it generates an ephemeral key with a warning in dev,
// and fails in production. Generate a key for ops with GenerateOPRFKeyHex.
func NewOPRFService() (*OPRFService, error) {
	key, persistent, err := loadOrGenerateOPRFKey()
	if err != nil {
		return nil, err
	}
	if !persistent {
		log.Println("⚠ contacts: CONTACT_OPRF_KEY unset — using an EPHEMERAL OPRF key; the contact-discovery index is invalidated on restart. Set CONTACT_OPRF_KEY (see GenerateOPRFKeyHex) for persistence.")
	}
	return &OPRFService{server: oprf.NewServer(oprfSuite, key)}, nil
}

func loadOrGenerateOPRFKey() (*oprf.PrivateKey, bool, error) {
	if h := os.Getenv("CONTACT_OPRF_KEY"); h != "" {
		raw, err := hex.DecodeString(h)
		if err != nil {
			return nil, false, fmt.Errorf("CONTACT_OPRF_KEY invalid hex: %w", err)
		}
		key := new(oprf.PrivateKey)
		if err := key.UnmarshalBinary(oprfSuite, raw); err != nil {
			return nil, false, fmt.Errorf("CONTACT_OPRF_KEY invalid: %w", err)
		}
		return key, true, nil
	}
	if os.Getenv("ENVIRONMENT") == "production" {
		return nil, false, errors.New("CONTACT_OPRF_KEY is required in production (generate with GenerateOPRFKeyHex)")
	}
	key, err := oprf.GenerateKey(oprfSuite, rand.Reader)
	if err != nil {
		return nil, false, fmt.Errorf("generate OPRF key: %w", err)
	}
	return key, false, nil
}

// IndexKey returns the canonical OPRF index key for a registered phone number:
// hex(OPRF_k(normalize(e164))). Computed server-side from the raw number at
// registration time; the raw number is never stored.
func (o *OPRFService) IndexKey(e164 string) (string, error) {
	out, err := o.server.FullEvaluate([]byte(normalizePhone(e164)))
	if err != nil {
		return "", fmt.Errorf("oprf full-evaluate: %w", err)
	}
	return hex.EncodeToString(out), nil
}

// Evaluate runs the oblivious server step over client-blinded elements. Inputs
// and outputs are base64-encoded group elements; the server learns nothing about
// the underlying phone numbers (it only sees blinded points).
func (o *OPRFService) Evaluate(blindedB64 []string) ([]string, error) {
	g := oprfSuite.Group()
	elems := make([]oprf.Blinded, 0, len(blindedB64))
	for _, b := range blindedB64 {
		raw, err := base64.StdEncoding.DecodeString(b)
		if err != nil {
			return nil, fmt.Errorf("invalid blinded element encoding: %w", err)
		}
		el := g.NewElement()
		if err := el.UnmarshalBinary(raw); err != nil {
			return nil, fmt.Errorf("invalid blinded element: %w", err)
		}
		elems = append(elems, el)
	}

	eval, err := o.server.Evaluate(&oprf.EvaluationRequest{Elements: elems})
	if err != nil {
		return nil, fmt.Errorf("oprf evaluate: %w", err)
	}

	out := make([]string, 0, len(eval.Elements))
	for _, e := range eval.Elements {
		raw, err := e.MarshalBinaryCompress()
		if err != nil {
			return nil, fmt.Errorf("marshal evaluated element: %w", err)
		}
		out = append(out, base64.StdEncoding.EncodeToString(raw))
	}
	return out, nil
}

// GenerateOPRFKeyHex returns a fresh marshaled OPRF private key as hex, suitable
// for the CONTACT_OPRF_KEY environment variable.
func GenerateOPRFKeyHex() (string, error) {
	key, err := oprf.GenerateKey(oprfSuite, rand.Reader)
	if err != nil {
		return "", err
	}
	raw, err := key.MarshalBinary()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
