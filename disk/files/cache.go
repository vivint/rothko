// Copyright (C) 2017. See AUTHORS.

package files

import (
	"github.com/spacemonkeygo/rothko/internal/pcg"
)

//
// this cache does a random eviction strategy. it is very cheap and requires
// almost no pointer chasing. since we frequently access every single file
// in the system during writes, i do not believe there is a better caching
// strategy.
//

// cacheToken is a token for retrieving an entry from the cache.
type cacheToken int64

// cacheEntry holds information about an item in the cache: it's location in
// the eviction order, the token that can retrieve it, and the value itself.
// it vary carefully does not contain any pointers.
type cacheEntry struct {
	loc int
	tok cacheToken
	f   file
}

// cache holds on to a set of files using an eviction strategy that would
// surprise you.
type cache struct {
	tok     cacheToken
	cap     int
	order   []cacheToken // evict order is last element first
	handles map[cacheToken]cacheEntry
	pcg     pcg.PCG
}

// newCache constructs a cache with the given capacity.
func newCache(cap int) *cache {
	return &cache{
		cap:     cap,
		order:   make([]cacheToken, 0, cap),
		handles: make(map[cacheToken]cacheEntry, cap),
		pcg:     pcg.New(0, 0), // we could introduce entropy here
	}
}

// Take retrieves the file from the cache. If the file has been evicted, then
// ok will be false.
func (c *cache) Take(tok cacheToken) (f file, ok bool) {
	entry, ok := c.handles[tok]
	if !ok {
		return file{}, false
	}

	// the last entry takes the found entry's position
	last := c.handles[c.order[len(c.order)-1]]
	last.loc = entry.loc
	c.order[entry.loc] = last.tok
	c.handles[last.tok] = last

	// entry is evacuated from the cache
	delete(c.handles, entry.tok)
	c.order = c.order[:len(c.order)-1]

	return entry.f, true
}

// Put sticks the file in the cache and returns a token that can be used to
// retrieve it later, a file that was evicted, and an ok boolean to signal if
// a file was actually evicted.
func (c *cache) Put(f file) (tok cacheToken, ev file, ok bool) {
	c.tok++
	entry := cacheEntry{
		loc: len(c.order),
		tok: c.tok,
		f:   f,
	}

	// if we have capacity, just add it and don't evict anything.
	if len(c.order) < cap(c.order) {
		c.handles[entry.tok] = entry
		c.order = append(c.order, entry.tok)

		return entry.tok, ev, false
	}

	// if we never have capacity, return that we evicted the file that was put.
	if cap(c.order) == 0 {
		return entry.tok, f, true
	}

	// if we don't have capacity, we will evict a random entry and this new
	// entry will take its spot.
	entry.loc = fastMod(c.pcg.Uint32(), c.cap)
	current := c.handles[c.order[entry.loc]]

	delete(c.handles, current.tok)
	c.handles[entry.tok] = entry
	c.order[current.loc] = entry.tok

	return entry.tok, current.f, true
}
