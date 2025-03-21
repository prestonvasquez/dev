#!/bin/bash
set -ex

cd ${DRIVERS_TOOLS}/.evergreen/docker
#ARCH=amd64 TOPOLOGY=replica_set MONGODB_VERSION="4.0" TARGET_IMAGE="ubuntu18.04" ./run-server.sh
TOPOLOGY=sharded_cluster ORCHESTRATION_FILE=basic-load-balancer.json ./run-server.sh
