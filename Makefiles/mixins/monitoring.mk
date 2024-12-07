# Monitoring mixin providing metrics and observability features
ifndef MONITORING_MK_INCLUDED
MONITORING_MK_INCLUDED := 1

include $(MAKEFILES_DIR)/core/helpers.mk

# Monitoring configuration
METRICS_PORT ?= 2112
METRICS_PATH ?= /metrics
HEALTH_PATH ?= /health
MONITORING_ENABLED ?= true

# Prometheus configuration
PROMETHEUS_JOB ?= $(COMPONENT_NAME)
PROMETHEUS_TARGETS ?= localhost:$(METRICS_PORT)

.PHONY: monitoring-info monitoring-status monitoring-check monitoring-metrics

# Show monitoring information
monitoring-info: ## Show monitoring configuration
	@echo "Monitoring Configuration:"
	@echo "  Enabled: $(MONITORING_ENABLED)"
	@echo "  Metrics port: $(METRICS_PORT)"
	@echo "  Metrics path: $(METRICS_PATH)"
	@echo "  Health path: $(HEALTH_PATH)"
	@echo "  Prometheus job: $(PROMETHEUS_JOB)"

# Check component health
monitoring-check: ## Check component health status
	$(call show_progress,"Checking health status")
	@curl -s -f http://localhost:$(METRICS_PORT)$(HEALTH_PATH) || \
		(echo "$(RED)Health check failed$(RESET)" && exit 1)
	@echo "$(GREEN)Service is healthy$(RESET)"

# Show component metrics
monitoring-metrics: ## Display current metrics
	$(call show_progress,"Fetching metrics")
	@curl -s http://localhost:$(METRICS_PORT)$(METRICS_PATH)

# Check monitoring status
monitoring-status: monitoring-info monitoring-check ## Show full monitoring status
	$(call show_progress,"Checking monitoring status")
	@if [ "$(MONITORING_ENABLED)" = "true" ]; then \
		$(MAKE) -s monitoring-metrics; \
	else \
		echo "$(YELLOW)Monitoring is disabled$(RESET)"; \
	fi

# Hook into the service lifecycle
ifdef SERVICE_TEMPLATE
run::
	@if [ "$(MONITORING_ENABLED)" = "true" ]; then \
		echo "Starting with monitoring on port $(METRICS_PORT)"; \
	fi

endif # SERVICE_TEMPLATE

endif # MONITORING_MK_INCLUDED