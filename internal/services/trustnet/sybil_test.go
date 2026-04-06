package trustnet

import (
	"testing"
)

func TestRegisterDevice(t *testing.T) {
	t.Run("register new device", func(t *testing.T) {
		svc := NewSybilService()
		err := svc.RegisterDevice("user1", "fp-abc", "192.168.1.1", "Mozilla/5.0")
		if err != nil {
			t.Fatalf("RegisterDevice failed: %v", err)
		}
	})

	t.Run("re-register same user updates", func(t *testing.T) {
		svc := NewSybilService()
		svc.RegisterDevice("user1", "fp-abc", "192.168.1.1", "Mozilla/5.0")
		err := svc.RegisterDevice("user1", "fp-abc", "10.0.0.1", "Chrome")
		if err != nil {
			t.Fatalf("re-register should succeed: %v", err)
		}
	})

	t.Run("max accounts per device", func(t *testing.T) {
		svc := NewSybilService()
		svc.RegisterDevice("user1", "fp-shared", "192.168.1.1", "Mozilla")
		svc.RegisterDevice("user2", "fp-shared", "192.168.1.2", "Mozilla")
		err := svc.RegisterDevice("user3", "fp-shared", "192.168.1.3", "Mozilla")
		if err != ErrDeviceClusterDetected {
			t.Errorf("expected ErrDeviceClusterDetected, got %v", err)
		}
	})
}

func TestAssessUser(t *testing.T) {
	t.Run("clean user", func(t *testing.T) {
		svc := NewSybilService()
		svc.RegisterDevice("user1", "fp-unique", "1.2.3.4", "Mozilla")
		// Add enough connections
		for i := 0; i < 5; i++ {
			svc.RecordConnection("user1", contactDID(i))
		}

		assessment := svc.AssessUser("user1")
		if assessment.RiskLevel != SybilRiskNone {
			t.Errorf("risk = %s, want none", assessment.RiskLevel)
		}
	})

	t.Run("shared device raises risk", func(t *testing.T) {
		svc := NewSybilService()
		svc.RegisterDevice("user1", "fp-shared", "1.2.3.4", "Mozilla")
		svc.RegisterDevice("user2", "fp-shared", "1.2.3.5", "Mozilla")
		for i := 0; i < 5; i++ {
			svc.RecordConnection("user1", contactDID(i))
		}

		assessment := svc.AssessUser("user1")
		if assessment.RiskLevel == SybilRiskNone {
			t.Error("shared device should raise risk")
		}
		found := false
		for _, f := range assessment.Flags {
			if f == "shared_device" {
				found = true
			}
		}
		if !found {
			t.Error("should have shared_device flag")
		}
	})

	t.Run("sparse graph raises risk", func(t *testing.T) {
		svc := NewSybilService()
		svc.RegisterDevice("user1", "fp-unique", "1.2.3.4", "Mozilla")
		// Only 1 connection, below MinGraphConnections
		svc.RecordConnection("user1", "friend1")

		assessment := svc.AssessUser("user1")
		found := false
		for _, f := range assessment.Flags {
			if f == "sparse_graph" {
				found = true
			}
		}
		if !found {
			t.Error("should have sparse_graph flag")
		}
	})

	t.Run("shared IP raises risk", func(t *testing.T) {
		svc := NewSybilService()
		svc.RegisterDevice("user1", "fp1", "1.1.1.1", "Mozilla")
		svc.RegisterDevice("user2", "fp2", "1.1.1.1", "Mozilla")
		svc.RegisterDevice("user3", "fp3", "1.1.1.1", "Mozilla")
		for i := 0; i < 5; i++ {
			svc.RecordConnection("user1", contactDID(i))
		}

		assessment := svc.AssessUser("user1")
		found := false
		for _, f := range assessment.Flags {
			if f == "shared_ip" {
				found = true
			}
		}
		if !found {
			t.Error("should have shared_ip flag")
		}
	})

	t.Run("high risk from multiple signals", func(t *testing.T) {
		svc := NewSybilService()
		svc.RegisterDevice("user1", "fp-shared", "1.1.1.1", "Mozilla")
		svc.RegisterDevice("user2", "fp-shared", "1.1.1.1", "Mozilla")
		// 3+ accounts on same IP
		svc.RegisterDevice("user3", "fp3", "1.1.1.1", "Chrome")
		// No connections (sparse graph)

		assessment := svc.AssessUser("user1")
		if assessment.RiskLevel != SybilRiskHigh {
			t.Errorf("risk = %s, want high (score=%f)", assessment.RiskLevel, assessment.Score)
		}
	})
}

func TestEndorsementRing(t *testing.T) {
	svc := NewSybilService()
	// Create a complete 3-node ring
	svc.RecordConnection("a", "b")
	svc.RecordConnection("a", "c")
	svc.RecordConnection("b", "c")
	svc.RegisterDevice("a", "fpa", "1.2.3.4", "Mozilla")

	assessment := svc.AssessUser("a")
	found := false
	for _, f := range assessment.Flags {
		if f == "endorsement_ring" {
			found = true
		}
	}
	if !found {
		t.Error("should detect endorsement ring")
	}
}

func TestDetectClusters(t *testing.T) {
	svc := NewSybilService()
	svc.RegisterDevice("user1", "fp-shared", "1.2.3.4", "Mozilla")
	svc.RegisterDevice("user2", "fp-shared", "1.2.3.5", "Mozilla")

	clusters := svc.DetectClusters()
	if len(clusters) < 1 {
		t.Error("should detect device cluster")
	}

	// Running again should not create duplicate clusters
	clusters2 := svc.DetectClusters()
	if len(clusters2) != 0 {
		t.Error("should not create duplicate clusters")
	}
}

func TestDetectIPClusters(t *testing.T) {
	svc := NewSybilService()
	for i := 0; i < 5; i++ {
		svc.RegisterDevice(contactDID(i), contactDID(i)+"-fp", "10.0.0.1", "Mozilla")
	}

	clusters := svc.DetectClusters()
	foundIP := false
	for _, c := range clusters {
		for _, sig := range c.Signals {
			if sig == "shared_ip_address" {
				foundIP = true
			}
		}
	}
	if !foundIP {
		t.Error("should detect IP cluster")
	}
}

func TestMarkClusterReviewed(t *testing.T) {
	svc := NewSybilService()
	svc.RegisterDevice("user1", "fp-shared", "1.2.3.4", "Mozilla")
	svc.RegisterDevice("user2", "fp-shared", "1.2.3.5", "Mozilla")

	clusters := svc.DetectClusters()
	if len(clusters) == 0 {
		t.Fatal("no clusters detected")
	}

	svc.MarkClusterReviewed(clusters[0].ID)
	c := svc.GetCluster(clusters[0].ID)
	if !c.Reviewed {
		t.Error("cluster should be marked as reviewed")
	}
}

func TestEndorsementWeightPenalty(t *testing.T) {
	svc := NewSybilService()

	// No assessment = full weight
	if p := svc.EndorsementWeightPenalty("unknown"); p != 1.0 {
		t.Errorf("penalty = %f, want 1.0", p)
	}

	// Assess a high-risk user
	svc.RegisterDevice("risky", "fp-shared", "1.1.1.1", "Mozilla")
	svc.RegisterDevice("risky2", "fp-shared", "1.1.1.1", "Mozilla")
	svc.RegisterDevice("risky3", "fp3", "1.1.1.1", "Chrome")
	svc.AssessUser("risky")

	penalty := svc.EndorsementWeightPenalty("risky")
	if penalty >= 1.0 {
		t.Errorf("high risk user should have reduced weight, got %f", penalty)
	}
}

func TestGetConnectionCount(t *testing.T) {
	svc := NewSybilService()
	svc.RecordConnection("user1", "a")
	svc.RecordConnection("user1", "b")
	svc.RecordConnection("user1", "c")

	if n := svc.GetConnectionCount("user1"); n != 3 {
		t.Errorf("connections = %d, want 3", n)
	}

	// Bidirectional
	if n := svc.GetConnectionCount("a"); n != 1 {
		t.Errorf("reverse connections = %d, want 1", n)
	}
}
