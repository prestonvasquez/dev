#!/bin/bash
set -ex

cd ${DRIVERS_TOOLS}/.evergreen/docker

TOPOLOGY=replica_set \
  ORCHESTRATION_FILE=basic-ssl.json \
  KEYFILE="/Users/preston.vasquez/Developer/drivers-evergreen-tools/.evergreen/x509gen/server.pem" \
  CAFILE="/Users/preston.vasquez/Developer/drivers-evergreen-tools/.evergreen/x509gen/ca.pem" \
  bash ./run-server.sh
