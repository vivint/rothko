// Copyright (C) 2017. See AUTHORS.

package mmap

import "syscall"

const (
	PROT_READ  = syscall.PROT_READ
	PROT_WRITE = syscall.PROT_WRITE
	MAP_SHARED = syscall.MAP_SHARED
	MS_SYNC    = syscall.MS_SYNC
	MS_ASYNC   = syscall.MS_ASYNC
)

func Mmap(fd int, length int, prot int, flags int) (data uintptr, err error) {
	data, _, ec := syscall.Syscall6(syscall.SYS_MMAP, 0,
		uintptr(length), uintptr(prot), uintptr(flags), uintptr(fd), 0)
	if ec != 0 {
		return 0, Error.Wrap(ec)
	}
	return data, nil
}

func Munmap(data uintptr, length int) (err error) {
	_, _, ec := syscall.Syscall(syscall.SYS_MUNMAP, data, uintptr(length), 0)
	if ec != 0 {
		return Error.Wrap(err)
	}
	return nil
}

func Msync(data uintptr, length int, flags int) (err error) {
	_, _, ec := syscall.Syscall(syscall.SYS_MSYNC, data, uintptr(length),
		uintptr(flags))
	if ec != 0 {
		return Error.Wrap(err)
	}
	return nil
}
