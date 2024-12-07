# Daemon-specific template
include $(MAKEFILES_DIR)/templates/go.mk

# Define the daemon template
define DAEMON_TEMPLATE
$(GO_TEMPLATE)

.PHONY: test-hw test-sim sim-start sim-stop install uninstall

test-hw: ## Run hardware tests
	$(GOTEST) $(TESTFLAGS) $(HWFLAGS) ./hardware/...

test-sim: sim-start ## Run simulation tests
	$(GOTEST) $(TESTFLAGS) $(SIMFLAGS) ./...
	@make sim-stop

sim-start: ## Start simulation environment
	mkdir -p $(SIM_DIR)/{gpio,power,thermal,secure}
	@echo "Simulation environment created at $(SIM_DIR)"

sim-stop: ## Stop simulation environment
	rm -rf $(SIM_DIR)
	@echo "Simulation environment cleaned up"

install: build ## Install daemon
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo cp ./init/$(BINARY_NAME).service /etc/systemd/system/
	sudo systemctl daemon-reload

uninstall: ## Uninstall daemon
	sudo systemctl stop $(BINARY_NAME)
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -f /etc/systemd/system/$(BINARY_NAME).service
	sudo systemctl daemon-reload

endef