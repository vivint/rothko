#!/usr/bin/env bash

# this script is called by go generate with a line like:
#
#	//go:generate bash -c "`go list -f '{{ .Dir }}' github.com/vivint/rothko`/regen.sh"
#
# inside of packages that contain protobuf files.

set -e

PLUGIN=github.com/gogo/protobuf/protoc-gen-gogo
PLUGIN_PATH=$(vgo list -f '{{ .Target }}' "${PLUGIN}")
INCLUDE=$(dirname "$(vgo list -f '{{ .Dir }}' "${PLUGIN}")")

vgo install -v $PLUGIN
protoc --plugin=protoc-gen-gogo="${PLUGIN_PATH}" -I"${INCLUDE}" -I. --gogo_out=. ./*.proto

# strip out the proto imports because we don't need them and they're silly.
# we want protobuf as a serialization format, not some api runtime reflection,
# generic registry whackdoodlery.

SED="sed"
case $(uname) in
	Darwin )
		SED="gsed"
		;;
esac

$SED -i '/proto\./d' -- *.pb.go
$SED -i '/^import proto/d' -- *.pb.go
$SED -i '/gogoproto/d' -- *.pb.go
