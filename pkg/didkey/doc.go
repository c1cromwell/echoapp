// Package didkey implements the W3C did:key method for the P-256 curve.
//
// did:key is a self-certifying DID method that embeds the public key directly
// in the DID identifier, so resolution is purely local (no chain transaction,
// no network round-trip). This package supports the P-256 (secp256r1)
// elliptic curve, which is the curve used by the iOS Secure Enclave for the
// Echo app's user identities.
//
// Encoding:
//
//	did:key:z<multibase-base58btc(varint(0x1200) || compressed-public-key)>
//
//	where:
//	  - 0x1200             is the multicodec identifier for "p256-pub"
//	  - varint(0x1200)     encodes to the bytes [0x80, 0x24]
//	  - compressed-public  is the 33-byte SEC1 compressed P-256 point
//	                       (0x02 or 0x03 prefix + 32-byte big-endian X coord)
//	  - 'z' is the multibase prefix for base58btc
//
// References:
//
//   - https://w3c-ccg.github.io/did-method-key/
//   - https://github.com/multiformats/multicodec/blob/master/table.csv
//   - W3C did-key Method Specification §3.1.2 P-256
//
// This package is intentionally dependency-free (only the Go standard library)
// so it can be reused on the iOS side via gomobile if needed and so the
// derivation rule remains auditable.
package didkey
