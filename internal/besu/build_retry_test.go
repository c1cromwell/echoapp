package besu

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// fakeNode answers just enough JSON-RPC to prepare a transaction. While dropFirst is positive it
// cuts the connection instead of answering any call other than the height probe, the way a
// validator does when it is restarted mid-request. With refuse set, gas estimation is refused the
// way a contract refusal is.
func fakeNode(t *testing.T, dropFirst *int32, refuse bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		_ = json.Unmarshal(body, &req)
		if req.Method != "eth_blockNumber" && atomic.AddInt32(dropFirst, -1) >= 0 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err == nil {
				conn.Close()
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if refuse && req.Method == "eth_estimateGas" {
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":` + string(req.ID) + `,"error":{"code":3,"message":"execution reverted"}}`))
			return
		}
		result := `"0x1"`
		switch req.Method {
		case "eth_blockNumber":
			result = `"0x10"`
		case "eth_chainId":
			result = `"0x4f25"`
		case "eth_estimateGas":
			result = `"0x5208"`
		case "eth_getCode":
			result = `"0x60"`
		}
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":` + string(req.ID) + `,"result":` + result + `}`))
	}))
}

func testClient(t *testing.T, url string) *Client {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(Config{RPCURLs: []string{url}, AnchorContract: common.HexToAddress("0x1000000000000000000000000000000000000001"), Signer: key, ReceiptTimeout: 20 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// A validator that drops the connection while a transaction is being prepared has said nothing
// about the transaction. The client must try again, not report the write as failed. (Found by the
// rolling-restart failure test: one anchor in a run was lost to exactly this.)
func TestBuildRetriesWhenTheValidatorDropsTheConnection(t *testing.T) {
	drops := int32(3)
	node := fakeNode(t, &drops, false)
	defer node.Close()
	c := testClient(t, node.URL)

	tx, err := c.build(context.Background(), c.cfg.AnchorContract, anchorABI, time.Now().Add(20*time.Second), "anchorRoot", common.HexToHash("0x01"), uint32(1))
	if err != nil {
		t.Fatalf("build gave up after dropped connections: %v", err)
	}
	if tx == nil || tx.To() == nil || *tx.To() != c.cfg.AnchorContract {
		t.Fatalf("unexpected transaction: %+v", tx)
	}
	if atomic.LoadInt32(&drops) >= 0 {
		t.Fatal("the fake node never dropped a connection; the test proved nothing")
	}
}

// A refusal is an answer. It must come back at once, not after retrying until the deadline.
func TestBuildReturnsARefusalAtOnce(t *testing.T) {
	none := int32(0)
	node := fakeNode(t, &none, true)
	defer node.Close()
	c := testClient(t, node.URL)

	start := time.Now()
	_, err := c.build(context.Background(), c.cfg.AnchorContract, anchorABI, time.Now().Add(20*time.Second), "anchorRoot", common.HexToHash("0x01"), uint32(1))
	if err == nil || !strings.Contains(err.Error(), "execution reverted") {
		t.Fatalf("err = %v, want the node's refusal", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("a refusal took %s to come back", time.Since(start))
	}
}

// With no validator answering at all, the client stops at the deadline and says why.
func TestBuildStopsAtTheDeadline(t *testing.T) {
	always := int32(1 << 30)
	node := fakeNode(t, &always, false)
	defer node.Close()
	c := testClient(t, node.URL)

	start := time.Now()
	_, err := c.build(context.Background(), c.cfg.AnchorContract, anchorABI, time.Now().Add(2*time.Second), "anchorRoot", common.HexToHash("0x01"), uint32(1))
	if err == nil || !strings.Contains(err.Error(), "no validator could prepare") {
		t.Fatalf("err = %v", err)
	}
	if d := time.Since(start); d < 2*time.Second || d > 8*time.Second {
		t.Fatalf("stopped after %s, want about 2s", d)
	}
}
