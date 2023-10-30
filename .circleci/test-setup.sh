#!/bin/bash

# Set up test environment

set -ex 

# Install Ponzu CMS
go get -u github.com/sdassow/ponzu/...

# test install
ponzu

# create a project and generate code
if [ $CIRCLE_BRANCH = "ponzu-dev" ]; then
        # ensure we have the latest from ponzu-dev branch
        cd /go/src/github.com/sdassow/ponzu
        git checkout ponzu-dev
        git pull origin ponzu-dev

        # create new project using the ponzu-dev branch
        ponzu new --dev github.com/tomma-a/ci/test-project
else 
        ponzu new github.com/tomma-a/ci/test-project
fi

cd /go/src/github.com/tomma-a/ci/test-project

ponzu gen content person name:string hashed_secret:string
ponzu gen content message from:@person,hashed_secret to:@person,hashed_secret

# build and run dev http/2 server with TLS
ponzu build

