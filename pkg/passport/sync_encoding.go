package passport

import (
	"encoding/base64"
	"strings"
)

func decodeStdBase64(raw string) ([]byte, error) {
	if strings.ContainsAny(raw, "+/") {
		return base64.StdEncoding.DecodeString(raw)
	}
	return base64.RawURLEncoding.DecodeString(raw)
}

func EncodeBase64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// DecodeCiphertextBase64 decodes client backup/sync ciphertext (Std or RawURL base64).
func DecodeCiphertextBase64(raw string) ([]byte, error) {
	return decodeStdBase64(strings.TrimSpace(raw))
}
