#!/bin/bash
#
# Run a subset of tests that strongly indicate driver health.

# Check if DRIVER_PATH is set
if [ -z "$DRIVER_PATH" ]; then
  echo "Error: DRIVER_PATH environment variable is not set."
  exit 1
fi

LOG_FILE="/var/log/test.log"

cd "$DRIVER_PATH" || exit
{
  make
  make build-compile-check
  go test ./mongo -v
  go test ./internal/integration/unified -v 
  go test ./internal/integration -v 
  go test ./internal/integration -v -tags=cse
  go test ./x/mongo/driver -v 
  go test ./x/mongo/driver/topology -v 
  go test ./x/mongo/driver/connstring -v 
} > "$LOG_FILE" 2>&1
