package models

import (
	"dhcp-monitoring-app/interfaces"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type DHCPEventProcessor struct {
	// alertProducer will be set later, after Kafka modularization
	alertProducer   interfaces.EventProducer
	eventCache      map[string][]DHCPSecurityEvent
	cacheMutex      sync.RWMutex
	alertRules      []SecurityRule
	websocketServer interfaces.WebSocketServer
}

// NewDHCPEventProcessor creates a new event processor
func NewDHCPEventProcessor(alertProducer interfaces.EventProducer, websocketServer interfaces.WebSocketServer) *DHCPEventProcessor {
	processor := &DHCPEventProcessor{
		alertProducer:   alertProducer,
		eventCache:      make(map[string][]DHCPSecurityEvent),
		websocketServer: websocketServer,
		alertRules: []SecurityRule{
			{
				Name:        "DHCP Starvation Attack",
				EventTypes:  []DHCPEventType{DHCPDiscover, DHCPRequest},
				Threshold:   10,
				TimeWindow:  30 * time.Second,
				Severity:    "HIGH",
				Description: "Potential DHCP starvation attack detected - multiple DHCP requests from same client",
			},
			{
				Name:        "Rogue DHCP Server",
				EventTypes:  []DHCPEventType{DHCPRogueServer},
				Threshold:   1,
				TimeWindow:  1 * time.Minute,
				Severity:    "CRITICAL",
				Description: "Rogue DHCP server detected on network",
			},
			{
				Name:        "DHCP Spoofing",
				EventTypes:  []DHCPEventType{DHCPSpoofing},
				Threshold:   2,
				TimeWindow:  1 * time.Minute,
				Severity:    "HIGH",
				Description: "DHCP spoofing attack detected - unauthorized DHCP responses",
			},
			{
				Name:        "DHCP Pool Exhaustion",
				EventTypes:  []DHCPEventType{DHCPExhaustion},
				Threshold:   1,
				TimeWindow:  1 * time.Minute,
				Severity:    "HIGH",
				Description: "DHCP address pool exhaustion detected",
			},
			{
				Name:        "Suspicious DHCP Decline",
				EventTypes:  []DHCPEventType{DHCPDecline},
				Threshold:   3,
				TimeWindow:  1 * time.Minute,
				Severity:    "MEDIUM",
				Description: "Multiple DHCP decline messages detected - possible address conflict",
			},
			{
				Name:        "Rapid DHCP Release",
				EventTypes:  []DHCPEventType{DHCPRelease},
				Threshold:   5,
				TimeWindow:  30 * time.Second,
				Severity:    "MEDIUM",
				Description: "Rapid DHCP release messages detected - possible network scanning",
			},
			{
				Name:        "DHCP Discover Flood",
				EventTypes:  []DHCPEventType{DHCPDiscover},
				Threshold:   15,
				TimeWindow:  1 * time.Minute,
				Severity:    "HIGH",
				Description: "DHCP discover flood detected - potential network reconnaissance",
			},
			{
				Name:        "Multiple DHCP Offers",
				EventTypes:  []DHCPEventType{DHCPOffer},
				Threshold:   3,
				TimeWindow:  30 * time.Second,
				Severity:    "MEDIUM",
				Description: "Multiple DHCP offers detected - possible unauthorized DHCP server",
			},
			{
				Name:        "DHCP Request Storm",
				EventTypes:  []DHCPEventType{DHCPRequest},
				Threshold:   20,
				TimeWindow:  1 * time.Minute,
				Severity:    "HIGH",
				Description: "DHCP request storm detected - potential DoS attack",
			},
			{
				Name:        "Excessive DHCP ACKs",
				EventTypes:  []DHCPEventType{DHCPAck},
				Threshold:   25,
				TimeWindow:  1 * time.Minute,
				Severity:    "MEDIUM",
				Description: "Excessive DHCP ACK messages - unusual network activity",
			},
			{
				Name:        "DHCP NAK Flood",
				EventTypes:  []DHCPEventType{DHCPNak},
				Threshold:   10,
				TimeWindow:  1 * time.Minute,
				Severity:    "HIGH",
				Description: "DHCP NAK flood detected - possible configuration issues or attack",
			},
			{
				Name:        "DHCP Decline Storm",
				EventTypes:  []DHCPEventType{DHCPDecline},
				Threshold:   8,
				TimeWindow:  30 * time.Second,
				Severity:    "HIGH",
				Description: "DHCP decline storm detected - possible address conflict attack",
			},
			{
				Name:        "Rapid IP Cycling",
				EventTypes:  []DHCPEventType{DHCPRequest, DHCPRelease},
				Threshold:   1,
				TimeWindow:  1 * time.Minute,
				Severity:    "MEDIUM",
				Description: "Rapid IP address cycling detected - possible network scanning",
			},
			{
				Name:        "DHCP Man-in-the-Middle",
				EventTypes:  []DHCPEventType{DHCPOffer, DHCPSpoofing},
				Threshold:   2,
				TimeWindow:  30 * time.Second,
				Severity:    "CRITICAL",
				Description: "Potential DHCP man-in-the-middle attack detected",
			},
			{
				Name:        "Network Reconnaissance",
				EventTypes:  []DHCPEventType{DHCPDiscover, DHCPRequest, DHCPRelease},
				Threshold:   30,
				TimeWindow:  2 * time.Minute,
				Severity:    "HIGH",
				Description: "Network reconnaissance activity detected through DHCP patterns",
			},
			{
				Name:        "DHCP Amplification Attack",
				EventTypes:  []DHCPEventType{DHCPDiscover, DHCPOffer},
				Threshold:   50,
				TimeWindow:  1 * time.Minute,
				Severity:    "CRITICAL",
				Description: "DHCP amplification attack detected - high volume of discover/offer messages",
			},
			{
				Name:        "Suspicious Lease Patterns",
				EventTypes:  []DHCPEventType{DHCPRequest, DHCPAck},
				Threshold:   15,
				TimeWindow:  30 * time.Second,
				Severity:    "MEDIUM",
				Description: "Suspicious DHCP lease patterns detected - possible automated attack",
			},
			{
				Name:        "DHCP Server Impersonation",
				EventTypes:  []DHCPEventType{DHCPOffer, DHCPAck, DHCPRogueServer},
				Threshold:   3,
				TimeWindow:  1 * time.Minute,
				Severity:    "CRITICAL",
				Description: "DHCP server impersonation detected - unauthorized server responses",
			},
			{
				Name:        "Address Pool Manipulation",
				EventTypes:  []DHCPEventType{DHCPRequest, DHCPDecline, DHCPExhaustion},
				Threshold:   20,
				TimeWindow:  2 * time.Minute,
				Severity:    "HIGH",
				Description: "Address pool manipulation detected - coordinated attack on DHCP pool",
			},
			{
				Name:        "DHCP Protocol Anomaly",
				EventTypes:  []DHCPEventType{DHCPDiscover, DHCPOffer, DHCPRequest, DHCPAck, DHCPNak, DHCPRelease, DHCPDecline},
				Threshold:   100,
				TimeWindow:  5 * time.Minute,
				Severity:    "MEDIUM",
				Description: "DHCP protocol anomaly detected - unusual combination of message types",
			},
		},
	}

	go processor.cleanupCache()

	return processor
}

// ProcessEvent processes incoming DHCP events
func (p *DHCPEventProcessor) ProcessEvent(event interface{}) error {
	dhcpEvent, ok := event.(DHCPSecurityEvent)
	if !ok {
		return fmt.Errorf("invalid event type: expected DHCPSecurityEvent, got %T", event)
	}

	p.cacheMutex.Lock()
	defer p.cacheMutex.Unlock()

	// Cache the event
	key := fmt.Sprintf("%s_%s", dhcpEvent.SourceIP, dhcpEvent.EventType)
	// Appending to a nil slice is safe in Go; no need to check if key exists
	p.eventCache[key] = append(p.eventCache[key], dhcpEvent)

	// Broadcast every event to WebSocket clients
	if p.websocketServer != nil {
		if data, err := json.Marshal(dhcpEvent); err == nil {
			log.Printf("Broadcasting event to WebSocket clients: %s", string(data))
			p.websocketServer.Broadcast(data)
		}
	}

	// Check security rules
	for _, rule := range p.alertRules {
		if p.checkRule(rule, dhcpEvent) {
			alert := p.generateAlert(rule, dhcpEvent)
			if err := p.publishAlert(alert); err != nil {
				log.Printf("Failed to publish alert: %v", err)
			}
		}
	}

	return nil
}

// checkRule checks if a security rule is triggered
func (p *DHCPEventProcessor) checkRule(rule SecurityRule, event DHCPSecurityEvent) bool {
	// Check if event type matches rule
	eventTypeMatch := false
	for _, eventType := range rule.EventTypes {
		if event.EventType == eventType {
			eventTypeMatch = true
			break
		}
	}
	if !eventTypeMatch {
		return false
	}

	// Count events within time window
	key := fmt.Sprintf("%s_%s", event.SourceIP, event.EventType)
	events := p.eventCache[key]

	count := 0
	cutoff := time.Now().Add(-rule.TimeWindow)

	for _, cachedEvent := range events {
		if cachedEvent.Timestamp.After(cutoff) {
			count++
		}
	}
	// log.Printf("count: %d, rule.Threshold: %d", count, rule.Threshold)
	return count >= rule.Threshold
}

// generateAlert generates a security alert
func (p *DHCPEventProcessor) generateAlert(rule SecurityRule, triggerEvent DHCPSecurityEvent) SecurityAlert {
	return SecurityAlert{
		ID:          fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		AlertType:   rule.Name,
		Severity:    rule.Severity,
		Source:      triggerEvent.SourceIP,
		Description: rule.Description,
		Events:      []DHCPSecurityEvent{triggerEvent},
		Metadata: map[string]interface{}{
			"rule_name":  rule.Name,
			"network_id": triggerEvent.NetworkID,
			"client_mac": triggerEvent.ClientMAC,
			"server_mac": triggerEvent.ServerMAC,
		},
	}
}

// publishAlert publishes a security alert
func (p *DHCPEventProcessor) publishAlert(alert SecurityAlert) error {
	if p.websocketServer != nil {
		if data, err := json.Marshal(alert); err == nil {
			p.websocketServer.Broadcast(data)
		}
	}
	return p.publishToKafka(alert)
}

func (p *DHCPEventProcessor) publishToKafka(alert SecurityAlert) error {
	if p.alertProducer == nil {
		return fmt.Errorf("alert producer is not set")
	}
	if len(alert.Events) == 0 {
		log.Printf("Warning: SecurityAlert has no event to publish:%+v", alert)
	}
	return p.alertProducer.PublishEvent(alert.Events[0])
}

// cleanupCache removes old events from cache
func (p *DHCPEventProcessor) cleanupCache() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.cacheMutex.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)

		for key, events := range p.eventCache {
			filteredEvents := make([]DHCPSecurityEvent, 0)
			for _, event := range events {
				if event.Timestamp.After(cutoff) {
					filteredEvents = append(filteredEvents, event)
				}
			}

			if len(filteredEvents) == 0 {
				delete(p.eventCache, key)
			} else {
				p.eventCache[key] = filteredEvents
			}
		}
		p.cacheMutex.Unlock()
	}
}
