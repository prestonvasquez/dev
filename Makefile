# Define the path to the script and cron job details
SCRIPT_PATH := os/run-trashdop.sh
CRON_EXPRESSION := "*/30 * * * * $(SCRIPT_PATH)"

# The target to add the cron job
install-cron:
	@echo "Adding cron job for trashdop..."
	@crontab -l | { cat; echo $(CRON_EXPRESSION); } | crontab -
	@echo "Cron job added successfully."

# Ensure the script is executable
install-cron: $(SCRIPT_PATH)
	@chmod +x $(SCRIPT_PATH)

.PHONY: driver-sanity-test
driver-sanity-test:
	scripts/mongodb/driver-sanity-test.sh

.PHONY: install-diskhop-trasher
install-diskhop-trasher:
	scripts/os/install-diskhop-trasher.sh
	chmod +x scripts/os/install-diskhop-trasher.sh
