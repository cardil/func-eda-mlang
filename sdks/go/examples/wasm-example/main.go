package main

import (
	"log/slog"
	"os"

	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/wasm"
)

func main() {
	// Run the EDA consumer with WASM core
	if err := wasm.Run(Handle); err != nil {
		slog.Error("Fatal error", "error", err)
		os.Exit(1)
	}
}
