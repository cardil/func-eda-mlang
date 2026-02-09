.PHONY: serve open help

PORT ?= 8000

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

serve: ## Serve slides locally (default port 8000)
	@echo "Starting slides server at http://localhost:$(PORT)"
	@python3 -m http.server $(PORT)

open: ## Open slides in browser
	@xdg-open http://localhost:$(PORT) 2>/dev/null || open http://localhost:$(PORT) 2>/dev/null || echo "Open http://localhost:$(PORT) in your browser"

run: serve ## Alias for serve
