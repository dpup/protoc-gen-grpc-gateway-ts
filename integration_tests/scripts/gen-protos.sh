#!/bin/bash
USE_PROTO_NAMES=${1:-"false"}
EMIT_UNPOPULATED=${2:-"false"}
ENABLE_STYLING_CHECK=${3:-"false"}

GOBIN=$(go env GOPATH)/bin
PATH=$GOBIN:$PATH

cd .. && go install && cd integration_tests && \
	protoc -I .  -I ../.. \
	--grpc-gateway-ts_out ./ \
	--grpc-gateway-ts_opt logtostderr=true \
	--grpc-gateway-ts_opt loglevel=debug \
	--grpc-gateway-ts_opt use_proto_names=$USE_PROTO_NAMES \
	--grpc-gateway-ts_opt emit_unpopulated=$EMIT_UNPOPULATED \
	--grpc-gateway-ts_opt enable_styling_check=$ENABLE_STYLING_CHECK \
	service.proto msg.proto empty.proto
