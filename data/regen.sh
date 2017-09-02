#!/bin/bash

PLUGIN=github.com/spacemonkeygo/rothko/vendor/github.com/gogo/protobuf/protoc-gen-gogo
PLUGIN_PATH=$(go list -f '{{ .Target }}' ${PLUGIN})

go install -v $PLUGIN
protoc --plugin=protoc-gen-gogo=$PLUGIN_PATH -I../vendor -I. --gogo_out=. data.proto