// Copyright (C) 2017. See AUTHORS.

package system

import (
	"syscall"
	"unsafe"
)

const nameOffset = uint64(unsafe.Offsetof(syscall.Dirent{}.Name))

func direntReadIno(buf []byte) (uint64, bool) {
	return readInt(buf,
		unsafe.Offsetof(syscall.Dirent{}.Ino),
		unsafe.Sizeof(syscall.Dirent{}.Ino))
}

func direntReadReclen(buf []byte) (uint64, bool) {
	return readInt(buf,
		unsafe.Offsetof(syscall.Dirent{}.Reclen),
		unsafe.Sizeof(syscall.Dirent{}.Reclen))
}

func direntReadNamlen(buf []byte) (uint64, bool) {
	return readInt(buf,
		unsafe.Offsetof(syscall.Dirent{}.Namlen),
		unsafe.Sizeof(syscall.Dirent{}.Namlen))
}
