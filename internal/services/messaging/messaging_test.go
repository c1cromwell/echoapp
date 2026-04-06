package messaging

import (
	"testing"
	"time"
)

func TestCreateConversation(t *testing.T) {
	service := NewMessagingService()

	t.Run("valid conversation", func(t *testing.T) {
		conv, err := service.CreateConversation([]string{"user1", "user2"})
		if err != nil {
			t.Fatalf("CreateConversation failed: %v", err)
		}
		if conv == nil {
			t.Fatal("Conversation is nil")
		}
		if conv.ID == "" {
			t.Error("Conversation ID is empty")
		}
		if len(conv.Users) != 2 {
			t.Errorf("expected 2 users, got %d", len(conv.Users))
		}
	})

	t.Run("insufficient participants", func(t *testing.T) {
		_, err := service.CreateConversation([]string{"user1"})
		if err != ErrInvalidParticipants {
			t.Errorf("expected ErrInvalidParticipants, got %v", err)
		}
	})

	t.Run("empty participants", func(t *testing.T) {
		_, err := service.CreateConversation([]string{})
		if err != ErrInvalidParticipants {
			t.Errorf("expected ErrInvalidParticipants, got %v", err)
		}
	})
}

func TestSendMessage(t *testing.T) {
	service := NewMessagingService()
	conv, _ := service.CreateConversation([]string{"user1", "user2"})

	t.Run("valid message", func(t *testing.T) {
		msg, err := service.SendMessage("user1", conv.ID, TextMessage, []byte("hello"))
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}
		if msg == nil {
			t.Fatal("Message is nil")
		}
		if msg.Silent {
			t.Error("regular message should not be silent")
		}
		if string(msg.Content) != "hello" {
			t.Errorf("content = %s, want hello", string(msg.Content))
		}
	})

	t.Run("empty sender", func(t *testing.T) {
		_, err := service.SendMessage("", conv.ID, TextMessage, []byte("hello"))
		if err != ErrInvalidSender {
			t.Errorf("expected ErrInvalidSender, got %v", err)
		}
	})

	t.Run("invalid conversation", func(t *testing.T) {
		_, err := service.SendMessage("user1", "nonexistent", TextMessage, []byte("hello"))
		if err != ErrConvNotFound {
			t.Errorf("expected ErrConvNotFound, got %v", err)
		}
	})
}

func TestGetMessage(t *testing.T) {
	service := NewMessagingService()
	conv, _ := service.CreateConversation([]string{"user1", "user2"})
	sent, _ := service.SendMessage("user1", conv.ID, TextMessage, []byte("test"))

	t.Run("existing message", func(t *testing.T) {
		retrieved, err := service.GetMessage(sent.ID)
		if err != nil {
			t.Fatalf("GetMessage failed: %v", err)
		}
		if retrieved.ID != sent.ID {
			t.Errorf("Message ID mismatch: got %s, want %s", retrieved.ID, sent.ID)
		}
	})

	t.Run("nonexistent message", func(t *testing.T) {
		_, err := service.GetMessage("nonexistent")
		if err != ErrMessageNotFound {
			t.Errorf("expected ErrMessageNotFound, got %v", err)
		}
	})
}

func TestSendSilentMessage(t *testing.T) {
	service := NewMessagingService()
	conv, _ := service.CreateConversation([]string{"user1", "user2"})

	t.Run("with default flags", func(t *testing.T) {
		msg, err := service.SendSilentMessage("user1", conv.ID, TextMessage, []byte("quiet"), nil)
		if err != nil {
			t.Fatalf("SendSilentMessage failed: %v", err)
		}
		if !msg.Silent {
			t.Error("message should be silent")
		}
		if msg.SilentFlags == nil {
			t.Fatal("silent flags should not be nil")
		}
		if !msg.SilentFlags.SuppressPush {
			t.Error("should suppress push notifications")
		}
		if !msg.SilentFlags.SuppressBadge {
			t.Error("should suppress badge")
		}
		if !msg.SilentFlags.SuppressTyping {
			t.Error("should suppress typing indicator")
		}
		if !msg.SilentFlags.SuppressReceipts {
			t.Error("should suppress read receipts")
		}
		if !msg.SilentFlags.SuppressPreview {
			t.Error("should suppress preview")
		}
	})

	t.Run("with custom flags", func(t *testing.T) {
		flags := &SilentFlags{
			SuppressPush:     true,
			SuppressBadge:    false,
			SuppressTyping:   true,
			SuppressReceipts: false,
			SuppressPreview:  true,
		}
		msg, err := service.SendSilentMessage("user1", conv.ID, TextMessage, []byte("partial silent"), flags)
		if err != nil {
			t.Fatalf("SendSilentMessage failed: %v", err)
		}
		if !msg.SilentFlags.SuppressPush {
			t.Error("should suppress push")
		}
		if msg.SilentFlags.SuppressBadge {
			t.Error("should not suppress badge")
		}
	})

	t.Run("empty sender", func(t *testing.T) {
		_, err := service.SendSilentMessage("", conv.ID, TextMessage, []byte("quiet"), nil)
		if err != ErrInvalidSender {
			t.Errorf("expected ErrInvalidSender, got %v", err)
		}
	})

	t.Run("invalid conversation", func(t *testing.T) {
		_, err := service.SendSilentMessage("user1", "bad-conv", TextMessage, []byte("quiet"), nil)
		if err != ErrConvNotFound {
			t.Errorf("expected ErrConvNotFound, got %v", err)
		}
	})
}

func TestConversationSilentMode(t *testing.T) {
	service := NewMessagingService()
	conv, _ := service.CreateConversation([]string{"user1", "user2"})

	t.Run("enable silent mode", func(t *testing.T) {
		err := service.SetConversationSilent(conv.ID, Silent1Hour)
		if err != nil {
			t.Fatalf("SetConversationSilent failed: %v", err)
		}
		c, _ := service.GetConversation(conv.ID)
		if c.SilentSettings == nil {
			t.Fatal("silent settings should not be nil")
		}
		if !c.SilentSettings.Enabled {
			t.Error("silent mode should be enabled")
		}
		if c.SilentSettings.Duration != Silent1Hour {
			t.Errorf("duration = %d, want %d", c.SilentSettings.Duration, Silent1Hour)
		}
		if c.SilentSettings.ExpiresAt == nil {
			t.Error("expires at should be set for timed duration")
		}
	})

	t.Run("enable always silent", func(t *testing.T) {
		err := service.SetConversationSilent(conv.ID, SilentAlways)
		if err != nil {
			t.Fatalf("SetConversationSilent failed: %v", err)
		}
		c, _ := service.GetConversation(conv.ID)
		if c.SilentSettings.ExpiresAt != nil {
			t.Error("always-silent should not have expiry")
		}
	})

	t.Run("disable silent mode", func(t *testing.T) {
		err := service.SetConversationSilent(conv.ID, SilentOff)
		if err != nil {
			t.Fatalf("SetConversationSilent failed: %v", err)
		}
		c, _ := service.GetConversation(conv.ID)
		if c.SilentSettings != nil {
			t.Error("silent settings should be nil when disabled")
		}
	})

	t.Run("invalid conversation", func(t *testing.T) {
		err := service.SetConversationSilent("nonexistent", Silent1Hour)
		if err != ErrConvNotFound {
			t.Errorf("expected ErrConvNotFound, got %v", err)
		}
	})
}

func TestGetSilentMessages(t *testing.T) {
	service := NewMessagingService()
	conv, _ := service.CreateConversation([]string{"user1", "user2"})

	// Send a mix of regular and silent messages
	service.SendMessage("user1", conv.ID, TextMessage, []byte("regular1"))
	service.SendSilentMessage("user1", conv.ID, TextMessage, []byte("silent1"), nil)
	service.SendMessage("user1", conv.ID, TextMessage, []byte("regular2"))
	service.SendSilentMessage("user1", conv.ID, TextMessage, []byte("silent2"), nil)

	t.Run("filter silent messages", func(t *testing.T) {
		silent, err := service.GetSilentMessages(conv.ID)
		if err != nil {
			t.Fatalf("GetSilentMessages failed: %v", err)
		}
		if len(silent) != 2 {
			t.Errorf("expected 2 silent messages, got %d", len(silent))
		}
		for _, msg := range silent {
			if !msg.Silent {
				t.Error("returned non-silent message")
			}
		}
	})

	t.Run("all messages includes both", func(t *testing.T) {
		all, err := service.GetConversationMessages(conv.ID)
		if err != nil {
			t.Fatalf("GetConversationMessages failed: %v", err)
		}
		if len(all) != 4 {
			t.Errorf("expected 4 total messages, got %d", len(all))
		}
	})

	t.Run("invalid conversation", func(t *testing.T) {
		_, err := service.GetSilentMessages("bad-conv")
		if err != ErrConvNotFound {
			t.Errorf("expected ErrConvNotFound, got %v", err)
		}
	})
}

func TestDefaultSilentFlags(t *testing.T) {
	flags := DefaultSilentFlags()
	if !flags.SuppressPush || !flags.SuppressBadge || !flags.SuppressTyping || !flags.SuppressReceipts || !flags.SuppressPreview {
		t.Error("default flags should suppress all notifications")
	}
}

func TestSilentDurationValues(t *testing.T) {
	tests := []struct {
		name     string
		duration SilentDuration
		hours    int
	}{
		{"off", SilentOff, 0},
		{"1 hour", Silent1Hour, 1},
		{"8 hours", Silent8Hours, 8},
		{"24 hours", Silent24Hours, 24},
		{"7 days", Silent7Days, 168},
		{"always", SilentAlways, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.duration) != tt.hours {
				t.Errorf("%s: got %d, want %d", tt.name, int(tt.duration), tt.hours)
			}
		})
	}
}

func TestConversationSilentExpiry(t *testing.T) {
	service := NewMessagingService()
	conv, _ := service.CreateConversation([]string{"user1", "user2"})

	// Set silent mode with 1 hour duration
	service.SetConversationSilent(conv.ID, Silent1Hour)
	c, _ := service.GetConversation(conv.ID)

	if c.SilentSettings == nil {
		t.Fatal("silent settings should be set")
	}

	// The expiry should be approximately 1 hour from now
	expectedExpiry := time.Now().Add(1 * time.Hour)
	diff := c.SilentSettings.ExpiresAt.Sub(expectedExpiry)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("expiry time off by more than 1 second: diff=%v", diff)
	}
}
