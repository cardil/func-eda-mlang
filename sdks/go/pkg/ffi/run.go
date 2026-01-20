package ffi

import (
	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/sdk"
)

// DefaultCoreConstructor is the default FFI core constructor
var DefaultCoreConstructor sdk.CoreConstructor = func() (sdk.Core, error) {
	return NewCore()
}

// Run starts the EDA consumer using FFI core with the given handler
// This is the main entry point for FFI-based functions
// Handler can be either SimpleHandler or OutputHandler signature
func Run[H sdk.Handler](handler H, opts ...sdk.Option) error {
	// Check if a core constructor is already provided
	hasConstructor := false
	for _, opt := range opts {
		// Apply option to a temporary Options to check
		tempOpts := &sdk.Options{}
		opt(tempOpts)
		if tempOpts.CoreConstructor != nil {
			hasConstructor = true
			break
		}
	}

	// Add default FFI core constructor if not provided
	allOpts := opts
	if !hasConstructor {
		allOpts = append([]sdk.Option{sdk.WithCoreConstructor(DefaultCoreConstructor)}, opts...)
	}

	// Run with constructor
	return sdk.RunWithConstructor(handler, allOpts...)
}
