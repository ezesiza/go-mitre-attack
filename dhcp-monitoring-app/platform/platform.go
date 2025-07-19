package platform

import (
	"context"
	"dhcp-monitoring-app/config"
	"dhcp-monitoring-app/interfaces"
	"dhcp-monitoring-app/kafka"
	"dhcp-monitoring-app/models"
	"dhcp-monitoring-app/simulator"
	"dhcp-monitoring-app/websockets"
	"fmt"
	"log"
	"sync"
)

// DefaultComponentFactory implements ComponentFactory with default implementations
type DefaultComponentFactory struct{}

// DHCPSecurityPlatform implements the Platform interface with dependency injection
type DHCPSecurityPlatform struct {
	container ServiceContainer
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// DefaultServiceContainer implements ServiceContainer with dependency injection
type DefaultServiceContainer struct {
	config          ConfigProvider
	factory         ComponentFactory
	eventProducer   EventProducer
	alertProducer   EventProducer
	eventConsumer   EventConsumer
	eventProcessor  EventProcessor
	eventSimulator  EventSimulator
	websocketServer interfaces.WebSocketServer
}

// NewDHCPSecurityPlatform creates a new platform instance with dependency injection
// func NewDHCPSecurityPlatform(cfg *config.AppConfig) (*DHCPSecurityPlatform, error) {
func NewDHCPSecurityPlatform(container ServiceContainer) *DHCPSecurityPlatform {

	ctx, cancel := context.WithCancel(context.Background())
	return &DHCPSecurityPlatform{
		container: container,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start initializes and starts all platform components
func (p *DHCPSecurityPlatform) Start(ctx context.Context) error {
	log.Println("Starting DHCP Security Platform")

	// Start websocket server if present
	if ws := p.container.GetWebSocketServer(); ws != nil {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			if err := ws.Start(":8080", "/ws"); err != nil {
				log.Printf("Websocket server error: %v", err)
			}
		}()
	}

	// Start consumer
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		// if err := p.consumer.Start(ctx); err != nil {
		if err := p.container.GetEventConsumer().Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	if p.container.GetEventSimulator() != nil {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.container.GetEventSimulator().SimulateEvents(p.ctx)
		}()
	}

	log.Println("DHCP Security Platform started successfully")
	<-ctx.Done()
	return ctx.Err()
}

// Stop gracefully shuts down all platform components
func (p *DHCPSecurityPlatform) Stop() error {

	log.Println("Stopping DHCP Security Platform...")

	// Stop websocket server if present
	if ws := p.container.GetWebSocketServer(); ws != nil {
		ws.Stop()
	}

	// Cancel context to stop all goroutines
	p.cancel()

	// Wait for all goroutines to finish
	p.wg.Wait()

	// Close all components
	if err := p.container.Close(); err != nil {
		log.Printf("Error closing service container: %v", err)
	}

	log.Println("DHCP Security Platform stopped successfully")
	return nil
}

// CreateEventProducer creates a new event producer
func (f *DefaultComponentFactory) CreateEventProducer(brokers []string, topic string) (EventProducer, error) {
	return kafka.NewDHCPEventProducer(brokers, topic)
}

// CreateEventConsumer creates a new event consumer
func (f *DefaultComponentFactory) CreateEventConsumer(brokers []string, groupID string, topics []string, processor EventProcessor) (EventConsumer, error) {

	return kafka.NewDHCPEventConsumer(brokers, groupID, topics, processor)
}

// CreateEventProcessor creates a new event processor
func (f *DefaultComponentFactory) CreateEventProcessor(alertProducer EventProducer, websocketServer interfaces.WebSocketServer) EventProcessor {
	return models.NewDHCPEventProcessor(alertProducer, websocketServer)
}

// CreateEventSimulator creates a new event simulator
func (f *DefaultComponentFactory) CreateEventSimulator(producer EventProducer, simCfg interface{}) EventSimulator {
	// Type assertion to get the concrete producer
	// kafkaProducer, ok := producer.(*kafka.DHCPEventProducer)

	// Type assertion to get the simulation config
	cfg, ok := simCfg.(*config.SimulationConfig)
	if !ok {
		return simulator.NewNetworkMonitoringSimulator(producer)
	}

	return simulator.NewNetworkMonitoringSimulatorWithConfig(producer, cfg)
}

// NewDefaultServiceContainer creates a new service container with all dependencies
func NewDefaultServiceContainer(cfg ConfigProvider) (*DefaultServiceContainer, error) {
	container := &DefaultServiceContainer{
		config:  cfg,
		factory: &DefaultComponentFactory{},
	}

	if err := container.initialize(); err != nil {
		return nil, err
	}

	return container, nil
}

// initialize sets up all dependencies in the correct order
func (c *DefaultServiceContainer) initialize() error {
	kafkaConfig := c.config.GetKafkaConfig()

	// Create websocket server first
	c.websocketServer = websockets.NewWebSocketServer()

	// Create event producer
	eventProducer, err := c.factory.CreateEventProducer(kafkaConfig.Brokers, kafkaConfig.DHCPEventsTopic)
	if err != nil {
		return fmt.Errorf("failed to create event producer: %w", err)
	}
	c.eventProducer = eventProducer

	// Create alert producer
	alertProducer, err := c.factory.CreateEventProducer(kafkaConfig.Brokers, kafkaConfig.AlertsTopic)
	if err != nil {
		return fmt.Errorf("failed to create alert producer: %w", err)
	}
	c.alertProducer = alertProducer

	// Create processor with correct websocketServer type
	c.eventProcessor = c.factory.CreateEventProcessor(c.alertProducer, c.websocketServer)

	// Create consumer
	consumer, err := c.factory.CreateEventConsumer(
		kafkaConfig.Brokers,
		kafkaConfig.ConsumerGroup,
		[]string{kafkaConfig.DHCPEventsTopic},
		c.eventProcessor,
	)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	c.eventConsumer = consumer

	// Create simulator if enabled
	appSettings := c.config.GetAppSettings()
	if appSettings.Simulation.Enabled {
		c.eventSimulator = c.factory.CreateEventSimulator(c.eventProducer, &appSettings.Simulation)
	}

	return nil
}

// GetEventProducer returns the event producer
func (c *DefaultServiceContainer) GetEventProducer() EventProducer {
	return c.eventProducer
}

// GetAlertProducer returns the alert producer
func (c *DefaultServiceContainer) GetAlertProducer() EventProducer {
	return c.alertProducer
}

// GetEventConsumer returns the event consumer
func (c *DefaultServiceContainer) GetEventConsumer() EventConsumer {
	return c.eventConsumer
}

// GetEventProcessor returns the event processor
func (c *DefaultServiceContainer) GetEventProcessor() EventProcessor {
	return c.eventProcessor
}

// GetEventSimulator returns the event simulator
func (c *DefaultServiceContainer) GetEventSimulator() EventSimulator {
	return c.eventSimulator
}

// GetWebSocketServer returns the websocket server
func (c *DefaultServiceContainer) GetWebSocketServer() interfaces.WebSocketServer {
	return c.websocketServer
}

// GetConfig returns the configuration provider
func (c *DefaultServiceContainer) GetConfig() ConfigProvider {
	return c.config
}

// Close closes all components in the container
func (c *DefaultServiceContainer) Close() error {
	var errors []error

	if c.eventConsumer != nil {
		if err := c.eventConsumer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("error closing consumer: %w", err))
		}
	}

	if c.eventProducer != nil {
		if err := c.eventProducer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("error closing event producer: %w", err))
		}
	}

	if c.alertProducer != nil {
		if err := c.alertProducer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("error closing alert producer: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	return nil
}

// PlatformBuilder provides a fluent interface for building platform instances
type PlatformBuilder struct {
	container ServiceContainer
}

// NewPlatformBuilder creates a new platform builder
func NewPlatformBuilder() *PlatformBuilder {
	return &PlatformBuilder{}
}

// WithServiceContainer sets the service container
func (b *PlatformBuilder) WithServiceContainer(container ServiceContainer) *PlatformBuilder {
	b.container = container
	return b
}

// Build creates a new platform instance
func (b *PlatformBuilder) Build() (*DHCPSecurityPlatform, error) {
	if b.container == nil {
		return nil, fmt.Errorf("service container is required")
	}

	return NewDHCPSecurityPlatform(b.container), nil
}
