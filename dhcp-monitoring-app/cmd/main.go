package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dhcp-monitoring-app/kafka"
	"dhcp-monitoring-app/platform"
)

func main() {
	// Kafka configuration
	config := kafka.KafkaConfig{
		Brokers:         []string{"localhost:9092"},
		DHCPEventsTopic: "dhcp-security-events",
		AlertsTopic:     "dhcp-security-alerts",
		ConsumerGroup:   "dhcp-security-processors",
	}

	// Create platform
	p, err := platform.NewDHCPSecurityPlatform(config)
	if err != nil {
		log.Fatalf("Failed to create platform: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down platform...")
		cancel()
		p.Stop()
	}()

	// Start platform
	log.Println("Starting DHCP Security Event Streaming Platform...")
	if err := p.Start(ctx); err != nil {
		log.Fatalf("Platform error: %v", err)
	}

	log.Println("Platform shutdown complete")
}
