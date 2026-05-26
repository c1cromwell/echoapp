// Package echooprf exposes RFC 9497 OPRF (ristretto255-SHA512) client operations
// for iOS via gomobile bind. Matches internal/services/contacts/oprf.go server suite.
package echooprf

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"sync"

	"github.com/cloudflare/circl/oprf"
	"github.com/google/uuid"
)

var suite = oprf.SuiteRistretto255

// Client performs blind → finalize for contact discovery PSI (WO-221).
type Client struct {
	mu       sync.Mutex
	sessions map[string]*oprf.FinalizeData
}

// NewClient creates an OPRF client with in-memory finalize sessions.
func NewClient() *Client {
	return &Client{sessions: make(map[string]*oprf.FinalizeData)}
}

// BlindResult holds blinded elements and a session token for finalize.
type BlindResult struct {
	SessionID string
	Blinded   []string
}

// BlindPhones blinds phone strings for POST /v3/contacts/psi.
func (c *Client) BlindPhones(phones []string) (*BlindResult, error) {
	if len(phones) == 0 {
		return nil, errors.New("no inputs")
	}
	oprfClient := oprf.NewClient(suite)
	inputs := make([][]byte, len(phones))
	for i, p := range phones {
		inputs[i] = []byte(normalizePhone(p))
	}
	fin, evalReq, err := oprfClient.Blind(inputs)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	c.mu.Lock()
	c.sessions[id] = fin
	c.mu.Unlock()

	blinded := make([]string, len(evalReq.Elements))
	for i, el := range evalReq.Elements {
		raw, err := el.MarshalBinaryCompress()
		if err != nil {
			return nil, err
		}
		blinded[i] = base64.StdEncoding.EncodeToString(raw)
	}
	return &BlindResult{SessionID: id, Blinded: blinded}, nil
}

// FinalizePhones completes OPRF after server evaluation; returns hex index keys.
func (c *Client) FinalizePhones(sessionID string, evaluatedB64 []string) ([]string, error) {
	c.mu.Lock()
	fin, ok := c.sessions[sessionID]
	if ok {
		delete(c.sessions, sessionID)
	}
	c.mu.Unlock()
	if !ok {
		return nil, errors.New("unknown or expired OPRF session")
	}

	oprfClient := oprf.NewClient(suite)
	g := suite.Group()
	evalElems := make([]oprf.Evaluated, len(evaluatedB64))
	for i, b := range evaluatedB64 {
		raw, err := base64.StdEncoding.DecodeString(b)
		if err != nil {
			return nil, err
		}
		el := g.NewElement()
		if err := el.UnmarshalBinary(raw); err != nil {
			return nil, err
		}
		evalElems[i] = el
	}
	outputs, err := oprfClient.Finalize(fin, &oprf.Evaluation{Elements: evalElems})
	if err != nil {
		return nil, err
	}
	hexOut := make([]string, len(outputs))
	for i, o := range outputs {
		hexOut[i] = hex.EncodeToString(o)
	}
	return hexOut, nil
}

func normalizePhone(phone string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	return replacer.Replace(phone)
}
