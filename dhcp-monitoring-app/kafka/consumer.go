package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"dhcp-monitoring-app/interfaces"
	"dhcp-monitoring-app/models"

	"github.com/IBM/sarama"
)

type DHCPEventConsumer struct {
	consumer  sarama.ConsumerGroup
	processor interfaces.EventProcessor
	topics    []string
}

func NewDHCPEventConsumer(brokers []string, groupID string, topics []string, processor interfaces.EventProcessor) (*DHCPEventConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &DHCPEventConsumer{
		consumer:  consumer,
		processor: processor,
		topics:    topics,
	}, nil
}

func (c *DHCPEventConsumer) Start(ctx context.Context) error {
	handler := &ConsumerGroupHandler{processor: c.processor}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.consumer.Consume(ctx, c.topics, handler); err != nil {
				log.Printf("Error consuming messages: %v", err)
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func (c *DHCPEventConsumer) Close() error {
	return c.consumer.Close()
}

type ConsumerGroupHandler struct {
	processor interfaces.EventProcessor
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var event models.DHCPSecurityEvent
		if err := json.Unmarshal(message.Value, &event); err == nil {
			log.Printf("Consumer received event from Kafka: type=%s, source_ip=%s", event.EventType, event.SourceIP)
		}
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		if err := h.processor.ProcessEvent(event); err != nil {
			log.Printf("Failed to process event: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}
