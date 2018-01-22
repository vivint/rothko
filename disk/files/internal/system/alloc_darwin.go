// Copyright (C) 2018. See AUTHORS.

package system

import (
	"syscall"
)

func Allocate(fd int, length int64) (err error) {
	// there may be way to create sparse files on os x, but i don't know it.
	if err := syscall.Ftruncate(fd, length); err != nil {
		return Error.Wrap(err)
	}

	return nil
}
