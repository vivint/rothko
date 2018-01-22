// Copyright (C) 2018. See AUTHORS.

package system

import (
	"encoding/binary"
)

func readInt(b []byte, off, size uintptr) (u uint64, ok bool) {
	if len(b) < int(off+size) {
		return 0, false
	}
	switch size {
	case 1:
		return uint64(b[off]), true
	case 2:
		return uint64(binary.LittleEndian.Uint16(b[off:])), true
	case 4:
		return uint64(binary.LittleEndian.Uint32(b[off:])), true
	case 8:
		return uint64(binary.LittleEndian.Uint64(b[off:])), true
	default:
		return 0, false
	}
}

func NextDirent(buf []byte) (out_buf []byte, name []byte, ok bool) {
	// if we can't determine a valid record length, we have fully consumed
	// the buffer.
	rec_len, ok := direntReadReclen(buf)
	if !ok || rec_len > uint64(len(buf)) {
		return nil, nil, false
	}

	// pull off the record and reslice the buffer
	rec := buf[:rec_len]
	buf = buf[rec_len:]

	// check that the inode is non-zero
	ino, ok := direntReadIno(rec)
	if !ok || ino == 0 {
		return buf, nil, false
	}

	// determine the name
	name_len, ok := direntReadNamlen(rec)
	if !ok || nameOffset+name_len > uint64(len(rec)) {
		return buf, nil, false
	}

	// truncate the name to the first zero byte
	name = rec[nameOffset : nameOffset+name_len]
	for i, c := range name {
		if c == 0 {
			name = name[:i]
			break
		}
	}

	return buf, name, string(name) != "." && string(name) != ".."
}
