#!/usr/bin/env bash

GRPC_GW_PATH=`go list -f '{{ .Dir }}' github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway`
GRPC_GW_PATH="${GRPC_GW_PATH}/../third_party/googleapis"

# generate the gRPC code
protoc -I../vendor -I${GRPC_GW_PATH} -I. --go_out=plugins=grpc:. \
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
    deviceProfile.proto \
    gatewayProfile.proto \
    multicastGroup.proto \
    internal.proto

# generate the JSON interface code
protoc -I../vendor -I${GRPC_GW_PATH} -I. --grpc-gateway_out=logtostderr=true:. \
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
    deviceProfile.proto \
    gatewayProfile.proto \
    multicastGroup.proto \
    internal.proto

# generate the swagger definitions
protoc -I../vendor -I${GRPC_GW_PATH} -I. --swagger_out=json_names_for_fields=true:./swagger \
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
    deviceProfile.proto \
    gatewayProfile.proto \
    multicastGroup.proto \
    internal.proto

# merge the swagger code into one file
go run swagger/main.go swagger > ../static/swagger/api.swagger.json
