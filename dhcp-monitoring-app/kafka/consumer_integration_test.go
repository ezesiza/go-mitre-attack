//go:build integration
// +build integration

package kafka

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"dhcp-monitoring-app/models"
)

type mockProcessor struct {
	received []models.DHCPSecurityEvent
	mu       sync.Mutex
}

func (m *mockProcessor) ProcessEvent(event models.DHCPSecurityEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.received = append(m.received, event)
	return nil
}

func TestKafkaProducerConsumerIntegration(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topic := "test-dhcp-events"
	groupID := "test-group"

	// Check if Kafka is available
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Setup mock processor
	processor := &mockProcessor{}

	// Start consumer
	consumer, err := NewDHCPEventConsumer(brokers, groupID, []string{topic}, processor)
	if err != nil {
		t.Skipf("Kafka not available: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = consumer.Start(ctx)
	}()
	defer consumer.Close()

	// Give the consumer a moment to start
	time.Sleep(2 * time.Second)

	// Create a producer (assume you have a NewProducer function)
	producer, err := NewProducer(brokers)
	if err != nil {
		t.Fatalf("Failed to create producer: %v", err)
	}

	event := models.DHCPSecurityEvent{
		EventType:   models.DHCPDiscover,
		SourceIP:    "192.168.1.100",
		Timestamp:   time.Now(),
		Description: "Integration test event",
	}
	value, _ := json.Marshal(event)
	if err := producer.Send(topic, value, 5*time.Second); err != nil {
		t.Fatalf("Failed to send event: %v", err)
	}

	// Wait for the consumer to process the event
	time.Sleep(3 * time.Second)

	processor.mu.Lock()
	received := len(processor.received)
	processor.mu.Unlock()
	if received == 0 {
		t.Error("Expected to receive at least one event, got 0")
	}
}
