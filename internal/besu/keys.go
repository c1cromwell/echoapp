package besu

import (
	"crypto/ecdsa"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// sec1Key is the ASN.1 layout of an "EC PRIVATE KEY" PEM block (RFC 5915).
type sec1Key struct {
	Version    int
	PrivateKey []byte
	Parameters asn1.RawValue  `asn1:"optional,explicit,tag:0"`
	PublicKey  asn1.BitString `asn1:"optional,explicit,tag:1"`
}

// pkcs8Key is the ASN.1 layout of a "PRIVATE KEY" PEM block wrapping a SEC1 key.
type pkcs8Key struct {
	Version    int
	Algorithm  pkix.AlgorithmIdentifier
	PrivateKey []byte
}

// ParseKeyPEM reads a secp256k1 private key from PEM, as written by
// `openssl ecparam -name secp256k1 -genkey -noout`. This is the gateway's signing key
// (IDENTITY_SERVICE_KEY_PEM); on Besu its address is what the contracts authorise.
func ParseKeyPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("besu: no PEM block in signing key")
	}
	der := block.Bytes
	var p8 pkcs8Key
	if _, err := asn1.Unmarshal(der, &p8); err == nil && len(p8.PrivateKey) > 0 {
		der = p8.PrivateKey
	}
	var sec1 sec1Key
	if _, err := asn1.Unmarshal(der, &sec1); err != nil || len(sec1.PrivateKey) == 0 {
		return nil, errors.New("besu: signing key is not a SEC1 or PKCS#8 EC private key")
	}
	key, err := crypto.ToECDSA(common.LeftPadBytes(sec1.PrivateKey, 32))
	if err != nil {
		return nil, fmt.Errorf("besu: signing key is not a valid secp256k1 key: %w", err)
	}
	return key, nil
}

// LoadKeyFile reads a private key from a PEM file, or from a file holding 64 hex characters.
func LoadKeyFile(path string) (*ecdsa.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if text := strings.TrimPrefix(strings.TrimSpace(string(raw)), "0x"); len(text) == 64 {
		if b, err := hex.DecodeString(text); err == nil {
			return crypto.ToECDSA(b)
		}
	}
	return ParseKeyPEM(raw)
}

// ParseAddresses reads a comma-separated list of addresses, rejecting anything malformed.
func ParseAddresses(list string) ([]common.Address, error) {
	var out []common.Address
	for _, item := range strings.Split(list, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if !common.IsHexAddress(item) {
			return nil, fmt.Errorf("besu: %q is not an address", item)
		}
		out = append(out, common.HexToAddress(item))
	}
	return out, nil
}
