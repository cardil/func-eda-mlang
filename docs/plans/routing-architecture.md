# Output Event Routing Architecture

This document describes the routing architecture for EDA Functions SDK. Routing determines where events produced by handlers are published.

---

## Overview

EDA functions follow a simple input-process-output pattern:

1. **Input**: Events arrive from a configured source (e.g., Kafka topic)
2. **Processing**: User's handler function processes the event
3. **Output**: Handler may produce zero or more response events

Routing is for **output events** - determining where events produced by handlers should be published based on configurable rules and filters.

### Default Behavior

When no routing rules are defined, all output events are routed to the **default cluster** and **default topic** as configured in the function's Kafka configuration.

---

## Handler Function Signatures

The SDK supports multiple handler function signatures to accommodate different use cases:

### Go SDK Handler Types

```go
// Type 1: Simple handler - no output events
type SimpleHandler func(event cloudevents.Event) error

// Type 2: Single output - returns one event
type SingleOutputHandler func(event cloudevents.Event) (cloudevents.Event, error)

// Type 3: Multi output - returns multiple events  
type MultiOutputHandler func(event cloudevents.Event) ([]cloudevents.Event, error)

// Type 4: Context-based - uses context for publishing
type ContextHandler func(ctx sdk.Context, event cloudevents.Event) error
```

### Handler Registration

```go
// Register with automatic signature detection
sdk.RegisterHandler(myHandler)

// Or explicit registration
sdk.RegisterSimpleHandler(func(e cloudevents.Event) error { ... })
sdk.RegisterOutputHandler(func(e cloudevents.Event) (cloudevents.Event, error) { ... })
```

---

## Filtering

Routing rules use filter expressions to match events. The SDK implements the **CloudEvents Subscriptions API filter dialects**.

For complete filter dialect specification, see:
- [CloudEvents Subscriptions API - Filters](https://github.com/cloudevents/spec/blob/main/subscriptions/spec.md#324-filters)
- [CloudEvents SQL Expression Language (CESQL)](https://github.com/cloudevents/spec/blob/main/cesql/spec.md)

### Supported Filter Dialects

| Dialect | Required | Description |
|---------|----------|-------------|
| `exact` | Yes | Exact match on CloudEvents attributes |
| `prefix` | Yes | Prefix match on CloudEvents attributes |
| `suffix` | Yes | Suffix match on CloudEvents attributes |
| `all` | Yes | All nested filters must match |
| `any` | Yes | At least one nested filter must match |
| `not` | Yes | Negates nested filter |
| `sql` | Optional | CESQL expression for complex filtering |

---

## WIT Interface

```wit
package eda:core@0.2.0;

interface types {
    record kafka-config {
        broker: string,
        topic: string,
        group-id: string,
    }
    
    enum error-category {
        transient,
        permanent,
        unknown,
    }
    
    record retry-decision {
        should-retry: bool,
        backoff-ms: u64,
        send-to-dlq: bool,
    }
    
    enum destination-type {
        kafka,
        rabbitmq,
        http,
        discard,
    }
    
    record output-destination {
        dest-type: destination-type,
        target: string,
        cluster: option<string>,
    }
    
    /// Filter expression following CloudEvents Subscriptions API format
    /// Serialized as JSON to support nested filter structures
    /// See: https://github.com/cloudevents/spec/blob/main/subscriptions/spec.md#324-filters
    type filter-expression = string;
    
    record routing-rule {
        name: string,
        filter: filter-expression,
        destination: output-destination,
    }
}

interface config {
    use types.{kafka-config};
    get-kafka-config: func() -> kafka-config;
}

interface retry {
    use types.{error-category, retry-decision};
    classify-error: func(error-message: string) -> error-category;
    get-retry-decision: func(error-category: error-category, attempt: u32, max-attempts: u32) -> retry-decision;
}

interface routing {
    use types.{output-destination, routing-rule};
    
    /// Evaluate routing rules against an event and return the destination
    /// The event-json parameter contains the serialized CloudEvent
    get-output-destination: func(event-json: string) -> output-destination;
    
    /// Add a routing rule
    add-routing-rule: func(rule: routing-rule);
    
    /// Clear all routing rules
    clear-routing-rules: func();
    
    /// Get the default destination used when no rule matches
    get-default-destination: func() -> output-destination;
    
    /// Set the default destination for unmatched events
    set-default-destination: func(dest: output-destination);
}

interface telemetry {
    record-event-received: func(event-type: string);
    record-event-processed: func(event-type: string, success: bool, duration-ms: u64);
    record-retry-attempt: func(attempt: u32, backoff-ms: u64);
    get-event-count: func() -> u64;
}

world eda-core {
    export config;
    export retry;
    export routing;
    export telemetry;
}
```

**Note:** The `filter-expression` is serialized as JSON to support the full CloudEvents Subscriptions API filter structure, including nested `all`, `any`, and `not` compositions.

---

## Configuration

Routing rules are configured via YAML:

```yaml
routing:
  # Default destination when no rule matches
  # If not specified, uses the function's default Kafka cluster and topic
  default:
    type: kafka
    target: unrouted-events
    cluster: default
  
  rules:
    # Exact match on type
    - name: order-events
      filter:
        exact:
          type: "com.example.order.created"
      destination:
        type: kafka
        target: order-notifications
        cluster: orders
    
    # Prefix match
    - name: all-orders
      filter:
        prefix:
          type: "com.example.order."
      destination:
        type: kafka
        target: order-events
        cluster: orders
    
    # CESQL for complex filtering
    - name: high-priority
      filter:
        sql: "type LIKE 'com.example.%' AND EXISTS priority"
      destination:
        type: http
        target: https://webhooks.example.com/urgent
    
    # Combined filters using all
    - name: analytics
      filter:
        all:
          - prefix:
              source: "/production/"
          - not:
              exact:
                type: "com.example.debug"
      destination:
        type: kafka
        target: analytics-stream
    
    # Discard internal events
    - name: drop-internal
      filter:
        suffix:
          type: ".internal"
      destination:
        type: discard
```

---

## Go SDK Context Interface

```go
package sdk

import cloudevents "github.com/cloudevents/sdk-go/v2"

// Context provides event handling capabilities to user handlers
type Context interface {
    // Publish sends an event to its routed destination
    Publish(event cloudevents.Event) error
    
    // PublishTo sends an event to a specific destination, bypassing routing
    PublishTo(event cloudevents.Event, destination OutputDestination) error
    
    // Logger returns the configured logger
    Logger() Logger
}

// OutputDestination specifies where to send an event
type OutputDestination struct {
    Type    DestinationType
    Target  string
    Cluster string
}

type DestinationType int

const (
    DestinationKafka DestinationType = iota
    DestinationRabbitMQ
    DestinationHTTP
    DestinationDiscard
)
```

---

## Summary

| Concept | Description |
|---------|-------------|
| **Routing purpose** | Route output events from handlers to destinations |
| **Handler signatures** | Simple, single-output, multi-output, context-based |
| **Filter dialects** | CloudEvents Subscriptions API compliant (exact, prefix, suffix, all, any, not, sql) |
| **Destinations** | Kafka topics, RabbitMQ queues, HTTP endpoints, discard |
| **Default behavior** | No rules defined → events go to default cluster/topic |
| **Configuration** | YAML file with optional default destination and routing rules |
