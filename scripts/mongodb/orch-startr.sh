#!/bin/bash
set -ex

export KEYFILE="/Users/preston.vasquez/Developer/drivers-evergreen-tools/.evergreen/x509gen/server.pem"
export CAFILE="/Users/preston.vasquez/Developer/drivers-evergreen-tools/.evergreen/x509gen/ca.pem"
export MONGODB_URI="mongodb://127.0.0.1:27017/?ssl=true&tlsCAFile=${CAFILE}&tlsCertificateKeyFile=${KEYFILE}"

export MONGO_ORCHESTRATION_HOME=/Users/preston.vasquez/Developer/mongo-orchestration-home

mongo-orchestration stop --pidfile=${MONGO_ORCHESTRATION_HOME}/server.pid
${DRIVERS_TOOLS}/.evergreen/start-orchestration.sh ${MONGO_ORCHESTRATION_HOME}

sudo cp -r /users/preston.vasquez/.local/m/versions/8.1.1/bin /Users/preston.vasquez/Developer/drivers-evergreen-tools/mongodb/bin

echo "sending security token configuraiton to the server..."

curl -X POST http://localhost:8889/v1/replica_sets \
  -d @${DRIVERS_TOOLS}/.evergreen/orchestration/configs/replica_sets/auth-ssl.json
