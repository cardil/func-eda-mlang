package sdk

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// RunWithCore starts the EDA consumer with an explicit core instance
func RunWithCore(core Core, handler HandlerFunc, opts ...Option) error {
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

// RunWithConstructor starts the EDA consumer using a core constructor from options
func RunWithConstructor(handler HandlerFunc, opts ...Option) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// Apply options
	options := applyOptions(opts)

	if options.CoreConstructor == nil {
		return fmt.Errorf("core constructor must be provided via WithCoreConstructor option")
	}

	// Create core using constructor
	core, err := options.CoreConstructor()
	if err != nil {
		return fmt.Errorf("failed to create core: %w", err)
	}
	defer core.Close()

	// Run with the created core
	return RunWithCore(core, handler, opts...)
}
