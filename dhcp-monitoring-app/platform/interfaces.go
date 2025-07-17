package platform

import (
	"context"
	"dhcp-monitoring-app/config"
	"dhcp-monitoring-app/models"
)

// EventProducer defines the interface for publishing DHCP events
type EventProducer interface {
	PublishEvent(event models.DHCPSecurityEvent) error
	Close() error
}

// EventConsumer defines the interface for consuming DHCP events
type EventConsumer interface {
	Start(ctx context.Context) error
	Close() error
}

// EventProcessor defines the interface for processing DHCP events
type EventProcessor interface {
	ProcessEvent(event models.DHCPSecurityEvent) error
}

// EventSimulator defines the interface for simulating DHCP events
type EventSimulator interface {
	SimulateEvents(ctx context.Context)
}

// ConfigProvider defines the interface for providing configuration
type ConfigProvider interface {
	GetKafkaConfig() config.KafkaConfig
	GetAppSettings() config.AppSettings
	IsDevelopment() bool
	IsProduction() bool
}

// Platform defines the main platform interface
type Platform interface {
	Start(ctx context.Context) error
	Stop() error
}

// ComponentFactory defines the interface for creating platform components
type ComponentFactory interface {
	CreateEventProducer(brokers []string, topic string) (EventProducer, error)
	CreateEventConsumer(brokers []string, groupID string, topics []string, processor EventProcessor) (EventConsumer, error)
	CreateEventProcessor(alertProducer EventProducer) EventProcessor
	CreateEventSimulator(producer EventProducer, config interface{}) EventSimulator
}

// ServiceContainer defines the interface for managing service dependencies
type ServiceContainer interface {
	GetEventProducer() EventProducer
	GetAlertProducer() EventProducer
	GetEventConsumer() EventConsumer
	GetEventProcessor() EventProcessor
	GetEventSimulator() EventSimulator
	GetConfig() ConfigProvider
	Close() error
}
