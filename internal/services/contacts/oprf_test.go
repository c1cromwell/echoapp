package contacts

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/cloudflare/circl/oprf"
)

// newTestOPRFService builds an OPRFService with a fixed key so results are stable.
func newTestOPRFService(t *testing.T) *OPRFService {
	t.Helper()
	keyHex, err := GenerateOPRFKeyHex()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONTACT_OPRF_KEY", keyHex)
	svc, err := NewOPRFService()
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

// clientFinalize simulates the iOS client: blind the inputs, run them through the
// server's oblivious Evaluate, and finalize — returning hex(OPRF_k(input)) for
// each input, exactly as a client would compute them locally.
func clientFinalize(t *testing.T, svc *Service, inputs []string) []string {
	t.Helper()
	client := oprf.NewClient(oprfSuite)

	in := make([][]byte, len(inputs))
	for i, s := range inputs {
		in[i] = []byte(normalizePhone(s))
	}
	finData, evalReq, err := client.Blind(in)
	if err != nil {
		t.Fatalf("blind: %v", err)
	}

	blindedB64 := make([]string, len(evalReq.Elements))
	for i, el := range evalReq.Elements {
		raw, err := el.MarshalBinaryCompress()
		if err != nil {
			t.Fatal(err)
		}
		blindedB64[i] = base64.StdEncoding.EncodeToString(raw)
	}

	evaluatedB64, err := svc.OPRFEvaluate(blindedB64)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	g := oprfSuite.Group()
	evalElems := make([]oprf.Evaluated, len(evaluatedB64))
	for i, b := range evaluatedB64 {
		raw, err := base64.StdEncoding.DecodeString(b)
		if err != nil {
			t.Fatal(err)
		}
		el := g.NewElement()
		if err := el.UnmarshalBinary(raw); err != nil {
			t.Fatal(err)
		}
		evalElems[i] = el
	}

	outputs, err := client.Finalize(finData, &oprf.Evaluation{Elements: evalElems})
	if err != nil {
		t.Fatalf("finalize: %v", err)
	}
	hexOut := make([]string, len(outputs))
	for i, o := range outputs {
		hexOut[i] = hex.EncodeToString(o)
	}
	return hexOut
}

// TestOPRF_ObliviousEqualsServerEvaluate is the correctness proof: the oblivious
// client path (blind -> server Evaluate -> finalize) yields exactly the same
// value as the server computing OPRF_k(input) directly — so client outputs can
// be matched against the server-built index.
func TestOPRF_ObliviousEqualsServerEvaluate(t *testing.T) {
	oprfSvc := newTestOPRFService(t)
	svc := NewService(nil)
	svc.SetOPRF(oprfSvc)

	phone := "+1 555-0100"
	got := clientFinalize(t, svc, []string{phone})[0]
	want, err := svc.DiscoveryKey(phone)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("oblivious output %s != server FullEvaluate %s", got, want)
	}
}

// TestOPRF_DiscoveryFindsRegisteredOnly verifies a registered number is found via
// the client-side index match and an unregistered one is not.
func TestOPRF_DiscoveryFindsRegisteredOnly(t *testing.T) {
	svc := NewService(nil)
	svc.SetOPRF(newTestOPRFService(t))

	if err := svc.RegisterPhoneForDiscovery("did:key:zAlice", "+15551111"); err != nil {
		t.Fatal(err)
	}

	outputs := clientFinalize(t, svc, []string{"+15551111", "+15559999"})
	index := svc.DiscoveryIndex()

	if did, ok := index[outputs[0]]; !ok || did != "did:key:zAlice" {
		t.Fatalf("registered number should resolve to did:key:zAlice, got ok=%v did=%q", ok, did)
	}
	if _, ok := index[outputs[1]]; ok {
		t.Fatal("unregistered number must not be in the discovery index")
	}
}

func TestOPRF_EvaluateRequiresService(t *testing.T) {
	svc := NewService(nil) // no OPRF wired
	if _, err := svc.OPRFEvaluate([]string{"x"}); err != ErrOPRFUnavailable {
		t.Fatalf("expected ErrOPRFUnavailable, got %v", err)
	}
}

func TestOPRF_KeyPersistsAcrossInstances(t *testing.T) {
	keyHex, err := GenerateOPRFKeyHex()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONTACT_OPRF_KEY", keyHex)

	s1, err := NewOPRFService()
	if err != nil {
		t.Fatal(err)
	}
	s2, err := NewOPRFService() // simulated restart, same key
	if err != nil {
		t.Fatal(err)
	}
	k1, _ := s1.IndexKey("+15551234")
	k2, _ := s2.IndexKey("+15551234")
	if k1 != k2 || k1 == "" {
		t.Fatalf("same key must yield same index key across instances: %q vs %q", k1, k2)
	}
}

func TestOPRF_ProductionRequiresKey(t *testing.T) {
	t.Setenv("CONTACT_OPRF_KEY", "")
	t.Setenv("ENVIRONMENT", "production")
	if _, err := NewOPRFService(); err == nil {
		t.Fatal("production without CONTACT_OPRF_KEY must fail")
	}
}
