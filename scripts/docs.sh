#!/usr/bin/env bash

set -e

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
"${SCRIPTDIR}/check-vgo.sh" || exit $?

PACKAGES=$(vgo list github.com/vivint/rothko/...)

IMPORT_PATH=github.com/robertkrimen/godocdown/godocdown
GODOCDOWN=$(vgo list -f '{{ .Target }}' "${IMPORT_PATH}")

vgo install -v $IMPORT_PATH

for PACKAGE in $PACKAGES; do
	if [ "$PACKAGE" == "github.com/vivint/rothko" ]; then
		continue
	fi
	if [ "$PACKAGE" == "github.com/vivint/rothko/ui" ]; then
		continue
	fi

	DIR=$(vgo list -f '{{.Dir}}' "$PACKAGE")
	"${GODOCDOWN}" -template "$SCRIPTDIR/godocdown.template" "$PACKAGE" > "$DIR/README.md"
done
