SHELL := /bin/bash

.PHONY: all
all: testdata lint test

.PHONY: testdata
testdata:
	go install
	@export PATH=$$PATH:$$(go env GOPATH)/bin; \
	cd testdata && protoc -I . \
	--grpc-gateway-ts_out ./ \
	--grpc-gateway-ts_opt logtostderr=true \
	log.proto environment.proto ./datasource/datasource.proto

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test: go-tests integration-tests

.PHONY: go-tests
go-tests:
	go test ./...

.PHONY: integration-tests
integration-tests: integration-tests-gengo
	cd integration_tests && ./scripts/test-ci.sh

.PHONY: integration-tests-gengo
integration-tests-gengo:
	go install; \
	cd integration_tests && ./scripts/gen-server-proto.sh

.PHONY: fmt 
fmt: fmt-go fmt-ts

.PHONY: fmt-go 
fmt-go:
	go fmt ./...

.PHONY: fmt-ts 
fmt-ts:
	prettier **/*.ts --write