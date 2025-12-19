# Multi-Language EDA Library - Research Document

## Goal

Build a multi-language EDA SDK that **maximizes code reuse** across Java, Python, JavaScript/TypeScript, and Go. The SDK provides a configuration layer to simplify writing event-driven functions:

```go
// Go example
func Handle(event cloudevents.Event) error {
    // User's business logic
}
```

The SDK handles transport configuration (Kafka, HTTP), CloudEvents conversion, and error handling.

## What Can Be Shared?

| Component              | Shareable? | Notes |
|------------------------|--------|-------|
| Configuration parsing  | ✅ Yes | Parse YAML/JSON config |
| Event routing logic    | ✅ Yes | Match events to handlers |
| Retry/backoff policies | ✅ Yes | Algorithm is language-agnostic |
| Dead-letter decisions  | ✅ Yes | Policy logic |
| Transport clients      | ❌ No  | Kafka clients differ per language |
| CloudEvents parsing    | ❌ No  | Use existing CE SDKs |
| Handler invocation     | ❌ No  | Language-specific |
| Framework integration  | ❌ No  | Quarkus vs Spring vs Flask |

---

## Approach 1: WebAssembly (WASM) Embedded Core

Write shareable core logic in Rust, compile to WASM, embed in thin language-specific wrappers.

### Architecture

```
┌──────────────────────────────────────────────────────┐
│              Language Wrapper (thin)                 │
│  - Kafka/HTTP transport (native library)             │
│  - CloudEvents SDK                                   │
│  - Handler invocation                                │
│  - Hosts WASM runtime                                │
│                      │                               │
│            calls for decisions                       │
│                      ▼                               │
│  ┌────────────────────────────────────────────────┐  │
│  │         WASM Core Module (Rust)                │  │
│  │  - parse_config(yaml) -> Config                │  │
│  │  - should_retry(error, attempt) -> bool        │  │
│  │  - calculate_backoff(attempt) -> duration      │  │
│  │  - route_event(type) -> handler_id             │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

### WASM Runtimes per Language

| Language | Runtime | Maturity |
|----------|---------|----------|
| Go | wazero | ⭐⭐⭐⭐⭐ Pure Go, zero deps |
| Java | GraalWasm, Chicory | ⭐⭐⭐⭐ |
| Python | wasmtime-py | ⭐⭐⭐⭐ |
| JavaScript | Native V8 | ⭐⭐⭐⭐⭐ |

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Code Reuse** | ⭐⭐⭐⭐ | ~40-50% logic shared |
| Consistency | ⭐⭐⭐⭐⭐ | Core logic guaranteed identical |
| Maintenance | ⭐⭐⭐⭐ | Single core implementation |
| Debug Experience | ⭐⭐⭐ | Cross-boundary debugging harder |
| Performance | ⭐⭐⭐⭐ | Near-native, small overhead |

### Pros
- Core logic (config, retry, routing) is write-once
- Wrappers use native transport libraries
- WASM modules are small (~100KB)

### Cons
- Adds WASM runtime dependency
- Requires Rust expertise
- GraalWasm needs GraalVM (Chicory for standard JVM)

---

## Approach 2: FFI Bindings (Native Shared Library)

Build core logic in Rust/Go, compile to native shared library, generate bindings for each language.

### Architecture

```
┌───────────────────────────────────────────────────────┐
│       Core Library (Rust with #[repr(C)])             │
│  - Configuration parsing                              │
│  - Retry logic                                        │
│  - Event routing                                      │
└───────────────────────────────────────────────────────┘
                         │
              FFI bindings generation
                         │
     ┌───────────────────┼───────────────────┐
     ▼                   ▼                   ▼
┌─────────┐        ┌─────────┐         ┌─────────┐
│  Java   │        │ Python  │         │   Go    │
│  (JNI)  │        │ (cffi)  │         │  (cgo)  │
└─────────┘        └─────────┘         └─────────┘
```

### Binding Tools

| Tool | Source | Targets |
|------|--------|---------|
| UniFFI | Rust | Python, Kotlin, Swift, Ruby |
| cbindgen | Rust | C headers → JNI/ctypes/cgo |
| PyO3 | Rust | Python (native) |
| cgo | Go | C-compatible shared lib |

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Code Reuse** | ⭐⭐⭐⭐⭐ | ~50-60% logic shared |
| Consistency | ⭐⭐⭐⭐⭐ | Native code, identical behavior |
| Maintenance | ⭐⭐⭐⭐ | Single core + binding layer |
| Debug Experience | ⭐⭐⭐ | Cross-language debugging hard |
| Performance | ⭐⭐⭐⭐⭐ | Native performance |

### Pros
- Best raw performance
- Mature tooling (FFI is well-understood)
- Single source of truth

### Cons
- Platform-specific builds (Linux/macOS/Windows × x86/ARM)
- Distribution complexity
- Memory management across boundary

---

## Approach 3: Code Generation from Schema

Define SDK structure in a schema, generate scaffolding per language.

### Architecture

```
┌───────────────────────────────────────────────────────┐
│              SDK Schema (JSON Schema/OpenAPI)         │
│  - Configuration types (?)                            │
│  - Handler interface (?)                              │
└───────────────────────────────────────────────────────┘
                         │
                    Code Generator
                         │
     ┌───────────────────┼───────────────────┐
     ▼                   ▼                   ▼
┌─────────────┐   ┌─────────────┐    ┌─────────────┐
│  Go SDK     │   │ Python SDK  │    │  Java SDK   │
│  Implement: │   │  Implement: │    │  Implement: │
│  - All logic│   │  - All logic│    │  - All logic│
└─────────────┘   └─────────────┘    └─────────────┘
```

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Code Reuse** | ⭐ | CE SDK already provides types; unclear what to generate |
| Consistency | ⭐⭐ | Only structural consistency, behavior may vary |
| Maintenance | ⭐⭐ | Must maintain N implementations |
| Debug Experience | ⭐⭐⭐⭐⭐ | Pure native code |
| Performance | ⭐⭐⭐⭐⭐ | Native |

### Pros
- No foreign runtime dependencies
- Full native debugging

### Cons
- **Minimal code reuse** - CE SDK already provides event types
- All business logic duplicated per language
- High maintenance burden
- Risk of behavioral divergence

---

## Approach 4: Sidecar Process

Run shared logic as a separate process, language SDKs communicate via IPC/HTTP.

### Architecture

```
┌───────────────────────────────────────────────────────┐
│                      Pod/Container                    │
│  ┌─────────────────┐     ┌─────────────────────────┐  │
│  │   User App      │◄───►│   EDA Sidecar (Go/Rust) │  │
│  │   (any lang)    │ IPC │   - Config parsing      │  │
│  │                 │     │   - Transport (Kafka)   │  │
│  │                 │     │   - Retry logic         │  │
│  └─────────────────┘     └─────────────────────────┘  │
└───────────────────────────────────────────────────────┘
```

### Evaluation

| Criterion | Rating | Notes |
|-----------|--------|-------|
| **Code Reuse** | ⭐⭐⭐⭐ | ~50-60% in sidecar, SDK still needed per language |
| Consistency | ⭐⭐⭐⭐⭐ | Core logic in single implementation |
| Maintenance | ⭐⭐ | High operational complexity |
| Debug Experience | ⭐⭐⭐ | Two processes to debug |
| Performance | ⭐⭐⭐ | IPC overhead (~1ms) |

### Pros
- Core logic (transport, retry, routing) centralized
- Already proven (Dapr, Envoy)

### Cons
- SDK still needed per language (receive from sidecar, parse CE, invoke handler)
- High operational complexity (deploy sidecar, routing infrastructure)
- Not suitable for all deployment models
- IPC latency for every event

---

## Comparison Matrix

| Approach | Code Reuse | Consistency | Maintenance | Debug | Performance |
|----------|-----------|-------------|-------------|-------|-------------|
| **FFI Bindings** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **WASM Core** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ |
| **Sidecar** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| **Code Gen** | ⭐ | ⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

---

## Weighted Evaluation

**Weights:** Code Reuse 40%, Debug 25%, Consistency 15%, Maintenance 10%, Performance 10%

| Approach | Weighted Score | Ranking |
|----------|---------------|---------|
| **FFI Bindings** | 5×.40 + 2×.25 + 5×.15 + 4×.10 + 5×.10 = **3.75** | 🥇 |
| **WASM Core** | 4×.40 + 2×.25 + 5×.15 + 4×.10 + 4×.10 = **3.45** | 🥈 |
| **Sidecar** | 4×.40 + 3×.25 + 5×.15 + 2×.10 + 3×.10 = **3.40** | 🥉 |
| **Code Gen** | 1×.40 + 5×.25 + 2×.15 + 2×.10 + 5×.10 = **2.45** | 4th |

### Debug Experience Details

| Approach | Debug Notes |
|----------|-------------|
| **Code Gen** ⭐⭐⭐⭐⭐ | Pure native code, standard debuggers work perfectly |
| **Sidecar** ⭐⭐⭐ | Two processes but clear boundary, debug independently |
| **WASM/FFI** ⭐⭐ | Stack traces don't cross boundary, need separate tooling |

---

## Recommendations for PoC

Implement **both top approaches** (FFI and WASM) with a real EDA function consuming from Kafka.

### PoC Goal: Showcase Developer UX (DUX)

The PoC demonstrates how a developer writes an EDA function:

**Go:**
```go
package main

import (
    "fmt"
    cloudevents "github.com/cloudevents/sdk-go/v2"
)

func Handle(event cloudevents.Event) error {
    fmt.Printf("Received: %s\n", event.Type())
    return nil
}
```

**Python:**
```python
from cloudevents.http import CloudEvent

def handle(event: CloudEvent) -> None:
    print(f"Received: {event['type']}")
```

The SDK (with shared core) handles Kafka subscription, CE parsing, and handler invocation.

### Shared Core (Mock/Static for PoC)

```rust
// Rust - static config, noop retry
fn get_kafka_config() -> KafkaConfig {
    KafkaConfig { broker: "localhost:9092", topic: "events", group: "poc" }
}

fn should_retry(_error: &str, _attempt: u32) -> bool {
    false  // noop
}
```

### Approach A: FFI Bindings

1. Compile Rust core to shared library (.so/.dylib)
2. Go SDK: cgo bindings + confluent-kafka-go
3. Python SDK: cffi bindings + confluent-kafka-python
4. Run function consuming real Kafka messages

### Approach B: WASM Core

1. Compile Rust core to WASM
2. Go SDK: wazero + confluent-kafka-go
3. Python SDK: wasmtime-py + confluent-kafka-python
4. Run function consuming real Kafka messages

### PoC Comparison Criteria

| Criterion | Evaluate |
|-----------|----------|
| **DUX simplicity** | How clean is the user-facing code? |
| Build complexity | FFI platform builds vs WASM single artifact |
| Integration effort | Wiring shared core + Kafka client |
| Debugging | Error messages, stack traces |

### Not Recommended for PoC

- **Code Generation:** Minimal code reuse (2.45 score)
- **Sidecar:** Different deployment model, not embedded

---

## References

- [WebAssembly Component Model](https://component-model.bytecodealliance.org/)
- [wazero](https://wazero.io/)
- [wasmtime-py](https://github.com/bytecodealliance/wasmtime-py)
- [UniFFI](https://mozilla.github.io/uniffi-rs/latest/)
- [CloudEvents SDK Requirements](https://github.com/cloudevents/spec/blob/main/cloudevents/SDK.md)