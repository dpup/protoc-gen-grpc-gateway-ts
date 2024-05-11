#!/bin/bash

GOBIN=$(go env GOPATH)/bin
PATH=$GOBIN:$PATH

function runTest {
  PROTO_NAMES=${1:="false"}
  UNPOPULATED=${2:="true"}
  CONFIG_NAME=test/integration/${3:="karma.conf.js"}
  ./scripts/gen-protos.sh $PROTO_NAMES $UNPOPULATED true
  go run ./test/integration/. --use_proto_names=$PROTO_NAMES --emit_unpopulated=$UNPOPULATED &
  pid=$!

  USE_PROTO_NAMES=$PROTO_NAMES EMIT_UNPOPULATED=$UNPOPULATED ./test/integration/node_modules/.bin/karma start $CONFIG_NAME
  TEST_EXIT=$?
  if [[ $TEST_EXIT -ne 0 ]]; then
    pkill -P $pid
    exit $TEST_EXIT
  fi

  pkill -P $pid
}


