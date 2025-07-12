package simulator

import (
	"context"
	"fmt"
	"log"
	"time"

	"dhcp-monitoring-app/dhcp"
	"dhcp-monitoring-app/kafka"
)

type NetworkMonitoringSimulator struct {
	producer *kafka.DHCPEventProducer
}

func NewNetworkMonitoringSimulator(producer *kafka.DHCPEventProducer) *NetworkMonitoringSimulator {
	return &NetworkMonitoringSimulator{producer: producer}
}

func (s *NetworkMonitoringSimulator) SimulateEvents(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	eventID := 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			event := s.generateRandomEvent(eventID)
			if err := s.producer.PublishEvent(event); err != nil {
				log.Printf("Failed to publish simulated event: %v", err)
			}
			eventID++
		}
	}
}

func (s *NetworkMonitoringSimulator) generateRandomEvent(id int) dhcp.DHCPSecurityEvent {
	eventTypes := []dhcp.DHCPEventType{
		dhcp.DHCPDiscover, dhcp.DHCPOffer, dhcp.DHCPRequest, dhcp.DHCPAck,
		dhcp.DHCPRogueServer, dhcp.DHCPStarvation, dhcp.DHCPSpoofing,
	}

	eventType := eventTypes[id%len(eventTypes)]
	severity := "LOW"

	switch eventType {
	case dhcp.DHCPRogueServer, dhcp.DHCPSpoofing:
		severity = "HIGH"
	case dhcp.DHCPStarvation:
		severity = "MEDIUM"
	}

	return dhcp.DHCPSecurityEvent{
		ID:            fmt.Sprintf("event_%d", id),
		Timestamp:     time.Now(),
		EventType:     eventType,
		Severity:      severity,
		SourceIP:      fmt.Sprintf("192.168.1.%d", (id%254)+1),
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
