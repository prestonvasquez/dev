#!/bin/bash
set -ex

cd ${DRIVERS_TOOLS}/.evergreen/docker
TOPOLOGY=sharded_cluster \
  bash ./run-server.sh
