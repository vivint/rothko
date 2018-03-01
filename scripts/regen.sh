#!/usr/bin/env bash

# this script is called by go generate with a line like:
#
#	//go:generate bash -c "`vgo list -f '{{ .Dir }}' github.com/vivint/rothko`/regen.sh"
#
# inside of packages that contain protobuf files.

set -e

log() {
	echo "---" "$@"
}

IMPORT=github.com/gogo/protobuf/protoc-gen-gogo
PROTOC_GEN_GOGO=$(vgo list -f '{{ .Target }}' "${IMPORT}")
vgo install -v $IMPORT

log "generating protobufs for $(vgo list .)..."

INCLUDE=$(dirname "$(vgo list -f '{{ .Dir }}' "${IMPORT}")")
protoc --plugin=protoc-gen-gogo="${PROTOC_GEN_GOGO}" -I"${INCLUDE}" -I. --gogo_out=. ./*.proto

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
