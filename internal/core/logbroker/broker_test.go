package logbroker

import (
	"testing"
	"time"
)

func TestBroker_PublishAndSubscribe(t *testing.T) {
	broker := NewBroker()
	jobID := "test-job-123"

	// Subscribe to job logs
	ch := broker.Subscribe(jobID)

	// Publish a log message
	msg := LogMessage{
		Timestamp: time.Now(),
		Level:     "INFO",
		Component: "test",
		Message:   "Test message",
	}

	broker.Publish(jobID, msg)

	// Receive the message
	select {
	case received := <-ch:
		if received.Message != msg.Message {
			t.Errorf("Message = %v, want %v", received.Message, msg.Message)
		}
		if received.Level != msg.Level {
			t.Errorf("Level = %v, want %v", received.Level, msg.Level)
		}
	case <-time.After(time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	broker := NewBroker()
	jobID := "test-job-456"

	// Create multiple subscribers
	ch1 := broker.Subscribe(jobID)
	ch2 := broker.Subscribe(jobID)
	ch3 := broker.Subscribe(jobID)

	// Publish a message
	msg := LogMessage{
		Timestamp: time.Now(),
		Level:     "WARN",
		Component: "test",
		Message:   "Warning message",
	}

	broker.Publish(jobID, msg)

	// All subscribers should receive the message
	received := 0
	timeout := time.After(time.Second)

	for i := 0; i < 3; i++ {
		select {
		case <-ch1:
			received++
		case <-ch2:
			received++
		case <-ch3:
			received++
		case <-timeout:
			t.Errorf("Only received %d messages, want 3", received)
			return
		}
	}

	if received != 3 {
		t.Errorf("Received = %d, want 3", received)
	}
}

func TestBroker_History(t *testing.T) {
	broker := NewBroker()
	jobID := "test-job-789"

	// Publish multiple messages
	for i := 0; i < 5; i++ {
		msg := LogMessage{
			Timestamp: time.Now(),
			Level:     "INFO",
			Component: "test",
			Message:   "Message " + string(rune('0'+i)),
		}
		broker.Publish(jobID, msg)
	}

	// Get history
	history := broker.GetHistory(jobID)

	if len(history) != 5 {
		t.Errorf("History length = %d, want 5", len(history))
	}
}

func TestBroker_HistoryLimit(t *testing.T) {
	broker := NewBroker()
	jobID := "test-job-limit"

	// Publish more than 1000 messages
	for i := 0; i < 1500; i++ {
		msg := LogMessage{
			Timestamp: time.Now(),
			Level:     "INFO",
			Component: "test",
			Message:   "Message",
		}
		broker.Publish(jobID, msg)
	}

	// History should be limited to 1000
	history := broker.GetHistory(jobID)

	if len(history) != 1000 {
		t.Errorf("History length = %d, want 1000 (limit)", len(history))
	}
}

func TestBroker_SubscribeWithHistory(t *testing.T) {
	broker := NewBroker()
	jobID := "test-job-history"

	// Publish some messages before subscribing
	for i := 0; i < 3; i++ {
		msg := LogMessage{
			Timestamp: time.Now(),
			Level:     "INFO",
			Component: "test",
			Message:   "Historical message",
		}
		broker.Publish(jobID, msg)
	}

	// Subscribe (should receive history)
	ch := broker.Subscribe(jobID)

	// Should receive historical messages
	received := 0
	timeout := time.After(time.Second)

	for received < 3 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Errorf("Only received %d historical messages, want 3", received)
			return
		}
	}

	if received != 3 {
		t.Errorf("Received = %d historical messages, want 3", received)
	}
}

func TestBroker_ClearHistory(t *testing.T) {
	broker := NewBroker()
	jobID := "test-job-clear"

	// Publish messages
	for i := 0; i < 5; i++ {
		msg := LogMessage{
			Timestamp: time.Now(),
			Level:     "INFO",
			Component: "test",
			Message:   "Message",
		}
		broker.Publish(jobID, msg)
	}

	// Clear history
	broker.ClearHistory(jobID)

	// History should be empty
	history := broker.GetHistory(jobID)

	if len(history) != 0 {
		t.Errorf("History length = %d after clear, want 0", len(history))
	}
}

func TestLogMessage_FormatAsHTML(t *testing.T) {
	tests := []struct {
		name      string
		msg       LogMessage
		wantClass string
	}{
		{
			name: "ERROR level",
			msg: LogMessage{
				Timestamp: time.Now(),
				Level:     "ERROR",
				Message:   "Error message",
			},
			wantClass: "text-rose-500",
		},
		{
			name: "WARN level",
			msg: LogMessage{
				Timestamp: time.Now(),
				Level:     "WARN",
				Message:   "Warning message",
			},
			wantClass: "text-amber-500",
		},
		{
			name: "INFO level",
			msg: LogMessage{
				Timestamp: time.Now(),
				Level:     "INFO",
				Message:   "Info message",
			},
			wantClass: "text-emerald-500",
		},
		{
			name: "DEBUG level",
			msg: LogMessage{
				Timestamp: time.Now(),
				Level:     "DEBUG",
				Message:   "Debug message",
			},
			wantClass: "text-slate-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := tt.msg.FormatAsHTML()

			// Check if HTML contains the expected class
			if !contains(html, tt.wantClass) {
				t.Errorf("FormatAsHTML() missing class %v in output: %v", tt.wantClass, html)
			}

			// Check if HTML contains the message
			if !contains(html, tt.msg.Message) {
				t.Errorf("FormatAsHTML() missing message %v", tt.msg.Message)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
