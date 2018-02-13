// Copyright (C) 2018. See AUTHORS.

package sset

// Set represents an ordered set of strings.
type Set struct {
	set   map[string]struct{}
	order []string
}

// New constructs a Set with the given initial capacity.
func New(cap int) *Set {
	return &Set{
		set:   make(map[string]struct{}, cap),
		order: make([]string, 0, cap),
	}
}

// Len returns the amount of elements in the set.
func (s *Set) Len() int { return len(s.order) }

// Has returns if the set has the key.
func (s *Set) Has(x string) bool {
	_, ok := s.set[x]
	return ok
}

func (s *Set) Add(x string) {
	// if it's already in the set, do nothing
	if _, ok := s.set[x]; ok {
		return
	}

	// add it to the set
	s.set[x] = struct{}{}

	// determine the insertion point for x
	i, j := 0, len(s.order)
	for i < j {
		h := int(uint(i+j) >> 1)
		if s.order[h] < x {
			i = h + 1
		} else {
			j = h
		}
	}

	// insert it
	s.order = append(s.order, x)
	copy(s.order[i+1:], s.order[i:])
	s.order[i] = x
}

// Copy returns a copy of the set.
func (s *Set) Copy() *Set {
	// TODO(jeff): is it pathalogical to add the keys in sorted order or hash
	// order? which is better?

	out := New(len(s.order))
	for _, key := range s.order {
		out.order = append(out.order, key)
		out.set[key] = struct{}{}
	}
	return out
}

// Merge inserts all of the values in o into the set.
func (s *Set) Merge(o *Set) {
	// TODO(jeff): is doing a merge sort here better than just calling Add
	// for every key?

	for _, key := range o.order {
		s.Add(key)
	}
}

// Iter iterates over all of the keys in the set.
func (s *Set) Iter(cb func(name string) bool) {
	for _, key := range s.order {
		if !cb(key) {
			return
		}
	}
}
