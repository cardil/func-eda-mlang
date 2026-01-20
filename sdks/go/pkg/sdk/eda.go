package sdk

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

// RunWithCore starts the EDA consumer with an explicit core instance
// Handler can be either SimpleHandler or OutputHandler
func RunWithCore[H Handler](core Core, handler H, opts ...Option) error {
	if core == nil {
		return fmt.Errorf("core cannot be nil")
	}

	// Apply options
	options := applyOptions(opts)

	// Try to load routing configuration from the caller's directory
	// Walk up the call stack to find the first caller outside the SDK
	var callerDir string
	for i := 1; i < 10; i++ {
		_, file, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Skip SDK internal files
		if filepath.Base(filepath.Dir(file)) != "sdk" &&
			filepath.Base(filepath.Dir(file)) != "ffi" &&
			filepath.Base(filepath.Dir(file)) != "wasm" {
			callerDir = filepath.Dir(file)
			break
		}
	}

	if callerDir != "" {
		routingConfigPath := filepath.Join(callerDir, "routing.yaml")
		if _, err := os.Stat(routingConfigPath); err == nil {
			slog.Info("Loading routing configuration", "path", routingConfigPath)
			if err := core.LoadRoutingConfig(routingConfigPath); err != nil {
				return fmt.Errorf("failed to load routing config: %w", err)
			}
		}
	}

	// Create consumer (NewConsumer accepts interface{} and does runtime type checking)
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

// RunWithConstructor starts the EDA consumer using a core constructor from options
// Handler can be either SimpleHandler or OutputHandler
func RunWithConstructor[H Handler](handler H, opts ...Option) error {
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
