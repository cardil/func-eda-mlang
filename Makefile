# Root Makefile - delegates to component Makefiles

include common.mk

.PHONY: help all build clean check test e2e run

help:  ## Show this help
	@echo -e "$(CYAN)$(BOOK) Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;36m%-20s\033[0m %s\n", $$1, $$2}'

all:  ## Build everything
	@echo -e "$(BLUE)$(GEAR) Building all components...$(RESET)"
	@$(MAKE) -C core all
	@$(MAKE) -C sdks/go build
	@$(MAKE) -C sdks/python build
	@echo -e "$(GREEN)$(ROCKET) All builds complete$(RESET)"

build:  ## Build for current platform
	@echo -e "$(BLUE)$(GEAR) Building for current platform...$(RESET)"
	@$(MAKE) -C core build
	@$(MAKE) -C sdks/go build
	@$(MAKE) -C sdks/python build
	@echo -e "$(GREEN)$(ROCKET) Build complete$(RESET)"

clean:  ## Clean all build artifacts
	@echo -e "$(YELLOW)$(CLEAN) Cleaning all artifacts...$(RESET)"
	@$(MAKE) -C core clean
	@$(MAKE) -C sdks/go clean
	@$(MAKE) -C sdks/python clean
	@$(MAKE) -C infra clean
	@rm -rf .tools
	@echo -e "$(GREEN)$(CHECK) All artifacts cleaned$(RESET)"

check:  ## Run all linters and formatters
	@echo -e "$(BLUE)$(GEAR) Running checks...$(RESET)"
	@$(MAKE) -C core check
	@$(MAKE) -C sdks/go check
	@$(MAKE) -C sdks/python check
	@echo -e "$(GREEN)$(CHECK) All checks passed$(RESET)"

test:  ## Run all unit tests
	@echo -e "$(BLUE)$(GEAR) Running tests...$(RESET)"
	@$(MAKE) -C core test
	@$(MAKE) -C sdks/go test
	@$(MAKE) -C sdks/python test
	@echo -e "$(GREEN)$(CHECK) All tests passed$(RESET)"

e2e: build  ## Run end-to-end tests
	@echo -e "$(BLUE)$(GEAR) Running e2e tests...$(RESET)"
	@sdks/python/.venv/bin/pytest tests/e2e
	@echo -e "$(GREEN)$(CHECK) E2E tests passed$(RESET)"

run:  ## Run an example (interactive selection)
	@echo -e "$(CYAN)Select an example to run:$(RESET)"
	@CHOICE=$$(go run github.com/charmbracelet/gum@latest choose \
		"Go FFI Example" \
		"Go FFI Output Example" \
		"Python FFI Example" \
		"Python FFI Output Example"); \
	case "$$CHOICE" in \
		"Go FFI Example") \
			$(MAKE) run-example EXAMPLE=sdks/go/examples/ffi-example ;; \
		"Go FFI Output Example") \
			$(MAKE) run-example EXAMPLE=sdks/go/examples/ffi-output-example ;; \
		"Python FFI Example") \
			$(MAKE) run-example EXAMPLE=sdks/python/examples/ffi ;; \
		"Python FFI Output Example") \
			$(MAKE) run-example EXAMPLE=sdks/python/examples/ffi-output ;; \
	esac

.PHONY: run-example
run-example:
	@echo -e "$(BLUE)$(ROCKET) Starting infrastructure...$(RESET)"
	@$(MAKE) -C infra up
	@echo -e "$(BLUE)$(ROCKET) Sending test events...$(RESET)"
	@$(MAKE) -C infra send-events
	@echo -e "$(BLUE)$(ROCKET) Running example: $(EXAMPLE)$(RESET)"
	@$(MAKE) -C $(EXAMPLE) run
	@echo -e "$(YELLOW)Stopping infrastructure...$(RESET)"
	@$(MAKE) -C infra down
	@echo -e "$(GREEN)$(CHECK) Example run complete$(RESET)"
