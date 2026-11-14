package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The embedded trust anchor must be exactly Apple Root CA - G3.
func TestAppleRootCAG3Fingerprint(t *testing.T) {
	block, rest := pem.Decode(appleRootCAG3PEM)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatal("embedded Apple root is not a PEM certificate")
	}
	if len(bytes.TrimSpace(rest)) != 0 {
		t.Fatal("embedded Apple root file must contain exactly one certificate")
	}
	sum := sha256.Sum256(block.Bytes)
	const want = "63343abfb89a6a03ebb57e9b3f5fa7be7c4f5c756f3017b3a8c488c3653e9179"
	if got := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("Apple Root CA G3 fingerprint = %s, want %s", got, want)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != "Apple Root CA - G3" {
		t.Fatalf("unexpected subject %q", cert.Subject.CommonName)
	}
	if _, err := appleStoreKitRoots(); err != nil {
		t.Fatal(err)
	}
}

// testStoreKitChain is a throwaway root -> intermediate -> leaf chain shaped
// like Apple's App Store receipt-signing chain.
type testStoreKitChain struct {
	roots   *x509.CertPool
	x5c     []string
	leafKey *ecdsa.PrivateKey
}

type testChainOptions struct {
	omitLeafMarker         bool
	omitIntermediateMarker bool
}

func newTestStoreKitChain(t *testing.T, opts testChainOptions) testStoreKitChain {
	t.Helper()
	marker := func(oid asn1.ObjectIdentifier) []pkix.Extension {
		return []pkix.Extension{{Id: oid, Value: []byte{0x05, 0x00}}}
	}
	newKey := func(curve elliptic.Curve) *ecdsa.PrivateKey {
		key, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		return key
	}
	notBefore, notAfter := time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour)

	rootKey := newKey(elliptic.P384())
	rootTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Root CA"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTmpl, rootTmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	rootCert, _ := x509.ParseCertificate(rootDER)

	interKey := newKey(elliptic.P384())
	interTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "Test WWDR Intermediate"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	if !opts.omitIntermediateMarker {
		interTmpl.ExtraExtensions = marker(oidAppleWWDRIntermediate)
	}
	interDER, err := x509.CreateCertificate(rand.Reader, interTmpl, rootCert, &interKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	interCert, _ := x509.ParseCertificate(interDER)

	leafKey := newKey(elliptic.P256())
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "Test Receipt Signing"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	if !opts.omitLeafMarker {
		leafTmpl.ExtraExtensions = marker(oidAppleStoreReceiptLeaf)
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, interCert, &leafKey.PublicKey, interKey)
	if err != nil {
		t.Fatal(err)
	}

	roots := x509.NewCertPool()
	roots.AddCert(rootCert)
	return testStoreKitChain{
		roots:   roots,
		leafKey: leafKey,
		x5c: []string{
			base64.StdEncoding.EncodeToString(leafDER),
			base64.StdEncoding.EncodeToString(interDER),
			base64.StdEncoding.EncodeToString(rootDER),
		},
	}
}

// sign returns a StoreKit-style JWS over claims.
func (c testStoreKitChain) sign(t *testing.T, claims map[string]interface{}) string {
	t.Helper()
	header, _ := json.Marshal(map[string]interface{}{"alg": "ES256", "x5c": c.x5c})
	payload, _ := json.Marshal(claims)
	signingInput := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, c.leafKey, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func vipClaims(txnID string) map[string]interface{} {
	return map[string]interface{}{
		"transactionId": txnID,
		"bundleId":      vipBundleID,
		"productId":     vipMonthlyProductID,
		"expiresDate":   time.Now().Add(30 * 24 * time.Hour).UnixMilli(),
		"environment":   "Production",
	}
}

func postVIPSubscription(rt *Router, did string, tier int, jws string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]interface{}{
		"did":                did,
		"trust_tier":         tier,
		"evidence_type":      "vip_subscription",
		"signed_transaction": jws,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/vip-verify", bytes.NewReader(body))
	w := httptest.NewRecorder()
	rt.handleVIPVerify(w, asCaller(req, did))
	return w
}

func TestVIPVerify_SubscriptionAcceptsVerifiedTransaction(t *testing.T) {
	chain := newTestStoreKitChain(t, testChainOptions{})
	rt := &Router{storeKitRoots: chain.roots}
	// Asking for tier 5 with a real subscription still records the VIP tier.
	w := postVIPSubscription(rt, "did:key:zVIP", 5, chain.sign(t, vipClaims("verified-1")))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		TrustTier int `json:"trust_tier"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.TrustTier != vipSubscriptionTier {
		t.Fatalf("trust_tier = %d, want %d", resp.TrustTier, vipSubscriptionTier)
	}
}

// The payload-only "JWS" the endpoint used to accept: an empty header, any
// payload and a junk signature.
func forgedStoreKitJWS(txnID string) string {
	payload, _ := json.Marshal(vipClaims(txnID))
	return "e30." + base64.RawURLEncoding.EncodeToString(payload) + ".c2ln"
}

func TestVIPVerify_SubscriptionRejectsForgedTransaction(t *testing.T) {
	trusted := newTestStoreKitChain(t, testChainOptions{})
	other := newTestStoreKitChain(t, testChainOptions{})

	valid := trusted.sign(t, vipClaims("forged-valid"))
	parts := strings.Split(valid, ".")
	swapped, _ := json.Marshal(vipClaims("forged-swapped"))
	tamperedPayload := parts[0] + "." + base64.RawURLEncoding.EncodeToString(swapped) + "." + parts[2]

	noneHeader, _ := json.Marshal(map[string]interface{}{"alg": "none", "x5c": trusted.x5c})
	algNone := base64.RawURLEncoding.EncodeToString(noneHeader) + "." + parts[1] + "." + parts[2]

	type forgery struct {
		roots *x509.CertPool
		jws   string
	}
	cases := map[string]forgery{
		"payload only, no chain":        {trusted.roots, forgedStoreKitJWS("forged-1")},
		"chain from an untrusted root":  {trusted.roots, other.sign(t, vipClaims("forged-2"))},
		"payload changed after signing": {trusted.roots, tamperedPayload},
		"alg none":                      {trusted.roots, algNone},
		// With no injected root the embedded Apple root is the anchor, and a
		// self-made chain must not verify against it.
		"self-made chain against Apple root": {nil, trusted.sign(t, vipClaims("forged-5"))},
	}
	noLeaf := newTestStoreKitChain(t, testChainOptions{omitLeafMarker: true})
	cases["leaf without App Store marker"] = forgery{noLeaf.roots, noLeaf.sign(t, vipClaims("forged-3"))}
	noInter := newTestStoreKitChain(t, testChainOptions{omitIntermediateMarker: true})
	cases["intermediate without WWDR mark"] = forgery{noInter.roots, noInter.sign(t, vipClaims("forged-4"))}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rt := &Router{storeKitRoots: tc.roots}
			w := postVIPSubscription(rt, "did:key:zVIP", 3, tc.jws)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestVIPVerify_SubscriptionStillChecksClaims(t *testing.T) {
	chain := newTestStoreKitChain(t, testChainOptions{})
	rt := &Router{storeKitRoots: chain.roots}

	wrongProduct := vipClaims("claims-1")
	wrongProduct["productId"] = "com.echo.app.other"
	wrongBundle := vipClaims("claims-2")
	wrongBundle["bundleId"] = "com.someone.else"
	expired := vipClaims("claims-3")
	expired["expiresDate"] = time.Now().Add(-time.Hour).UnixMilli()

	revoked := vipClaims("claims-4")
	revoked["revocationDate"] = time.Now().Add(-time.Minute).UnixMilli()
	sandbox := vipClaims("claims-5")
	sandbox["environment"] = "Sandbox"

	for name, claims := range map[string]map[string]interface{}{
		"wrong product": wrongProduct, "wrong bundle": wrongBundle, "expired": expired,
		"refunded or revoked": revoked, "sandbox without the flag": sandbox,
	} {
		t.Run(name, func(t *testing.T) {
			w := postVIPSubscription(rt, "did:key:zVIP", 3, chain.sign(t, claims))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("want 400, got %d %s", w.Code, w.Body.String())
			}
		})
	}
}

// The unverified path exists for Xcode's local .storekit signer. It needs the
// explicit flag and is never honoured in production.
func TestVIPVerify_UnverifiedStoreKitNeedsFlagAndNonProduction(t *testing.T) {
	cases := []struct {
		name, environment, flag string
		want                    int
	}{
		{"default verifies", "", "", http.StatusBadRequest},
		{"development without flag verifies", "development", "", http.StatusBadRequest},
		{"development with flag accepts", "development", "true", http.StatusOK},
		{"production ignores the flag", "production", "true", http.StatusBadRequest},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ENVIRONMENT", tc.environment)
			t.Setenv(storeKitUnverifiedEnv, tc.flag)
			rt := &Router{}
			w := postVIPSubscription(rt, "did:key:zVIP", 3, forgedStoreKitJWS("unverified-"+string(rune('a'+i))))
			if w.Code != tc.want {
				t.Fatalf("want %d, got %d %s", tc.want, w.Code, w.Body.String())
			}
		})
	}
}

func TestVIPVerify_TransactionCannotMoveToAnotherAccount(t *testing.T) {
	chain := newTestStoreKitChain(t, testChainOptions{})
	rt := &Router{storeKitRoots: chain.roots}
	jws := chain.sign(t, vipClaims("replay-1"))

	if w := postVIPSubscription(rt, "did:key:zFirst", 3, jws); w.Code != http.StatusOK {
		t.Fatalf("first claim: %d %s", w.Code, w.Body.String())
	}
	if w := postVIPSubscription(rt, "did:key:zFirst", 3, jws); w.Code != http.StatusOK {
		t.Fatalf("same account again: %d %s", w.Code, w.Body.String())
	}
	if w := postVIPSubscription(rt, "did:key:zSecond", 3, jws); w.Code != http.StatusConflict {
		t.Fatalf("second account: want 409, got %d %s", w.Code, w.Body.String())
	}
}

// fakeVIPTxnStore stands in for Redis.
type fakeVIPTxnStore struct {
	data map[string][]byte
	err  error
}

func (f *fakeVIPTxnStore) SetNX(_ context.Context, key string, value []byte, _ time.Duration) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	if _, ok := f.data[key]; ok {
		return false, nil
	}
	f.data[key] = value
	return true, nil
}

func (f *fakeVIPTxnStore) CacheGet(_ context.Context, key string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.data["cache:"+key], nil
}

func TestClaimVIPTransaction_UsesSharedStore(t *testing.T) {
	ctx := context.Background()
	expires := time.Now().Add(time.Hour)
	store := &fakeVIPTxnStore{data: map[string][]byte{}}

	if !claimVIPTransactionIn(ctx, store, "shared-1", "did:key:zA", expires) {
		t.Fatal("first claim should succeed")
	}
	if !claimVIPTransactionIn(ctx, store, "shared-1", "did:key:zA", expires) {
		t.Fatal("the owner may present the transaction again")
	}
	if claimVIPTransactionIn(ctx, store, "shared-1", "did:key:zB", expires) {
		t.Fatal("another account must not take a claimed transaction")
	}
	// The binding lives in the shared store, not in this process.
	vipTxnMu.Lock()
	_, local := vipTxnOwner["shared-1"]
	vipTxnMu.Unlock()
	if local {
		t.Fatal("the shared store should hold the binding")
	}
	if string(store.data["cache:vip:txn:shared-1"]) != "did:key:zA" {
		t.Fatalf("store = %v", store.data)
	}

	// With the store down the process-local guard still applies.
	down := &fakeVIPTxnStore{err: errors.New("redis down")}
	if !claimVIPTransactionIn(ctx, down, "shared-2", "did:key:zA", expires) {
		t.Fatal("fallback first claim should succeed")
	}
	if claimVIPTransactionIn(ctx, down, "shared-2", "did:key:zB", expires) {
		t.Fatal("fallback must still refuse a second account")
	}
}

// TestFlight and App Review buy in Sandbox against the production backend, so
// an Apple-signed Sandbox transaction is accepted only when explicitly enabled.
func TestVIPVerify_SandboxTransactionNeedsFlag(t *testing.T) {
	chain := newTestStoreKitChain(t, testChainOptions{})
	rt := &Router{storeKitRoots: chain.roots}
	claims := vipClaims("sandbox-1")
	claims["environment"] = "Sandbox"

	t.Setenv("ENVIRONMENT", "production")
	t.Setenv(storeKitSandboxEnv, "true")
	if w := postVIPSubscription(rt, "did:key:zVIP", 3, chain.sign(t, claims)); w.Code != http.StatusOK {
		t.Fatalf("sandbox with flag: want 200, got %d %s", w.Code, w.Body.String())
	}
}
