package sdk

import (
	"context"
	"fmt"
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// HandlerFunc is the user's event handler function
type HandlerFunc func(cloudevents.Event) error

// Consumer manages Kafka consumption and event processing
type Consumer struct {
	core     Core
	consumer *kafka.Consumer
	handler  HandlerFunc
}

// NewConsumer creates a new consumer with the given core implementation
func NewConsumer(core Core, handler HandlerFunc) (*Consumer, error) {
	if core == nil {
		return nil, fmt.Errorf("core cannot be nil")
	}
	if handler == nil {
		return nil, fmt.Errorf("handler cannot be nil")
	}

	// Get Kafka config from core
	config, err := core.GetKafkaConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kafka config: %w", err)
	}

	log.Printf("Kafka config - Broker: %s, Topic: %s, Group: %s", config.Broker, config.Topic, config.Group)

	// Create Kafka consumer
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.Broker,
		"group.id":          config.Group,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	// Subscribe to topic with rebalance callback that starts from beginning
	rebalanceCb := func(c *kafka.Consumer, event kafka.Event) error {
		switch e := event.(type) {
		case kafka.AssignedPartitions:
			log.Printf("Partitions assigned: %v", e.Partitions)
			// Set offset to beginning before assigning to replay all messages
			for i := range e.Partitions {
				e.Partitions[i].Offset = kafka.OffsetBeginning
			}
			log.Printf("Starting from beginning for all partitions")
			return c.Assign(e.Partitions)
		case kafka.RevokedPartitions:
			log.Printf("Partitions revoked: %v", e.Partitions)
			return c.Unassign()
		}
		return nil
	}

	if err := kafkaConsumer.Subscribe(config.Topic, rebalanceCb); err != nil {
		kafkaConsumer.Close()
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", config.Topic, err)
	}

	return &Consumer{
		core:     core,
		consumer: kafkaConsumer,
		handler:  handler,
	}, nil
}

// Start begins consuming events (blocking)
func (c *Consumer) Start(ctx context.Context) error {
	log.Printf("Starting consumer...")

	consecutiveErrors := 0
	maxConsecutiveErrors := 5
	pollTimeout := 100 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			log.Printf("Consumer stopping...")
			return ctx.Err()
		default:
			// Poll for messages with timeout to allow context cancellation
			msg, err := c.consumer.ReadMessage(pollTimeout)
			if err != nil {
				// Timeout is expected when no messages, not an error
				if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
					continue
				}
				log.Printf("Error reading message: %v", err)
				consecutiveErrors++
				if consecutiveErrors >= maxConsecutiveErrors {
					return fmt.Errorf("too many consecutive errors (%d), giving up: %w", maxConsecutiveErrors, err)
				}
				continue
			}

			// Reset error counter on successful read
			consecutiveErrors = 0

			// Parse CloudEvent
			event, err := c.parseCloudEvent(msg)
			if err != nil {
				log.Printf("Error parsing CloudEvent: %v", err)
				continue
			}

			// Route event using core
			handlerID, err := c.core.RouteEvent(event.Type())
			if err != nil {
				log.Printf("Error routing event: %v", err)
				continue
			}
			log.Printf("Event routed to handler ID: %d", handlerID)

			// Call user handler
			if err := c.handler(*event); err != nil {
				log.Printf("Handler error: %v", err)

				// Check if we should retry using core
				shouldRetry, retryErr := c.core.ShouldRetry(err.Error(), 1)
				if retryErr != nil {
					log.Printf("Error checking retry: %v", retryErr)
					continue
				}

				if shouldRetry {
					backoff, backoffErr := c.core.CalculateBackoff(1)
					if backoffErr != nil {
						log.Printf("Error calculating backoff: %v", backoffErr)
						continue
					}
					log.Printf("Would retry after %dms (not implemented in PoC)", backoff)
				}
			}
		}
	}
}

// parseCloudEvent converts Kafka message to CloudEvent
func (c *Consumer) parseCloudEvent(msg *kafka.Message) (*cloudevents.Event, error) {
	event := cloudevents.NewEvent()

	// Try to parse as structured CloudEvent (JSON)
	if err := event.UnmarshalJSON(msg.Value); err == nil {
		return &event, nil
	}

	// Fallback: create simple CloudEvent from message
	event.SetID(string(msg.Key))
	event.SetSource("kafka")
	event.SetType("kafka.message")
	if err := event.SetData(cloudevents.ApplicationJSON, msg.Value); err != nil {
		return nil, fmt.Errorf("failed to set event data: %w", err)
	}

	return &event, nil
}

// Close releases resources
func (c *Consumer) Close() error {
	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			return fmt.Errorf("failed to close Kafka consumer: %w", err)
		}
	}
	if c.core != nil {
		if err := c.core.Close(); err != nil {
			return fmt.Errorf("failed to close core: %w", err)
		}
	}
	return nil
}
