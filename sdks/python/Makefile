# Python SDK Makefile

include ../../common.mk

# Directories
CORE_DIR := ../../core
BINDINGS_FFI_DIR := ../../bindings/ffi
VENV := .venv
PYTHON := $(CURDIR)/$(VENV)/bin/python
PIP := $(CURDIR)/$(VENV)/bin/pip

# Artifacts
C_HEADER := $(BINDINGS_FFI_DIR)/include/eda_core.h
EMBEDDED_HEADER := eda_sdk/ffi/include/eda_core.h
EMBEDDED_LIBS := $(wildcard eda_sdk/ffi/libs/*/*.so eda_sdk/ffi/libs/*/*.dylib eda_sdk/ffi/libs/*/*.dll)

.PHONY: all build test clean check venv embed-libs install

all: build  ## Build everything

venv: $(VENV)  ## Create Python virtual environment

$(VENV):
	@echo -e "$(BLUE)$(GEAR) Creating Python virtual environment...$(RESET)"
	python3 -m venv $(VENV)
	@echo -e "$(GREEN)$(CHECK) Virtual environment created$(RESET)"

embed-libs: $(EMBEDDED_HEADER)  ## Copy FFI libraries and headers to Python SDK (current platform)

$(EMBEDDED_HEADER): $(C_HEADER)
	@echo -e "$(BLUE)$(GEAR) Copying FFI library and headers to Python SDK...$(RESET)"
	@mkdir -p eda_sdk/ffi/libs/linux_amd64
	@mkdir -p eda_sdk/ffi/libs/linux_arm64
	@mkdir -p eda_sdk/ffi/libs/darwin_amd64
	@mkdir -p eda_sdk/ffi/libs/darwin_arm64
	@mkdir -p eda_sdk/ffi/libs/windows_amd64
	@mkdir -p eda_sdk/ffi/include
	@if [ -f $(CORE_DIR)/target/release/libeda_core.so ]; then \
		cp $(CORE_DIR)/target/release/libeda_core.so eda_sdk/ffi/libs/linux_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied libeda_core.so to linux_amd64$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/release/libeda_core.dylib ]; then \
		cp $(CORE_DIR)/target/release/libeda_core.dylib eda_sdk/ffi/libs/darwin_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied libeda_core.dylib to darwin_amd64$(RESET)"; \
	fi
	@if [ -f $(CORE_DIR)/target/release/eda_core.dll ]; then \
		cp $(CORE_DIR)/target/release/eda_core.dll eda_sdk/ffi/libs/windows_amd64/; \
		echo -e "$(GREEN)$(CHECK) Copied eda_core.dll to windows_amd64$(RESET)"; \
	fi
	@cp $(C_HEADER) eda_sdk/ffi/include/
	@echo -e "$(GREEN)$(CHECK) Copied C header to Python SDK$(RESET)"

install: $(VENV) embed-libs  ## Install Python SDK in development mode
	@echo -e "$(BLUE)$(GEAR) Installing Python SDK in development mode...$(RESET)"
	$(PIP) install -e ".[dev]"
	@echo -e "$(GREEN)$(CHECK) Python SDK installed$(RESET)"

build: install  ## Build Python SDK (install in dev mode)

test: build  ## Run tests
	@echo -e "$(BLUE)$(GEAR) Testing Python SDK...$(RESET)"
	$(PYTHON) -m pytest tests/ -v
	@echo -e "$(GREEN)$(CHECK) Tests passed$(RESET)"

check: $(VENV)  ## Run linters and formatters
	@echo -e "$(BLUE)$(GEAR) Running black...$(RESET)"
	$(PIP) install black ruff mypy 2>/dev/null || true
	$(VENV)/bin/black --check eda_sdk/
	@echo -e "$(BLUE)$(GEAR) Running ruff...$(RESET)"
	$(VENV)/bin/ruff check eda_sdk/
	@echo -e "$(BLUE)$(GEAR) Running mypy...$(RESET)"
	$(VENV)/bin/mypy eda_sdk/
	@echo -e "$(GREEN)$(CHECK) Checks passed$(RESET)"

clean:  ## Clean build artifacts
	@echo -e "$(YELLOW)Cleaning Python SDK artifacts...$(RESET)"
	rm -rf $(VENV)
	rm -rf build
	rm -rf dist
	rm -rf *.egg-info
	rm -rf eda_sdk/ffi/libs
	rm -rf eda_sdk/ffi/include
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete 2>/dev/null || true
	@echo -e "$(GREEN)$(CHECK) Python SDK cleaned$(RESET)"
