# Common Makefile variables and functions
# Include this in component Makefiles with: include ../common.mk or include ../../common.mk

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

# Configuration (can be overridden from environment)
EXAMPLE_TIMEOUT ?= 30

# Detect container engine (podman or docker)
CONTAINER_ENGINE := $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null)

# Detect OS for proper library extension
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	LIB_EXT := so
endif
ifeq ($(UNAME_S),Darwin)
	LIB_EXT := dylib
endif
