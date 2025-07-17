#!/bin/bash
set -ex

#cd ${DRIVERS_TOOLS}/.evergreen/docker
#
##aws sso login --profile $AWS_PROFILE
##bash setup.sh
#
##ARCH=amd64 TOPOLOGY=replica_set MONGODB_VERSION="4.0" TARGET_IMAGE="ubuntu18.04" ./run-server.sh
#TOPOLOGY=replica_set ./run-server.sh

pushd "${DRIVERS_TOOLS}/.evergreen/docker"

TOPOLOGY=replica_set ./run-server.sh

popd
