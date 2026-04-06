package auth

import (
	"testing"
)

func TestStepUpService_GetRequirement(t *testing.T) {
	s := NewStepUpService()

	tests := []struct {
		action         StepUpAction
		expectedMethod StepUpMethod
		exists         bool
	}{
		{StepUpChangePhone, StepUpMethodOTP, true},
		{StepUpRevokeDevice, StepUpMethodPasskey, true},
		{StepUpLargePayment, StepUpMethodPasskey, true},
		{StepUpExportRecovery, StepUpMethodPasskeyOTP, true},
		{StepUpDeleteAccount, StepUpMethodPasskeyOTP, true},
		{StepUpAction("unknown"), "", false},
	}

	for _, tt := range tests {
		req := s.GetRequirement(tt.action)
		if tt.exists && req == nil {
			t.Errorf("expected requirement for %s", tt.action)
		}
		if !tt.exists && req != nil {
			t.Errorf("expected nil for unknown action")
		}
		if tt.exists && req != nil && req.RequiredMethod != tt.expectedMethod {
			t.Errorf("action %s: expected method %s, got %s", tt.action, tt.expectedMethod, req.RequiredMethod)
		}
	}
}

func TestStepUpService_RequiresStepUp(t *testing.T) {
	s := NewStepUpService()

	if !s.RequiresStepUp("change_phone") {
		t.Error("change_phone should require step-up")
	}
	if !s.RequiresStepUp("delete_account") {
		t.Error("delete_account should require step-up")
	}
	if s.RequiresStepUp("send_message") {
		t.Error("send_message should not require step-up")
	}
}

func TestStepUpService_ValidateMethod(t *testing.T) {
	s := NewStepUpService()

	// Exact match
	if !s.ValidateMethod(StepUpChangePhone, StepUpMethodOTP) {
		t.Error("OTP should satisfy change_phone")
	}

	// Wrong method
	if s.ValidateMethod(StepUpChangePhone, StepUpMethodPasskey) {
		t.Error("passkey alone should not satisfy change_phone (requires OTP)")
	}

	// passkey+otp satisfies any requirement
	if !s.ValidateMethod(StepUpChangePhone, StepUpMethodPasskeyOTP) {
		t.Error("passkey+otp should satisfy any action")
	}
	if !s.ValidateMethod(StepUpRevokeDevice, StepUpMethodPasskeyOTP) {
		t.Error("passkey+otp should satisfy revoke_device")
	}
	if !s.ValidateMethod(StepUpExportRecovery, StepUpMethodPasskeyOTP) {
		t.Error("passkey+otp should satisfy export_recovery")
	}

	// Unknown action
	if s.ValidateMethod(StepUpAction("nonexistent"), StepUpMethodPasskey) {
		t.Error("unknown action should fail validation")
	}
}

func TestStepUpService_DeleteAccountHasZeroTTL(t *testing.T) {
	s := NewStepUpService()
	req := s.GetRequirement(StepUpDeleteAccount)
	if req == nil {
		t.Fatal("delete_account requirement should exist")
	}
	if req.TokenTTL != 0 {
		t.Errorf("delete_account should have 0 TTL (queued), got %v", req.TokenTTL)
	}
}

func TestStepUpService_ExportRecoveryShortTTL(t *testing.T) {
	s := NewStepUpService()
	req := s.GetRequirement(StepUpExportRecovery)
	if req == nil {
		t.Fatal("export_recovery requirement should exist")
	}
	// export_recovery should have the shortest non-zero TTL (2 min)
	if req.TokenTTL.Minutes() != 2 {
		t.Errorf("export_recovery should have 2-minute TTL, got %v", req.TokenTTL)
	}
}

func TestIsElevatedForAction(t *testing.T) {
	// Nil claims
	if IsElevatedForAction(nil, "change_phone") {
		t.Error("nil claims should not be elevated")
	}

	// Not elevated
	claims := &TokenClaims{Elevated: false}
	if IsElevatedForAction(claims, "change_phone") {
		t.Error("non-elevated claims should not pass")
	}

	// Elevated but wrong action
	claims = &TokenClaims{Elevated: true, ElevatedAction: "revoke_device"}
	if IsElevatedForAction(claims, "change_phone") {
		t.Error("elevated for wrong action should not pass")
	}

	// Elevated for correct action
	claims = &TokenClaims{Elevated: true, ElevatedAction: "change_phone"}
	if !IsElevatedForAction(claims, "change_phone") {
		t.Error("elevated for correct action should pass")
	}
}
