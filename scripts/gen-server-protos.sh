#!/bin/bash

GOBIN=$(go env GOPATH)/bin
PATH=$GOBIN:$PATH

# remove binaries to ensure that binaries present in tools.go are installed
rm -f $GOBIN/protoc-gen-go $GOBIN/protoc-gen-grpc $GOBIN/protoc-gen-grpc-gateway $GOBIN/protoc-gen-openapiv2

# remove old generated files
rm test/integration/*.pb.go test/integration/*.pb.gw.go

go install \
	google.golang.org/protobuf/cmd/protoc-gen-go \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc \
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

# TODO(dan): Why is the filename different?
mv $GOBIN/protoc-gen-go-grpc $GOBIN/protoc-gen-grpc

mkdir -p test/integration/tmp/openapiv2

echo $(pwd)

protoc -I ./test/integration/. \
  -I ../. \
  --go_out ./test/integration/. \
  --go_opt paths=source_relative \
  --grpc_out ./test/integration/. \
	--grpc-gateway_out ./test/integration/. \
	--grpc-gateway_opt logtostderr=true \
	--grpc-gateway_opt paths=source_relative \
	--grpc-gateway_opt generate_unbound_methods=true \
	--grpc-gateway_opt logtostderr=true \
	--grpc-gateway_opt generate_unbound_methods=true \
	--openapiv2_out ./test/integration/tmp/openapiv2 \
	--openapiv2_opt logtostderr=true \
	--openapiv2_opt use_go_templates=true \
	--openapiv2_opt simple_operation_ids=true \
	--openapiv2_opt openapi_naming_strategy=fqn \
	--openapiv2_opt disable_default_errors=true \
	service.proto msg.proto
