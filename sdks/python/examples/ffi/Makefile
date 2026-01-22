# Python FFI Example Makefile

include ../../../../common.mk

SDK_DIR := ../..
PYTHON := $(CURDIR)/$(SDK_DIR)/.venv/bin/python
FUNC := func.py

.PHONY: all run clean sdk-install

all: sdk-install  ## Build everything (install SDK)

sdk-install:
	@$(MAKE) -C $(SDK_DIR) install

run: sdk-install  ## Run the example
	@echo -e "$(BLUE)$(ROCKET) Running Python FFI example...$(RESET)"
	timeout $(EXAMPLE_TIMEOUT) $(PYTHON) $(FUNC)

clean:  ## Clean build artifacts
	@echo -e "$(YELLOW)Cleaning Python FFI example...$(RESET)"
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete 2>/dev/null || true
	@echo -e "$(GREEN)$(CHECK) Cleaned$(RESET)"
