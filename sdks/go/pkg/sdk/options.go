package sdk

import "context"

// CoreConstructor is a function that creates a Core instance
type CoreConstructor func() (Core, error)

// Options holds configuration for the consumer
type Options struct {
	// Context for the consumer (if nil, a default context will be created)
	Context context.Context

	// CoreConstructor creates the Core instance (if nil, must be provided externally)
	CoreConstructor CoreConstructor

	// Additional options can be added here in the future
	// e.g., custom logger, error handlers, etc.
}

// Option is a function that modifies Options
type Option func(*Options)

// WithContext sets a custom context for the consumer
func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// WithCoreConstructor sets a custom core constructor
func WithCoreConstructor(constructor CoreConstructor) Option {
	return func(o *Options) {
		o.CoreConstructor = constructor
	}
}

// applyOptions applies the given options and returns the final Options
func applyOptions(opts []Option) *Options {
	options := &Options{
		Context: context.Background(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
