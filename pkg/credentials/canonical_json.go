package credentials

import (
	"bytes"
	"encoding/json"
	"sort"
)

// canonicalJSON returns a deterministic JSON representation suitable for signing
// DataIntegrity proofs: object keys are sorted lexicographically; arrays preserve
// order. Primitive values use encoding/json (same as VC producers expect).
func canonicalJSON(v interface{}) ([]byte, error) {
	return marshalCanonical(v)
}

func marshalCanonical(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			buf.Write(keyBytes)
			buf.WriteByte(':')
			inner, err := marshalCanonical(val[k])
			if err != nil {
				return nil, err
			}
			buf.Write(inner)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil
	case []interface{}:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, e := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			inner, err := marshalCanonical(e)
			if err != nil {
				return nil, err
			}
			buf.Write(inner)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	case nil:
		return []byte("null"), nil
	default:
		return json.Marshal(v)
	}
}
