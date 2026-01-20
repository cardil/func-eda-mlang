package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// handlerSignature represents the detected handler signature type
type handlerSignature int

const (
	handlerSignatureSimple handlerSignature = iota
	handlerSignatureOutput
)

// Consumer manages Kafka consumption and event processing
type Consumer struct {
	core         Core
	consumer     *kafka.Consumer
	producer     *kafka.Producer
	handlerValue reflect.Value
	handlerSig   handlerSignature
	logger       *slog.Logger
}

// NewConsumer creates a new consumer with the given core implementation
// Accepts either SimpleHandler or OutputHandler signatures
func NewConsumer(core Core, handler interface{}) (*Consumer, error) {
	if core == nil {
		return nil, fmt.Errorf("core cannot be nil")
	}
	if handler == nil {
		return nil, fmt.Errorf("handler cannot be nil")
	}

	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Detect handler signature
	handlerValue := reflect.ValueOf(handler)
	handlerKind := handlerValue.Kind()
	if handlerKind != reflect.Func {
		return nil, fmt.Errorf("handler must be a function")
	}

	handlerType := handlerValue.Type()
	var detectedSig handlerSignature

	// Check for SimpleHandler: func(cloudevents.Event) error
	if handlerType.NumIn() == 1 && handlerType.NumOut() == 1 {
		if handlerType.Out(0).String() == "error" {
			detectedSig = handlerSignatureSimple
		} else {
			return nil, fmt.Errorf("invalid handler signature")
		}
	} else if handlerType.NumIn() == 1 && handlerType.NumOut() == 2 {
		// Check for OutputHandler: func(cloudevents.Event) (*cloudevents.Event, error)
		if handlerType.Out(1).String() == "error" {
			detectedSig = handlerSignatureOutput
		} else {
			return nil, fmt.Errorf("invalid handler signature")
		}
	} else {
		return nil, fmt.Errorf("handler must have signature func(Event) error or func(Event) (*Event, error)")
	}

	// Get Kafka config from core
	config, err := core.GetKafkaConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kafka config: %w", err)
	}

	logger.Info("Kafka configuration loaded",
		"broker", config.Broker,
		"topic", config.Topic,
		"group", config.Group)

	// Create Kafka consumer
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.Broker,
		"group.id":          config.Group,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	// Create Kafka producer for output events (if handler returns events)
	var kafkaProducer *kafka.Producer
	if detectedSig == handlerSignatureOutput {
		kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": config.Broker,
		})
		if err != nil {
			kafkaConsumer.Close()
			return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
		}
		logger.Info("Kafka producer initialized for output events")
	}

	// Subscribe to topic with rebalance callback that starts from beginning
	rebalanceCb := func(c *kafka.Consumer, event kafka.Event) error {
		switch e := event.(type) {
		case kafka.AssignedPartitions:
			logger.Info("Partitions assigned", "partitions", e.Partitions)
			// Set offset to beginning before assigning to replay all messages
			for i := range e.Partitions {
				e.Partitions[i].Offset = kafka.OffsetBeginning
			}
			logger.Info("Starting from beginning for all partitions")
			return c.Assign(e.Partitions)
		case kafka.RevokedPartitions:
			logger.Info("Partitions revoked", "partitions", e.Partitions)
			return c.Unassign()
		}
		return nil
	}

	if err := kafkaConsumer.Subscribe(config.Topic, rebalanceCb); err != nil {
		kafkaConsumer.Close()
		if kafkaProducer != nil {
			kafkaProducer.Close()
		}
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", config.Topic, err)
	}

	return &Consumer{
		core:         core,
		consumer:     kafkaConsumer,
		producer:     kafkaProducer,
		handlerValue: handlerValue,
		handlerSig:   detectedSig,
		logger:       logger,
	}, nil
}

// Start begins consuming events (blocking)
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("Starting consumer")

	consecutiveErrors := 0
	maxConsecutiveErrors := 5
	pollTimeout := 100 * time.Millisecond

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer stopping")
			return ctx.Err()
		default:
			// Poll for messages with timeout to allow context cancellation
			msg, err := c.consumer.ReadMessage(pollTimeout)
			if err != nil {
				// Timeout is expected when no messages, not an error
				if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
					continue
				}
				c.logger.Error("Error reading message", "error", err)
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
				c.logger.Error("Error parsing CloudEvent", "error", err)
				continue
			}

			// Call user handler based on signature type
			if err := c.invokeHandler(event); err != nil {
				c.logger.Error("Handler error", "error", err, "event_type", event.Type())

				// Check if we should retry using core
				shouldRetry, retryErr := c.core.ShouldRetry(err.Error(), 1)
				if retryErr != nil {
					c.logger.Error("Error checking retry", "error", retryErr)
					continue
				}

				if shouldRetry {
					backoff, backoffErr := c.core.CalculateBackoff(1)
					if backoffErr != nil {
						c.logger.Error("Error calculating backoff", "error", backoffErr)
						continue
					}
					c.logger.Warn("Would retry after backoff", "backoff_ms", backoff, "note", "not implemented in PoC")
				}
			}
		}
	}
}

// invokeHandler calls the user's handler function and handles output events
func (c *Consumer) invokeHandler(event *cloudevents.Event) error {
	// Prepare arguments
	args := []reflect.Value{reflect.ValueOf(*event)}

	// Call handler
	results := c.handlerValue.Call(args)

	// Handle results based on handler signature
	switch c.handlerSig {
	case handlerSignatureSimple:
		// func(Event) error
		if !results[0].IsNil() {
			return results[0].Interface().(error)
		}
		return nil

	case handlerSignatureOutput:
		// func(Event) (*Event, error)
		// Check error first
		if !results[1].IsNil() {
			return results[1].Interface().(error)
		}

		// Handle output event if present
		if !results[0].IsNil() {
			outputEvent := results[0].Interface().(*cloudevents.Event)
			if err := c.publishOutputEvent(outputEvent); err != nil {
				return fmt.Errorf("failed to publish output event: %w", err)
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown handler signature")
	}
}

// publishOutputEvent routes and publishes an output event
func (c *Consumer) publishOutputEvent(event *cloudevents.Event) error {
	// Serialize event to JSON for routing
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Get output destination from core routing
	dest, err := c.core.GetOutputDestination(string(eventJSON))
	if err != nil {
		return fmt.Errorf("failed to get output destination: %w", err)
	}

	c.logger.Info("Routing output event",
		"event_type", event.Type(),
		"dest_type", dest.Type,
		"dest_target", dest.Target)

	// Handle different destination types
	switch dest.Type {
	case DestinationKafka:
		return c.publishToKafka(event, dest)
	case DestinationDiscard:
		c.logger.Info("Discarding output event", "event_type", event.Type())
		return nil
	case DestinationHTTP, DestinationRabbitMQ:
		// TODO: Implement HTTP and RabbitMQ publishing
		c.logger.Warn("Destination type not yet implemented, discarding event",
			"dest_type", dest.Type,
			"event_type", event.Type())
		return nil
	default:
		return fmt.Errorf("unknown destination type: %d", dest.Type)
	}
}

// publishToKafka publishes an event to a Kafka topic
func (c *Consumer) publishToKafka(event *cloudevents.Event, dest *OutputDestination) error {
	if c.producer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Produce to Kafka
	topic := dest.Target
	err = c.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          eventJSON,
		Key:            []byte(event.ID()),
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	c.logger.Info("Published output event to Kafka", "topic", topic, "event_type", event.Type())
	return nil
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
	if c.producer != nil {
		c.producer.Flush(5000)
		c.producer.Close()
	}
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
