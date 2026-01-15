package sdk

// Core is the interface that abstracts FFI and WASM implementations
// of the EDA core functionality
type Core interface {
	// GetKafkaConfig retrieves the Kafka connection configuration
	GetKafkaConfig() (*KafkaConfig, error)

	// ShouldRetry checks if an error should be retried
	ShouldRetry(error string, attempt uint32) (bool, error)

	// CalculateBackoff calculates backoff duration in milliseconds
	CalculateBackoff(attempt uint32) (uint64, error)

	// RouteEvent routes an event based on its type and returns handler ID
	RouteEvent(eventType string) (uint32, error)

	// Close releases resources held by the Core implementation
	Close() error
}
