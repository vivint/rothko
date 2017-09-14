// Copyright (C) 2017. See AUTHORS.

package files

import (
	"os"
	"syscall"

	"github.com/spacemonkeygo/errors"
)

// file represents a buffer of records mmaped into memory
type file struct {
	fh   *os.File // used to remap
	data []byte   // mmap'd data
	size int      // alignment size of each record
	len  int      // length (in records) of the data excluding metadata
	buf  []byte   // buffer that can hold a record
}

// open returns a file for the given path
func open(path string, size int) (f file, err error) {
	fh, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return f, Error.Wrap(err)
	}

	fi, err := fh.Stat()
	if err != nil {
		fh.Close()
		return f, Error.Wrap(err)
	}

	f_size := int(fi.Size())
	if f_size < size {
		f_size = size
	}

	data, err := syscall.Mmap(int(fh.Fd()), 0, f_size,
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		fh.Close()
		return f, Error.Wrap(err)
	}

	return file{
		fh:   fh,
		data: data,
		size: size,
		len:  len(data)/size - 1,
		buf:  make([]byte, size),
	}, nil
}

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

// metadata returns the metadata record.
func (f file) metadata() (out record, err error) {
	return parse(f.data[:f.size])
}

// get returns the nth record.
func (f file) get(n int) (out record, err error) {
	if n >= f.len {
		return out, Error.New("file: out of bounds")
	}
	off := f.offset(n)
	return parse(f.data[off : off+f.size])
}

// put stores the record in the nth slot.
func (f *file) put(n int, rec record) (err error) {
	if n >= f.len {
		if err := f.truncate(n + 1); err != nil {
			return err
		}
	}

	// TODO(jeff): this is either safe with a copy, or dangerous without one.
	// let's go with dangerous for now. as long as the record marshals to less
	// than the size, we're good! maybe we can just error if that won't be the
	// case... should be cheap.

	off := f.offset(n)
	rec.Marshal(f.data[off:off])
	return nil
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
