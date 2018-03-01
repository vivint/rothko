#!/usr/bin/env bash

set -e

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PACKAGES=$(vgo list github.com/vivint/rothko/...)

IMPORT=github.com/robertkrimen/godocdown/godocdown
GODOCDOWN=$(vgo list -f '{{ .Target }}' "${IMPORT}")
vgo install -v "${IMPORT}"

SED="sed"
case $(uname) in
	Darwin )
		SED="gsed"
		;;
esac

for PACKAGE in $PACKAGES; do
	if [ "$PACKAGE" == "github.com/vivint/rothko" ]; then
		continue
	fi
	if [ "$PACKAGE" == "github.com/vivint/rothko/ui" ]; then
		continue
	fi

	DIR=$(vgo list -f '{{.Dir}}' "$PACKAGE")
	"${GODOCDOWN}" -template "$SCRIPTDIR/godocdown.template" "$DIR" > "$DIR/README.md"
	# because we used the dir rather than the package (for vgo reasons), we
	# need to fix up the import line
	"$SED" -i "s#import \"\.\"#import \"$PACKAGE\"#g" -- "$DIR/README.md"
done
