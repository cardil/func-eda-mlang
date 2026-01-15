package sdk

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// Run starts the EDA consumer with the given core and handler
// This is the main entry point that users call from their main()
func Run(core Core, handler HandlerFunc, opts ...Option) error {
	if core == nil {
		return fmt.Errorf("core cannot be nil")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// Apply options
	options := applyOptions(opts)

	// Create consumer
	consumer, err := NewConsumer(core, handler)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer consumer.Close()

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(options.Context)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Start consuming
	log.Println("Starting EDA consumer...")
	if err := consumer.Start(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("consumer error: %w", err)
	}

	log.Println("Consumer stopped")
	return nil
}
