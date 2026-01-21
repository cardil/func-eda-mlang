# EDA Functions SDK - Python

Python SDK for building Event-Driven Architecture (EDA) functions with shared Rust core logic.

## Features

- **FFI Backend**: Uses cffi to call Rust shared library (`libeda_core.so`)
- **WASM Backend**: Placeholder for future implementation
- **Unified Interface**: Same API regardless of backend
- **Kafka Integration**: Built-in Kafka consumer with CloudEvents support
- **Simple API**: Minimal boilerplate for function developers

## Installation

```bash
# Install in development mode
pip install -e .

# Or install from PyPI (when published)
pip install eda-sdk
```

## Quick Start

### Simple Handler Example

```python
from cloudevents.http import CloudEvent
from eda_sdk.ffi import run

def handle(event: CloudEvent) -> None:
    print(f"Received: {event['id']} (type: {event['type']})")

if __name__ == "__main__":
    run(handle)
```

### Output Handler Example

```python
from datetime import datetime, timezone
from typing import Optional
from cloudevents.http import CloudEvent
from eda_sdk.ffi import run

def handle(event: CloudEvent) -> Optional[CloudEvent]:
    # Process event and return output event
    output = CloudEvent(
        {
            "specversion": "1.0",
            "type": "com.example.processed",
            "source": "my-function",
            "id": f"processed-{event['id']}",
        },
        data={"processed_at": datetime.now(timezone.utc).isoformat()},
    )
    return output

if __name__ == "__main__":
    run(handle)
```

## Configuration

The SDK reads configuration from environment variables:

- `KAFKA_BROKER` - Kafka broker address (default: `localhost:9092`)
- `KAFKA_TOPIC` - Input topic to consume from (default: `events`)
- `KAFKA_GROUP` - Consumer group ID (default: `eda-consumer`)

## Output Event Routing

Create a `routing.yaml` file in the same directory as your handler:

```yaml
routing:
  default:
    type: kafka
    target: events
    cluster: default
  
  rules:
    - name: processed-events
      filter:
        exact:
          type: "com.example.processed"
      destination:
        type: kafka
        target: processed-events
        cluster: default
```

## Development

### Prerequisites

- Python 3.10+
- Rust toolchain (for building core)
- librdkafka (for Kafka client)

### Build Core Library

```bash
# From repository root
make core-build-ffi    # Build FFI library
make core-headers      # Generate C headers
```

### Copy Library to SDK

```bash
# From repository root
make sdk-python-deps   # Install Python package and copy libraries
```

### Run Examples

```bash
# Start Kafka infrastructure
make kafka-up

# Run simple example
make example-python-ffi

# Run output example
cd sdks/python/examples
python ffi_output_example.py
```

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  User Function                      │
│              def handle(event) -> ...               │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│                   SDK Layer                         │
│  - Consumer (Kafka + CloudEvents)                   │
│  - run() orchestration                              │
└─────────────────────────────────────────────────────┘
                         │
          ┌──────────────┴──────────────┐
          ▼                             ▼
┌──────────────────┐          ┌──────────────────┐
│   FFI Backend    │          │  WASM Backend    │
│   (cffi)         │          │  (placeholder)   │
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

- `eda_sdk/` - Core SDK package
  - `types.py` - Type definitions
  - `core.py` - Core protocol interface
  - `consumer.py` - Kafka consumer with CloudEvents
  - `run.py` - Main run function
  - `ffi/` - FFI implementation
    - `core.py` - FFI core implementation
    - `loader.py` - Library loading with cffi
    - `run.py` - FFI run function
    - `libs/` - Embedded native libraries
    - `include/` - C header files
  - `wasm/` - WASM implementation (placeholder)
- `examples/` - Example functions

## License

Apache-2.0
