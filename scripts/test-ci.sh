#!/bin/bash
source ./scripts/source.sh

CONF="karma.conf.ci.js"

# runTest [use proto names] [emit unpopulated] [config]
runTest false false $CONF
runTest true false $CONF
runTest false true $CONF
runTest true true $CONF
