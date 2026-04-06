package messaging

import (
	"testing"
	"time"
)

func newTestScheduledService() (*ScheduledMessageService, *MessagingService) {
	ms := NewMessagingService()
	ss := NewScheduledMessageService(ms)
	return ss, ms
}

func TestScheduleMessage(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	t.Run("valid schedule", func(t *testing.T) {
		msg, err := ss.Schedule("user1", conv.ID, []byte("hello future"), TextMessage, futureTime, "America/New_York")
		if err != nil {
			t.Fatalf("Schedule failed: %v", err)
		}
		if msg.ID == "" {
			t.Error("scheduled message ID is empty")
		}
		if msg.Status != DeliveryPending {
			t.Errorf("status = %s, want %s", msg.Status, DeliveryPending)
		}
		if msg.SenderID != "user1" {
			t.Errorf("senderID = %s, want user1", msg.SenderID)
		}
		if string(msg.Content) != "hello future" {
			t.Errorf("content = %s, want 'hello future'", string(msg.Content))
		}
		if msg.Timezone != "America/New_York" {
			t.Errorf("timezone = %s, want America/New_York", msg.Timezone)
		}
	})

	t.Run("past time rejected", func(t *testing.T) {
		pastTime := time.Now().Add(-1 * time.Hour)
		_, err := ss.Schedule("user1", conv.ID, []byte("too late"), TextMessage, pastTime, "")
		if err != ErrScheduledTimeInPast {
			t.Errorf("expected ErrScheduledTimeInPast, got %v", err)
		}
	})

	t.Run("empty sender rejected", func(t *testing.T) {
		_, err := ss.Schedule("", conv.ID, []byte("no sender"), TextMessage, futureTime, "")
		if err != ErrInvalidSender {
			t.Errorf("expected ErrInvalidSender, got %v", err)
		}
	})
}

func TestScheduleSilentMessage(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	t.Run("with default flags", func(t *testing.T) {
		msg, err := ss.ScheduleSilent("user1", conv.ID, []byte("quiet future"), TextMessage, futureTime, "", nil)
		if err != nil {
			t.Fatalf("ScheduleSilent failed: %v", err)
		}
		if !msg.Silent {
			t.Error("scheduled message should be silent")
		}
		if msg.SilentFlags == nil {
			t.Fatal("silent flags should not be nil")
		}
		if !msg.SilentFlags.SuppressPush {
			t.Error("should suppress push")
		}
	})

	t.Run("with custom flags", func(t *testing.T) {
		flags := &SilentFlags{SuppressPush: true, SuppressBadge: false}
		msg, err := ss.ScheduleSilent("user1", conv.ID, []byte("custom"), TextMessage, futureTime, "", flags)
		if err != nil {
			t.Fatalf("ScheduleSilent failed: %v", err)
		}
		if !msg.SilentFlags.SuppressPush {
			t.Error("should suppress push")
		}
		if msg.SilentFlags.SuppressBadge {
			t.Error("should not suppress badge")
		}
	})
}

func TestEditScheduledMessage(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	t.Run("valid edit", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("original"), TextMessage, futureTime, "")
		edited, err := ss.Edit(msg.ID, "user1", []byte("updated"))
		if err != nil {
			t.Fatalf("Edit failed: %v", err)
		}
		if string(edited.Content) != "updated" {
			t.Errorf("content = %s, want 'updated'", string(edited.Content))
		}
	})

	t.Run("nonexistent message", func(t *testing.T) {
		_, err := ss.Edit("nonexistent", "user1", []byte("update"))
		if err != ErrScheduledNotFound {
			t.Errorf("expected ErrScheduledNotFound, got %v", err)
		}
	})

	t.Run("wrong owner", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("mine"), TextMessage, futureTime, "")
		_, err := ss.Edit(msg.ID, "user2", []byte("hijack"))
		if err != ErrScheduledNotOwner {
			t.Errorf("expected ErrScheduledNotOwner, got %v", err)
		}
	})

	t.Run("too close to delivery", func(t *testing.T) {
		soonTime := time.Now().Add(2 * time.Minute)
		msg, _ := ss.Schedule("user1", conv.ID, []byte("soon"), TextMessage, soonTime, "")
		_, err := ss.Edit(msg.ID, "user1", []byte("too late"))
		if err != ErrScheduledEditTooLate {
			t.Errorf("expected ErrScheduledEditTooLate, got %v", err)
		}
	})

	t.Run("already delivered", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("done"), TextMessage, futureTime, "")
		// Manually mark as delivered
		ss.mu.Lock()
		msg.Status = DeliveryDelivered
		ss.mu.Unlock()

		_, err := ss.Edit(msg.ID, "user1", []byte("retry"))
		if err != ErrScheduledAlreadyDelivered {
			t.Errorf("expected ErrScheduledAlreadyDelivered, got %v", err)
		}
	})
}

func TestRescheduleMessage(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)
	newTime := time.Now().Add(2 * time.Hour)

	t.Run("valid reschedule", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("test"), TextMessage, futureTime, "")
		rescheduled, err := ss.Reschedule(msg.ID, "user1", newTime)
		if err != nil {
			t.Fatalf("Reschedule failed: %v", err)
		}
		if !rescheduled.ScheduledAt.Equal(newTime) {
			t.Error("scheduled time was not updated")
		}
		if rescheduled.Status != DeliveryPending {
			t.Errorf("status = %s, want pending", rescheduled.Status)
		}
	})

	t.Run("past time rejected", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("test"), TextMessage, futureTime, "")
		_, err := ss.Reschedule(msg.ID, "user1", time.Now().Add(-1*time.Hour))
		if err != ErrScheduledTimeInPast {
			t.Errorf("expected ErrScheduledTimeInPast, got %v", err)
		}
	})

	t.Run("wrong owner", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("test"), TextMessage, futureTime, "")
		_, err := ss.Reschedule(msg.ID, "user2", newTime)
		if err != ErrScheduledNotOwner {
			t.Errorf("expected ErrScheduledNotOwner, got %v", err)
		}
	})
}

func TestCancelScheduledMessage(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	t.Run("valid cancel", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("cancel me"), TextMessage, futureTime, "")
		err := ss.Cancel(msg.ID, "user1")
		if err != nil {
			t.Fatalf("Cancel failed: %v", err)
		}
		cancelled, _ := ss.Get(msg.ID)
		if cancelled.Status != DeliveryCancelled {
			t.Errorf("status = %s, want cancelled", cancelled.Status)
		}
	})

	t.Run("wrong owner", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("mine"), TextMessage, futureTime, "")
		err := ss.Cancel(msg.ID, "user2")
		if err != ErrScheduledNotOwner {
			t.Errorf("expected ErrScheduledNotOwner, got %v", err)
		}
	})

	t.Run("already delivered", func(t *testing.T) {
		msg, _ := ss.Schedule("user1", conv.ID, []byte("done"), TextMessage, futureTime, "")
		ss.mu.Lock()
		msg.Status = DeliveryDelivered
		ss.mu.Unlock()

		err := ss.Cancel(msg.ID, "user1")
		if err != ErrScheduledAlreadyDelivered {
			t.Errorf("expected ErrScheduledAlreadyDelivered, got %v", err)
		}
	})
}

func TestSendNow(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	t.Run("deliver immediately", func(t *testing.T) {
		scheduled, _ := ss.Schedule("user1", conv.ID, []byte("now!"), TextMessage, futureTime, "")
		delivered, err := ss.SendNow(scheduled.ID, "user1")
		if err != nil {
			t.Fatalf("SendNow failed: %v", err)
		}
		if delivered == nil {
			t.Fatal("delivered message is nil")
		}
		if string(delivered.Content) != "now!" {
			t.Errorf("content = %s, want 'now!'", string(delivered.Content))
		}

		// Verify the scheduled message is marked delivered
		s, _ := ss.Get(scheduled.ID)
		if s.Status != DeliveryDelivered {
			t.Errorf("status = %s, want delivered", s.Status)
		}
		if s.DeliveredAt == nil {
			t.Error("deliveredAt should be set")
		}
	})

	t.Run("wrong owner", func(t *testing.T) {
		scheduled, _ := ss.Schedule("user1", conv.ID, []byte("nope"), TextMessage, futureTime, "")
		_, err := ss.SendNow(scheduled.ID, "user2")
		if err != ErrScheduledNotOwner {
			t.Errorf("expected ErrScheduledNotOwner, got %v", err)
		}
	})
}

func TestProcessDue(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})

	// Create messages: one past due, one future
	pastDue := time.Now().Add(-1 * time.Second)
	future := time.Now().Add(1 * time.Hour)

	// Manually create a past-due message (bypass time validation)
	ss.mu.Lock()
	pastMsg := &ScheduledMessage{
		ID:          "past-msg",
		SenderID:    "user1",
		ConvID:      conv.ID,
		Content:     []byte("overdue"),
		ContentType: TextMessage,
		ScheduledAt: pastDue,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      DeliveryPending,
	}
	ss.messages[pastMsg.ID] = pastMsg
	ss.mu.Unlock()

	// Create a future message normally
	ss.Schedule("user1", conv.ID, []byte("later"), TextMessage, future, "")

	delivered, errs := ss.ProcessDue()
	if len(errs) > 0 {
		t.Fatalf("ProcessDue had errors: %v", errs)
	}
	if len(delivered) != 1 {
		t.Fatalf("expected 1 delivered, got %d", len(delivered))
	}
	if string(delivered[0].Content) != "overdue" {
		t.Errorf("content = %s, want 'overdue'", string(delivered[0].Content))
	}

	// Verify past message is now delivered
	past, _ := ss.Get("past-msg")
	if past.Status != DeliveryDelivered {
		t.Errorf("past message status = %s, want delivered", past.Status)
	}
}

func TestGetPending(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	ss.Schedule("user1", conv.ID, []byte("msg1"), TextMessage, futureTime, "")
	ss.Schedule("user1", conv.ID, []byte("msg2"), TextMessage, futureTime, "")
	ss.Schedule("user2", conv.ID, []byte("msg3"), TextMessage, futureTime, "")

	pending := ss.GetPending("user1")
	if len(pending) != 2 {
		t.Errorf("expected 2 pending for user1, got %d", len(pending))
	}

	pending2 := ss.GetPending("user2")
	if len(pending2) != 1 {
		t.Errorf("expected 1 pending for user2, got %d", len(pending2))
	}
}

func TestCountPending(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	ss.Schedule("user1", conv.ID, []byte("msg1"), TextMessage, futureTime, "")
	ss.Schedule("user1", conv.ID, []byte("msg2"), TextMessage, futureTime, "")

	if count := ss.CountPending("user1"); count != 2 {
		t.Errorf("CountPending = %d, want 2", count)
	}
	if count := ss.CountPending("user3"); count != 0 {
		t.Errorf("CountPending for unknown user = %d, want 0", count)
	}
}

func TestCountPendingForRecipient(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv1, _ := ms.CreateConversation([]string{"user1", "user2"})
	conv2, _ := ms.CreateConversation([]string{"user1", "user3"})
	futureTime := time.Now().Add(1 * time.Hour)

	ss.Schedule("user1", conv1.ID, []byte("msg1"), TextMessage, futureTime, "")
	ss.Schedule("user1", conv1.ID, []byte("msg2"), TextMessage, futureTime, "")
	ss.Schedule("user1", conv2.ID, []byte("msg3"), TextMessage, futureTime, "")

	if count := ss.CountPendingForRecipient("user1", conv1.ID); count != 2 {
		t.Errorf("CountPendingForRecipient conv1 = %d, want 2", count)
	}
	if count := ss.CountPendingForRecipient("user1", conv2.ID); count != 1 {
		t.Errorf("CountPendingForRecipient conv2 = %d, want 1", count)
	}
}

func TestGetByConversation(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	ss.Schedule("user1", conv.ID, []byte("msg1"), TextMessage, futureTime, "")
	ss.Schedule("user1", conv.ID, []byte("msg2"), TextMessage, futureTime, "")

	msgs := ss.GetByConversation("user1", conv.ID)
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestDeliverySilentScheduled(t *testing.T) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})

	// Create a past-due silent scheduled message
	ss.mu.Lock()
	msg := &ScheduledMessage{
		ID:          "silent-scheduled",
		SenderID:    "user1",
		ConvID:      conv.ID,
		Content:     []byte("quiet delivery"),
		ContentType: TextMessage,
		ScheduledAt: time.Now().Add(-1 * time.Second),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      DeliveryPending,
		Silent:      true,
		SilentFlags: DefaultSilentFlags(),
	}
	ss.messages[msg.ID] = msg
	ss.mu.Unlock()

	delivered, errs := ss.ProcessDue()
	if len(errs) > 0 {
		t.Fatalf("ProcessDue had errors: %v", errs)
	}
	if len(delivered) != 1 {
		t.Fatalf("expected 1 delivered, got %d", len(delivered))
	}
	if !delivered[0].Silent {
		t.Error("delivered message should be silent")
	}
	if delivered[0].SilentFlags == nil {
		t.Error("delivered message should have silent flags")
	}
}

func BenchmarkScheduleMessage(b *testing.B) {
	ss, ms := newTestScheduledService()
	conv, _ := ms.CreateConversation([]string{"user1", "user2"})
	futureTime := time.Now().Add(1 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss.Schedule("user1", conv.ID, []byte("bench"), TextMessage, futureTime, "")
	}
}
