#!/bin/bash
# 
# Export env variables required to test on a RHEL VM.

GO_VERSION="go1.20"

# Move to the root director.
cd

git clone git@github.com:prestonvasquez/mongo-go-driver.git

export PATH="$PATH:/opt/golang/$GO_VERSION/bin"
export GOROOT="/opt/golang/$GO_VERSION"
