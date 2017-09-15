// Copyright (C) 2017. See AUTHORS.

// +build ignore

package files

type cacheToken int64

type cacheNode struct {
	loc int
	tok cacheToken
	f   *file
}

type cache struct {
	tok     cacheToken
	cap     int
	order   []cacheToken // evict order is last element first
	handles map[cacheToken]cacheNode
}

func newCache(cap int) *cache {
	return &cache{
		num:     0,
		cap:     cap,
		order:   make([]cacheToken, 0, cap),
		handles: make(map[cacheToken]*file, cap),
	}
}

func (c *cache) last() (cacheNode, bool) {
	last := len(c.order)
	if len(c.order) == 0 {
		return cacheNode{}, false
	}
	return c.handles[c.order[last-1]], true
}

func (c *cache) get(tok int) *file {
	node, ok := c.handles[tok]
	if !ok {
		return nil
	}
	last, ok := c.last()
	if !ok {
		return nil
	}

	// swap the locations of the last and the node
	last.loc, node.loc = node.loc, last.loc
	c.order[last.loc], c.order[node.loc] = c.order[node.loc], c.order[last.loc]

	// update the cacheNode inside of the map
	c.handles[last.id] = last
	c.handles[node.id] = node

	return node.f
}
