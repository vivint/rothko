#!/usr/bin/env bash

if [ ! -z "$( go list github.com/vivint/rothko 2>/dev/null )" ]; then
	exit 0
fi

cat <<EOF
An attempt to use vgo and contribute to its development has put
this project in a weird migration spot. While building a large part
of the project is possible with vgo, tools such as godocdown and
gopherjs don't yet know how to work with vgo. Thus, in order for
these tools to work, you must create a GOPATH with both the
github.com/vivint/rothko and github.com/gopherjs/gopherjs repos
installed. For example, running

	mkdir rothko
	cd rothko
	export GOPATH=\`pwd\`
	go get github.com/vivint/rothko github.com/gopherjs/gopherjs
	cd src/github.com/vivint/rothko
	vgo vendor
	rm -rf vendor/github.com/gopherjs/gopherjs

will get you into an appropriate state.

Sorry for the inconvenience.
EOF

exit 1
