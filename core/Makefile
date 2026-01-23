# Core Rust library Makefile

include ../common.mk

# Directories
BINDINGS_FFI_DIR := ../bindings/ffi
BINDINGS_WASM_DIR := ../bindings/wasm
TOOLS_DIR := ../.tools
TOOLS_BIN := $(TOOLS_DIR)/bin
CBINDGEN := $(TOOLS_BIN)/cbindgen

# Cross-compilation sysroot (auto-detect Fedora version)
AARCH64_SYSROOT := $(shell ls -d /usr/aarch64-redhat-linux/sys-root/fc* 2>/dev/null | head -1)

# Artifacts
FFI_LIB := target/release/libeda_core.$(LIB_EXT)
WASM_COMPONENT := $(BINDINGS_WASM_DIR)/target/wasm32-wasip2/release/eda_wasm.wasm
C_HEADER := $(BINDINGS_FFI_DIR)/include/eda_core.h

# Source tracking
SRC_FILES := $(shell find src -type f -name '*.rs' 2>/dev/null)

.PHONY: all build build-ffi build-wasm headers test clean check

all: build-ffi-linux-amd64 build-ffi-windows-amd64 build-wasm headers  ## Build everything (Linux x86_64 + Windows + WASM + headers)

build: build-ffi build-wasm  ## Build for current platform (FFI + WASM)

build-ffi: $(FFI_LIB) headers  ## Build Rust core as shared library with headers

$(FFI_LIB): $(SRC_FILES) Cargo.toml
	@echo -e "$(BLUE)$(GEAR) Building Rust core as shared library...$(RESET)"
	cargo build --release --lib
	@echo -e "$(GREEN)$(CHECK) FFI library: $@$(RESET)"

build-wasm: $(WASM_COMPONENT)  ## Build Rust core as WASM Component

$(WASM_COMPONENT): $(SRC_FILES) $(BINDINGS_WASM_DIR)/Cargo.toml
	@echo -e "$(BLUE)$(GEAR) Building Rust core as WASM Component (wasm32-wasip2)...$(RESET)"
	cd $(BINDINGS_WASM_DIR) && cargo build --release --target wasm32-wasip2
	@echo -e "$(GREEN)$(CHECK) WASM Component: $@$(RESET)"

# Install cbindgen to local .tools directory
$(CBINDGEN):
	@echo -e "$(BLUE)$(GEAR) Installing cbindgen to $(TOOLS_DIR)...$(RESET)"
	@mkdir -p $(TOOLS_DIR)
	cargo install --locked --version 0.27.0 --root $(TOOLS_DIR) cbindgen
	@echo -e "$(GREEN)$(CHECK) cbindgen installed$(RESET)"

headers: $(C_HEADER)  ## Generate C headers with cbindgen

$(C_HEADER): $(FFI_LIB) $(CBINDGEN)
	@mkdir -p $(BINDINGS_FFI_DIR)/include
	@echo -e "$(BLUE)$(GEAR) Generating C headers with cbindgen...$(RESET)"
	$(CBINDGEN) --config $(BINDINGS_FFI_DIR)/cbindgen.toml --crate eda-core \
		--output $@ .
	@echo -e "$(GREEN)$(CHECK) Headers: $@$(RESET)"

test:  ## Run tests
	@echo -e "$(BLUE)$(GEAR) Testing Rust core...$(RESET)"
	cargo test -- --test-threads=1
	@echo -e "$(GREEN)$(CHECK) Tests passed$(RESET)"

check:  ## Run linters and formatters
	@echo -e "$(BLUE)$(GEAR) Running cargo fmt...$(RESET)"
	cargo fmt --check
	@echo -e "$(BLUE)$(GEAR) Running cargo clippy...$(RESET)"
	cargo clippy -- -D warnings
	@echo -e "$(GREEN)$(CHECK) Checks passed$(RESET)"

clean:  ## Clean build artifacts
	@echo -e "$(YELLOW)Cleaning Rust core artifacts...$(RESET)"
	cargo clean
	rm -rf $(BINDINGS_WASM_DIR)/target
	rm -rf $(BINDINGS_FFI_DIR)/include
	@echo -e "$(GREEN)$(CHECK) Core cleaned$(RESET)"

# Cross-compilation targets
.PHONY: build-ffi-linux-amd64 build-ffi-linux-arm64 build-ffi-darwin-amd64 build-ffi-darwin-arm64 build-ffi-windows-amd64 build-ffi-all

build-ffi-linux-amd64:  ## Build FFI library for Linux x86_64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for Linux x86_64...$(RESET)"
	cargo build --release --target x86_64-unknown-linux-gnu --lib
	@echo -e "$(GREEN)$(CHECK) Linux x86_64 FFI library built$(RESET)"

build-ffi-linux-arm64:  ## Build FFI library for Linux ARM64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for Linux ARM64...$(RESET)"
	@if [ -z "$(AARCH64_SYSROOT)" ]; then \
		echo -e "$(RED)$(CROSS) ARM64 sysroot not found. Install with: sudo dnf install sysroot-aarch64-fc*-glibc$(RESET)"; \
		exit 1; \
	fi
	CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_RUSTFLAGS="-C link-arg=--sysroot=$(AARCH64_SYSROOT) -C link-arg=-static-libgcc" \
		cargo build --release --target aarch64-unknown-linux-gnu --lib
	@echo -e "$(GREEN)$(CHECK) Linux ARM64 FFI library built$(RESET)"

build-ffi-darwin-amd64:  ## Build FFI library for macOS x86_64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for macOS x86_64...$(RESET)"
	cargo build --release --target x86_64-apple-darwin --lib
	@echo -e "$(GREEN)$(CHECK) macOS x86_64 FFI library built$(RESET)"

build-ffi-darwin-arm64:  ## Build FFI library for macOS ARM64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for macOS ARM64...$(RESET)"
	cargo build --release --target aarch64-apple-darwin --lib
	@echo -e "$(GREEN)$(CHECK) macOS ARM64 FFI library built$(RESET)"

build-ffi-windows-amd64:  ## Build FFI library for Windows x86_64
	@echo -e "$(BLUE)$(GEAR) Building FFI library for Windows x86_64...$(RESET)"
	cargo build --release --target x86_64-pc-windows-gnu --lib
	@echo -e "$(GREEN)$(CHECK) Windows x86_64 FFI library built$(RESET)"

build-ffi-all: build-ffi-linux-amd64 build-ffi-linux-arm64 build-ffi-windows-amd64  ## Build FFI libraries for all supported platforms (Linux + Windows)
	@echo -e "$(GREEN)$(CHECK) All platform FFI libraries built$(RESET)"

build-ffi-all-with-macos: build-ffi-all build-ffi-darwin-amd64 build-ffi-darwin-arm64  ## Build FFI libraries including macOS (requires OSXCross)
	@echo -e "$(GREEN)$(CHECK) All platform FFI libraries built (including macOS)$(RESET)"
