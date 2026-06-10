package metagraph

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// Tessellation 4.x data-application signing (see tessellation Signed.scala + JsonBrotliBinarySerializer).
const tessellationSecp256k1SPKIPrefix = "3056301006072a8648ce3d020106052b8104000a03420004"

// IdentitySigningConfig holds Tessellation-compatible signing material for POST /data.
type IdentitySigningConfig struct {
	Secp256k1Private *secp256k1.PrivateKey
	PublicHex        string // Tessellation proof id (stripped secp256k1 SPKI hex)
	DID              string // Phase-1 authorized sender DID (env-only; not used in crypto proof)
}

// LoadIdentitySigningFromPEM loads a secp256k1 key and Tessellation proof id from PEM.
func LoadIdentitySigningFromPEM(pemBytes []byte) (IdentitySigningConfig, error) {
	return LoadTessellationSigningFromPEM(pemBytes)
}

// LoadTessellationSigningFromPEM loads a secp256k1 private key and derives the proof id Tessellation expects.
func LoadTessellationSigningFromPEM(pemBytes []byte) (IdentitySigningConfig, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return IdentitySigningConfig{}, fmt.Errorf("metagraph: no PEM block in signing key")
	}
	secpPriv, err := secp256k1PrivateKeyFromPEMBlock(block.Bytes)
	if err != nil {
		return IdentitySigningConfig{}, err
	}
	proofID, err := tessellationProofIDFromSecp256k1(secpPriv.PubKey())
	if err != nil {
		return IdentitySigningConfig{}, err
	}
	return IdentitySigningConfig{
		Secp256k1Private: secpPriv,
		PublicHex:        proofID,
	}, nil
}

func (c IdentitySigningConfig) secp256k1PrivateKey() (*secp256k1.PrivateKey, error) {
	if c.Secp256k1Private == nil {
		return nil, fmt.Errorf("metagraph: identity signing key is not configured")
	}
	return c.Secp256k1Private, nil
}

type sec1ECPrivateKey struct {
	Version    int
	PrivateKey []byte
}

func secp256k1PrivateKeyFromPEMBlock(der []byte) (*secp256k1.PrivateKey, error) {
	var sec1 sec1ECPrivateKey
	if _, err := asn1.Unmarshal(der, &sec1); err == nil && len(sec1.PrivateKey) > 0 {
		d := sec1.PrivateKey
		if len(d) > 32 {
			d = d[len(d)-32:]
		}
		return secp256k1.PrivKeyFromBytes(d), nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		if ec, ok := key.(*ecdsa.PrivateKey); ok {
			return secp256k1KeyFromECDSA(ec)
		}
	}
	return nil, fmt.Errorf("metagraph: unsupported secp256k1 PEM (openssl ecparam -name secp256k1 expected)")
}

func secp256k1KeyFromECDSA(key *ecdsa.PrivateKey) (*secp256k1.PrivateKey, error) {
	if key == nil || key.D == nil || key.PublicKey.X == nil || key.PublicKey.Y == nil {
		return nil, fmt.Errorf("metagraph: incomplete ECDSA private key")
	}
	enc := make([]byte, 32)
	key.D.FillBytes(enc)
	priv := secp256k1.PrivKeyFromBytes(enc)
	pubX, pubY := key.PublicKey.X, key.PublicKey.Y
	if priv.PubKey().X().Cmp(pubX) != 0 || priv.PubKey().Y().Cmp(pubY) != 0 {
		return nil, fmt.Errorf("metagraph: signing key must be secp256k1 (generate with: openssl ecparam -name secp256k1 -genkey)")
	}
	return priv, nil
}

func tessellationProofIDFromSecp256k1(pub *secp256k1.PublicKey) (string, error) {
	uncompressed := pub.SerializeUncompressed()
	if len(uncompressed) != 65 || uncompressed[0] != 0x04 {
		return "", fmt.Errorf("metagraph: invalid secp256k1 public key encoding")
	}
	// Tessellation Id hex is SPKI without PublicKeyHexPrefix; prefix ends with 0x04.
	return strings.ToLower(hex.EncodeToString(uncompressed[1:])), nil
}

// tessellationDataHash mirrors Identity L1 POST /data verification:
// SHA-256(IdentitySerializers.serializeUpdate) → hex → SHA-512 ECDSA sign of hex UTF-8 bytes.
// See tessellation DataApplicationRoutes.toHashedWithSignatureCheck(serializeUpdate).
func tessellationDataHash(update interface{}) ([]byte, string, error) {
	canonical, err := json.Marshal(update)
	if err != nil {
		return nil, "", err
	}
	canonical = bytes.TrimSpace(canonical)
	sum := sha256.Sum256(canonical)
	hashHex := hex.EncodeToString(sum[:])
	return canonical, hashHex, nil
}

func signTessellationDataHash(hashHex string, priv *secp256k1.PrivateKey) ([]byte, error) {
	digest := sha512.Sum512([]byte(hashHex))
	sig := secp256k1ecdsa.Sign(priv, digest[:])
	return sig.Serialize(), nil
}

