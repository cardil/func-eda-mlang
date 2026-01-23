# Go FFI Example Makefile

include ../../../../common.mk

BINARY := ffi-consumer
SOURCES := $(shell find . -name '*.go')
SDK_DIR := ../..
SDK_SOURCES := $(shell find $(SDK_DIR)/pkg -name '*.go' 2>/dev/null)
EMBEDDED_LIBS := $(wildcard $(SDK_DIR)/pkg/ffi/libs/*/*.so $(SDK_DIR)/pkg/ffi/libs/*/*.dylib $(SDK_DIR)/pkg/ffi/libs/*/*.dll)

.PHONY: all build run clean sdk-build

all: build  ## Build everything

build: $(BINARY)  ## Build the example

sdk-build:
	@$(MAKE) -C $(SDK_DIR) build

$(BINARY): $(SOURCES) $(SDK_SOURCES) $(EMBEDDED_LIBS) | sdk-build
	@echo -e "$(BLUE)$(GEAR) Building Go FFI example...$(RESET)"
	CGO_ENABLED=1 go build -o $(BINARY) .
	@echo -e "$(GREEN)$(CHECK) Built: $(BINARY)$(RESET)"

run: $(BINARY)  ## Run the example
	@echo -e "$(BLUE)$(ROCKET) Running Go FFI example...$(RESET)"
	timeout $(EXAMPLE_TIMEOUT) ./$(BINARY)

clean:  ## Clean build artifacts
	@echo -e "$(YELLOW)Cleaning Go FFI example...$(RESET)"
	rm -f $(BINARY)
	@echo -e "$(GREEN)$(CHECK) Cleaned$(RESET)"
