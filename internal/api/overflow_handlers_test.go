package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

func TestOverflowBlob_RetrieveByURI(t *testing.T) {
	stub := encblob.NewStubStorage()
	uri, err := stub.Store(t.Context(), []byte(`{"type":"text","payload":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}

	h := &V3Handlers{OverflowStorage: stub}
	mux := http.NewServeMux()
	mux.HandleFunc("/v3/relay/overflow/", h.handleOverflowBlob)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v3/relay/overflow/"+uri, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		StorageURI       string `json:"storage_uri"`
		CiphertextBase64 string `json:"ciphertext_base64"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.StorageURI != uri || resp.CiphertextBase64 == "" {
		t.Fatalf("unexpected response %+v", resp)
	}
}
