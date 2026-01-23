# EDA Functions SDK - Go

Go SDK for building Event-Driven Architecture (EDA) functions with shared Rust core logic.

## Features

- **FFI Backend**: Uses cgo to call Rust shared library (`libeda_core.so`)
- **WASM Backend**: Uses wazero to run Rust WASM module (`eda_wasm.wasm`)
- **Unified Interface**: Same API regardless of backend
- **Kafka Integration**: Built-in Kafka consumer with CloudEvents support
- **Simple API**: Minimal boilerplate for function developers

## Quick Start

### FFI Example

```go
package main

import (
    "fmt"
    "log"

    cloudevents "github.com/cloudevents/sdk-go/v2"
    "github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/ffi"
)

func Handle(event cloudevents.Event) error {
    fmt.Printf("Received: %s (type: %s)\n", event.ID(), event.Type())
    return nil
}

func main() {
    if err := ffi.Run(Handle); err != nil {
        log.Fatalf("Error: %v", err)
    }
}
```

### WASM Example

```go
package main

import (
    "fmt"
    "log"

    cloudevents "github.com/cloudevents/sdk-go/v2"
    "github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/wasm"
)

func Handle(event cloudevents.Event) error {
    fmt.Printf("Received: %s (type: %s)\n", event.ID(), event.Type())
    return nil
}

func main() {
    if err := wasm.Run(Handle); err != nil {
        log.Fatalf("Error: %v", err)
    }
}
```

## Building

### Prerequisites

- Go 1.21+
- Rust toolchain (for building core)
- librdkafka (for Kafka client)

### Build Core

```bash
# From repository root
make core-build-ffi    # Build FFI library
make core-build-wasm   # Build WASM module
make core-headers      # Generate C headers
```

### Build SDK

```bash
make sdk-go-build      # Build Go SDK
```

### Build Examples

```bash
make example-go-ffi    # Build FFI example
make example-go-wasm   # Build WASM example
```

## Running Examples

### FFI Example

```bash
make run-go-ffi
```

### WASM Example

```bash
make run-go-wasm
```

## Advanced Usage

### Custom Options

```go
import (
    "context"
    "time"

    "github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/ffi"
    "github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/sdk"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    err := ffi.Run(Handle, sdk.WithContext(ctx))
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
}
```

### Custom Core Constructor

```go
import (
    "github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/sdk"
    "github.com/openshift-knative/func-eda-mlang/sdks/go/pkg/wasm"
)

func main() {
    customConstructor := func() (sdk.Core, error) {
        return wasm.NewCore(context.Background(), "/custom/path/to/module.wasm")
    }

    err := wasm.Run(Handle, sdk.WithCoreConstructor(customConstructor))
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
}
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  User Function                      │
│              func Handle(event) error               │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│                   SDK Layer                         │
│  - Consumer (Kafka + CloudEvents)                   │
│  - Run() orchestration                              │
└─────────────────────────────────────────────────────┘
                         │
          ┌──────────────┴──────────────┐
          ▼                             ▼
┌──────────────────┐          ┌──────────────────┐
│   FFI Backend    │          │  WASM Backend    │
│   (cgo)          │          │  (wazero)        │
└──────────────────┘          └──────────────────┘
          │                             │
          └──────────────┬──────────────┘
                         ▼
              ┌──────────────────┐
              │   Rust Core      │
              │  - Config        │
              │  - Retry logic   │
              │  - Routing       │
              └──────────────────┘
```

## Package Structure

- `pkg/sdk/` - Core SDK interfaces and consumer logic
- `pkg/ffi/` - FFI (cgo) implementation
- `pkg/wasm/` - WASM (wazero) implementation
- `examples/` - Example functions

## Testing

```bash
make sdk-go-test
```

## License

See repository root LICENSE file.
