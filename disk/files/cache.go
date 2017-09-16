// Copyright (C) 2017. See AUTHORS.

package files

//
// this cache has some peculiar properties. it's close in spirit to an MRU
// cache in that the most recently put entry will be evicted when a new put
// is performed and there is no extra capacity, but we do not attempt to
// strictly follow that rule recursively. that allows us to avoid much of the
// pointer chasing required when using a linked list to store the order, and
// avoid copying using a backing slice because we just swap whoever is the
// newest value with whatever value is taken, rather than move the entire
// section down.
//
// For example, if we had the order
//
//    1 2 3 4 5
//
// Then 5 would be first to be evicted. If we then took 2 from that list,
// the order, in an MRU would be
//
//    1 3 4 5
//
// but with this implementation, it will be
//
//    1 5 3 4
//
// which is cheaper to perform, but causes 5 to potentially live longer.
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
}

// newCache constructs a cache with the given capacity.
func newCache(cap int) *cache {
	return &cache{
		cap:     cap,
		order:   make([]cacheToken, 0, cap),
		handles: make(map[cacheToken]cacheEntry, cap),
	}
}

// last returns the last cacheEntry: the one that would be evicted first.
func (c *cache) last() cacheEntry {
	return c.handles[c.order[len(c.order)-1]]
}

// Take retrieves the file from the cache. If the file has been evicted, then
// ok will be false.
func (c *cache) Take(tok cacheToken) (f file, ok bool) {
	entry, ok := c.handles[tok]
	if !ok {
		return file{}, false
	}
	last := c.last()

	// last takes the found entry's position
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

	// if we don't have capacity, we reduce one from the entry's location
	// because it will be assigned to the last slot in the list.
	entry.loc--

	// if we never have capacity, return that we evicted the file that was put.
	if cap(c.order) == 0 {
		return entry.tok, f, true
	}

	// otherwise, we take the place of and evict the last entry.
	last := c.last()

	delete(c.handles, last.tok)
	c.handles[entry.tok] = entry
	c.order[last.loc] = entry.tok

	return entry.tok, last.f, true
}
