# Multi-Language EDA SDK: FFI & WASI Proof of Concept
## Presentation Transcript

---

## Slide 1: The Problem

**[SPEAKER - 1 min]**

Hey everyone! Quick context: I've been working on a PoC for Functions 2.0 EDA that solves a real problem.

Look at CloudEvents SDKs - each language got its own implementation. Result? **CESQL only exists in Go and Java.** Python and JavaScript users? Out of luck. Same pattern everywhere: inconsistent implementations, maintenance nightmares, slow feature rollout.

For Functions 2.0 EDA, I wanted to avoid this. What if we wrote the logic once and shared it across all languages?

---

## Slide 2: The Solution - FFI & WASI

**[SPEAKER - 1.5 min]**

I evaluated four approaches. The winners:

**FFI (Foreign Function Interface)** - scored 3.75. Compile Rust to native libraries (.so, .dylib, .dll), generate bindings. Works with any language today. Downside: need macOS and Windows builders - it's cumbersome.

**WASI (WebAssembly System Interface)** - scored 3.45. Compile once to WASM, runs everywhere. One 100KB artifact. The catch? Host bindings generation only supports Rust right now. Other languages coming this year or next.

**My PoC:** Implement both to validate the architecture. **Recommendation:** Start with FFI (works today), migrate to WASI when it matures (Preview 3 this year, GA next year).

---

## Slide 3: Architecture

**[SPEAKER - 1 min]**

**Rust core** - shared logic lives here. Config, retry, routing, telemetry.

**Two compilation paths:**
- **FFI**: cbindgen generates C headers → native libraries for each platform
- **WASI**: compile to wasm32-wasip2 with WIT interfaces → one portable artifact

**Language SDKs** - thin wrappers. Handle Kafka/Camel transport, CloudEvents parsing, call user handlers. Delegate decisions to the Rust core.

**Key insight:** Share algorithms and decision logic. Keep I/O and native libraries in the SDKs.

---

## Slide 4: What's Actually Implemented

**[SPEAKER - 1.5 min]**

Let me be honest - this is a **proof of concept**, not production code.

**Rust core** - basically stubs:
- **Routing** - CloudEvents Subscriptions API filters (exact, prefix, suffix, all, any, not). This actually works and demonstrates code sharing.
- **Config, Retry, Telemetry** - placeholders with TODOs describing what production needs.

**Go SDK** - FFI works, calls into Rust core via CGO. WASI path exists but wasmtime-go doesn't support Component Model hosting yet, so it returns placeholders.

**Python SDK** - FFI works with cffi bindings. WASI experimental with wasmtime-py.

**The PoC is done** - it successfully showcases code reuse across Go and Python. That's the goal. If the team approves, we build proper implementations.

---

## Slide 5: What This Proves

**[SPEAKER - 1 min]**

**Code sharing works** - routing implementation is shared between Go and Python via FFI. Same code, same behavior.

**The potential is clear** - if we implement retry logic in the core, every language gets it. CESQL support? Every language gets it. Comprehensive telemetry? Every language gets it.

**Architecture is sound** - FFI works today. WASI will work when tooling matures (WASI Preview 2 only supports Rust hosts by design; Preview 3 onwards will support more languages).

**Type safety** - WIT gives compile-time checking for WASI. FFI with C headers gives similar benefits, though more manual.

---

## Slide 6: FFI vs WASI Trade-offs

**[SPEAKER - 1.5 min]**

**FFI:**
- ✅ Works with any language today, native performance
- 🟡 Type safety via C headers (manual - types not auto-generated like WIT)
- ❌ Need platform-specific builds (macOS, Windows, Linux × x86_64, ARM64), complex CI
- ❌ Debugging across boundary is tricky (C debugging)

**WASI:**
- ✅ Compile once, runs everywhere. One 100KB file. Easy distribution.
- ✅ Type safety via WIT (auto-generated bindings)
- ❌ Debugging across boundary is tricky now (future: DWARF debug info + source maps will enable proper debugging)
- ❌ Host bindings only support Rust right now (WASI Preview 2 by design; Preview 3+ will support more languages)

**Verdict:** FFI is practical today. WASI is the future - simpler builds, easier distribution. Start with FFI, migrate to WASI when Preview 3 lands.

---

## Slide 7: Developer Experience

**[SPEAKER - 1 min]**

Here's what developers write:

**Go:**
```go
func Handle(event cloudevents.Event) error {
    // Business logic here
}
```

**Python:**
```python
def handle(event: CloudEvent) -> None:
    # Business logic here
```

That's it. SDK handles Kafka, CloudEvents, retries, routing, telemetry. Routing config is YAML - exact matches, prefix patterns. When we add CESQL to the core, every language gets it automatically.

Developers write business logic, not infrastructure plumbing.

---

## Slide 7.5: Live Demo

**[SPEAKER - 2 min]**

Let me show you this working. I'll run the Go FFI example consuming from Kafka...

*[Demo: `make -C sdks/go/examples/ffi-output-example run`]*
*[Show: events being consumed, routing logic working, output events being produced]*

You can see the routing configuration in the YAML file, and the shared Rust core making routing decisions for both Go and Python.

---

## Slide 8: Decision Point & Next Steps

**[SPEAKER - 1.5 min]**

**This is a decision point.** The PoC validates the architecture. The team needs to decide: pursue this for Functions 2.0 EDA?

**If yes, we need to implement it properly** - not just fill in TODOs. Following my recommendation:
- Build FFI-based SDKs for all Functions 2.0 target languages
- Integrate into the Func tooling
- Implement full production-grade features:
  - Complete routing with all CloudEvents Subscriptions API filters
  - Configuration management (not hardcoded values)
  - Comprehensive telemetry with Prometheus metrics
  - Proper error handling and classification
  - Retry logic with exponential backoff and jitter
  - Dead-letter queue support
  - CESQL when CloudEvents Rust SDK supports it
- When WASI Preview 3 lands with broader language support, begin migration
- When WASI is GA, complete migration

**Bottom line:** The PoC is successful. Architecture is sound, code sharing is proven with working routing implementation. If approved, we have a clear path to production.

Thanks! Questions?

**[END - Total: ~12 minutes with demo]**
