// Copyright (C) 2018. See AUTHORS.

package system

import (
	"syscall"
)

func Allocate(fd int, length int64) (err error) {
	const (
		FALLOC_FL_COLLAPSE_RANGE = 0x8
		FALLOC_FL_INSERT_RANGE   = 0x20
		FALLOC_FL_KEEP_SIZE      = 0x1
		FALLOC_FL_NO_HIDE_STALE  = 0x4
		FALLOC_FL_PUNCH_HOLE     = 0x2
		FALLOC_FL_UNSHARE_RANGE  = 0x40
		FALLOC_FL_ZERO_RANGE     = 0x10
	)

	const mode = 0
	if err := syscall.Fallocate(fd, mode, 0, length); err != nil {
		return Error.Wrap(err)
	}

	return nil
}
