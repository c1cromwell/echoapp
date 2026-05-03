package didkey

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestParseECPrivateKeyPEM_SEC1(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	blk := &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}
	pemBytes := pem.EncodeToMemory(blk)
	got, err := ParseECPrivateKeyPEM(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if got.PublicKey.X.Cmp(priv.PublicKey.X) != 0 {
		t.Fatal("x mismatch")
	}
}

func TestParseECPrivateKeyPEM_PKCS8(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	blk := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	pemBytes := pem.EncodeToMemory(blk)
	got, err := ParseECPrivateKeyPEM(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if got.D.Cmp(priv.D) != 0 {
		t.Fatal("d mismatch")
	}
}

func TestSignAndVerifyECDSAP256(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte(`{"hello":"world"}`)
	sig, err := SignECDSAP256SHA256ASN1(priv, msg)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyECDSAP256SHA256(&priv.PublicKey, msg, sig); err != nil {
		t.Fatal(err)
	}
}
