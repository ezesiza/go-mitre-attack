package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"dhcp-monitoring-app/models"

	"github.com/IBM/sarama"
)

type DHCPEventProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewDHCPEventProducer(brokers []string, topic string) (*DHCPEventProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &DHCPEventProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *DHCPEventProducer) PublishEvent(event models.DHCPSecurityEvent) error {
	// dhcpEvent, ok := event.(models.DHCPSecurityEvent)
	// if !ok {
	// 	return fmt.Errorf("invalid event type: expected DHCPSecurityEvent, got %T", event)
	// }

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(event.ID),
		Value: sarama.ByteEncoder(eventJSON),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event_type"),
				Value: []byte(event.EventType),
			},
			{
				Key:   []byte("severity"),
				Value: []byte(event.Severity),
			},
		},
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Event published to partition %d at offset %d", partition, offset)
	return nil
}

func (p *DHCPEventProducer) Close() error {
	return p.producer.Close()
}
