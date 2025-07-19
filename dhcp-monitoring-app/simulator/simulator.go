package simulator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"dhcp-monitoring-app/config"
	"dhcp-monitoring-app/interfaces"
	"dhcp-monitoring-app/models"
)

type NetworkMonitoringSimulator struct {
	producer interfaces.EventProducer
	config   *config.SimulationConfig
}

func NewNetworkMonitoringSimulator(producer interfaces.EventProducer) *NetworkMonitoringSimulator {
	return &NetworkMonitoringSimulator{
		producer: producer,
		config:   &config.SimulationConfig{},
	}
}

// NewNetworkMonitoringSimulatorWithConfig creates a simulator with custom configuration
func NewNetworkMonitoringSimulatorWithConfig(producer interfaces.EventProducer, cfg *config.SimulationConfig) *NetworkMonitoringSimulator {
	return &NetworkMonitoringSimulator{
		producer: producer,
		config:   cfg,
	}
}

func (s *NetworkMonitoringSimulator) SimulateEvents(ctx context.Context) {
	interval := s.config.Interval
	if interval == 0 {
		interval = 2 * time.Second // Default interval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	eventID := 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			event := s.generateRandomEvent(eventID)
			// Log event generation
			log.Printf("Simulator generated event: type=%s, source_ip=%s", event.EventType, event.SourceIP)
			if err := s.producer.PublishEvent(event); err != nil {
				log.Printf("Failed to publish simulated event: %v", err)
			} else {
				log.Printf("Simulator published event: type=%s, source_ip=%s", event.EventType, event.SourceIP)
			}
			eventID++
		}
	}
}

func (s *NetworkMonitoringSimulator) generateRandomEvent(id int) models.DHCPSecurityEvent {
	// Use configured event types if available, otherwise use defaults
	eventTypes := s.config.EventTypes
	if len(eventTypes) == 0 {
		eventTypes = []string{
			"DHCP_DISCOVER", "DHCP_OFFER", "DHCP_REQUEST", "DHCP_ACK",
			"DHCP_ROGUE_SERVER", "DHCP_STARVATION", "DHCP_SPOOFING",
		}
	}

	// Convert string event types to DHCPEventType
	dhcpEventTypes := make([]models.DHCPEventType, len(eventTypes))
	for i, eventType := range eventTypes {
		dhcpEventTypes[i] = models.DHCPEventType(eventType)
	}

	eventType := dhcpEventTypes[id%len(dhcpEventTypes)]
	severity := "LOW"

	switch eventType {
	case models.DHCPRogueServer, models.DHCPSpoofing:
		severity = "HIGH"
	case models.DHCPStarvation:
		severity = "MEDIUM"
	}

	// Use configured source IPs if available, otherwise use defaults
	sourceIPs := s.config.SourceIPs
	if len(sourceIPs) == 0 {
		sourceIPs = []string{"192.168.1.100", "192.168.1.101", "192.168.1.102"}
	}

	// Randomly select a source IP from the configured list
	selectedIP := sourceIPs[rand.Intn(len(sourceIPs))]

	return models.DHCPSecurityEvent{
		ID:            fmt.Sprintf("event_%d", id),
		Timestamp:     time.Now(),
		EventType:     eventType,
		Severity:      severity,
		SourceIP:      selectedIP,
		DestIP:        "192.168.1.1",
		ClientMAC:     fmt.Sprintf("aa:bb:cc:dd:ee:%02x", id%256),
		ServerMAC:     "00:11:22:33:44:55",
		TransactionID: uint32(id),
		RequestedIP:   fmt.Sprintf("192.168.1.%d", (id%100)+100),
		OfferedIP:     fmt.Sprintf("192.168.1.%d", (id%100)+100),
		LeaseTime:     3600,
		NetworkID:     "network_001",
		Description:   fmt.Sprintf("DHCP %s event from monitoring tool", eventType),
		Metadata: map[string]interface{}{
			"monitoring_tool": "NetworkAnalyzer",
			"interface":       "eth0",
			"vlan_id":         100,
		},
	}
}
