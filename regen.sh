#!/bin/bash

# this script is called by go generate with a line like:
#
#	//go:generate bash -c "`go list -f '{{ .Dir }}' github.com/spacemonkeygo/rothko`/regen.sh"
#
# inside of packages that contain protobuf files.

PLUGIN=github.com/spacemonkeygo/rothko/vendor/github.com/gogo/protobuf/protoc-gen-gogo
PLUGIN_PATH=$(go list -f '{{ .Target }}' ${PLUGIN})
VENDOR=$(go list -f '{{ .Dir }}' github.com/spacemonkeygo/rothko)/vendor

go install -v $PLUGIN
protoc --plugin=protoc-gen-gogo=${PLUGIN_PATH} -I${VENDOR} -I. --gogo_out=. *.proto
