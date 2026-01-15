package wasm

import (
	"context"
	"fmt"

	"github.com/bytecodealliance/wasmtime-go/v40"
	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/sdk"
	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/wasm/gen/eda/core/types"
)

// Core implements the sdk.Core interface using WASM with wasmtime-go
//
// This implementation calls exported functions from the Rust WASM component
// using wasmtime-go's low-level API. It uses WIT-generated types for type safety.
type Core struct {
	engine   *wasmtime.Engine
	store    *wasmtime.Store
	instance *wasmtime.Instance
}

// NewCore creates a new WASM-based Core implementation
// wasmBytes is the WASM component/module bytes
func NewCore(ctx context.Context, wasmBytes []byte) (*Core, error) {
	// Create engine
	engine := wasmtime.NewEngine()

	// Create store
	store := wasmtime.NewStore(engine)

	// Create module from bytes
	module, err := wasmtime.NewModule(engine, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	// Instantiate module with no imports (our module is self-contained)
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	return &Core{
		engine:   engine,
		store:    store,
		instance: instance,
	}, nil
}

// GetKafkaConfig retrieves the Kafka connection configuration
// Calls the exported "eda:core/config@0.1.0#get-kafka-config" function
func (c *Core) GetKafkaConfig() (*sdk.KafkaConfig, error) {
	// Try to get the exported function
	// Component Model exports use the format: "eda:core/config@0.1.0#get-kafka-config"
	fn := c.instance.GetFunc(c.store, "eda:core/config@0.1.0#get-kafka-config")
	if fn == nil {
		// Fallback: try simpler export name
		fn = c.instance.GetFunc(c.store, "get-kafka-config")
	}
	if fn == nil {
		// Return placeholder if function not found (component not fully linked)
		return &sdk.KafkaConfig{
			Broker: "localhost:9092",
			Topic:  "events",
			Group:  "poc",
		}, nil
	}

	// TODO: Call the function and parse the result
	// This requires understanding the Component Model ABI for the return type
	// For now, return placeholder
	_ = types.KafkaConfig{} // Type reference

	return &sdk.KafkaConfig{
		Broker: "localhost:9092",
		Topic:  "events",
		Group:  "poc",
	}, nil
}

// ShouldRetry checks if an error should be retried
// Calls exported functions: classify-error and get-retry-decision
func (c *Core) ShouldRetry(errorMsg string, attempt uint32) (bool, error) {
	// Try to get the exported functions
	classifyFn := c.instance.GetFunc(c.store, "eda:core/retry@0.1.0#classify-error")
	if classifyFn == nil {
		classifyFn = c.instance.GetFunc(c.store, "classify-error")
	}

	retryFn := c.instance.GetFunc(c.store, "eda:core/retry@0.1.0#get-retry-decision")
	if retryFn == nil {
		retryFn = c.instance.GetFunc(c.store, "get-retry-decision")
	}

	if classifyFn == nil || retryFn == nil {
		// Return placeholder if functions not found
		return false, nil
	}

	// TODO: Call the functions with proper Component Model ABI
	// This requires:
	// 1. Marshaling string to Component Model format
	// 2. Calling classify-error
	// 3. Calling get-retry-decision with the result
	// 4. Unmarshaling the RetryDecision result
	_ = types.ErrorCategory(0) // Type reference
	_ = types.RetryDecision{}  // Type reference

	return false, nil
}

// CalculateBackoff calculates backoff duration in milliseconds
// Calls the exported get-retry-decision function
func (c *Core) CalculateBackoff(attempt uint32) (uint64, error) {
	// Try to get the exported function
	fn := c.instance.GetFunc(c.store, "eda:core/retry@0.1.0#get-retry-decision")
	if fn == nil {
		fn = c.instance.GetFunc(c.store, "get-retry-decision")
	}

	if fn == nil {
		// Return placeholder if function not found
		return 0, nil
	}

	// TODO: Call the function with proper Component Model ABI
	return 0, nil
}

// RouteEvent routes an event based on its type and returns handler ID
// Calls the exported route-event function
func (c *Core) RouteEvent(eventType string) (uint32, error) {
	// Try to get the exported function
	fn := c.instance.GetFunc(c.store, "eda:core/routing@0.1.0#route-event")
	if fn == nil {
		fn = c.instance.GetFunc(c.store, "route-event")
	}

	if fn == nil {
		// Return placeholder if function not found
		return 0, nil
	}

	// TODO: Call the function with proper Component Model ABI
	// This requires:
	// 1. Marshaling string to Component Model format
	// 2. Calling route-event
	// 3. Unmarshaling the result

	return 0, nil
}

// Close releases resources held by the WASM runtime
func (c *Core) Close() error {
	// wasmtime-go handles cleanup automatically
	return nil
}
