package ffi

import (
	"fmt"
	"runtime"

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
	cError, buf := cStringWithBuf(error)
	result := edaShouldRetry(cError, attempt)
	runtime.KeepAlive(buf)
	return result != 0, nil
}

// CalculateBackoff calculates backoff duration in milliseconds
func (c *Core) CalculateBackoff(attempt uint32) (uint64, error) {
	result := edaCalculateBackoff(attempt)
	return result, nil
}

// GetOutputDestination routes an output event to its destination
func (c *Core) GetOutputDestination(eventJSON string) (*sdk.OutputDestination, error) {
	cEventJSON, buf := cStringWithBuf(eventJSON)
	cDest := edaGetOutputDestination(cEventJSON)
	runtime.KeepAlive(buf)
	if cDest == nil {
		return nil, fmt.Errorf("failed to get output destination")
	}
	defer edaFreeOutputDestination(cDest)

	// Convert C destination to Go destination
	dest := &sdk.OutputDestination{
		Type: sdk.DestinationType(cDest.DestType),
	}

	// Get target string
	if cDest.Target != nil {
		dest.Target = goString(cDest.Target)
	}

	// Get cluster string (optional)
	if cDest.Cluster != nil {
		cluster := goString(cDest.Cluster)
		dest.Cluster = &cluster
	}

	return dest, nil
}

// LoadRoutingConfig loads routing configuration from a YAML file
func (c *Core) LoadRoutingConfig(filePath string) error {
	cPath, buf := cStringWithBuf(filePath)
	success := edaLoadRoutingConfig(cPath)
	runtime.KeepAlive(buf)
	if !success {
		return fmt.Errorf("failed to load routing config from %s", filePath)
	}
	return nil
}

// Close releases resources (no-op for FFI implementation)
func (c *Core) Close() error {
	return nil
}
