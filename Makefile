.PHONY: driver-sanity-test
driver-sanity-test:
	scripts/mongodb/driver-sanity-test.sh

.PHONY: install-diskhop-trasher
install-diskhop-trasher:
	scripts/os/install-diskhop-trasher.sh
	chmod +x scripts/os/install-diskhop-trasher.sh
