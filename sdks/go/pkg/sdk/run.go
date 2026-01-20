package sdk

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// Run starts the EDA consumer with the given core and handler
// This is the main entry point that users call from their main()
// Handler can be either SimpleHandler or OutputHandler
func Run[H Handler](core Core, handler H, opts ...Option) error {
	if core == nil {
		return fmt.Errorf("core cannot be nil")
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
		slog.Info("Shutting down...")
		cancel()
	}()

	// Start consuming
	slog.Info("Starting EDA consumer...")
	if err := consumer.Start(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("consumer error: %w", err)
	}

	slog.Info("Consumer stopped")
	return nil
}
