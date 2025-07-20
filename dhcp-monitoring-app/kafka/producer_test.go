package kafka

import (
	"testing"
	"time"
)

type mockProducer struct {
	sent bool
}

func (m *mockProducer) Send(topic string, message []byte, timeout time.Duration) error {
	m.sent = true
	return nil
}

func TestProducer_Send(t *testing.T) {
	mock := &mockProducer{}
	err := mock.Send("test-topic", []byte("hello"), time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !mock.sent {
		t.Error("expected message to be sent")
	}
}
