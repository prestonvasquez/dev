#!/bin/bash

# Move to the root director.
cd $HOME

git clone git@github.com:prestonvasquez/mongo-go-driver.git
git clone git@github.com:mongodb-labs/drivers-evergreen-tools.git

export MONGO_ORCHESTRATION_HOME=$HOME
export STPATH=$HOME/mongo-go-driver
export DRIVERS_TOOLS=$HOME/drivers-evergreen-tools
export PROJECT_DIRECTORY=$HOME/mongo-go-driver
export MONGO_ORCHESTRATION_HOME=$DRIVERS_TOOLS/.evergreen/orchestration

# Remove the mongodb dir if it already exists in driver tools 
rm -rf $DRIVERS_TOOLS/mongodb

# Start the server on 27017
sh $DRIVERS_TOOLS/.evergreen/run-orchestration.sh
sudo netstat -tuln | grep 27017

echo "export the following: 
GO_VERSION="go1.20"

export PATH="$PATH:/opt/golang/$GO_VERSION/bin"
export GOROOT="/opt/golang/$GO_VERSION"
"
