#!/bin/bash

USE_PROTO_NAMES=${1:-"false"}
EMIT_UNPOPULATED=${2:-"false"}
ENABLE_STYLING_CHECK=${3:-"false"}

GOBIN=$(go env GOPATH)/bin
PATH=$GOBIN:$PATH

# NOTE: The parent directory is brought in because google/protobuf/empty.proto
# has been updated to reference "protoc-gen-grpc-gateway-ts/options/ts_package.protoprotoc-gen-grpc-gateway-ts/options/ts_package.proto"
# this seems somewhat risky.

go install &&  \
	protoc -I ./test/integration/ \
	-I ../ \
	--grpc-gateway-ts_out ./test/integration/ \
	--grpc-gateway-ts_opt logtostderr=true \
	--grpc-gateway-ts_opt loglevel=debug \
	--grpc-gateway-ts_opt use_proto_names=$USE_PROTO_NAMES \
	--grpc-gateway-ts_opt emit_unpopulated=$EMIT_UNPOPULATED \
	--grpc-gateway-ts_opt enable_styling_check=$ENABLE_STYLING_CHECK \
	service.proto msg.proto empty.proto
