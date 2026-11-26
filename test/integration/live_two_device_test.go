package integration

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thechadcromwell/echoapp/internal/wallet"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// Live tests: two simulated devices against a RUNNING gateway, over real HTTP, WebSocket and
// Postgres. Nothing here is in-process.
//
//	ECHO_LIVE_URL=http://localhost:8000 \
//	DATABASE_URL=postgres://echoapp:echoapp_dev@localhost:5432/echoapp?sslmode=disable \
//	ECHO_LIVE_RESTART_CMD="docker restart echoapp-testnet" \
//	go test ./test/integration -run Live -count=1 -v
//
// They speak the same protocol the app does, but they are not the app: they prove the server's
// side of account restore and sealed delivery. The Swift code is covered by its own unit tests
// and still needs a run on two phones.

func liveURL(t *testing.T) string {
	t.Helper()
	u := os.Getenv("ECHO_LIVE_URL")
	if u == "" {
		t.Skip("ECHO_LIVE_URL not set")
	}
	return strings.TrimRight(u, "/")
}

// liveDevice is one device: a P-256 key the gateway knows as a did:key.
type liveDevice struct {
	base string
	priv *ecdsa.PrivateKey
	did  string
}

func newDeviceKey(t *testing.T) (*ecdsa.PrivateKey, string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return priv, hex.EncodeToString(elliptic.Marshal(elliptic.P256(), priv.X, priv.Y))
}

func registerDevice(t *testing.T, base string) *liveDevice {
	t.Helper()
	priv, pubHex := newDeviceKey(t)
	did, err := didkey.Derive(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]string{"did": did, "public_key_hex": pubHex})
	resp, err := http.Post(base+"/identity/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("register: %d %s", resp.StatusCode, raw)
	}
	return &liveDevice{base: base, priv: priv, did: did}
}

// signed sends a request signed the way the app signs: X-Sender-DID, X-Timestamp, X-Signature.
func (d *liveDevice) signed(t *testing.T, method, path string, payload interface{}) (int, map[string]interface{}) {
	t.Helper()
	var body []byte
	if payload != nil {
		body, _ = json.Marshal(payload)
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	bodyHash := sha256.Sum256(body)
	msg := []byte(method + "\n" + path + "\n" + ts + "\n" + hex.EncodeToString(bodyHash[:]))
	sig, err := didkey.SignECDSAP256SHA256ASN1(d.priv, msg)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(method, d.base+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sender-DID", d.did)
	req.Header.Set("X-Timestamp", ts)
	req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(sig))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]interface{}
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func (d *liveDevice) session(t *testing.T) string {
	t.Helper()
	code, out := d.signed(t, http.MethodPost, "/v1/auth/session", map[string]string{"device_id": "live-test-device"})
	token, _ := out["access_token"].(string)
	if code != http.StatusOK || token == "" {
		t.Fatalf("open session: %d %v", code, out)
	}
	return token
}

func (d *liveDevice) socket(t *testing.T) *websocket.Conn {
	t.Helper()
	header := http.Header{"Authorization": []string{"Bearer " + d.session(t)}}
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(d.base, "http")+"/ws", header)
	if err != nil {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		t.Fatalf("socket: %v (HTTP %d)", err, code)
	}
	return conn
}

// nextOfType reads frames until one of the wanted type arrives.
func nextOfType(t *testing.T, conn *websocket.Conn, wanted string, timeout time.Duration) map[string]interface{} {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		_ = conn.SetReadDeadline(deadline)
		_, raw, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("waiting for a %q frame: %v", wanted, err)
		}
		var frame map[string]interface{}
		if json.Unmarshal(raw, &frame) == nil && frame["type"] == wanted {
			return frame
		}
	}
}

// anonymousPost sends a request with no identity at all: no token, no signature, no cookie jar.
func anonymousPost(t *testing.T, url string, payload interface{}) (int, string) {
	t.Helper()
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func waitHealthy(t *testing.T, base string) {
	t.Helper()
	for i := 0; i < 90; i++ {
		resp, err := http.Get(base + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("gateway did not come back")
}

// Account restore, end to end: an account is created on one device and its recovery address is
// linked; the gateway is restarted; a second device, holding only the recovery key, restores the
// account and can then act as it.
func TestLive_AccountRestoreOnANewDeviceAcrossARestart(t *testing.T) {
	base := liveURL(t)
	phone := registerDevice(t, base)

	// The recovery key: what the recovery phrase derives. secp256k1, "DAG" address.
	recovery, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	recoveryPub := hex.EncodeToString(recovery.PubKey().SerializeUncompressed())
	address := wallet.DagAddressFromPubKey(recoveryPub)

	if code, out := phone.signed(t, http.MethodPost, "/v1/identity/link-wallet",
		map[string]string{"did": phone.did, "wallet_address": address}); code != http.StatusOK {
		t.Fatalf("link recovery address: %d %v", code, out)
	}
	// Someone else cannot take the address.
	thief := registerDevice(t, base)
	if code, _ := thief.signed(t, http.MethodPost, "/v1/identity/link-wallet",
		map[string]string{"did": thief.did, "wallet_address": address}); code != http.StatusConflict {
		t.Fatalf("a second account linking the same address: %d, want 409", code)
	}

	if cmd := os.Getenv("ECHO_LIVE_RESTART_CMD"); cmd != "" {
		if out, err := exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
			t.Fatalf("restart: %v: %s", err, out)
		}
		time.Sleep(3 * time.Second)
		waitHealthy(t, base)
		t.Log("gateway restarted")
	} else {
		t.Log("ECHO_LIVE_RESTART_CMD not set: the restart is NOT exercised in this run")
	}

	// A new phone: a fresh device key, and the recovery key from the phrase. Nothing else.
	newPriv, newPubHex := newDeviceKey(t)
	code, raw := anonymousPost(t, base+"/v1/auth/restore-challenge", map[string]string{"wallet_address": address})
	if code != http.StatusOK {
		t.Fatalf("restore challenge: %d %s", code, raw)
	}
	var ch struct {
		Challenge string `json:"challenge"`
	}
	_ = json.Unmarshal([]byte(raw), &ch)
	if ch.Challenge == "" {
		t.Fatalf("no challenge in %s", raw)
	}
	digest := sha512.Sum512([]byte(ch.Challenge))
	signature := hex.EncodeToString(secp256k1ecdsa.Sign(recovery, digest[:]).Serialize())

	code, raw = anonymousPost(t, base+"/v1/auth/restore-did", map[string]string{
		"wallet_address": address, "wallet_public_key": recoveryPub,
		"new_device_public_key": newPubHex, "wallet_signature": signature,
	})
	if code != http.StatusOK {
		t.Fatalf("restore: %d %s", code, raw)
	}
	var restored struct {
		DID string `json:"did"`
	}
	_ = json.Unmarshal([]byte(raw), &restored)
	if restored.DID != phone.did {
		t.Fatalf("restored account %q, want %q", restored.DID, phone.did)
	}

	// The new phone can now act as the account, signing with its own key.
	newPhone := &liveDevice{base: base, priv: newPriv, did: phone.did}
	if token := newPhone.session(t); token == "" {
		t.Fatal("the restored device could not open a session")
	}

	// A wrong recovery key gets nothing.
	other, _ := secp256k1.GeneratePrivateKey()
	code, raw = anonymousPost(t, base+"/v1/auth/restore-challenge", map[string]string{"wallet_address": address})
	_ = json.Unmarshal([]byte(raw), &ch)
	digest = sha512.Sum512([]byte(ch.Challenge))
	code, _ = anonymousPost(t, base+"/v1/auth/restore-did", map[string]string{
		"wallet_address": address, "wallet_public_key": hex.EncodeToString(other.PubKey().SerializeUncompressed()),
		"new_device_public_key": newPubHex, "wallet_signature": hex.EncodeToString(secp256k1ecdsa.Sign(other, digest[:]).Serialize()),
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("restore with someone else's key: %d, want 401", code)
	}
}

// Sealed delivery, end to end: Bob registers a delivery key; a sender with no identity at all
// delivers to it; Bob receives the message with no sender and no conversation on it, both while
// online and after being offline; and what the database held while he was offline names only him.
func TestLive_SealedDeliveryNamesNoSender(t *testing.T) {
	base := liveURL(t)
	bob := registerDevice(t, base)

	key := make([]byte, 32)
	_, _ = rand.Read(key)
	hash := sha256.Sum256(key)
	if code, out := bob.signed(t, http.MethodPut, "/v3/messages/delivery-key",
		map[string]string{"key_hash": hex.EncodeToString(hash[:])}); code != http.StatusOK {
		t.Fatalf("register delivery key: %d %v", code, out)
	}

	deliver := func(marker string) (int, string) {
		return anonymousPost(t, base+"/v1/sealed/deliver", map[string]interface{}{
			"delivery_key": hex.EncodeToString(key),
			"payload":      map[string]string{"ciphertext": base64.StdEncoding.EncodeToString([]byte(marker))},
		})
	}
	check := func(frame map[string]interface{}, marker string) {
		t.Helper()
		if from, _ := frame["from"].(string); from != "" {
			t.Fatalf("the delivered message names a sender: %q", from)
		}
		if conv, _ := frame["conversation_id"].(string); conv != "dm:sealed" {
			t.Fatalf("conversation on the wire = %q, want the constant dm:sealed", conv)
		}
		if to, _ := frame["to"].(string); to != bob.did {
			t.Fatalf("delivered to %q", to)
		}
		payload, _ := frame["payload"].(map[string]interface{})
		got, _ := base64.StdEncoding.DecodeString(payload["ciphertext"].(string))
		if string(got) != marker {
			t.Fatalf("payload %q, want %q", got, marker)
		}
	}

	// Online.
	conn := bob.socket(t)
	if code, raw := deliver("while-online"); code != http.StatusAccepted {
		t.Fatalf("sealed deliver: %d %s", code, raw)
	}
	check(nextOfType(t, conn, "sealed_text", 15*time.Second), "while-online")
	conn.Close()
	time.Sleep(time.Second)

	// Offline: the message waits in the database.
	if code, raw := deliver("while-offline"); code != http.StatusAccepted {
		t.Fatalf("sealed deliver to an offline recipient: %d %s", code, raw)
	}
	var pool *pgxpool.Pool
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		var err error
		if pool, err = pgxpool.New(context.Background(), dsn); err != nil {
			t.Fatal(err)
		}
		defer pool.Close()
		var sender, conversation string
		var payload []byte
		err = pool.QueryRow(context.Background(),
			`SELECT sender_did, conversation_id, encrypted_payload FROM message_queue WHERE recipient_did = $1 AND status = 'queued'`,
			bob.did).Scan(&sender, &conversation, &payload)
		if err != nil {
			t.Fatalf("the waiting message is not in the queue: %v", err)
		}
		if sender != "" || conversation != "" {
			t.Fatalf("the stored row names sender %q / conversation %q", sender, conversation)
		}
		if bytes.Contains(payload, []byte(`"from":"did:`)) {
			t.Fatalf("the stored envelope names a sender: %s", payload)
		}
	} else {
		t.Log("DATABASE_URL not set: the stored row is NOT inspected in this run")
	}

	conn = bob.socket(t)
	defer conn.Close()
	check(nextOfType(t, conn, "sealed_text", 15*time.Second), "while-offline")

	if pool != nil {
		var pending int
		_ = pool.QueryRow(context.Background(), `SELECT count(*) FROM message_queue WHERE recipient_did = $1 AND status = 'queued'`, bob.did).Scan(&pending)
		if pending != 0 {
			t.Fatalf("%d message(s) still pending after delivery", pending)
		}
		var leftover int
		_ = pool.QueryRow(context.Background(), `SELECT count(*) FROM message_queue WHERE recipient_did = $1 AND octet_length(encrypted_payload) > 0`, bob.did).Scan(&leftover)
		if leftover != 0 {
			t.Fatalf("%d delivered row(s) still hold their envelope", leftover)
		}
	}

	// Blocking: a new key with every earlier key revoked. The old key stops at once.
	fresh := make([]byte, 32)
	_, _ = rand.Read(fresh)
	freshHash := sha256.Sum256(fresh)
	if code, out := bob.signed(t, http.MethodPut, "/v3/messages/delivery-key",
		map[string]interface{}{"key_hash": hex.EncodeToString(freshHash[:]), "revoke_others": true}); code != http.StatusOK {
		t.Fatalf("replace delivery key: %d %v", code, out)
	}
	if code, _ := deliver("after-revocation"); code != http.StatusNotFound {
		t.Fatalf("delivery with a revoked key: %d, want 404", code)
	}
	// Registering a key needs the account's own signature.
	if code, _ := anonymousPost(t, base+"/v3/messages/delivery-key", map[string]string{"key_hash": hex.EncodeToString(hash[:])}); code != http.StatusUnauthorized {
		t.Fatalf("registering a delivery key with no identity: %d, want 401", code)
	}
}
