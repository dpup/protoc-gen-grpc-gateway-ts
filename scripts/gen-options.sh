#!/bin/bash
# Use to regenerate the options package.

GOBIN=$(go env GOPATH)/bin
PATH=$GOBIN:$PATH

protoc --go_out=paths=source_relative:./ ./options/*.proto
