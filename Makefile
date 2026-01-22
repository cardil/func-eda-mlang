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

# Local tool executables
CBINDGEN := $(TOOLS_BIN)/cbindgen
WASM_PACK := $(TOOLS_BIN)/wasm-pack

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

.PHONY: core-build-ffi
core-build-ffi: $(GUARDS)/core-ffi.done  ## Build Rust core as shared library
	@echo -e "$(WRENCH) FFI library ......................... $(GREEN)BUILT$(RESET)"

$(GUARDS)/core-ffi.done: $(CORE_SRC_FILES) $(CORE_CARGO)
	@mkdir -p $(GUARDS)
	@echo -e "$(BLUE)$(GEAR) Building Rust core as shared library...$(RESET)"
	cd $(CORE_DIR) && cargo build --release --lib
	@echo -e "$(GREEN)$(CHECK) FFI library: $(CORE_DIR)/target/release/libeda_core.$(LIB_EXT)$(RESET)"
	@touch $@

.PHONY: core-build-wasm
core-build-wasm: $(GUARDS)/core-wasm.done  ## Build Rust core as WASM module
	@echo -e "$(WRENCH) WASM module ......................... $(GREEN)BUILT$(RESET)"

# Install wasm-pack to local .tools directory
$(WASM_PACK):
	@echo -e "$(BLUE)$(GEAR) Installing wasm-pack to $(TOOLS_DIR)...$(RESET)"
	@mkdir -p $(TOOLS_DIR)
	cargo install --locked --version 0.13.1 --root $(TOOLS_DIR) wasm-pack
	@echo -e "$(GREEN)$(CHECK) wasm-pack installed$(RESET)"

$(BINDINGS_WASM_DIR)/Cargo.toml:
	@echo -e "$(YELLOW)WASM bindings not yet configured$(RESET)"
	@exit 1

$(GUARDS)/core-wasm.done: $(CORE_SRC_FILES) $(BINDINGS_WASM_DIR)/Cargo.toml $(WASM_PACK)
	@mkdir -p $(GUARDS)
	@echo -e "$(BLUE)$(GEAR) Building Rust core as WASM module...$(RESET)"
	cd $(BINDINGS_WASM_DIR) && $(CURDIR)/$(WASM_PACK) build --target web --release
	@echo -e "$(GREEN)$(CHECK) WASM module: $(BINDINGS_WASM_DIR)/pkg/$(RESET)"
	@touch $@

.PHONY: core-headers
core-headers: $(GUARDS)/core-headers.done  ## Generate C headers with cbindgen
	@echo -e "$(WRENCH) C headers ........................... $(GREEN)GENERATED$(RESET)"

# Install cbindgen to local .tools directory
$(CBINDGEN):
	@echo -e "$(BLUE)$(GEAR) Installing cbindgen to $(TOOLS_DIR)...$(RESET)"
	@mkdir -p $(TOOLS_DIR)
	cargo install --locked --version 0.27.0 --root $(TOOLS_DIR) cbindgen
	@echo -e "$(GREEN)$(CHECK) cbindgen installed$(RESET)"

$(GUARDS)/core-headers.done: $(GUARDS)/core-ffi.done $(CBINDGEN)
	@mkdir -p $(GUARDS)
	@mkdir -p $(BINDINGS_FFI_DIR)/include
	@echo -e "$(BLUE)$(GEAR) Generating C headers with cbindgen...$(RESET)"
	$(CBINDGEN) --config $(BINDINGS_FFI_DIR)/cbindgen.toml --crate eda-core \
		--output $(BINDINGS_FFI_DIR)/include/eda_core.h $(CORE_DIR)
	@echo -e "$(GREEN)$(CHECK) Headers: $(BINDINGS_FFI_DIR)/include/eda_core.h$(RESET)"
	@touch $@

##@ Build All

.PHONY: all
all: core-build-ffi core-build-wasm core-headers  ## Build everything
	@echo -e "$(GREEN)$(ROCKET) All builds complete$(RESET)"

##@ Development

.PHONY: clean
clean:  ## Clean build artifacts and caches
	@echo -e "$(YELLOW)$(CLEAN) Cleaning build artifacts...$(RESET)"
	@rm -rf $(GUARDS)/
	@cd $(CORE_DIR) && cargo clean
	@rm -rf $(BINDINGS_WASM_DIR)/pkg
	@rm -rf $(BINDINGS_FFI_DIR)/include
	@echo -e "$(GREEN)$(CHECK) Build artifacts cleaned$(RESET)"

.PHONY: distclean
distclean: clean  ## Complete cleanup including tools and Cargo cache
	@echo -e "$(YELLOW)$(CLEAN) Removing tools and Cargo target directory...$(RESET)"
	@rm -rf $(TOOLS_DIR)
	@rm -rf $(CORE_DIR)/target
	@echo -e "$(GREEN)$(CHECK) Complete cleanup done$(RESET)"
