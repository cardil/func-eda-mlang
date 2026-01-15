package ffi

import (
	"fmt"

	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/sdk"
)

// Core implements the sdk.Core interface using FFI (purego)
type Core struct{}

// NewCore creates a new FFI-based Core implementation
func NewCore() (*Core, error) {
	// Load the embedded library on first use
	if err := loadLibrary(); err != nil {
		return nil, fmt.Errorf("failed to load FFI library: %w", err)
	}
	return &Core{}, nil
}

// GetKafkaConfig retrieves the Kafka connection configuration
func (c *Core) GetKafkaConfig() (*sdk.KafkaConfig, error) {
	// Get broker
	brokerPtr := edaGetKafkaBroker()
	if brokerPtr == nil {
		return nil, fmt.Errorf("failed to get Kafka broker")
	}
	broker := goString(brokerPtr)
	edaFreeString(brokerPtr)

	// Get topic
	topicPtr := edaGetKafkaTopic()
	if topicPtr == nil {
		return nil, fmt.Errorf("failed to get Kafka topic")
	}
	topic := goString(topicPtr)
	edaFreeString(topicPtr)

	// Get group
	groupPtr := edaGetKafkaGroup()
	if groupPtr == nil {
		return nil, fmt.Errorf("failed to get Kafka group")
	}
	group := goString(groupPtr)
	edaFreeString(groupPtr)

	return &sdk.KafkaConfig{
		Broker: broker,
		Topic:  topic,
		Group:  group,
	}, nil
}

// ShouldRetry checks if an error should be retried
func (c *Core) ShouldRetry(error string, attempt uint32) (bool, error) {
	cError := cString(error)
	result := edaShouldRetry(cError, attempt)
	return result != 0, nil
}

// CalculateBackoff calculates backoff duration in milliseconds
func (c *Core) CalculateBackoff(attempt uint32) (uint64, error) {
	result := edaCalculateBackoff(attempt)
	return result, nil
}

// RouteEvent routes an event based on its type and returns handler ID
func (c *Core) RouteEvent(eventType string) (uint32, error) {
	cEventType := cString(eventType)
	handlerID := edaRouteEvent(cEventType)
	return handlerID, nil
}

// Close releases resources (no-op for FFI implementation)
func (c *Core) Close() error {
	return nil
}
