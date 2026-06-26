package crypto

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

// handshake sets up an initiator (Alice) and responder (Bob) sharing a 32-byte
// secret, with Bob publishing a ratchet pre-key — mirroring the X3DH bootstrap.
func handshake(t *testing.T) (alice, bob *RatchetSession) {
	t.Helper()
	ss := make([]byte, 32)
	if _, err := rand.Read(ss); err != nil {
		t.Fatalf("rand: %v", err)
	}
	bobPre, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("bob prekey: %v", err)
	}
	bobPub := bobPre.PublicKey().Bytes()
	if len(bobPub) == 65 && bobPub[0] == 0x04 {
		bobPub = bobPub[1:]
	}
	alice, err = NewRatchetSessionInitiator(ss, bobPub)
	if err != nil {
		t.Fatalf("init alice: %v", err)
	}
	bob, err = NewRatchetSessionResponder(ss, bobPre)
	if err != nil {
		t.Fatalf("init bob: %v", err)
	}
	return alice, bob
}

func TestRatchet_BasicRoundTrip(t *testing.T) {
	alice, bob := handshake(t)

	msg, err := alice.Encrypt([]byte("hello bob"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	pt, err := bob.Decrypt(msg)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(pt) != "hello bob" {
		t.Fatalf("got %q", pt)
	}
}

func TestRatchet_FullPingPongConversation(t *testing.T) {
	alice, bob := handshake(t)

	// Several messages each way, alternating — forces repeated DH ratchet steps.
	for round := 0; round < 5; round++ {
		a2b := []byte("a->b #" + string(rune('0'+round)))
		m, err := alice.Encrypt(a2b)
		if err != nil {
			t.Fatalf("alice enc r%d: %v", round, err)
		}
		got, err := bob.Decrypt(m)
		if err != nil {
			t.Fatalf("bob dec r%d: %v", round, err)
		}
		if !bytes.Equal(got, a2b) {
			t.Fatalf("r%d a->b mismatch: %q", round, got)
		}

		b2a := []byte("b->a #" + string(rune('0'+round)))
		m2, err := bob.Encrypt(b2a)
		if err != nil {
			t.Fatalf("bob enc r%d: %v", round, err)
		}
		got2, err := alice.Decrypt(m2)
		if err != nil {
			t.Fatalf("alice dec r%d: %v", round, err)
		}
		if !bytes.Equal(got2, b2a) {
			t.Fatalf("r%d b->a mismatch: %q", round, got2)
		}
	}
}

func TestRatchet_OutOfOrderWithinChain(t *testing.T) {
	alice, bob := handshake(t)

	// Alice sends 3 messages; deliver them to Bob as 0, 2, 1.
	m0, _ := alice.Encrypt([]byte("zero"))
	m1, _ := alice.Encrypt([]byte("one"))
	m2, _ := alice.Encrypt([]byte("two"))

	if pt, err := bob.Decrypt(m0); err != nil || string(pt) != "zero" {
		t.Fatalf("m0: %q %v", pt, err)
	}
	// m2 arrives before m1 — Bob must cache the skipped key for n=1.
	if pt, err := bob.Decrypt(m2); err != nil || string(pt) != "two" {
		t.Fatalf("m2: %q %v", pt, err)
	}
	if pt, err := bob.Decrypt(m1); err != nil || string(pt) != "one" {
		t.Fatalf("m1 (skipped): %q %v", pt, err)
	}
}

func TestRatchet_OutOfOrderAcrossDHRatchet(t *testing.T) {
	alice, bob := handshake(t)

	// Alice -> Bob establishes chains.
	m, _ := alice.Encrypt([]byte("hi"))
	if _, err := bob.Decrypt(m); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	// Bob sends two (new sending chain via his DH key); Alice receives 2nd first.
	b0, _ := bob.Encrypt([]byte("b0"))
	b1, _ := bob.Encrypt([]byte("b1"))
	if pt, err := alice.Decrypt(b1); err != nil || string(pt) != "b1" {
		t.Fatalf("b1 first: %q %v", pt, err)
	}
	if pt, err := alice.Decrypt(b0); err != nil || string(pt) != "b0" {
		t.Fatalf("b0 skipped: %q %v", pt, err)
	}
}

func TestRatchet_TamperDetected(t *testing.T) {
	alice, bob := handshake(t)
	m, _ := alice.Encrypt([]byte("authentic"))

	// Flip a byte in the ciphertext — GCM auth must fail.
	raw, _ := base64.StdEncoding.DecodeString(m.Ciphertext)
	if len(raw) == 0 {
		// short plaintext may produce empty ct; tamper the tag instead.
		tag, _ := base64.StdEncoding.DecodeString(m.Tag)
		tag[0] ^= 0xff
		m.Tag = base64.StdEncoding.EncodeToString(tag)
	} else {
		raw[0] ^= 0xff
		m.Ciphertext = base64.StdEncoding.EncodeToString(raw)
	}
	if _, err := bob.Decrypt(m); err == nil {
		t.Fatal("expected auth failure on tampered ciphertext")
	}
}

func TestRatchet_MaxSkipEnforced(t *testing.T) {
	alice, bob := handshake(t)
	bob.maxSkip = 5 // tighten for the test

	// Bootstrap a receiving chain on Bob.
	m0, _ := alice.Encrypt([]byte("0"))
	if _, err := bob.Decrypt(m0); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	// Alice sends many; deliver one far ahead so Bob must skip > maxSkip.
	var far *RatchetMessage
	for i := 0; i < 10; i++ {
		far, _ = alice.Encrypt([]byte("x"))
	}
	if _, err := bob.Decrypt(far); err == nil {
		t.Fatal("expected error skipping more than maxSkip messages")
	}
}

func TestRatchet_ForwardSecrecyChainAdvances(t *testing.T) {
	// Property check: the sending chain key must change after each message, so a
	// later chain key cannot reproduce an earlier message key.
	alice, _ := handshake(t)
	first := append([]byte(nil), alice.sendChainKey...)
	if _, err := alice.Encrypt([]byte("m1")); err != nil {
		t.Fatalf("enc: %v", err)
	}
	if bytes.Equal(first, alice.sendChainKey) {
		t.Fatal("sending chain key did not advance — no forward secrecy")
	}
}

func TestRatchet_ResponderCannotSendBeforeReceiving(t *testing.T) {
	_, bob := handshake(t)
	if _, err := bob.Encrypt([]byte("nope")); err == nil {
		t.Fatal("responder should not have a sending chain before first receive")
	}
}
