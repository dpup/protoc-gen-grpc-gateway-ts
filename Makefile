SHELL := /bin/bash

.PHONY: all
all: testdata lint test

.PHONY: testdata
testdata:
	go install
	@export PATH=$$PATH:$$(go env GOPATH)/bin; \
	cd testdata && protoc -I . \
	--grpc-gateway-ts_out=logtostderr=true:./ \
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
integration-tests:
	@export PATH=$$PATH:$$(go env GOPATH)/bin; \
	cd integration_tests && ./scripts/test-ci.sh
