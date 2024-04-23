#!/bin/bash
USE_PROTO_NAMES=${1:-"false"}
EMIT_UNPOPULATED=${2:-"false"}
ENABLE_STYLING_CHECK=${3:-"false"}

GOBIN=$(go env GOPATH)/bin
PATH=$GOBIN:$PATH

cd .. && go install && cd integration_tests && \
	protoc -I .  -I ../.. \
	--grpc-gateway-ts_out=logtostderr=true,use_proto_names=$USE_PROTO_NAMES,emit_unpopulated=$EMIT_UNPOPULATED,enable_styling_check=$ENABLE_STYLING_CHECK,loglevel=debug:./ \
	service.proto msg.proto empty.proto
