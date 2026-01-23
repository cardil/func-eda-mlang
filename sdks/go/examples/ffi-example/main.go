package main

import (
	"log/slog"
	"os"

	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/ffi"
)

func main() {
	// Run the EDA consumer with FFI core
	if err := ffi.Run(Handle); err != nil {
		slog.Error("Fatal error", "error", err)
		os.Exit(1)
	}
}
