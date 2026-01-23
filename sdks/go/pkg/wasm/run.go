package wasm

//go:generate sh -c "cp ../../../../bindings/wasm/target/wasm32-wasip2/release/eda_wasm.wasm eda_core.wasm"
//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go@v0.7.0 generate --out gen --package-root github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/wasm/gen ../../../../wit

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/sdk"
)

//go:embed eda_core.wasm
var embeddedWasm []byte

// DefaultCoreConstructor creates a WASM core with the embedded WASM component
var DefaultCoreConstructor sdk.CoreConstructor = func() (sdk.Core, error) {
	return NewCoreFromBytes(context.Background(), embeddedWasm)
}

// NewCoreFromBytes creates a WASM core from WASM component bytes
func NewCoreFromBytes(ctx context.Context, wasmBytes []byte) (*Core, error) {
	if len(wasmBytes) == 0 {
		return nil, fmt.Errorf("WASM bytes are empty")
	}

	return NewCore(ctx, wasmBytes)
}

// NewCoreFromFile creates a WASM core from a file path (for custom WASM components)
func NewCoreFromFile(ctx context.Context, wasmPath string) (*Core, error) {
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read WASM file: %w", err)
	}
	return NewCoreFromBytes(ctx, wasmBytes)
}

// Run starts the EDA consumer using WASM core with the given handler
// This is the main entry point for WASM-based functions
// Handler can be either SimpleHandler or OutputHandler
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

	// Add default WASM core constructor if not provided
	allOpts := opts
	if !hasConstructor {
		allOpts = append([]sdk.Option{sdk.WithCoreConstructor(DefaultCoreConstructor)}, opts...)
	}

	// Run with constructor
	return sdk.RunWithConstructor(handler, allOpts...)
}
