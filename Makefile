# Makefile for func-eda-mlang
# Using guard files for idempotent operations

# Colors
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
CYAN := \033[0;36m
RESET := \033[0m

# Emojis
CHECK := ✅
CROSS := ❌
ROCKET := 🚀
GEAR := "⚙️ "
CLEAN := 🧹
BOOK := 📚
WRENCH := 🔧

# Detect container engine (podman or docker)
CONTAINER_ENGINE := $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null)
ifeq ($(CONTAINER_ENGINE),)
$(error Neither podman nor docker found. Please install one of them.)
endif

.PHONY: help
help:  ## Show this help
	@echo -e "$(CYAN)$(BOOK) Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;36m%-20s\033[0m %s\n", $$1, $$2}'

# Directories
GUARDS := .make
TOOLS_DIR := .tools
TOOLS_BIN := $(TOOLS_DIR)/bin
CORE_DIR := core
BINDINGS_FFI_DIR := bindings/ffi
BINDINGS_WASM_DIR := bindings/wasm
SDK_GO_DIR := sdks/go

# Local tool executables
CBINDGEN := $(TOOLS_BIN)/cbindgen

# Detect OS for proper library extension
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	LIB_EXT := so
endif
ifeq ($(UNAME_S),Darwin)
	LIB_EXT := dylib
endif

# Source tracking
CORE_SRC_FILES := $(shell find $(CORE_DIR)/src -type f -name '*.rs' 2>/dev/null)
CORE_CARGO := $(CORE_DIR)/Cargo.toml

##@ Core Builds

# FFI library artifact
FFI_LIB := $(CORE_DIR)/target/release/libeda_core.$(LIB_EXT)

.PHONY: core-build-ffi
core-build-ffi: $(FFI_LIB)  ## Build Rust core as shared library

$(FFI_LIB): $(CORE_SRC_FILES) $(CORE_CARGO)
	@echo -e "$(BLUE)$(GEAR) Building Rust core as shared library...$(RESET)"
	cd $(CORE_DIR) && cargo build --release --lib
	@echo -e "$(GREEN)$(CHECK) FFI library: $@$(RESET)"

# Cross-compilation targets for FFI libraries
.PHONY: core-build-ffi-linux-amd64
core-build-ffi-linux-amd64:  ## Build FFI library for Linux x86_64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for Linux x86_64...$(RESET)"
	cd $(CORE_DIR) && cargo build --release --target x86_64-unknown-linux-gnu --lib
	@echo -e "$(GREEN)$(CHECK) Linux x86_64 FFI library built$(RESET)"

.PHONY: core-build-ffi-linux-arm64
core-build-ffi-linux-arm64:  ## Build FFI library for Linux ARM64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for Linux ARM64...$(RESET)"
	cd $(CORE_DIR) && cargo build --release --target aarch64-unknown-linux-gnu --lib
	@echo -e "$(GREEN)$(CHECK) Linux ARM64 FFI library built$(RESET)"

.PHONY: core-build-ffi-darwin-amd64
core-build-ffi-darwin-amd64:  ## Build FFI library for macOS x86_64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for macOS x86_64...$(RESET)"
	cd $(CORE_DIR) && cargo build --release --target x86_64-apple-darwin --lib
	@echo -e "$(GREEN)$(CHECK) macOS x86_64 FFI library built$(RESET)"

.PHONY: core-build-ffi-darwin-arm64
core-build-ffi-darwin-arm64:  ## Build FFI library for macOS ARM64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for macOS ARM64...$(RESET)"
	cd $(CORE_DIR) && cargo build --release --target aarch64-apple-darwin --lib
	@echo -e "$(GREEN)$(CHECK) macOS ARM64 FFI library built$(RESET)"

.PHONY: core-build-ffi-windows-amd64
core-build-ffi-windows-amd64:  ## Build FFI library for Windows x86_64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for Windows x86_64...$(RESET)"
	cd $(CORE_DIR) && cargo build --release --target x86_64-pc-windows-gnu --lib
	@echo -e "$(GREEN)$(CHECK) Windows x86_64 FFI library built$(RESET)"

.PHONY: core-build-ffi-all
core-build-ffi-all: core-build-ffi-linux-amd64 core-build-ffi-linux-arm64 core-build-ffi-darwin-amd64 core-build-ffi-darwin-arm64 core-build-ffi-windows-amd64  ## Build FFI libraries for all platforms
	@echo -e "$(GREEN)$(ROCKET) All platform FFI libraries built$(RESET)"

# WASM Component artifact
WASM_COMPONENT := $(BINDINGS_WASM_DIR)/target/wasm32-wasip2/release/eda_wasm.wasm

.PHONY: core-build-component
core-build-component: $(WASM_COMPONENT)  ## Build Rust core as WASM Component

$(WASM_COMPONENT): $(CORE_SRC_FILES) $(BINDINGS_WASM_DIR)/Cargo.toml
	@echo -e "$(BLUE)$(GEAR) Building Rust core as WASM Component (wasm32-wasip2)...$(RESET)"
	cd $(BINDINGS_WASM_DIR) && cargo build --release --target wasm32-wasip2
	@echo -e "$(GREEN)$(CHECK) WASM Component: $@$(RESET)"

.PHONY: core-build-wasm
core-build-wasm: $(WASM_COMPONENT)  ## Build Rust core as WASM Component (wasm32-wasip2)
	@echo -e "$(WRENCH) WASM Component ...................... $(GREEN)BUILT$(RESET)"

# C headers artifact
C_HEADER := $(BINDINGS_FFI_DIR)/include/eda_core.h

.PHONY: core-headers
core-headers: $(C_HEADER)  ## Generate C headers with cbindgen

# Install cbindgen to local .tools directory
$(CBINDGEN):
	@echo -e "$(BLUE)$(GEAR) Installing cbindgen to $(TOOLS_DIR)...$(RESET)"
	@mkdir -p $(TOOLS_DIR)
	cargo install --locked --version 0.27.0 --root $(TOOLS_DIR) cbindgen
	@echo -e "$(GREEN)$(CHECK) cbindgen installed$(RESET)"

$(C_HEADER): $(FFI_LIB) $(CBINDGEN)
	@mkdir -p $(BINDINGS_FFI_DIR)/include
	@echo -e "$(BLUE)$(GEAR) Generating C headers with cbindgen...$(RESET)"
	$(CBINDGEN) --config $(BINDINGS_FFI_DIR)/cbindgen.toml --crate eda-core \
		--output $@ $(CORE_DIR)
	@echo -e "$(GREEN)$(CHECK) Headers: $@$(RESET)"

##@ Go SDK

.PHONY: sdk-go-embed-libs
sdk-go-embed-libs: core-build-ffi  ## Copy FFI libraries to Go SDK for embedding (current platform)
	@echo -e "$(BLUE)$(GEAR) Copying FFI library to Go SDK...$(RESET)"
	@mkdir -p $(SDK_GO_DIR)/pkg/ffi/libs/linux_amd64
	@mkdir -p $(SDK_GO_DIR)/pkg/ffi/libs/linux_arm64
	@mkdir -p $(SDK_GO_DIR)/pkg/ffi/libs/darwin_amd64
	@mkdir -p $(SDK_GO_DIR)/pkg/ffi/libs/darwin_arm64
	@mkdir -p $(SDK_GO_DIR)/pkg/ffi/libs/windows_amd64
	@if [ -f $(CORE_DIR)/target/release/libeda_core.so ]; then \
		cp $(CORE_DIR)/target/release/libeda_core.so $(SDK_GO_DIR)/pkg/ffi/libs/linux_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied libeda_core.so to linux_amd64$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/release/libeda_core.dylib ]; then \
		cp $(CORE_DIR)/target/release/libeda_core.dylib $(SDK_GO_DIR)/pkg/ffi/libs/darwin_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied libeda_core.dylib to darwin_amd64$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/release/eda_core.dll ]; then \
		cp $(CORE_DIR)/target/release/eda_core.dll $(SDK_GO_DIR)/pkg/ffi/libs/windows_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied eda_core.dll to windows_amd64$(RESET)"; \
	fi

.PHONY: sdk-go-embed-libs-all
sdk-go-embed-libs-all:  ## Copy all cross-compiled FFI libraries to Go SDK
	@echo -e "$(BLUE)$(GEAR) Copying all FFI libraries to Go SDK...$(RESET)"
	@mkdir -p $(SDK_GO_DIR)/pkg/ffi/libs/{linux_amd64,linux_arm64,darwin_amd64,darwin_arm64,windows_amd64}
	@if [ -f $(CORE_DIR)/target/x86_64-unknown-linux-gnu/release/libeda_core.so ]; then \
		cp $(CORE_DIR)/target/x86_64-unknown-linux-gnu/release/libeda_core.so $(SDK_GO_DIR)/pkg/ffi/libs/linux_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied Linux x86_64 library$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/aarch64-unknown-linux-gnu/release/libeda_core.so ]; then \
		cp $(CORE_DIR)/target/aarch64-unknown-linux-gnu/release/libeda_core.so $(SDK_GO_DIR)/pkg/ffi/libs/linux_arm64/; \
		echo -e "$(GREEN)$(CHECK) Copied Linux ARM64 library$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/x86_64-apple-darwin/release/libeda_core.dylib ]; then \
		cp $(CORE_DIR)/target/x86_64-apple-darwin/release/libeda_core.dylib $(SDK_GO_DIR)/pkg/ffi/libs/darwin_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied macOS x86_64 library$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/aarch64-apple-darwin/release/libeda_core.dylib ]; then \
		cp $(CORE_DIR)/target/aarch64-apple-darwin/release/libeda_core.dylib $(SDK_GO_DIR)/pkg/ffi/libs/darwin_arm64/; \
		echo -e "$(GREEN)$(CHECK) Copied macOS ARM64 library$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/x86_64-pc-windows-gnu/release/eda_core.dll ]; then \
		cp $(CORE_DIR)/target/x86_64-pc-windows-gnu/release/eda_core.dll $(SDK_GO_DIR)/pkg/ffi/libs/windows_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied Windows x86_64 library$(RESET)"; \
	fi
	@echo -e "$(GREEN)$(ROCKET) All libraries copied to Go SDK$(RESET)"

.PHONY: sdk-go-deps
sdk-go-deps:  ## Download Go SDK dependencies
	@echo -e "$(BLUE)$(GEAR) Downloading Go dependencies...$(RESET)"
	cd $(SDK_GO_DIR) && go mod download
	@echo -e "$(GREEN)$(CHECK) Go dependencies downloaded$(RESET)"

.PHONY: sdk-go-generate
sdk-go-generate: core-build-wasm  ## Generate Go SDK embedded assets
	@echo -e "$(BLUE)$(GEAR) Generating Go SDK assets...$(RESET)"
	cd $(SDK_GO_DIR) && go generate ./...
	@echo -e "$(GREEN)$(CHECK) Go SDK assets generated$(RESET)"

.PHONY: sdk-go-build
sdk-go-build: sdk-go-embed-libs core-build-wasm sdk-go-deps sdk-go-generate  ## Build Go SDK
	@echo -e "$(BLUE)$(GEAR) Building Go SDK...$(RESET)"
	cd $(SDK_GO_DIR) && CGO_ENABLED=1 go build ./...
	@echo -e "$(GREEN)$(CHECK) Go SDK built (FFI uses embedded libs via purego)$(RESET)"

.PHONY: sdk-go-test
sdk-go-test: sdk-go-build  ## Test Go SDK
	@echo -e "$(BLUE)$(GEAR) Testing Go SDK...$(RESET)"
	cd $(SDK_GO_DIR) && CGO_ENABLED=1 go test -v ./...
	@echo -e "$(GREEN)$(CHECK) Go SDK tests passed$(RESET)"

# Build Go examples
.PHONY: example-go-ffi
example-go-ffi: sdk-go-build  ## Build Go FFI example
	@echo -e "$(BLUE)$(GEAR) Building Go FFI example...$(RESET)"
	cd $(SDK_GO_DIR)/examples/ffi-example && CGO_ENABLED=1 go build -o ffi-consumer .
	@echo -e "$(GREEN)$(CHECK) Built: $(SDK_GO_DIR)/examples/ffi-example/ffi-consumer$(RESET)"

.PHONY: example-go-ffi-output
example-go-ffi-output: sdk-go-build  ## Build Go FFI output example
	@echo -e "$(BLUE)$(GEAR) Building Go FFI output example...$(RESET)"
	cd $(SDK_GO_DIR)/examples/ffi-output-example && CGO_ENABLED=1 go build -o ffi-output-consumer .
	@echo -e "$(GREEN)$(CHECK) Built: $(SDK_GO_DIR)/examples/ffi-output-example/ffi-output-consumer$(RESET)"

.PHONY: example-go-wasm
example-go-wasm: sdk-go-build  ## Build Go WASM example
	@echo -e "$(BLUE)$(GEAR) Building Go WASM example...$(RESET)"
	cd $(SDK_GO_DIR)/examples/wasm-example && go build -o wasm-consumer .
	@echo -e "$(GREEN)$(CHECK) Built: $(SDK_GO_DIR)/examples/wasm-example/wasm-consumer$(RESET)"

.PHONY: run-go-ffi
run-go-ffi: example-go-ffi  ## Run Go FFI example
	@echo -e "$(BLUE)$(ROCKET) Running Go FFI example...$(RESET)"
	cd $(SDK_GO_DIR)/examples/ffi-example && timeout 30 ./ffi-consumer

.PHONY: run-go-ffi-output
run-go-ffi-output: example-go-ffi-output  ## Run Go FFI output example
	@echo -e "$(BLUE)$(ROCKET) Running Go FFI output example...$(RESET)"
	cd $(SDK_GO_DIR)/examples/ffi-output-example && timeout 30 ./ffi-output-consumer

.PHONY: run-go-wasm
run-go-wasm: example-go-wasm  ## Run Go WASM example
	@echo -e "$(BLUE)$(ROCKET) Running Go WASM example...$(RESET)"
	cd $(SDK_GO_DIR)/examples/wasm-example && \
		WASM_PATH=$(CURDIR)/$(BINDINGS_WASM_DIR)/target/wasm32-unknown-unknown/release/eda_wasm.wasm \
		timeout 30 ./wasm-consumer

##@ Build All

.PHONY: all
all: core-build-ffi core-build-wasm core-headers sdk-go-build examples  ## Build everything
	@echo -e "$(GREEN)$(ROCKET) All builds complete$(RESET)"

.PHONY: examples
examples: example-go-ffi example-go-wasm  ## Build all examples
	@echo -e "$(GREEN)$(ROCKET) All examples built$(RESET)"

##@ Development

.PHONY: clean
clean:  ## Clean build artifacts and caches
	@echo -e "$(YELLOW)$(CLEAN) Cleaning build artifacts...$(RESET)"
	@rm -rf $(GUARDS)/
	@cd $(CORE_DIR) && cargo clean
	@rm -rf $(BINDINGS_WASM_DIR)/target
	@rm -rf $(BINDINGS_FFI_DIR)/include
	@cd $(SDK_GO_DIR) && go clean
	@rm -f $(SDK_GO_DIR)/examples/ffi-example/ffi-consumer
	@rm -f $(SDK_GO_DIR)/examples/wasm-example/wasm-consumer
	@rm -f $(SDK_GO_DIR)/pkg/wasm/*.wasm
	@rm -rf $(SDK_GO_DIR)/pkg/wasm/gen
	@rm -rf $(SDK_GO_DIR)/pkg/ffi/libs
	@echo -e "$(GREEN)$(CHECK) Build artifacts cleaned$(RESET)"

.PHONY: distclean
distclean: clean  ## Complete cleanup including tools and Cargo cache
	@echo -e "$(YELLOW)$(CLEAN) Removing tools and Cargo target directory...$(RESET)"
	@rm -rf $(TOOLS_DIR)
	@rm -rf $(CORE_DIR)/target
	@cd $(SDK_GO_DIR) && go clean -modcache
	@echo -e "$(GREEN)$(CHECK) Complete cleanup done$(RESET)"

##@ Kafka Infrastructure

.PHONY: kafka-up
kafka-up:  ## Start Kafka/Redpanda infrastructure
	@echo -e "$(BLUE)$(ROCKET) Starting Kafka/Redpanda...$(RESET)"
	cd infra && $(CONTAINER_ENGINE) compose up -d
	@echo -e "$(GREEN)$(CHECK) Kafka/Redpanda started$(RESET)"
	@echo -e "$(CYAN)Waiting for Redpanda to be healthy...$(RESET)"
	@sleep 5
	@echo -e "$(CYAN)Creating topics...$(RESET)"
	@$(CONTAINER_ENGINE) exec redpanda rpk topic create events --partitions 1 --replicas 1 2>/dev/null || echo -e "$(YELLOW)Topic 'events' may already exist$(RESET)"
	@$(CONTAINER_ENGINE) exec redpanda rpk topic create processed-events --partitions 1 --replicas 1 2>/dev/null || echo -e "$(YELLOW)Topic 'processed-events' may already exist$(RESET)"
	@echo -e "$(GREEN)$(CHECK) Redpanda Console available at http://localhost:8080$(RESET)"

.PHONY: kafka-down
kafka-down:  ## Stop Kafka/Redpanda infrastructure
	@echo -e "$(YELLOW)Stopping Kafka/Redpanda...$(RESET)"
	cd infra && $(CONTAINER_ENGINE) compose down
	@echo -e "$(GREEN)$(CHECK) Kafka/Redpanda stopped$(RESET)"

.PHONY: kafka-logs
kafka-logs:  ## View Kafka/Redpanda logs
	@echo -e "$(CYAN)Showing Kafka/Redpanda logs...$(RESET)"
	cd infra && $(CONTAINER_ENGINE) compose logs -f redpanda

.PHONY: kafka-topics
kafka-topics:  ## List Kafka topics
	@echo -e "$(CYAN)Listing Kafka topics...$(RESET)"
	$(CONTAINER_ENGINE) exec redpanda rpk topic list

.PHONY: test-send-event
test-send-event:  ## Send a test CloudEvent to Kafka
	@echo -e "$(BLUE)$(ROCKET) Sending test CloudEvent...$(RESET)"
	@bash infra/scripts/send-test-event.sh
	@echo -e "$(GREEN)$(CHECK) Test event sent$(RESET)"
