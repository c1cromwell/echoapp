package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/media"
	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

func TestMediaUploadAndChunkDownload(t *testing.T) {
	db := database.NewMemoryDB()
	store := media.NewMemoryStorage()
	svc := media.NewService(db, store)
	h := &V3Handlers{DB: db, Media: svc}
	mux := http.NewServeMux()
	h.RegisterV3Routes(mux)

	ciphertext := []byte("encrypted-media-bytes")
	req := httptest.NewRequest(http.MethodPost, "/v3/media/upload", bytes.NewReader(ciphertext))
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("X-Encrypted-Size", "23")
	req.Header.Set("X-Trust-Tier", "3")
	req.Header.Set("X-Sender-DID", "did:key:alice")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload status %d: %s", rec.Code, rec.Body.String())
	}
	var upload media.UploadResult
	if err := json.Unmarshal(rec.Body.Bytes(), &upload); err != nil {
		t.Fatal(err)
	}
	if upload.FileID == "" || upload.ChunkCount != 1 {
		t.Fatalf("unexpected upload result: %+v", upload)
	}

	chunkReq := httptest.NewRequest(http.MethodGet, "/v3/media/"+upload.FileID+"/chunks/0", nil)
	chunkReq.Header.Set("X-Sender-DID", "did:key:alice")
	chunkRec := httptest.NewRecorder()
	mux.ServeHTTP(chunkRec, chunkReq)
	if chunkRec.Code != http.StatusOK {
		t.Fatalf("chunk status %d: %s", chunkRec.Code, chunkRec.Body.String())
	}
	body, _ := io.ReadAll(chunkRec.Body)
	if !bytes.Equal(body, ciphertext) {
		t.Fatalf("chunk bytes mismatch: got %q", body)
	}
}

func TestWSOfflineQueue_OverflowToEncblob(t *testing.T) {
	q := newWSOfflineQueue()
	stub := encblob.NewStubStorage()
	q.SetOverflowStorage(stub)

	for i := 0; i < wsOfflineMaxDepth; i++ {
		q.Enqueue("did:key:bob", []byte(`{"type":"text"}`), wsOfflineRetention)
	}
	overflow := []byte(`{"type":"overflow-blob"}`)
	q.Enqueue("did:key:bob", overflow, wsOfflineRetention)

	drain := q.DequeueAll("did:key:bob")
	if len(drain.Blobs) != wsOfflineMaxDepth {
		t.Fatalf("blobs %d, want %d", len(drain.Blobs), wsOfflineMaxDepth)
	}
	if len(drain.OverflowURIs) != 1 {
		t.Fatalf("overflow URIs %d, want 1", len(drain.OverflowURIs))
	}
	got, err := stub.Retrieve(t.Context(), drain.OverflowURIs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, overflow) {
		t.Fatalf("overflow payload mismatch")
	}
}

func TestFlushOffline_SendsOverflowManifest(t *testing.T) {
	h := NewHub()
	stub := encblob.NewStubStorage()
	h.offlineQueue.SetOverflowStorage(stub)
	for i := 0; i < wsOfflineMaxDepth; i++ {
		h.offlineQueue.Enqueue("did:key:bob", []byte(`{"queued":true}`), wsOfflineRetention)
	}
	h.offlineQueue.Enqueue("did:key:bob", []byte(`{"overflow":true}`), wsOfflineRetention)

	bob := &Client{hub: h, userID: "did:key:bob", send: make(chan []byte, wsOfflineMaxDepth+2)}
	h.flushOffline(bob)

	received := 0
	var manifest *WSMessage
	for {
		select {
		case raw := <-bob.send:
			received++
			var msg WSMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				t.Fatal(err)
			}
			if msg.Type == "overflow_manifest" {
				manifest = &msg
			}
		default:
			goto done
		}
	}
done:
	if received != wsOfflineMaxDepth+1 {
		t.Fatalf("received %d messages, want %d", received, wsOfflineMaxDepth+1)
	}
	if manifest == nil {
		t.Fatal("expected overflow_manifest")
	}
	var body map[string]interface{}
	if err := json.Unmarshal(manifest.Payload, &body); err != nil {
		t.Fatal(err)
	}
	uris, ok := body["storage_uris"].([]interface{})
	if !ok || len(uris) != 1 {
		t.Fatalf("unexpected manifest payload: %+v", body)
	}
}
