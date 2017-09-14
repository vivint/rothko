// Copyright (C) 2017. See AUTHORS.

package files

import (
	"os"
	"syscall"

	"github.com/spacemonkeygo/errors"
	"github.com/spacemonkeygo/rothko/disk/files/internal/meta"
)

// file represents a buffer of records mmaped into memory
type file struct {
	fh   *os.File // used to remap
	data []byte   // mmap'd data
	size int      // alignment size of each record
	len  int      // length (in records) of the data excluding metadata
	buf  []byte   // buffer that can hold a record
}

// create creates a file at the given path with the given size and metadata.
// the file is allocated with the ability to store cap records.
func create(path string, size, cap int) (f file, err error) {
	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return f, Error.Wrap(err)
	}

	// TODO(jeff): overflow detection?
	trunc := size * (cap + 1)

	if err := fh.Truncate(int64(trunc)); err != nil {
		fh.Close()
		return f, Error.Wrap(err)
	}

	data, err := syscall.Mmap(int(fh.Fd()), 0, trunc,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		fh.Close()
		return f, Error.Wrap(err)
	}

	err = writeMetadata(data[:size], meta.Metadata{
		Size_: size,
		Head:  0,
	})
	if err != nil {
		fh.Close()
		return f, err
	}

	return file{
		fh:   fh,
		data: data,
		size: size,
		len:  0,
		buf:  make([]byte, size),
	}, nil
}

// open returns a file for the given path.
func open(path string) (f file, err error) {
	fh, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return f, Error.Wrap(err)
	}

	fi, err := fh.Stat()
	if err != nil {
		fh.Close()
		return f, Error.Wrap(err)
	}

	if fi.Size() < recordHeaderSize {
		fh.Close()
		return f, Error.New("file is too small to contain metadata")
	}

	data, err := syscall.Mmap(int(fh.Fd()), 0, int(fi.Size()),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		fh.Close()
		return f, Error.Wrap(err)
	}

	// read the metadata record to determine the size of the records in this
	// file.
	meta, err := readMetadata(data)
	if err != nil {
		fh.Close()
		return f, err
	}

	if meta.Size_ < recordHeaderSize {
		fh.Close()
		return f, Error.New("possible corruption: invalid size")
	}

	return file{
		fh:   fh,
		data: data,
		size: meta.Size_,
		len:  len(data)/meta.Size_ - 1,
		buf:  make([]byte, meta.Size_),
	}, nil
}

// Size returns the maximum size of a record.
func (f file) Size() int { return f.size }

// Close releases all the of resources for the file.
func (f file) Close() error {
	var eg errors.ErrorGroup
	eg.Add(f.fh.Close())
	eg.Add(syscall.Munmap(f.data))
	return eg.Finalize()
}

// offset computes the byte offset for the nth record.
func (f file) offset(n int) int {
	return (n + 1) * f.size
}

// Metadata returns the metadata record.
func (f file) Metadata() (m meta.Metadata, err error) {
	return readMetadata(f.data[:f.size])
}

// SetMetadata sets the metadata record.
func (f file) SetMetadata(m meta.Metadata) (err error) {
	return writeMetadata(f.data[:f.size], m)
}

// Record returns the nth record.
func (f file) Record(n int) (out record, err error) {
	if n >= f.len {
		return out, Error.New("record out of bounds")
	}
	off := f.offset(n)
	return readRecord(f.data[off : off+f.size])
}

// SetRecrod stores the record in the nth slot.
func (f *file) SetRecord(n int, rec record) (err error) {
	if n >= f.len {
		if err := f.truncate(n + 1); err != nil {
			return err
		}
	}

	off := f.offset(n)
	return writeRecord(f.data[off:off+f.size], rec)
}

// truncate causes the file to accomodate n records (and one metadata record).
func (f *file) truncate(n int) (err error) {
	size := f.offset(n) + f.size

	if err := syscall.Munmap(f.data); err != nil {
		return Error.Wrap(err)
	}

	if err := f.fh.Truncate(int64(size)); err != nil {
		return Error.Wrap(err)
	}

	data, err := syscall.Mmap(int(f.fh.Fd()), 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return Error.Wrap(err)
	}

	f.data = data
	f.len = n
	return nil
}
