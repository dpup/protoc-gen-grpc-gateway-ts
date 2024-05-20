SHELL := /bin/bash

GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin

export PATH := $(GOBIN):$(PATH)

.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: options/ts_package.pb.go ## Builds the typescript generator plugin.
	@go install

options/ts_package.pb.go: options/ts_package.proto tools
	@protoc --go_out=paths=source_relative:./ ./options/*.proto

.PHONY: lint
lint: ## Runs the go linter.
	@golangci-lint run 

.PHONY: test
test: go-tests integration-tests ## Runs all the tests.

.PHONY: go-tests
go-tests: testdata integration-test-server ## Runs the unit tests.
	@go test ./...

.PHONY: integration-tests
integration-tests: integration-test-deps integration-test-client integration-test-server ## Runs the integration tests.
	@cd test/integration && ./run.js

.PHONY: fmt 
fmt: fmt-go fmt-ts ## Auto-format the Go and TypeScript files.

.PHONY: fmt-go 
fmt-go: ## Auto-format the Go files.
	@go fmt ./...

.PHONY: fmt-ts 
fmt-ts: ## Auto-format the TypeScript files.
	@prettier **/*.ts --write

.PHONY: testdata
testdata: build tools ## Generates test data used by the unit tests.
	@protoc -I ./test/testdata/. \
		--grpc-gateway-ts_out ./test/testdata/ \
		--grpc-gateway-ts_opt logtostderr=true \
		log.proto environment.proto names.proto datasource/datasource.proto

.PHONY: integration-test-client
integration-test-client: build ## Generates the typescript client code used by each test case.
	@protoc -I ./test/integration/protos \
		-I ../ \
		--grpc-gateway-ts_out ./test/integration/defaultConfig \
		--grpc-gateway-ts_opt logtostderr=true \
		--grpc-gateway-ts_opt loglevel=debug \
		--grpc-gateway-ts_opt use_proto_names=false \
		--grpc-gateway-ts_opt emit_unpopulated=false \
		--grpc-gateway-ts_opt enable_styling_check=true \
		--grpc-gateway-ts_opt use_static_classes=true \
		service.proto msg.proto empty.proto

	@# Use proto names (i.e. foo_bar instead of fooBar) for fields and messages.
	@protoc -I ./test/integration/protos \
		-I ../ \
		--grpc-gateway-ts_out ./test/integration/useProtoNames \
		--grpc-gateway-ts_opt logtostderr=true \
		--grpc-gateway-ts_opt loglevel=debug \
		--grpc-gateway-ts_opt use_proto_names=true \
		--grpc-gateway-ts_opt emit_unpopulated=false \
		--grpc-gateway-ts_opt enable_styling_check=true \
		--grpc-gateway-ts_opt use_static_classes=true \
		service.proto msg.proto empty.proto

	@# Emit unpopulated fields.
	@protoc -I ./test/integration/protos \
		-I ../ \
		--grpc-gateway-ts_out ./test/integration/emitUnpopulated \
		--grpc-gateway-ts_opt logtostderr=true \
		--grpc-gateway-ts_opt loglevel=debug \
		--grpc-gateway-ts_opt use_proto_names=false \
		--grpc-gateway-ts_opt emit_unpopulated=true \
		--grpc-gateway-ts_opt enable_styling_check=true \
		--grpc-gateway-ts_opt use_static_classes=true \
		service.proto msg.proto empty.proto

	@# Disable static classes, export functions and a client instead.
		@protoc -I ./test/integration/protos \
		-I ../ \
		--grpc-gateway-ts_out ./test/integration/noStaticClasses \
		--grpc-gateway-ts_opt logtostderr=true \
		--grpc-gateway-ts_opt loglevel=debug \
		--grpc-gateway-ts_opt use_proto_names=false \
		--grpc-gateway-ts_opt emit_unpopulated=false \
		--grpc-gateway-ts_opt enable_styling_check=true \
		--grpc-gateway-ts_opt use_static_classes=false \
		service.proto msg.proto empty.proto

.PHONY: integration-test-server
integration-test-server: tools ## Generates the server dependencies used by the integration tests.
	@mkdir -p test/integration/tmp/openapiv2/
	@protoc -I ./test/integration/protos/. \
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

.PHONY: tools
tools: tools.touchfile ## Installs the other proto plugins used by the project.
tools.touchfile: go.mod go.sum
	@touch tools.touchfile
	@go get
	@go install \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	@cp -f $(GOBIN)/protoc-gen-go-grpc $(GOBIN)/protoc-gen-grpc

.PHONY: integration-test-deps
integration-test-deps: node-deps.touchfile ## Installs NPM deps required for the integration tests.
node-deps.touchfile: test/integration/package.json test/integration/package-lock.json
	@cd test/integration && npm install

.PHONY: clean
clean: ## Removes all generated files.
	@rm -r coverage || true
	@rm -r test/integration/coverage || true
	@rm -r test/integration/tmp || true
	@find . -name "*.pb.go" -type f -delete
	@find . -name "*.pb.gw.go" -type f -delete
	@find . -name "*.pb.ts" -type f -delete
	@echo "👷🏽‍♀️ Generated proto files removed"