#!/bin/bash
set -ex

cd ${DRIVERS_TOOLS}/.evergreen/docker
TOPOLOGY=server ORCHESTRATION_FILE=auth.json ./run-server.sh
