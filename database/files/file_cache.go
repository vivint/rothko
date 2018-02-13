// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"
	"sync"
)

// fileCacheOptions describes the options a file cache can be made with.
type fileCacheOptions struct {
	// Handles is the number of handles in the cache.
	Handles int

	// Size is the size of each record in a newly allocated file.
	Size int

	// Cap is the number of records a file will be allocated with.
	Cap int
}

// fileCache is a cache on files that maps paths to files with acquire and
// release semantics.
type fileCache struct {
	opts fileCacheOptions

	mu   sync.Mutex
	toks map[string]cacheToken
	ch   *cache
}

// newFileCache constructs a file cache.
func newFileCache(opts fileCacheOptions) *fileCache {
	return &fileCache{
		toks: make(map[string]cacheToken),
		ch:   newCache(opts.Handles),

		opts: opts,
	}
}

// releaseFile puts the file back into the cache, closing any evicted file.
func (fch *fileCache) releaseFile(path string, f file) {
	fch.mu.Lock()
	tok, ev, ok := fch.ch.Put(f)
	fch.toks[path] = tok
	fch.mu.Unlock()

	// TODO(jeff): do we want to call sync or anything?
	if ok {
		ev.Close()
	}
}

// acquireFile opens or creates the file at the path. it is expected to be
// called exclusive to all others that might be interested in the path.
func (fch *fileCache) acquireFile(ctx context.Context, path string,
	exists bool) (f file, err error) {

	fch.mu.Lock()
	tok, ok := fch.toks[path]
	if ok {
		f, ok = fch.ch.Take(tok)
	}
	fch.mu.Unlock()

	if ok {
		return f, nil
	}

	if exists {
		return openFile(ctx, path)
	}
	return createFile(ctx, path, fch.opts.Size, fch.opts.Cap)
}

// evictFile removes the file from the cache if it exists.
func (fch *fileCache) evictFile(path string) {
	var f file

	fch.mu.Lock()
	tok, ok := fch.toks[path]
	if ok {
		delete(fch.toks, path)
		f, ok = fch.ch.Take(tok)
	}
	fch.mu.Unlock()

	if ok {
		f.Close()
	}
}
