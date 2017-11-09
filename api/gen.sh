#!/usr/bin/env bash

# Since GOPATH can be a path, we can't just use it as a variable.  Split it up 
# to the various paths, and append the subpaths.
GOSUBPATHS="/src:/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis"
GOPATHLIST=""
OIFS=$IFS
IFS=':' 
for GOBASEPATH in $GOPATH; do
    for GOSUBPATH in $GOSUBPATHS; do
    	if [ -e ${GOBASEPATH}${GOSUBPATH} ]; then
        	GOPATHLIST="${GOPATHLIST} -I${GOBASEPATH}${GOSUBPATH}"
        fi
    done
done
IFS=$OIFS

# generate the gRPC code
protoc  -I/usr/local/include -I. ${GOPATHLIST} --go_out=plugins=grpc:. \
    device.proto \
    application.proto \
    deviceQueue.proto \
    common.proto \
    user.proto \
    gateway.proto \
    organization.proto \
    profiles.proto \
    networkServer.proto \
    serviceProfile.proto \
    deviceProfile.proto

# generate the JSON interface code
protoc -I/usr/local/include -I. ${GOPATHLIST} --grpc-gateway_out=logtostderr=true:. \
    device.proto \
    application.proto \
    deviceQueue.proto \
    common.proto \
    user.proto \
    gateway.proto \
    organization.proto \
    profiles.proto \
    networkServer.proto \
    serviceProfile.proto \
    deviceProfile.proto

# generate the swagger definitions
protoc -I/usr/local/include -I. ${GOPATHLIST} --swagger_out=logtostderr=true:./swagger \
    device.proto \
    application.proto \
    deviceQueue.proto \
    common.proto \
    user.proto \
    gateway.proto \
    organization.proto \
    profiles.proto \
    networkServer.proto \
    serviceProfile.proto \
    deviceProfile.proto

# merge the swagger code into one file
go run swagger/main.go swagger > ../static/swagger/api.swagger.json
