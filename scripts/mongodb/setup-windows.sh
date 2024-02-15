#!/bin/bash
# 
# Export env variables required to test on a windows VM.

GO_VERSION="go1.20"

git clone git@github.com:prestonvasquez/mongo-go-driver.git

export GOROOT="C:\\golang\\$GO_VERSION"
export PATH="$GOROOT/bin:$PATH"
export GOPATH="C:\\home\\Administrator\\mongo-go-driver"
export GOCACHE="C:\\home\\Administrator\\mongo-go-driver\\mongo\\integration\\unified\\.cache"
