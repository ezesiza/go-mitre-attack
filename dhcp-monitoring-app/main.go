package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// DHCPEventType represents different types of DHCP security events
type DHCPEventType string

const (
	DHCPDiscover    DHCPEventType = "DHCP_DISCOVER"
	DHCPOffer       DHCPEventType = "DHCP_OFFER"
	DHCPRequest     DHCPEventType = "DHCP_REQUEST"
	DHCPAck         DHCPEventType = "DHCP_ACK"
	DHCPNak         DHCPEventType = "DHCP_NAK"
	DHCPRelease     DHCPEventType = "DHCP_RELEASE"
	DHCPDecline     DHCPEventType = "DHCP_DECLINE"
	DHCPRogueServer DHCPEventType = "DHCP_ROGUE_SERVER"
	DHCPStarvation  DHCPEventType = "DHCP_STARVATION"
	DHCPSpoofing    DHCPEventType = "DHCP_SPOOFING"
	DHCPExhaustion  DHCPEventType = "DHCP_POOL_EXHAUSTION"
)

// DHCPSecurityEvent represents a DHCP-related security event
type DHCPSecurityEvent struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	EventType     DHCPEventType          `json:"event_type"`
	Severity      string                 `json:"severity"`
	SourceIP      string                 `json:"source_ip"`
	DestIP        string                 `json:"dest_ip"`
	ClientMAC     string                 `json:"client_mac"`
	ServerMAC     string                 `json:"server_mac"`
	TransactionID uint32                 `json:"transaction_id"`
	RequestedIP   string                 `json:"requested_ip"`
	OfferedIP     string                 `json:"offered_ip"`
	LeaseTime     int                    `json:"lease_time"`
	NetworkID     string                 `json:"network_id"`
	Description   string                 `json:"description"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// SecurityAlert represents processed security alerts
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	AlertType   string                 `json:"alert_type"`
	Severity    string                 `json:"severity"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Events      []DHCPSecurityEvent    `json:"events"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers         []string
	DHCPEventsTopic string
	AlertsTopic     string
	ConsumerGroup   string
}

// DHCPEventProducer handles publishing DHCP events to Kafka
type DHCPEventProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewDHCPEventProducer creates a new DHCP event producer
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

// PublishEvent publishes a DHCP security event to Kafka
func (p *DHCPEventProducer) PublishEvent(event DHCPSecurityEvent) error {
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

// Close closes the producer
func (p *DHCPEventProducer) Close() error {
	return p.producer.Close()
}

// DHCPEventProcessor processes DHCP events and generates security alerts
type DHCPEventProcessor struct {
	alertProducer *DHCPEventProducer
	eventCache    map[string][]DHCPSecurityEvent
	cacheMutex    sync.RWMutex
	alertRules    []SecurityRule
}

// SecurityRule defines rules for generating security alerts
type SecurityRule struct {
	Name        string
	EventTypes  []DHCPEventType
	Threshold   int
	TimeWindow  time.Duration
	Severity    string
	Description string
}

// NewDHCPEventProcessor creates a new event processor
func NewDHCPEventProcessor(alertProducer *DHCPEventProducer) *DHCPEventProcessor {
	processor := &DHCPEventProcessor{
		alertProducer: alertProducer,
		eventCache:    make(map[string][]DHCPSecurityEvent),
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
				Threshold:   12,
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
func (p *DHCPEventProcessor) ProcessEvent(event DHCPSecurityEvent) error {
	p.cacheMutex.Lock()
	defer p.cacheMutex.Unlock()

	// Cache the event
	key := fmt.Sprintf("%s_%s", event.SourceIP, event.EventType)
	// Appending to a nil slice is safe in Go; no need to check if key exists
	p.eventCache[key] = append(p.eventCache[key], event)

	// Check security rules
	for _, rule := range p.alertRules {
		if p.checkRule(rule, event) {
			alert := p.generateAlert(rule, event)
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
	log.Printf("count: %d, rule.Threshold: %d", count, rule.Threshold)
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
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "dhcp-security-alerts",
		Key:   sarama.StringEncoder(alert.ID),
		Value: sarama.ByteEncoder(alertJSON),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("alert_type"),
				Value: []byte(alert.AlertType),
			},
			{
				Key:   []byte("severity"),
				Value: []byte(alert.Severity),
			},
		},
	}

	_, _, err = p.alertProducer.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send alert: %w", err)
	}

	log.Printf("Security alert published: %s - %s", alert.AlertType, alert.Description)
	return nil
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

type DHCPEventConsumer struct {
	consumer  sarama.ConsumerGroup
	processor *DHCPEventProcessor
	topics    []string
}

// NewDHCPEventConsumer creates a new DHCP event consumer
func NewDHCPEventConsumer(brokers []string, groupID string, topics []string, processor *DHCPEventProcessor) (*DHCPEventConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
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

// Start starts consuming messages
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

// Close closes the consumer
func (c *DHCPEventConsumer) Close() error {
	return c.consumer.Close()
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	processor *DHCPEventProcessor
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var event DHCPSecurityEvent
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

// NetworkMonitoringSimulator simulates network monitoring tools generating DHCP events
type NetworkMonitoringSimulator struct {
	producer *DHCPEventProducer
}

// NewNetworkMonitoringSimulator creates a new simulator
func NewNetworkMonitoringSimulator(producer *DHCPEventProducer) *NetworkMonitoringSimulator {
	return &NetworkMonitoringSimulator{producer: producer}
}

// SimulateEvents generates sample DHCP security events
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

// generateRandomEvent generates a random DHCP security event
func (s *NetworkMonitoringSimulator) generateRandomEvent(id int) DHCPSecurityEvent {
	eventTypes := []DHCPEventType{
		DHCPDiscover, DHCPOffer, DHCPRequest, DHCPAck,
		DHCPRogueServer, DHCPStarvation, DHCPSpoofing,
	}

	eventType := eventTypes[id%len(eventTypes)]
	severity := "LOW"

	switch eventType {
	case DHCPRogueServer, DHCPSpoofing:
		severity = "HIGH"
	case DHCPStarvation:
		severity = "MEDIUM"
	}

	return DHCPSecurityEvent{
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

// DHCPSecurityPlatform orchestrates the entire platform
type DHCPSecurityPlatform struct {
	config        KafkaConfig
	eventProducer *DHCPEventProducer
	alertProducer *DHCPEventProducer
	processor     *DHCPEventProcessor
	consumer      *DHCPEventConsumer
	simulator     *NetworkMonitoringSimulator
}

// NewDHCPSecurityPlatform creates a new DHCP security platform
func NewDHCPSecurityPlatform(config KafkaConfig) (*DHCPSecurityPlatform, error) {
	// Create event producer
	eventProducer, err := NewDHCPEventProducer(config.Brokers, config.DHCPEventsTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create event producer: %w", err)
	}

	// Create alert producer
	alertProducer, err := NewDHCPEventProducer(config.Brokers, config.AlertsTopic)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert producer: %w", err)
	}

	// Create processor
	processor := NewDHCPEventProcessor(alertProducer)

	// Create consumer
	consumer, err := NewDHCPEventConsumer(
		config.Brokers,
		config.ConsumerGroup,
		[]string{config.DHCPEventsTopic},
		processor,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Create simulator
	simulator := NewNetworkMonitoringSimulator(eventProducer)

	return &DHCPSecurityPlatform{
		config:        config,
		eventProducer: eventProducer,
		alertProducer: alertProducer,
		processor:     processor,
		consumer:      consumer,
		simulator:     simulator,
	}, nil
}

// Start starts the platform
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

// Stop stops the platform
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

func main() {
	// Kafka configuration
	config := KafkaConfig{
		Brokers:         []string{"localhost:9092"},
		DHCPEventsTopic: "dhcp-security-events",
		AlertsTopic:     "dhcp-security-alerts",
		ConsumerGroup:   "dhcp-security-processors",
	}

	// Create platform
	platform, err := NewDHCPSecurityPlatform(config)
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
		platform.Stop()
	}()

	// Start platform
	log.Println("Starting DHCP Security Event Streaming Platform...")
	if err := platform.Start(ctx); err != nil {
		log.Fatalf("Platform error: %v", err)
	}

	log.Println("Platform shutdown complete")
}
