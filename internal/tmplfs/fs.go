// Copyright (C) 2018. See AUTHORS.

package tmplfs

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// FS wraps a http.FileSystem to add templates to files that end in .html.
type FS struct {
	fs http.FileSystem
}

// New constructs a FS around the http.FileSystem. It implements http.Handler.
func New(fs http.FileSystem) *FS {
	return &FS{
		fs: fs,
	}
}

// ServeHTTP conforms to the http.Handler interface.
func (s *FS) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.FileServer(&wrapper{
		fs:  s.fs,
		req: req,
	}).ServeHTTP(w, req)
}

// wrapper wraps the http.FileSystem in the context of a request.
type wrapper struct {
	fs  http.FileSystem
	req *http.Request
}

// Open implements http.FileSystem and processes any html files as a template.
func (w *wrapper) Open(name string) (f http.File, err error) {
	f, err = w.fs.Open(name)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(name, ".html") {
		return f, nil
	}

	// close only if we have any errors
	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	// TODO(jeff): caching of the parsed template?

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("").Parse(string(data))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]interface{}{
		"query":  w.req.URL.RawQuery,
		"metric": w.req.FormValue("metric"),
		"dev":    os.Getenv("ROTHKO_DEV") != "",
	})
	if err != nil {
		return nil, err
	}

	return &fileWrapper{
		File: f,
		size: int64(buf.Len()),
		data: bytes.NewReader(buf.Bytes()),
	}, nil
}

// fileWrapper wraps an http.File with overridden size and data.
type fileWrapper struct {
	http.File
	size int64
	data *bytes.Reader
}

// Read dispatches to the *bytes.Reader.
func (f *fileWrapper) Read(p []byte) (int, error) {
	return f.data.Read(p)
}

// Seek dispatches to the *bytes.Reader.
func (f *fileWrapper) Seek(offset int64, whence int) (int64, error) {
	return f.data.Seek(offset, whence)
}

// Stat returns a wrapped os.FileInfo with an updated mtime and size.
func (f *fileWrapper) Stat() (os.FileInfo, error) {
	info, err := f.File.Stat()
	if err != nil {
		return nil, err
	}

	return fileInfoWrapper{
		FileInfo: info,
		size:     f.size,
		mod:      time.Now(),
	}, nil
}

// fileInfoWrapper wraps an os.FileInfo with a different size and mtime.
type fileInfoWrapper struct {
	os.FileInfo
	size int64
	mod  time.Time
}

// ModTime returns the overridden mtime.
func (f fileInfoWrapper) ModTime() time.Time { return f.mod }

// Size returns the overridden size.
func (f fileInfoWrapper) Size() int64 { return f.size }
