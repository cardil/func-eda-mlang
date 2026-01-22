# Go SDK Makefile

include ../../common.mk

# Directories
CORE_DIR := ../../core
BINDINGS_WASM_DIR := ../../bindings/wasm
EXAMPLES_DIR := examples

# Artifacts
WASM_COMPONENT := $(BINDINGS_WASM_DIR)/target/wasm32-wasip2/release/eda_wasm.wasm

.PHONY: all build test clean check embed-libs

all: build  ## Build everything

build: embed-libs generate  ## Build Go SDK
	@echo -e "$(BLUE)$(GEAR) Building Go SDK...$(RESET)"
	CGO_ENABLED=1 go build ./...
	@echo -e "$(GREEN)$(CHECK) Go SDK built$(RESET)"

embed-libs:  ## Copy FFI libraries to Go SDK for embedding (current platform)
	@echo -e "$(BLUE)$(GEAR) Copying FFI library to Go SDK...$(RESET)"
	@mkdir -p pkg/ffi/libs/linux_amd64
	@mkdir -p pkg/ffi/libs/linux_arm64
	@mkdir -p pkg/ffi/libs/darwin_amd64
	@mkdir -p pkg/ffi/libs/darwin_arm64
	@mkdir -p pkg/ffi/libs/windows_amd64
	@if [ -f $(CORE_DIR)/target/release/libeda_core.so ]; then \
		cp $(CORE_DIR)/target/release/libeda_core.so pkg/ffi/libs/linux_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied libeda_core.so to linux_amd64$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/release/libeda_core.dylib ]; then \
		cp $(CORE_DIR)/target/release/libeda_core.dylib pkg/ffi/libs/darwin_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied libeda_core.dylib to darwin_amd64$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/release/eda_core.dll ]; then \
		cp $(CORE_DIR)/target/release/eda_core.dll pkg/ffi/libs/windows_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied eda_core.dll to windows_amd64$(RESET)"; \
	fi

embed-libs-all:  ## Copy all cross-compiled FFI libraries to Go SDK
	@echo -e "$(BLUE)$(GEAR) Copying all FFI libraries to Go SDK...$(RESET)"
	@mkdir -p pkg/ffi/libs/{linux_amd64,linux_arm64,darwin_amd64,darwin_arm64,windows_amd64}
	@if [ -f $(CORE_DIR)/target/x86_64-unknown-linux-gnu/release/libeda_core.so ]; then \
		cp $(CORE_DIR)/target/x86_64-unknown-linux-gnu/release/libeda_core.so pkg/ffi/libs/linux_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied Linux x86_64 library$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/aarch64-unknown-linux-gnu/release/libeda_core.so ]; then \
		cp $(CORE_DIR)/target/aarch64-unknown-linux-gnu/release/libeda_core.so pkg/ffi/libs/linux_arm64/; \
		echo -e "$(GREEN)$(CHECK) Copied Linux ARM64 library$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/x86_64-apple-darwin/release/libeda_core.dylib ]; then \
		cp $(CORE_DIR)/target/x86_64-apple-darwin/release/libeda_core.dylib pkg/ffi/libs/darwin_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied macOS x86_64 library$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/aarch64-apple-darwin/release/libeda_core.dylib ]; then \
		cp $(CORE_DIR)/target/aarch64-apple-darwin/release/libeda_core.dylib pkg/ffi/libs/darwin_arm64/; \
		echo -e "$(GREEN)$(CHECK) Copied macOS ARM64 library$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/x86_64-pc-windows-gnu/release/eda_core.dll ]; then \
		cp $(CORE_DIR)/target/x86_64-pc-windows-gnu/release/eda_core.dll pkg/ffi/libs/windows_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied Windows x86_64 library$(RESET)"; \
	fi
	@echo -e "$(GREEN)$(ROCKET) All libraries copied to Go SDK$(RESET)"

generate: $(WASM_COMPONENT)  ## Generate Go SDK embedded assets
	@echo -e "$(BLUE)$(GEAR) Generating Go SDK assets...$(RESET)"
	go generate ./...
	@echo -e "$(GREEN)$(CHECK) Go SDK assets generated$(RESET)"

test: build  ## Run tests
	@echo -e "$(BLUE)$(GEAR) Testing Go SDK...$(RESET)"
	CGO_ENABLED=1 go test -v ./...
	@echo -e "$(GREEN)$(CHECK) Go SDK tests passed$(RESET)"

check:  ## Run linters and formatters
	@echo -e "$(BLUE)$(GEAR) Running go fmt...$(RESET)"
	go fmt ./...
	@echo -e "$(BLUE)$(GEAR) Running go vet...$(RESET)"
	go vet ./...
	@echo -e "$(GREEN)$(CHECK) Checks passed$(RESET)"

clean:  ## Clean build artifacts
	@echo -e "$(YELLOW)Cleaning Go SDK artifacts...$(RESET)"
	go clean
	rm -f $(EXAMPLES_DIR)/ffi-example/ffi-consumer
	rm -f $(EXAMPLES_DIR)/ffi-output-example/ffi-output-consumer
	rm -f $(EXAMPLES_DIR)/wasm-example/wasm-consumer
	rm -f pkg/wasm/*.wasm
	rm -rf pkg/wasm/gen
	rm -rf pkg/ffi/libs
	@echo -e "$(GREEN)$(CHECK) Go SDK cleaned$(RESET)"
