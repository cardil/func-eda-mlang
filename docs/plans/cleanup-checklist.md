# EDA Functions SDK PoC - Cleanup Checklist

This document identifies all loose ends in the PoC that need to be addressed before final documentation.

## Summary

| Category | Must Fix | Nice to Have | Known Limitation |
|----------|----------|--------------|------------------|
| Git/VCS | 1 | 0 | 0 |
| Build System | 2 | 1 | 0 |
| Code Quality | 4 | 3 | 4 |
| Documentation | 2 | 2 | 0 |
| Testing | 2 | 2 | 0 |
| Dependencies | 0 | 1 | 0 |

---

## 1. Git/VCS Status

### Must Fix

- [ ] **Python SDK not committed** - Verify Python SDK files are tracked in git
  - Files exist in `sdks/python/` but may not be committed
  - Run `git status` to check for uncommitted files
  - Commit all Python SDK files

### Current .gitignore Assessment

The `.gitignore` is comprehensive and correctly ignores:
- Build artifacts: `.make/`, `.tools/`, `**/target/`
- Generated libraries: `*.so`, `*.dylib`, `*.dll`
- Python venv and caches
- IDE files

No changes needed to `.gitignore`.

---

## 2. Build System - Makefile

### Must Fix

- [ ] **Split Makefile into smaller targeted Makefiles**
  - Current: 388 lines in single file
  - Proposed structure below

- [ ] **Inconsistent Makefile target naming** between SDKs
  - Go: `make run-go-ffi-output` (verb-prefix)
  - Python: `make example-python-ffi-output` (noun-prefix)
  - **Solution**:
    - Delegate standard tasks from root: `clean`, `build`, `check`, `test`, `e2e`
    - For `run`: use interactive selector (e.g., `gum choose`) at root level
    - `make run` → presents menu to choose which example to run
    - Each example has own Makefile with `run` target
    - **Run lifecycle**: `make run` should:
      1. Start infra (`make -C infra up`)
      2. Run selected example
      3. Stop infra on exit (`make -C infra down`) - use trap for cleanup
    - Note: Use generic `infra` target names (`up`/`down`) not `kafka-*` to support future transports (Camel, etc.)

### Nice to Have

- [ ] **Add `make check` target** - Run all linters and formatters
  - cargo fmt, cargo clippy for Rust
  - go fmt, go vet for Go
  - black, ruff, mypy for Python

### Proposed Makefile Structure

```
Makefile                    # Root Makefile (delegates to sub-makefiles)
├── core/Makefile           # Rust core build, FFI, WASM
├── sdks/go/Makefile        # Go SDK build, test, examples
├── sdks/python/Makefile    # Python SDK build, test, examples
└── infra/Makefile          # Kafka infrastructure
```

#### Root Makefile (Proposed)

```makefile
# Root Makefile - delegates to component Makefiles

SUBDIRS = core sdks/go sdks/python infra

.PHONY: help all clean test

help:
    @echo "Top-level targets:"
    @echo "  all        - Build everything"
    @echo "  clean      - Clean all artifacts"
    @echo "  test       - Run all tests"
    @echo ""
    @echo "Component targets (prefix with component name):"
    @echo "  core-*     - Rust core targets"
    @echo "  go-*       - Go SDK targets"
    @echo "  python-*   - Python SDK targets"
    @echo "  kafka-*    - Infrastructure targets"

all:
    $(MAKE) -C core all
    $(MAKE) -C sdks/go all
    $(MAKE) -C sdks/python all

clean:
    $(MAKE) -C core clean
    $(MAKE) -C sdks/go clean
    $(MAKE) -C sdks/python clean

test:
    $(MAKE) -C core test
    $(MAKE) -C sdks/go test
    $(MAKE) -C sdks/python test

# Delegate pattern: core-* targets
core-%:
    $(MAKE) -C core $*

# Delegate pattern: go-* targets
go-%:
    $(MAKE) -C sdks/go $*

# Delegate pattern: python-* targets
python-%:
    $(MAKE) -C sdks/python $*

# Delegate pattern: kafka-* targets
kafka-%:
    $(MAKE) -C infra $*
```

#### core/Makefile (Proposed)

```makefile
# Core Rust library Makefile

.PHONY: all build-ffi build-wasm headers test clean

all: build-ffi build-wasm headers

build-ffi:
    cargo build --release --lib

build-wasm:
    cd ../bindings/wasm && cargo build --release --target wasm32-wasip2

headers: build-ffi
    cbindgen --config ../bindings/ffi/cbindgen.toml --output ../bindings/ffi/include/eda_core.h .

test:
    cargo test

clean:
    cargo clean
```

#### sdks/go/Makefile (Proposed)

```makefile
# Go SDK Makefile

.PHONY: all build test examples clean

all: build examples

build:
    CGO_ENABLED=1 go build ./...

test:
    CGO_ENABLED=1 go test -v ./...

examples: example-ffi example-wasm

example-ffi:
    cd examples/ffi-example && CGO_ENABLED=1 go build -o ffi-consumer .

example-wasm:
    cd examples/wasm-example && go build -o wasm-consumer .

clean:
    go clean
    rm -f examples/*/ffi-consumer examples/*/wasm-consumer
```

---

## 3. Code Quality

### Must Fix

- [ ] **Event IDs not displayed correctly** - Bug in CloudEvent parsing
  - Go example shows `''` (empty string) for event ID
  - Python example shows `unknown` for event ID
  - Investigate CloudEvent parsing in both SDKs
  - Likely issue with how structured CloudEvents are parsed from Kafka messages

- [ ] **Inconsistent consumer group offset handling**
  - Python: subscribes to end of stream
  - Go: subscribes to beginning of stream
  - **Solution**: Add `auto_offset_reset` policy to core lib `KafkaConfig`
  - SDKs should read this from core instead of hardcoding

- [ ] **Inconsistent SDK consumer TODO for HTTP/RabbitMQ**
  - [`sdks/go/pkg/sdk/consumer.go:275`](sdks/go/pkg/sdk/consumer.go:275) - TODO for HTTP/RabbitMQ
  - [`sdks/python/eda_sdk/consumer.py:244`](sdks/python/eda_sdk/consumer.py:244) - TODO for HTTP/RabbitMQ
  - Both should either implement or document as known limitation

- [ ] **Go SDK README inaccuracy** - States "Uses cgo" but actually uses purego
  - [`sdks/go/README.md:7`](sdks/go/README.md:7) says "Uses cgo"
  - Implementation uses `github.com/ebitengine/purego` for FFI

### Nice to Have

- [ ] **Remove unused Go dependencies** - Review go.mod for unused imports
  - `wasmtime-go/v40` is used but WASM implementation is placeholder

- [ ] **Add Python type stubs** - Full mypy strict mode compatibility
  - pyproject.toml has mypy configured but no py.typed marker

- [ ] **Consolidate event serialization** - Duplicate code in Python consumer
  - `_publish_output_event` and `_publish_to_kafka` have duplicate dict building

### Known Limitations (Document, Don't Fix)

These are placeholder implementations acknowledged as PoC scope:

- [ ] **Rust retry logic placeholders** - Document in README
  - [`core/src/retry.rs:22`](core/src/retry.rs:22) - `classify_error` returns Unknown
  - [`core/src/retry.rs:35`](core/src/retry.rs:35) - `get_retry_decision` always returns no-retry
  - [`core/src/retry.rs:50`](core/src/retry.rs:50) - `calculate_backoff` returns 0

- [ ] **Rust telemetry placeholders** - Document in README
  - [`core/src/telemetry.rs:8`](core/src/telemetry.rs:8) - `record_event_received` is stub
  - [`core/src/telemetry.rs:18`](core/src/telemetry.rs:18) - `record_event_processed` is stub
  - [`core/src/telemetry.rs:28`](core/src/telemetry.rs:28) - `record_retry_attempt` is stub

- [ ] **WASM Component Model not fully implemented** - Document in README
  - [`sdks/go/pkg/wasm/wasm.go:70`](sdks/go/pkg/wasm/wasm.go:70) - Multiple TODOs for ABI
  - [`sdks/python/eda_sdk/wasm/core.py`](sdks/python/eda_sdk/wasm/core.py) - Raises NotImplementedError

- [ ] **CESQL routing not supported** - Document in README
  - [`core/src/routing.rs:9`](core/src/routing.rs:9) - TODO for CESQL support
  - CloudEvents Rust SDK v0.9 doesn't support CESQL

---

## 4. Documentation

### Must Fix

- [ ] **Expand root README.md** - Currently only 11 lines
  - Add project overview with architecture diagram
  - Add quick start guide
  - Add links to SDK-specific READMEs
  - Document known limitations
  - Add contributing section

- [ ] **Update Go README FFI description**
  - Change "Uses cgo" to "Uses purego" at line 7

### Nice to Have

- [ ] **Add ARCHITECTURE.md** - Document overall design decisions
  - Explain FFI vs WASM approaches
  - Explain routing system design
  - Explain multi-language strategy

- [ ] **Add CHANGELOG.md** - Track changes for PoC iterations

---

## 5. Testing

### Approach

**E2E tests only** - No unit tests to keep PoC simple. Complete end-to-end tests verify the functions work as expected with real Kafka infrastructure.

### Must Fix

- [ ] **Add e2e test script for Go SDK**
  - Uses `make kafka-up` to start infrastructure
  - Runs Go FFI example consumer in background
  - Sends test CloudEvent via `send-test-event.sh`
  - Verifies consumer received and processed event
  - Cleans up processes

- [ ] **Add e2e test script for Python SDK**
  - Same flow as Go e2e test
  - Runs Python FFI example in background
  - Sends test event and verifies processing

### Nice to Have

- [ ] **Add output routing e2e test**
  - Tests the output handler examples
  - Sends event, verifies output event appears on target topic

- [ ] **Add CI pipeline definition**
  - GitHub Actions or similar
  - Run e2e tests with docker-compose Kafka

### Proposed E2E Test Structure

```
tests/
├── e2e/
│   ├── test-go-ffi.sh           # Go FFI consumer e2e test
│   ├── test-go-ffi-output.sh    # Go FFI output handler e2e test
│   ├── test-python-ffi.sh       # Python FFI consumer e2e test
│   ├── test-python-ffi-output.sh # Python FFI output handler e2e test
│   └── run-all.sh               # Run all e2e tests
└── helpers/
    ├── wait-for-kafka.sh        # Wait for Kafka to be ready
    ├── verify-event-received.sh # Check consumer processed event
    └── verify-output-event.sh   # Check output event on target topic
```

### Sample E2E Test Script

```bash
#!/usr/bin/env bash
# tests/e2e/test-go-ffi.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="${SCRIPT_DIR}/../.."

echo "=== E2E Test: Go FFI Consumer ==="

# 1. Ensure Kafka is running
echo "Starting Kafka infrastructure..."
make -C "${ROOT_DIR}" kafka-up
sleep 5

# 2. Build the example
echo "Building Go FFI example..."
make -C "${ROOT_DIR}" example-go-ffi

# 3. Start consumer in background
echo "Starting consumer..."
cd "${ROOT_DIR}/sdks/go/examples/ffi-example"
timeout 30 ./ffi-consumer > /tmp/go-ffi-consumer.log 2>&1 &
CONSUMER_PID=$!
sleep 2

# 4. Send test event
echo "Sending test event..."
bash "${ROOT_DIR}/infra/scripts/send-test-event.sh"
sleep 2

# 5. Check consumer received event
echo "Checking consumer output..."
if grep -q "Received:" /tmp/go-ffi-consumer.log; then
    echo "✅ SUCCESS: Consumer received event"
else
    echo "❌ FAILED: Consumer did not receive event"
    cat /tmp/go-ffi-consumer.log
    kill $CONSUMER_PID 2>/dev/null || true
    exit 1
fi

# 6. Cleanup
kill $CONSUMER_PID 2>/dev/null || true
echo "=== Test completed successfully ==="
```

### Makefile Targets for E2E Tests

```makefile
.PHONY: test-e2e test-e2e-go test-e2e-python

test-e2e: test-e2e-go test-e2e-python  ## Run all e2e tests
    @echo "All e2e tests passed"

test-e2e-go: kafka-up example-go-ffi  ## Run Go e2e tests
    @bash tests/e2e/test-go-ffi.sh

test-e2e-python: kafka-up sdk-python-deps  ## Run Python e2e tests
    @bash tests/e2e/test-python-ffi.sh
```

---

## 6. Dependencies

### Nice to Have

- [ ] **Pin Python dependency versions for reproducibility**
  - Currently uses `>=` ranges in pyproject.toml
  - Consider creating `requirements.txt` or `poetry.lock` for exact pins

### Current Dependency Assessment

| Component | Status | Notes |
|-----------|--------|-------|
| Rust core | ✅ Good | Versions pinned in Cargo.toml |
| Go SDK | ⚠️ Check | go1.24.7 toolchain (future version?) |
| Python SDK | ✅ OK | Using ranges, acceptable for PoC |

---

## 7. Action Items Summary

### Priority 1: Must Complete Before Documentation

1. Verify and commit Python SDK files
2. Split Makefile into component Makefiles with consistent target naming
3. Fix event ID parsing bug in both SDKs
4. Add `auto_offset_reset` to core KafkaConfig for consistent consumer behavior
5. Fix Go README "cgo" -> "purego" inaccuracy
6. Expand root README.md with project overview
7. Add e2e test scripts for both SDKs

### Priority 2: Nice to Have

1. Add `make check` linting target
2. Remove unused dependencies
3. Add ARCHITECTURE.md
4. Pin Python dependencies
5. Add CI pipeline

### Priority 3: Document as Known Limitations

1. Retry/backoff logic is placeholder
2. Telemetry is placeholder (counter only)
3. WASM Component Model not fully implemented
4. CESQL routing not supported
5. HTTP/RabbitMQ publishing not implemented

---

## Appendix: Project TODO Comments

### Rust Core (Project Code)

| File | Line | Comment |
|------|------|---------|
| [`core/src/routing.rs`](core/src/routing.rs:9) | 9 | TODO: Add CESQL support |
| [`core/src/retry.rs`](core/src/retry.rs:22) | 22 | TODO: Implement error classification |
| [`core/src/retry.rs`](core/src/retry.rs:35) | 35 | TODO: Implement retry logic |
| [`core/src/retry.rs`](core/src/retry.rs:50) | 50 | TODO: Implement backoff calculation |
| [`core/src/telemetry.rs`](core/src/telemetry.rs:8) | 8 | TODO: Implement telemetry |
| [`core/src/telemetry.rs`](core/src/telemetry.rs:18) | 18 | TODO: Implement processing metrics |
| [`core/src/telemetry.rs`](core/src/telemetry.rs:28) | 28 | TODO: Implement retry metrics |

### Go SDK (Project Code)

| File | Line | Comment |
|------|------|---------|
| [`sdks/go/pkg/wasm/wasm.go`](sdks/go/pkg/wasm/wasm.go:70) | 70 | TODO: Call function and parse result |
| [`sdks/go/pkg/wasm/wasm.go`](sdks/go/pkg/wasm/wasm.go:101) | 101 | TODO: Component Model ABI |
| [`sdks/go/pkg/wasm/wasm.go`](sdks/go/pkg/wasm/wasm.go:127) | 127 | TODO: Component Model ABI |
| [`sdks/go/pkg/wasm/wasm.go`](sdks/go/pkg/wasm/wasm.go:150) | 150 | TODO: Component Model ABI |
| [`sdks/go/pkg/sdk/consumer.go`](sdks/go/pkg/sdk/consumer.go:275) | 275 | TODO: HTTP/RabbitMQ publishing |

### Python SDK (Project Code)

| File | Line | Comment |
|------|------|---------|
| [`sdks/python/eda_sdk/consumer.py`](sdks/python/eda_sdk/consumer.py:244) | 244 | TODO: HTTP/RabbitMQ publishing |
