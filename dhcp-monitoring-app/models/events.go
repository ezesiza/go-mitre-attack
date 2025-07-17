package models

import "time"

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

// SecurityRule defines rules for generating security alerts
type SecurityRule struct {
	Name        string
	EventTypes  []DHCPEventType
	Threshold   int
	TimeWindow  time.Duration
	Severity    string
	Description string
}
