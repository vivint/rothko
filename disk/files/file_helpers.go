// Copyright (C) 2017. See AUTHORS.

package files

import (
	"unsafe"

	"github.com/spacemonkeygo/rothko/disk/files/internal/meta"
)

// slice returns a slice of the data with the given length. the data MUST NOT
// point at go allocated memory.
func slice(data uintptr, length int) []byte {
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{data, length, length}

	return *(*[]byte)(unsafe.Pointer(&sl))
}

// writeMetadata writes a metadata value into the buffer, ensuring that it
// fits in the size.
func writeMetadata(buf []byte, m meta.Metadata) (err error) {
	// this is a little gnarly to avoid allocations.

	size := m.Size()
	if int(uint16(size)) != size {
		return Error.New("metadata too large")
	}

	rec := record{
		version: recordVersion,
		kind:    recordKind_complete,
		size:    uint16(size),
	}
	if rec.Size() > len(buf) {
		return Error.New("metadata too large")
	}

	out := rec.MarshalHeader(buf[:0])
	_, err = m.MarshalTo(buf[len(out):])

	return Error.Wrap(err)
}

// readMetadata reads a metadata value from the buffer.
func readMetadata(buf []byte) (m meta.Metadata, err error) {
	rec, err := readRecord(buf)
	if err != nil {
		return m, err
	}
	if err := m.Unmarshal(rec.data); err != nil {
		return m, Error.Wrap(err)
	}
	return m, nil
}

// writeRecord writes a record value into the buffer, ensuring that it fits in
// the buffer.
func writeRecord(buf []byte, rec record) (err error) {
	if rec.Size() > len(buf) {
		return Error.New("record too large")
	}
	rec.Marshal(buf[:0])
	return nil
}

// readRecord reads a record value from the buffer
func readRecord(buf []byte) (rec record, err error) {
	return parse(buf)
}
