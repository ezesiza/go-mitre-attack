package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"dhcp-monitoring-app/config"

	"dhcp-monitoring-app/platform"
)

func main() {
	// Load configuration
	cfg := config.NewDefaultConfig()
	if err := cfg.LoadFromEnvironment(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Log configuration summary
	log.Printf("Starting DHCP Security Platform with configuration:")
	log.Printf("  Environment: %s", cfg.App.Environment)
	log.Printf("  Kafka Brokers: %v", cfg.Kafka.Brokers)
	log.Printf("  DHCP Events Topic: %s", cfg.Kafka.DHCPEventsTopic)
	log.Printf("  Alerts Topic: %s", cfg.Kafka.AlertsTopic)
	log.Printf("  Simulation Enabled: %t", cfg.App.Simulation.Enabled)

	// Create platform with injected configuration
	// p, err := platform.NewDHCPSecurityPlatform(cfg)
	container, err := platform.NewDefaultServiceContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to create service container: %v", err)
	}

	// Build platform using the builder pattern
	p, err := platform.NewPlatformBuilder().WithServiceContainer(container).Build()
	// if err != nil {
	// log.Fatalf("Failed to build platform: %v", err)
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
	// }
}
