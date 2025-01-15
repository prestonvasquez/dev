#!/bin/bash
set -ex

cd ${DRIVERS_TOOLS}/.evergreen/docker
#ARCH=amd64 TOPOLOGY=replica_set MONGODB_VERSION="4.0" TARGET_IMAGE="ubuntu18.04" ./run-server.sh  
ARCH=amd64 MONGODB_VERSION="8.0" TOPOLOGY=replica_set ./run-server.sh  


