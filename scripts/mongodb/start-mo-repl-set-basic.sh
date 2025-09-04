export MONGO_ORCHESTRATION_HOME=/Users/preston.vasquez/Developer/mongo-orchestration-home

mongo-orchestration stop --pidfile=${MONGO_ORCHESTRATION_HOME}/server.pid
${DRIVERS_TOOLS}/.evergreen/start-orchestration.sh ${MONGO_ORCHESTRATION_HOME}

sudo cp -r /users/preston.vasquez/.local/m/versions/4.0.0/bin /Users/preston.vasquez/Developer/drivers-evergreen-tools/mongodb/bin

${DRIVERS_TOOLS}/.evergreen/start-orchestration.sh /Users/preston.vasquez/Developer/mongo-orchestration-home

curl -X POST http://localhost:8889/v1/sharded_clusters \
  -H 'Content-Type: application/json' \
  -d @${DRIVERS_TOOLS}/.evergreen/orchestration/configs/sharded_clusters/basic-load-balancer.json
