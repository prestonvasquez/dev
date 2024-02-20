#!/bin/bash
# 
# Export env variables required to test on a windows VM.

echo "!!!!!!THIS SCRIPT HAS TO BE RUN IN MICROSOFT REMOTE DESKTOP!!!!!!"
echo "See here for setup info: "
echo "https://docs.google.com/document/d/1rkB4N8C4e_mSTjlPG-ghX3MMRZSDSnZSeAOJ90pf5Us/edit#heading=h.b6g7b2kydeti"

# Move to the root director.
WD="/cygdrive/c/data"
cd $WD


git clone git@github.com:prestonvasquez/mongo-go-driver.git
git clone git@github.com:mongodb-labs/drivers-evergreen-tools.git

export DRIVERS_TOOLS=C:/data/drivers-evergreen-tools
export MONGO_ORCHESTRATION_HOME=C:/data/drivers-evergreen-tools/.evergreen/orchestration
export MONGODB_BINARIES=C:/data/drivers-evergreen-tools/mongodb/bin

# Remove the mongodb dir if it already exists in driver tools 
rm -rf $DRIVERS_TOOLS/mongodb

# Start the server on 27017
sh C:/data/drivers-evergreen-tools/.evergreen/run-orchestration.sh

echo "export the following: 
GO_VERSION="go1.20"

export GOROOT="C:\\golang\\$GO_VERSION"
export PATH="$GOROOT/bin:$PATH"
export GOPATH="C:\\home\\Administrator\\mongo-go-driver"
export GOCACHE="C:\\home\\Administrator\\mongo-go-driver\\mongo\\integration\\unified\\.cache"
"
