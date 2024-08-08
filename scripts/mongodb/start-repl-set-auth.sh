#!/bin/bash
set -ex

echo "run the following in the working dir:"
echo "export AUTH=auth"
echo "export MONGODB_URI=mongodb://bob:pwd123@localhost:27017"

cd ${DRIVERS_TOOLS}/.evergreen/docker
TOPOLOGY=replica_set ORCHESTRATION_FILE=auth.json bash ./run-server.sh  

