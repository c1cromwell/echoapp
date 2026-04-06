package auth

import (
	"encoding/json"
)

// jsonUnmarshal wraps encoding/json.Unmarshal for use in the auth package.
func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
