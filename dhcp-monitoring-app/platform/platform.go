package platform

import (
	"context"
	"fmt"
	"log"
	"sync"

	"dhcp-monitoring-app/dhcp"
	"dhcp-monitoring-app/kafka"
	"dhcp-monitoring-app/simulator"
)

type DHCPSecurityPlatform struct {
	config        kafka.KafkaConfig
	eventProducer *kafka.DHCPEventProducer
	alertProducer *kafka.DHCPEventProducer
	processor     *dhcp.DHCPEventProcessor
	consumer      *kafka.DHCPEventConsumer
	simulator     *simulator.NetworkMonitoringSimulator
}

func NewDHCPSecurityPlatform(config kafka.KafkaConfig) (*DHCPSecurityPlatform, error) {
	// Create event producer
	eventProducer, err := kafka.NewDHCPEventProducer(config.Brokers, config.DHCPEventsTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create event producer: %w", err)
	}

	// Create alert producer
	alertProducer, err := kafka.NewDHCPEventProducer(config.Brokers, config.AlertsTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert producer: %w", err)
	}

	// Create processor
	processor := dhcp.NewDHCPEventProcessor(alertProducer)

	// Create consumer
	consumer, err := kafka.NewDHCPEventConsumer(
		config.Brokers,
		config.ConsumerGroup,
		[]string{config.DHCPEventsTopic},
		processor,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Create simulator
	sim := simulator.NewNetworkMonitoringSimulator(eventProducer)

	return &DHCPSecurityPlatform{
		config:        config,
		eventProducer: eventProducer,
		alertProducer: alertProducer,
		processor:     processor,
		consumer:      consumer,
		simulator:     sim,
	}, nil
}

func (p *DHCPSecurityPlatform) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	// Start consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := p.consumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Start simulator (for demo purposes)
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.simulator.SimulateEvents(ctx)
	}()

	log.Println("DHCP Security Platform started successfully")
	wg.Wait()
	return nil
}

func (p *DHCPSecurityPlatform) Stop() error {
	if err := p.consumer.Close(); err != nil {
		log.Printf("Error closing consumer: %v", err)
	}
	if err := p.eventProducer.Close(); err != nil {
		log.Printf("Error closing event producer: %v", err)
	}
	if err := p.alertProducer.Close(); err != nil {
		log.Printf("Error closing alert producer: %v", err)
	}
	return nil
}
