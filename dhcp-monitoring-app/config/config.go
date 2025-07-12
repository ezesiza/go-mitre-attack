package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type AppConfig struct {
	Kafka KafkaConfig
	App   AppSettings
}

type KafkaConfig struct {
	Brokers         []string
	DHCPEventsTopic string
	AlertsTopic     string
	ConsumerGroup   string
	ProducerTimeout time.Duration
	ConsumerTimeout time.Duration
}

// AppSettings holds general application settings
type AppSettings struct {
	LogLevel    string
	Environment string
	Simulation  SimulationConfig
	Processing  ProcessingConfig
}

// SimulationConfig holds simulation-specific settings
type SimulationConfig struct {
	Enabled    bool
	Interval   time.Duration
	EventTypes []string
	SourceIPs  []string
}

// ProcessingConfig holds event processing settings
type ProcessingConfig struct {
	CacheTimeout    time.Duration
	CleanupInterval time.Duration
	MaxCacheSize    int
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		Kafka: KafkaConfig{
			Brokers:         []string{"localhost:9092"},
			DHCPEventsTopic: "dhcp-security-events",
			AlertsTopic:     "dhcp-security-alerts",
			ConsumerGroup:   "dhcp-security-processors",
			ProducerTimeout: 30 * time.Second,
			ConsumerTimeout: 10 * time.Second,
		},
		App: AppSettings{
			LogLevel:    "info",
			Environment: "development",
			Simulation: SimulationConfig{
				Enabled:    true,
				Interval:   2 * time.Second,
				EventTypes: []string{"DHCP_DISCOVER", "DHCP_OFFER", "DHCP_REQUEST", "DHCP_ACK", "DHCP_ROGUE_SERVER", "DHCP_STARVATION", "DHCP_SPOOFING"},
				SourceIPs:  []string{"192.168.1.100", "192.168.1.101", "192.168.1.102"},
			},
			Processing: ProcessingConfig{
				CacheTimeout:    10 * time.Minute,
				CleanupInterval: 1 * time.Minute,
				MaxCacheSize:    10000,
			},
		},
	}
}

// LoadFromEnvironment loads configuration from environment variables
func (c *AppConfig) LoadFromEnvironment() error {
	// Kafka configuration
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		c.Kafka.Brokers = strings.Split(brokers, ",")
	}
	if topic := os.Getenv("KAFKA_DHCP_EVENTS_TOPIC"); topic != "" {
		c.Kafka.DHCPEventsTopic = topic
	}
	if topic := os.Getenv("KAFKA_ALERTS_TOPIC"); topic != "" {
		c.Kafka.AlertsTopic = topic
	}
	if group := os.Getenv("KAFKA_CONSUMER_GROUP"); group != "" {
		c.Kafka.ConsumerGroup = group
	}

	// App settings
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.App.LogLevel = level
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		c.App.Environment = env
	}

	// Simulation settings
	if enabled := os.Getenv("SIMULATION_ENABLED"); enabled != "" {
		if enabled == "true" {
			c.App.Simulation.Enabled = true
		} else {
			c.App.Simulation.Enabled = false
		}
	}
	if interval := os.Getenv("SIMULATION_INTERVAL"); interval != "" {
		if duration, err := time.ParseDuration(interval); err == nil {
			c.App.Simulation.Interval = duration
		}
	}

	// Processing settings
	if timeout := os.Getenv("CACHE_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			c.App.Processing.CacheTimeout = duration
		}
	}
	if interval := os.Getenv("CLEANUP_INTERVAL"); interval != "" {
		if duration, err := time.ParseDuration(interval); err == nil {
			c.App.Processing.CleanupInterval = duration
		}
	}
	if size := os.Getenv("MAX_CACHE_SIZE"); size != "" {
		if maxSize, err := strconv.Atoi(size); err == nil {
			c.App.Processing.MaxCacheSize = maxSize
		}
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *AppConfig) Validate() error {
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker must be specified")
	}
	if c.Kafka.DHCPEventsTopic == "" {
		return fmt.Errorf("DHCP events topic must be specified")
	}
	if c.Kafka.AlertsTopic == "" {
		return fmt.Errorf("alerts topic must be specified")
	}
	if c.Kafka.ConsumerGroup == "" {
		return fmt.Errorf("consumer group must be specified")
	}
	return nil
}

// IsDevelopment returns true if the app is running in development mode
func (c *AppConfig) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if the app is running in production mode
func (c *AppConfig) IsProduction() bool {
	return c.App.Environment == "production"
}

// GetKafkaConfig returns the Kafka configuration
func (c *AppConfig) GetKafkaConfig() KafkaConfig {
	return c.Kafka
}

// GetAppSettings returns the application settings
func (c *AppConfig) GetAppSettings() AppSettings {
	return c.App
}
