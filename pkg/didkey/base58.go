package didkey

import (
	"errors"
	"math/big"
)

// base58btcAlphabet is the standard Bitcoin / multibase 'z' alphabet.
// Note: 0, O, I, l are intentionally absent.
const base58btcAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var base58btcDecode [256]int8

func init() {
	for i := range base58btcDecode {
		base58btcDecode[i] = -1
	}
	for i, c := range base58btcAlphabet {
		base58btcDecode[byte(c)] = int8(i)
	}
}

// base58Encode encodes input bytes as a base58btc string. Leading zero bytes
// are preserved as leading '1' characters.
func base58Encode(in []byte) string {
	if len(in) == 0 {
		return ""
	}

	zeros := 0
	for zeros < len(in) && in[zeros] == 0 {
		zeros++
	}

	x := new(big.Int).SetBytes(in)
	base := big.NewInt(58)
	mod := new(big.Int)

	out := make([]byte, 0, len(in)*138/100+1)
	for x.Sign() > 0 {
		x.DivMod(x, base, mod)
		out = append(out, base58btcAlphabet[mod.Int64()])
	}

	for i := 0; i < zeros; i++ {
		out = append(out, base58btcAlphabet[0])
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

// base58Decode decodes a base58btc string into its original byte slice.
// Returns an error if any character is outside the base58btc alphabet.
func base58Decode(s string) ([]byte, error) {
	if len(s) == 0 {
		return []byte{}, nil
	}

	zeros := 0
	for zeros < len(s) && s[zeros] == base58btcAlphabet[0] {
		zeros++
	}

	x := new(big.Int)
	base := big.NewInt(58)
	for i := 0; i < len(s); i++ {
		v := base58btcDecode[s[i]]
		if v < 0 {
			return nil, errors.New("didkey: invalid base58btc character")
		}
		x.Mul(x, base)
		x.Add(x, big.NewInt(int64(v)))
	}

	decoded := x.Bytes()
	out := make([]byte, zeros+len(decoded))
	copy(out[zeros:], decoded)
	return out, nil
}
