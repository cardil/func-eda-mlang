package sdk

// KafkaConfig holds the Kafka connection configuration
type KafkaConfig struct {
	Broker string
	Topic  string
	Group  string
}
