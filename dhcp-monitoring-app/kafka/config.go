package kafka

type KafkaConfig struct {
	Brokers         []string
	DHCPEventsTopic string
	AlertsTopic     string
	ConsumerGroup   string
}
