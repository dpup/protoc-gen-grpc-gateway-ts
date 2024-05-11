SHELL := /bin/bash

.PHONY: all
all: testdata lint test

.PHONY: testdata
testdata:
	go install
	@export PATH=$$PATH:$$(go env GOPATH)/bin; \
	protoc -I ./test/testdata/. \
	--grpc-gateway-ts_out ./test/testdata/ \
	--grpc-gateway-ts_opt logtostderr=true \
	log.proto environment.proto datasource/datasource.proto

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test: go-tests integration-tests

.PHONY: go-tests
go-tests: gen-go testdata
	go test ./...

.PHONY: integration-tests
integration-tests: gen-go
	./scripts/test-ci.sh

.PHONY: gen-go
gen-go:
	./scripts/gen-options.sh
	go install
	./scripts/gen-server-protos.sh

.PHONY: fmt 
fmt: fmt-go fmt-ts

.PHONY: fmt-go 
fmt-go:
	go fmt ./...

.PHONY: fmt-ts 
fmt-ts:
	prettier **/*.ts --write

.PHONY: clean
clean:
	@rm -r coverage
	@rm -r test/integration/coverage
	@rm -r test/integration/tmp
	@find . -name "*.pb.go" -type f -delete
	@find . -name "*.pb.gw.go" -type f -delete
	@find . -name "*.pb.ts" -type f -delete
	@echo "👷🏽‍♀️ Generated proto files removed"