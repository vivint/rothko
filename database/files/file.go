// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"
	"os"

	"github.com/spacemonkeygo/rothko/database/files/internal/meta"
	"github.com/spacemonkeygo/rothko/database/files/internal/system"
)

//
// the file implementation goes to great lengths to be efficient: it avoids
// allocations in common operations as much as possible, the struct layout
// does not contain any pointers, none of the methods are mutating so that
// they may be passed as values cheaply.
//

// file represents a buffer of records mmaped into memory
type file struct {
	data uintptr // mmap'd data. stored as a uintptr to avoid gc pressure.
	len  int     // length of mmap'd data
	cap  int     // capacity (in records) of the data excluding metadata
	size int     // alignment size of each record
	ref  ref     // for debugging leaks
}

// createFile creates a file at the given path with the given record size.
// the file is allocated with the ability to store cap records without a
// resize.
func createFile(ctx context.Context, path string, size, cap int) (
	f file, err error) {

	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return file{}, Error.Wrap(err)
	}
	defer fh.Close()
	fd := int(fh.Fd())

	// these overflow checks might be too restrictive, but i think it will
	// go up to 1GB at least, so meh that's probably good enough. we can
	// revisit making them up to a larger size size later, though the system
	// will probably struggle with that anyway.
	if cap+1 < cap ||
		int(int32(cap)) != cap ||
		int(int32(size)) != size ||
		int32(size)*int32(cap)/int32(cap) != int32(size) {

		return file{}, Error.New("capacity too large")
	}

	len := size * (cap + 1)
	if err := system.Allocate(fd, int64(len)); err != nil {
		return file{}, err
	}

	data, err := system.Mmap(fd, len,
		system.PROT_READ|system.PROT_WRITE,
		system.MAP_SHARED)
	if err != nil {
		return file{}, Error.Wrap(err)
	}

	err = writeMetadata(slice(data, len)[:size], meta.Metadata{
		Size_: size,
		Head:  cap - 1,
	})
	if err != nil {
		return file{}, err
	}

	return file{
		data: data,
		len:  len,
		cap:  cap,
		size: size,
		ref:  newRef(path),
	}, nil
}

// openFile returns a file for the given path.
func openFile(ctx context.Context, path string) (f file, err error) {
	fh, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return f, Error.Wrap(err)
	}
	defer fh.Close()
	fd := int(fh.Fd())

	fi, err := fh.Stat()
	if err != nil {
		return f, Error.Wrap(err)
	}
	len := int(fi.Size())

	if len < recordHeaderSize {
		return f, Error.New("file is too small to contain metadata")
	}

	data, err := system.Mmap(fd, len,
		system.PROT_READ|system.PROT_WRITE,
		system.MAP_SHARED)
	if err != nil {
		return f, Error.Wrap(err)
	}

	// read the metadata record to determine the size of the records in this
	// file.
	meta, err := readMetadata(slice(data, len))
	if err != nil {
		return f, err
	}

	if meta.Size_ < recordHeaderSize {
		return f, Error.New("possible corruption: invalid size")
	}

	return file{
		data: data,
		len:  len,
		cap:  len/meta.Size_ - 1,
		size: meta.Size_,
		ref:  newRef(path),
	}, nil
}

// slice returns the data slice for the file
func (f file) slice() []byte { return slice(f.data, f.len) }

// offset computes the byte offset for the nth record.
func (f file) offset(n int) int {
	return (n + 1) * f.size
}

// Nil returns if the file is nil.
func (f file) Nil() bool { return f.data != 0 }

// Size returns the maximum size of a record.
func (f file) Size() int { return f.size }

// Capacity returns the capacity of records in the file.
func (f file) Capacity() int { return f.cap }

// Close releases all the of resources for the file.
func (f file) Close() error {
	if f.data != 0 {
		f.ref.Close()
		return system.Munmap(f.data, f.len)
	}
	return nil
}

// Metadata returns the metadata record.
func (f file) Metadata(ctx context.Context) (m meta.Metadata, err error) {
	return readMetadata(f.slice()[:f.size])
}

// SetMetadata sets the metadata record.
func (f file) SetMetadata(ctx context.Context, m meta.Metadata) (err error) {
	return writeMetadata(f.slice()[:f.size], m)
}

// Record returns the nth record.
func (f file) Record(ctx context.Context, n int) (out record, err error) {
	if n >= f.cap || n < 0 {
		return out, Error.New("record out of bounds")
	}
	off := f.offset(n)
	return readRecord(f.slice()[off : off+f.size])
}

// SetRecord stores the record in the nth slot.
func (f file) SetRecord(ctx context.Context, n int, rec record) (err error) {
	if n >= f.cap || n < 0 {
		return Error.New("record out of bounds")
	}
	off := f.offset(n)
	return writeRecord(f.slice()[off:off+f.size], rec)
}

// HasRecord returns if there is a record stored at the index.
func (f file) HasRecord(ctx context.Context, n int) (ok bool) {
	if n >= f.cap || n < 0 {
		return false
	}

	// this relies on the first byte of the serialized record containing a
	// non-zero version.
	off := f.offset(n)
	return f.slice()[off] != 0
}

// FullSync causes the file's contents to be synced to disk.
func (f file) FullSync(ctx context.Context) (err error) {
	return system.Msync(f.data, f.len, system.MS_SYNC)
}

// FullAsync causes the file's contents to be synced to disk asynchronously.
func (f file) FullAsync(ctx context.Context) (err error) {
	return system.Msync(f.data, f.len, system.MS_ASYNC)
}
