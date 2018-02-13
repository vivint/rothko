// Copyright (C) 2018. See AUTHORS.

package tgzfs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/zeebo/errs"
)

// rootMtime is the mtime for the root file.
var rootMtime = time.Now()

// rootFileInfo is a fake file info for the root of the tree.
type rootFileInfo struct{}

func (r rootFileInfo) Name() string       { return "" }
func (r rootFileInfo) Size() int64        { return 0 }
func (r rootFileInfo) Mode() os.FileMode  { return os.ModeDir }
func (r rootFileInfo) ModTime() time.Time { return rootMtime }
func (r rootFileInfo) IsDir() bool        { return true }
func (r rootFileInfo) Sys() interface{}   { return nil }

// FS is an http.FileServer for a tarball.
type FS struct {
	root *file
	def  *file
}

// New constructs a FS from a gzip encoded tar ball in the data.
func New(data []byte) (*FS, error) {
	tardata, err := tar.NewReader(gzip.NewReader(bytes.NewReader(data)))
	if err != nil {
		return nil, errs.Wrap(err)
	}

	fs := &FS{
		root: newFile(),
	}
	fs.root.setFileInfo(rootFileInfo{})

	for {
		h, err := tardata.Next()
		if err == io.EOF {
			return fs, nil
		}
		if err != nil {
			return nil, err.Wrap(err)
		}

		file := fs.getFile(h.Name)
		file.setFileInfo(h.FileInfo())

		data, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, errs.Wrap(err)
		}
		file.setData(data)

		switch h.Name {
		case "index.html", "./index.html":
			fs.def = file
		}
		return nil
	}
}

// splitPath splits a cleaned version of the path. It assumes unix style path
// separators.
func splitPath(p string) []string {
	return strings.Split(path.Clean(p), "/")
}

// getFile returns the *file at the given path, allocating file structures
// along the way if alloc is true. If alloc is false, then the default file
// is returned if there is no file.
func (fs *FS) getFile(path string, alloc bool) *file {
	child := fs.root
	for _, part := range splitPath(path) {
		if len(part) == 0 {
			continue
		}
		children := child.children
		child = children[part]
		if child == nil {
			if !alloc {
				return fs.def
			}
			child = newFile()
			children[part] = child
		}
	}
	return child
}

// Open returns an http.File for the given path.
func (fs *FS) Open(name string) (http.File, error) {
	file := fs.lookupFile(name)
	if file == nil {
		return nil, os.ErrNotExist
	}
	return file.open(), nil
}
