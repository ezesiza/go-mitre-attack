package platform

import (
	"context"
	"dhcp-monitoring-app/config"
	"dhcp-monitoring-app/models"
)

// EventConsumer defines the interface for consuming DHCP events
type EventConsumer interface {
	Start(ctx context.Context) error
	Close() error
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
	CreateEventProducer(brokers []string, topic string) (models.EventProducer, error)
	CreateEventConsumer(brokers []string, groupID string, topics []string, processor models.EventProcessor) (EventConsumer, error)
	CreateEventProcessor(alertProducer models.EventProducer, websocketServer models.WebSocketServer) models.EventProcessor
	CreateEventSimulator(producer models.EventProducer, config interface{}) EventSimulator
}

// ServiceContainer defines the interface for managing service dependencies
type ServiceContainer interface {
	GetEventProducer() models.EventProducer
	GetAlertProducer() models.EventProducer
	GetEventConsumer() EventConsumer
	GetEventProcessor() models.EventProcessor
	GetEventSimulator() EventSimulator
	GetConfig() ConfigProvider
	GetWebSocketServer() models.WebSocketServer
	Close() error
}
