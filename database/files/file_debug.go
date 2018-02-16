// Copyright (C) 2018. See AUTHORS.

// +build debug

package files

import (
	"runtime"

	"github.com/spacemonkeygo/rothko/external"
)

// in a debug build, ref keeps track of if it was closed, and spews into the
// error log if it is finalized without being closed. in the non-debug build,
// it has zero size and does nothing, so that Go can inline and remove it from
// the compiled output.

type ref struct {
	path   string
	closed *bool
}

func newRef(path string) (r ref) {
	closed := new(bool)
	runtime.SetFinalizer(closed, func(closed *bool) {
		if !*closed {
			external.Errorw("leaked",
				"path", path,
			)
		}
	})

	return ref{
		closed: closed,
	}
}

func (r ref) Close() {
	runtime.SetFinalizer(r.closed, nil)
	*r.closed = true
}
