// Copyright (C) 2018. See AUTHORS.

package system

import (
	"os"
	"syscall"
	"unsafe"
)

func Open(path []byte) (fd uintptr, err error) {
	if len(path) == 0 || path[len(path)-1] != 0 {
		return 0, Error.New("invalid path: %q", path)
	}

	ptr := unsafe.Pointer(&path[0])
	fd, _, ec := syscall.Syscall(syscall.SYS_OPEN,
		uintptr(ptr), uintptr(os.O_RDONLY), 0)
	if ec != 0 {
		return 0, Error.Wrap(ec)
	}
	return fd, nil
}

func Close(fd uintptr) (err error) {
	return Error.Wrap(syscall.Close(int(fd)))
}
