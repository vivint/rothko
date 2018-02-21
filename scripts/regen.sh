#!/usr/bin/env bash

# this script is called by go generate with a line like:
#
#	//go:generate bash -c "`go list -f '{{ .Dir }}' github.com/vivint/rothko`/regen.sh"
#
# inside of packages that contain protobuf files.

PLUGIN=github.com/vivint/rothko/vendor/github.com/gogo/protobuf/protoc-gen-gogo
PLUGIN_PATH=$(go list -f '{{ .Target }}' ${PLUGIN})
VENDOR=$(go list -f '{{ .Dir }}' github.com/vivint/rothko)/vendor

go install -v $PLUGIN
protoc --plugin=protoc-gen-gogo="${PLUGIN_PATH}" -I"${VENDOR}" -I. --gogo_out=. ./*.proto

# strip out the proto imports because we don't need them and they're silly
SED=sed
case $(uname) in
	Darwin )
		SED=gsed
		;;
esac

$SED -i '/proto\./d' -- *.pb.go
$SED -i '/^import proto/d' -- *.pb.go
$SED -i '/gogoproto/d' -- *.pb.go
