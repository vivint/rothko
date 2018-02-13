// Copyright (C) 2018. See AUTHORS.

package tgzfs

import (
	"bytes"
	"io"
	"os"
)

// file represents a file in the tarball.
type file struct {
	fi       os.FileInfo
	data     []byte
	children map[string]*file
}

// newFile constructs a new file.
func newFile() *file {
	return &file{
		children: make(map[string]*file),
	}
}

// setFileInfo sets the info struct for the file.
func (f *file) setFileInfo(fi os.FileInfo) {
	f.fi = fi
}

// setData sets the contents of the file.
func (f *file) setData(data []byte) {
	f.data = data
}

// open allocates a file handle for the data, and keeps track of a directory
// listing as well as file info about the file.
func (f *file) open() *fileHandle {
	return &fileHandle{
		f:    f,
		data: bytes.NewReader(f.data),
	}
}

// fileHandle represents an open file handle.
type fileHandle struct {
	f    *file
	data *bytes.Reader
}

// Close closes the fileHandle.
func (h *fileHandle) Close() error {
	return nil
}

// Read reads data from the file into the provided buffer.
func (h *fileHandle) Read(p []byte) (int, error) {
	return h.data.Read(p)
}

// Readdir reads up to count entries from the listing and returns them.
func (h *fileHandle) Readdir(count int) (entries []os.FileInfo, err error) {
	if len(h.listing) == 0 {
		return nil, io.EOF
	}
	if len(h.listing) < count {
		count = len(h.listing)
	}
	listing := h.listing[:count]
	h.listing = h.listing[:len(h.listing)-len(listing)]
	return listing, nil
}

// Seek sets the position into the file handle.
func (h *fileHandle) Seek(offset int64, whence int) (int64, error) {
	return h.data.Seek(offset, whence)
}

// Stat returns information about the file handle.
func (f *fileHandle) Stat() (os.FileInfo, error) {
	return f.fi, nil
}
