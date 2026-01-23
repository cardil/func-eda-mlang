package ffi

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// COutputDestination matches the C struct from Rust FFI
type COutputDestination struct {
	DestType uint32
	Target   *byte
	Cluster  *byte
}

var (
	// Library handle
	libHandle uintptr

	// Function pointers
	edaGetKafkaBroker        func() *byte
	edaGetKafkaTopic         func() *byte
	edaGetKafkaGroup         func() *byte
	edaFreeString            func(*byte)
	edaShouldRetry           func(*byte, uint32) int32
	edaCalculateBackoff      func(uint32) uint64
	edaGetOutputDestination  func(*byte) *COutputDestination
	edaFreeOutputDestination func(*COutputDestination)
	edaLoadRoutingConfig     func(*byte) bool

	// Ensure library is loaded only once
	loadOnce sync.Once
	loadErr  error
)

// libName returns the platform-specific library name
func libName() string {
	switch runtime.GOOS {
	case "windows":
		return "eda_core.dll"
	case "darwin":
		return "libeda_core.dylib"
	default:
		return "libeda_core.so"
	}
}

// extractEmbeddedLib extracts the embedded library to a temporary file
func extractEmbeddedLib() (string, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "eda-core-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Write the embedded library to the temp directory
	libPath := filepath.Join(tmpDir, libName())
	if err := os.WriteFile(libPath, embeddedLib, 0755); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to write library to temp file: %w", err)
	}

	return libPath, nil
}

// loadLibrary loads the embedded FFI library and registers all functions
func loadLibrary() error {
	var err error
	loadOnce.Do(func() {
		// Extract embedded library to temp file
		libPath, extractErr := extractEmbeddedLib()
		if extractErr != nil {
			err = fmt.Errorf("failed to extract embedded library: %w", extractErr)
			return
		}

		// Load the library
		var openErr error
		libHandle, openErr = purego.Dlopen(libPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if openErr != nil {
			err = fmt.Errorf("failed to load library from %s: %w", libPath, openErr)
			return
		}

		// Register all FFI functions
		if err = registerFunctions(); err != nil {
			err = fmt.Errorf("failed to register functions: %w", err)
			return
		}

		loadErr = nil
	})

	if err != nil {
		loadErr = err
		return err
	}

	return loadErr
}

// registerFunctions registers all C function pointers using purego
func registerFunctions() error {
	// Register eda_get_kafka_broker
	purego.RegisterLibFunc(&edaGetKafkaBroker, libHandle, "eda_get_kafka_broker")

	// Register eda_get_kafka_topic
	purego.RegisterLibFunc(&edaGetKafkaTopic, libHandle, "eda_get_kafka_topic")

	// Register eda_get_kafka_group
	purego.RegisterLibFunc(&edaGetKafkaGroup, libHandle, "eda_get_kafka_group")

	// Register eda_free_string
	purego.RegisterLibFunc(&edaFreeString, libHandle, "eda_free_string")

	// Register eda_should_retry
	purego.RegisterLibFunc(&edaShouldRetry, libHandle, "eda_should_retry")

	// Register eda_calculate_backoff
	purego.RegisterLibFunc(&edaCalculateBackoff, libHandle, "eda_calculate_backoff")

	// Register eda_get_output_destination
	purego.RegisterLibFunc(&edaGetOutputDestination, libHandle, "eda_get_output_destination")

	// Register eda_free_output_destination
	purego.RegisterLibFunc(&edaFreeOutputDestination, libHandle, "eda_free_output_destination")

	// Register eda_load_routing_config
	purego.RegisterLibFunc(&edaLoadRoutingConfig, libHandle, "eda_load_routing_config")

	return nil
}

// cStringWithBuf converts a Go string to a C string (null-terminated byte array).
// Returns both the pointer and the backing slice. The caller must keep the slice
// alive (e.g., with runtime.KeepAlive) until the C function completes to prevent
// garbage collection.
func cStringWithBuf(s string) (*byte, []byte) {
	b := append([]byte(s), 0)
	return &b[0], b
}

// goString converts a C string pointer to a Go string
func goString(cstr *byte) string {
	if cstr == nil {
		return ""
	}

	var length int
	for {
		ptr := (*byte)(unsafe.Add(unsafe.Pointer(cstr), length))
		if *ptr == 0 {
			break
		}
		length++
	}

	return string(unsafe.Slice(cstr, length))
}
