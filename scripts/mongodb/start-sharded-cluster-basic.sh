#!/bin/bash
set -ex

cd ${DRIVERS_TOOLS}/.evergreen/docker
MONGODB_VERSION="7.0" TOPOLOGY=sharded_cluster ./run-server.sh
#ARCH=amd64 TOPOLOGY=replica_set MONGODB_VERSION="4.2" TARGET_IMAGE="ubuntu18.04" \
#  TOPOLOGY=sharded_cluster \
#  bash ./run-server.sh
