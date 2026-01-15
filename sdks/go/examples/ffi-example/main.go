package main

import (
	"log"

	"github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/ffi"
)

func main() {
	// Run the EDA consumer with FFI core
	if err := ffi.Run(Handle); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
