#!/usr/bin/env bash

# godocdown provided by https://github.com/robertkrimen/godocdown
# which sadly cannot be inserted into Gopkg.toml for some reason:
# 	github.com/robertkrimen/godocdown/godocdown has err (*pkgtree.LocalImportsError); required by (root).
# oh well.

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PACKAGES=$(go list github.com/spacemonkeygo/rothko/...)

for PACKAGE in $PACKAGES; do
	if [ "$PACKAGE" == "github.com/spacemonkeygo/rothko" ]; then
		continue
	fi
	DIR=$(go list -f '{{.Dir}}' "$PACKAGE")
	godocdown -template "$SCRIPTDIR/godocdown.template" "$PACKAGE" > "$DIR/README.md"
done
