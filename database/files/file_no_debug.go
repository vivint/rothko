// Copyright (C) 2018. See AUTHORS.

// +build !debug

package files

// in a debug build, ref keeps track of if it was closed, and spews into the
// error log if it is finalized without being closed. in the non-debug build,
// it has zero size and does nothing, so that Go can inline and remove it from
// the compiled output.

type ref struct{}

func newRef(path string) (r ref) { return r }

func (r ref) Close() {}
