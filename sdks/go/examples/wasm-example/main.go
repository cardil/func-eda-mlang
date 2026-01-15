package main

import (
	"log"

	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/wasm"
)

func main() {
	// Run the EDA consumer with WASM core
	if err := wasm.Run(Handle); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
