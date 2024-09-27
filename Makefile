# Define the path to the script and cron job details
SCRIPT_PATH := scripts/os/run-trashdop.sh
CRON_EXPRESSION := "*/30 * * * * $(SCRIPT_PATH)"

# Ensure the script is executable
$(SCRIPT_PATH):
	chmod +x $(SCRIPT_PATH)

# The target to add the cron job
install-cron: $(SCRIPT_PATH)
	@echo "Adding cron job for trashdop..."
	@crontab -l | { cat; echo $(CRON_EXPRESSION); } | crontab -
	@echo "Cron job added successfully."

.PHONY: driver-sanity-test
driver-sanity-test:
	scripts/mongodb/driver-sanity-test.sh

.PHONY: install-diskhop-trasher
install-diskhop-trasher:
	chmod +x scripts/os/install-diskhop-trasher.sh
	scripts/os/install-diskhop-trasher.sh

.PHONY: install-diskhop-compactor
install-diskhop-compactor:
	chmod +x scripts/os/install-diskhop-compactor.sh
	scripts/os/install-diskhop-compactor.sh
