package trustnet

import (
	"testing"
)

func newRequestTestSetup() (*ContactRequestService, *CircleService) {
	circles := NewCircleService()
	requests := NewContactRequestService(circles)
	return requests, circles
}

func TestSendRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		req, err := svc.SendRequest("alice", "bob", "Hey, met at the conference!", 67)
		if err != nil {
			t.Fatalf("SendRequest failed: %v", err)
		}
		if req.FromDID != "alice" || req.ToDID != "bob" {
			t.Error("incorrect DIDs")
		}
		if req.Status != RequestPending {
			t.Errorf("status = %s, want pending", req.Status)
		}
		if req.Message != "Hey, met at the conference!" {
			t.Error("message not preserved")
		}
		if req.FromTrustScore != 67 {
			t.Errorf("trust score = %d, want 67", req.FromTrustScore)
		}
	})

	t.Run("cannot request self", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		_, err := svc.SendRequest("alice", "alice", "", 50)
		if err != ErrRequestToSelf {
			t.Errorf("expected ErrRequestToSelf, got %v", err)
		}
	})

	t.Run("already contacts", func(t *testing.T) {
		svc, circles := newRequestTestSetup()
		circles.AddContact("alice", "bob", CircleAcquaintance)
		_, err := svc.SendRequest("alice", "bob", "", 50)
		if err != ErrAlreadyContacts {
			t.Errorf("expected ErrAlreadyContacts, got %v", err)
		}
	})

	t.Run("duplicate pending request", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		svc.SendRequest("alice", "bob", "", 50)
		_, err := svc.SendRequest("alice", "bob", "", 50)
		if err != ErrRequestAlreadySent {
			t.Errorf("expected ErrRequestAlreadySent, got %v", err)
		}
	})

	t.Run("cross request blocked", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		svc.SendRequest("alice", "bob", "", 50)
		// Bob tries to also request Alice while Alice's is pending
		_, err := svc.SendRequest("bob", "alice", "", 50)
		if err != ErrRequestAlreadySent {
			t.Errorf("expected ErrRequestAlreadySent for cross-request, got %v", err)
		}
	})

	t.Run("mutual count computed", func(t *testing.T) {
		svc, circles := newRequestTestSetup()
		// shared mutual: charlie
		circles.AddContact("alice", "charlie", CircleTrusted)
		circles.AddContact("bob", "charlie", CircleAcquaintance)

		req, _ := svc.SendRequest("alice", "bob", "", 50)
		if req.MutualCount != 1 {
			t.Errorf("mutual count = %d, want 1", req.MutualCount)
		}
	})
}

func TestAcceptRequest(t *testing.T) {
	t.Run("accept as trusted", func(t *testing.T) {
		svc, circles := newRequestTestSetup()
		req, _ := svc.SendRequest("alice", "bob", "", 50)

		contact, err := svc.AcceptRequest(req.ID, CircleTrusted)
		if err != nil {
			t.Fatalf("AcceptRequest failed: %v", err)
		}
		if contact.Tier != CircleTrusted {
			t.Errorf("tier = %s, want trusted", contact.Tier)
		}

		// Bob should have Alice as a contact
		if !circles.HasContact("bob", "alice") {
			t.Error("bob should have alice as contact")
		}

		// Alice should also have Bob (reverse direction)
		if !circles.HasContact("alice", "bob") {
			t.Error("alice should have bob as contact (reverse)")
		}

		// Request should be marked accepted
		got, _ := svc.GetRequest(req.ID)
		if got.Status != RequestAccepted {
			t.Errorf("status = %s, want accepted", got.Status)
		}
		if got.RespondedAt == nil {
			t.Error("RespondedAt should be set")
		}
	})

	t.Run("accept not found", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		_, err := svc.AcceptRequest("nonexistent", CircleTrusted)
		if err != ErrRequestNotFound {
			t.Errorf("expected ErrRequestNotFound, got %v", err)
		}
	})

	t.Run("accept already handled", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		req, _ := svc.SendRequest("alice", "bob", "", 50)
		svc.AcceptRequest(req.ID, CircleTrusted)

		_, err := svc.AcceptRequest(req.ID, CircleInner)
		if err != ErrRequestAlreadyHandled {
			t.Errorf("expected ErrRequestAlreadyHandled, got %v", err)
		}
	})
}

func TestDeclineRequest(t *testing.T) {
	t.Run("decline pending", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		req, _ := svc.SendRequest("alice", "bob", "", 50)

		err := svc.DeclineRequest(req.ID)
		if err != nil {
			t.Fatalf("DeclineRequest failed: %v", err)
		}

		got, _ := svc.GetRequest(req.ID)
		if got.Status != RequestDeclined {
			t.Errorf("status = %s, want declined", got.Status)
		}
	})

	t.Run("decline already handled", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		req, _ := svc.SendRequest("alice", "bob", "", 50)
		svc.DeclineRequest(req.ID)

		err := svc.DeclineRequest(req.ID)
		if err != ErrRequestAlreadyHandled {
			t.Errorf("expected ErrRequestAlreadyHandled, got %v", err)
		}
	})
}

func TestCancelRequest(t *testing.T) {
	t.Run("cancel own request", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		req, _ := svc.SendRequest("alice", "bob", "", 50)

		err := svc.CancelRequest(req.ID, "alice")
		if err != nil {
			t.Fatalf("CancelRequest failed: %v", err)
		}

		got, _ := svc.GetRequest(req.ID)
		if got.Status != RequestCancelled {
			t.Errorf("status = %s, want cancelled", got.Status)
		}
	})

	t.Run("cannot cancel others request", func(t *testing.T) {
		svc, _ := newRequestTestSetup()
		req, _ := svc.SendRequest("alice", "bob", "", 50)

		err := svc.CancelRequest(req.ID, "bob")
		if err != ErrRequestNotFound {
			t.Errorf("expected ErrRequestNotFound, got %v", err)
		}
	})
}

func TestGetPendingRequests(t *testing.T) {
	svc, _ := newRequestTestSetup()

	svc.SendRequest("alice", "bob", "hi", 50)
	svc.SendRequest("charlie", "bob", "hello", 60)
	req3, _ := svc.SendRequest("dave", "bob", "hey", 40)
	svc.DeclineRequest(req3.ID) // decline one

	received := svc.GetPendingReceived("bob")
	if len(received) != 2 {
		t.Errorf("pending received = %d, want 2", len(received))
	}

	sent := svc.GetPendingSent("alice")
	if len(sent) != 1 {
		t.Errorf("pending sent = %d, want 1", len(sent))
	}

	count := svc.CountPendingReceived("bob")
	if count != 2 {
		t.Errorf("count pending = %d, want 2", count)
	}
}

func TestLowTrustWarning(t *testing.T) {
	svc, _ := newRequestTestSetup()
	req, _ := svc.SendRequest("spammer", "bob", "", 12)

	if !IsLowTrustWarning(req) {
		t.Error("trust score 12 should trigger low trust warning")
	}

	req2, _ := svc.SendRequest("trusted", "charlie", "", 67)
	if IsLowTrustWarning(req2) {
		t.Error("trust score 67 should not trigger warning")
	}
}
