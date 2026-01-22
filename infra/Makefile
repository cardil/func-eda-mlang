# Infrastructure Makefile

include ../common.mk

.PHONY: all up down logs send-events clean

all: up  ## Start infrastructure

up:  ## Start Kafka/Redpanda infrastructure
	@echo -e "$(BLUE)$(ROCKET) Starting Kafka/Redpanda...$(RESET)"
	$(CONTAINER_ENGINE) compose up -d
	@printf "$(CYAN)Waiting for Redpanda to be ready$(RESET)"
	@for i in $$(seq 1 60); do \
		status="$$($(CONTAINER_ENGINE) container inspect --format '{{.State.Health.Status}}' redpanda 2>/dev/null)"; \
		if [ "$$status" = "healthy" ]; then \
			break; \
		fi; \
		printf "."; \
		sleep 1; \
	done
	@echo ""
	@echo -e "$(CYAN)Creating topics...$(RESET)"
	@$(CONTAINER_ENGINE) exec redpanda rpk topic create events --partitions 1 --replicas 1 2>/dev/null || echo -e "$(YELLOW)Topic 'events' may already exist$(RESET)"
	@$(CONTAINER_ENGINE) exec redpanda rpk topic create processed-events --partitions 1 --replicas 1 2>/dev/null || echo -e "$(YELLOW)Topic 'processed-events' may already exist$(RESET)"
	@echo -e "$(GREEN)$(CHECK) Redpanda Console available at http://localhost:8080$(RESET)"

down:  ## Stop Kafka/Redpanda infrastructure
	@echo -e "$(YELLOW)Stopping Kafka/Redpanda...$(RESET)"
	$(CONTAINER_ENGINE) compose down
	@echo -e "$(GREEN)$(CHECK) Kafka/Redpanda stopped$(RESET)"

logs:  ## View Kafka/Redpanda logs
	@echo -e "$(CYAN)Showing Kafka/Redpanda logs...$(RESET)"
	$(CONTAINER_ENGINE) compose logs -f redpanda

send-events:  ## Send a test CloudEvent to Kafka
	@echo -e "$(BLUE)$(ROCKET) Sending test CloudEvent...$(RESET)"
	@bash scripts/send-test-event.sh
	@echo -e "$(GREEN)$(CHECK) Test event sent$(RESET)"

clean: down  ## Clean infrastructure (stop containers)
	@echo -e "$(GREEN)$(CHECK) Infrastructure cleaned$(RESET)"
