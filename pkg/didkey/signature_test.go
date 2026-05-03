package didkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"
)

func p256Int32(i *big.Int) []byte {
	b := i.Bytes()
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

func TestVerifyECDSAP256SHA256_Raw64(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte(`{"subject_did":"did:key:zTest"}`)
	d := sha256.Sum256(msg)
	r, s, err := ecdsa.Sign(rand.Reader, priv, d[:])
	if err != nil {
		t.Fatal(err)
	}
	raw := append(p256Int32(r), p256Int32(s)...)
	if err := VerifyECDSAP256SHA256(&priv.PublicKey, msg, raw); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyECDSAP256SHA256_ASN1(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte(`hello`)
	d := sha256.Sum256(msg)
	sig, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyECDSAP256SHA256(&priv.PublicKey, msg, sig); err != nil {
		t.Fatal(err)
	}
	bad := append([]byte(nil), sig...)
	bad[3] ^= 0xff
	if VerifyECDSAP256SHA256(&priv.PublicKey, msg, bad) == nil {
		t.Fatal("expected verification failure")
	}
}

func TestVerifyECDSAP256SHA256_RoundTripWithDerivedKey(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	did, err := Derive(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"subject_did":"` + did + `","signing_did":"` + did + `"}`)
	d := sha256.Sum256(body)
	sig, err := ecdsa.SignASN1(rand.Reader, priv, d[:])
	if err != nil {
		t.Fatal(err)
	}
	parsedPub, err := Parse(did)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyECDSAP256SHA256(parsedPub, body, sig); err != nil {
		t.Fatal(err)
	}
}
