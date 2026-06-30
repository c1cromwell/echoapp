package wallet

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"math/big"
	"strconv"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// pkcsPrefix is dag4's CONSTANTS.PKCS_PREFIX prepended to the uncompressed
// public key before hashing (SPKI DER header up to, but not including, the 0x04
// point prefix — the key hex supplies that). Matches the metagraph signer's
// tessellationSecp256k1SPKIPrefix (which folds the 0x04 in).
const pkcsPrefix = "3056301006072a8648ce3d020106052b8104000a034200"

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// DagAddressFromPubKey ports dag4 keyStore.getDagAddressFromPublicKey: prepend
// the PKCS prefix to the uncompressed pubkey, SHA-256, base58-encode, take the
// last 36 chars, and prefix "DAG" + (digitSum % 9). This is the user-held
// address the backend must bind to; it is the single source of truth the proof
// verifier checks against.
func DagAddressFromPubKey(pubHex string) string {
	if len(pubHex) == 128 { // raw X||Y without the 0x04 point prefix
		pubHex = "04" + pubHex
	}
	raw, err := hex.DecodeString(pkcsPrefix + pubHex)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	encoded := base58Encode(sum[:])
	if len(encoded) < 36 {
		return ""
	}
	end := encoded[len(encoded)-36:]
	digitSum := 0
	for _, c := range end {
		if c >= '0' && c <= '9' {
			digitSum += int(c - '0')
		}
	}
	return "DAG" + strconv.Itoa(digitSum%9) + end
}

// VerifyDagMessageSignature verifies a dag4 keyStore.sign() signature: secp256k1
// ECDSA (DER) over SHA-512(message), matching the iOS bundle's signMessage and
// the metagraph signer's SHA-512 digest convention (signTessellationDataHash).
func VerifyDagMessageSignature(pubHex, message, sigDERHex string) bool {
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return false
	}
	pub, err := secp256k1.ParsePubKey(pubBytes)
	if err != nil {
		return false
	}
	sigBytes, err := hex.DecodeString(sigDERHex)
	if err != nil {
		return false
	}
	sig, err := secp256k1ecdsa.ParseDERSignature(sigBytes)
	if err != nil {
		return false
	}
	digest := sha512.Sum512([]byte(message))
	return sig.Verify(digest[:], pub)
}

func base58Encode(input []byte) string {
	x := new(big.Int).SetBytes(input)
	radix := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)
	var out []byte
	for x.Cmp(zero) > 0 {
		x.DivMod(x, radix, mod)
		out = append(out, base58Alphabet[mod.Int64()])
	}
	// Preserve leading zero bytes as '1'.
	for _, b := range input {
		if b != 0 {
			break
		}
		out = append(out, base58Alphabet[0])
	}
	// Reverse.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}
