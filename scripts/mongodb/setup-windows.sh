#!/bin/bash
# 
# Export env variables required to test on a windows VM.

git clone git@github.com:prestonvasquez/mongo-go-driver.git


# Move to the root director.
cd $HOME

git clone git@github.com:prestonvasquez/mongo-go-driver.git
git clone git@github.com:mongodb-labs/drivers-evergreen-tools.git

export MONGO_ORCHESTRATION_HOME=$(cygpath -m $HOME)
export STPATH=$(cygpath $HOME/mongo-go-driver)
export DRIVERS_TOOLS=$(cygpath $HOME/drivers-evergreen-tools)
export PROJECT_DIRECTORY=$(cygpath $HOME/mongo-go-driver)
export MONGO_ORCHESTRATION_HOME=$(cygpath $DRIVERS_TOOLS/.evergreen/orchestration)
export VENV_BIN_DIR=$(cygpath $DRIVERS_TOOLS/.evergreen/orchestration/venv/Scripts)
export PYTHONIOENCODING=UTF-8

# Remove the mongodb dir if it already exists in driver tools 
rm -rf $DRIVERS_TOOLS/mongodb

# Start the server on 27017
sh $DRIVERS_TOOLS/.evergreen/run-orchestration.sh
sudo netstat -tuln | grep 27017

echo "export the following: 
GO_VERSION="go1.20"

export GOROOT="C:\\golang\\$GO_VERSION"
export PATH="$GOROOT/bin:$PATH"
export GOPATH="C:\\home\\Administrator\\mongo-go-driver"
export GOCACHE="C:\\home\\Administrator\\mongo-go-driver\\mongo\\integration\\unified\\.cache"
"
