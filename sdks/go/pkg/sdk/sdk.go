package sdk

import cloudevents "github.com/cloudevents/sdk-go/v2"

// Core is the interface that abstracts FFI and WASM implementations
// of the EDA core functionality
type Core interface {
	// GetKafkaConfig retrieves the Kafka connection configuration
	GetKafkaConfig() (*KafkaConfig, error)

	// ShouldRetry checks if an error should be retried
	ShouldRetry(error string, attempt uint32) (bool, error)

	// CalculateBackoff calculates backoff duration in milliseconds
	CalculateBackoff(attempt uint32) (uint64, error)

	// GetOutputDestination routes an output event to its destination
	GetOutputDestination(eventJSON string) (*OutputDestination, error)

	// LoadRoutingConfig loads routing configuration from a YAML file
	LoadRoutingConfig(filePath string) error

	// Close releases resources held by the Core implementation
	Close() error
}

// OutputDestination specifies where to send an output event
type OutputDestination struct {
	Type    DestinationType
	Target  string
	Cluster *string
}

// DestinationType represents the type of destination
type DestinationType int

const (
	DestinationKafka DestinationType = iota
	DestinationRabbitMQ
	DestinationHTTP
	DestinationDiscard
)

// Handler function signatures

// SimpleHandler processes an event without returning output events
type SimpleHandler func(cloudevents.Event) error

// OutputHandler processes an event and returns a single output event
type OutputHandler func(cloudevents.Event) (*cloudevents.Event, error)

// Handler is a constraint for valid handler function types
type Handler interface {
	SimpleHandler | OutputHandler
}
