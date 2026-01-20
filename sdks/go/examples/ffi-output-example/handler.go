package main

import (
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/sdk"
)

// Handle is an event handler that produces output events
// This demonstrates the output event routing capability
var Handle sdk.OutputHandler = func(event cloudevents.Event) (*cloudevents.Event, error) {
	fmt.Printf("📨 Received event: %s\n", event.ID())
	fmt.Printf("   Type: %s\n", event.Type())
	fmt.Printf("   Source: %s\n", event.Source())

	// Only process events of type "kafka.message"
	// Produce a transformed event for other types
	if event.Type() != "kafka.message" {
		return nil, nil
	}

	// Create an output event
	outputEvent := cloudevents.NewEvent()
	outputEvent.SetID(fmt.Sprintf("processed-%s", event.ID()))
	outputEvent.SetSource("ffi-output-example")
	outputEvent.SetType("com.example.processed")
	outputEvent.SetTime(time.Now())

	// Copy data from input event
	if err := outputEvent.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"original_id":   event.ID(),
		"original_type": event.Type(),
		"processed_at":  time.Now().Format(time.RFC3339),
		"message":       "Event processed successfully",
	}); err != nil {
		return nil, fmt.Errorf("failed to set output event data: %w", err)
	}

	fmt.Printf("✅ Producing output event: %s (type: %s)\n", outputEvent.ID(), outputEvent.Type())

	return &outputEvent, nil
}
