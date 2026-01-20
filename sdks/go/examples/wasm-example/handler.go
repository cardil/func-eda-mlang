package main

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/sdk"
)

// Handle is the user's event handler function
// This is what a developer would write for their EDA function
var Handle sdk.SimpleHandler = func(event cloudevents.Event) error {
	fmt.Printf("📨 Received event: %s\n", event.ID())
	fmt.Printf("   Type: %s\n", event.Type())
	fmt.Printf("   Source: %s\n", event.Source())

	// User's business logic would go here
	// For this example, we just log the event

	return nil
}
