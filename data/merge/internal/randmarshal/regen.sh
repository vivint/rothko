#!/bin/bash

PLUGIN=github.com/spacemonkeygo/rothko/vendor/github.com/gogo/protobuf/protoc-gen-gogo
PLUGIN_PATH=$(go list -f '{{ .Target }}' ${PLUGIN})
VENDOR=$(go list -f '{{ .Dir }}' github.com/spacemonkeygo/rothko)/vendor

go install -v $PLUGIN
protoc --plugin=protoc-gen-gogo=${PLUGIN_PATH} -I${VENDOR} -I. --gogo_out=. *.proto
