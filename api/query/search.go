// Copyright (C) 2018. See AUTHORS.

package query

import (
	"strings"
)

// Search represents a metric search.
type Search struct {
	specs   []spec
	matched []string
}

// New constructs a metric searcher from the query string.
func New(query string, capacity int) *Search {
	parts := strings.Fields(query)
	specs := make([]spec, len(parts))
	for i, part := range parts {
		specs[i] = newSpec(part)
	}
	return &Search{
		specs:   specs,
		matched: make([]string, 0, capacity),
	}
}

// Match checks if the Search matches the metric.
func (s *Search) Match(metric string) bool {
	for _, spec := range s.specs {
		if !spec.Match(metric) {
			return false
		}
	}
	return true
}

// Add is meant to be passed to a disk.Metrics call.
func (s *Search) Add(name string) (bool, error) {
	if len(s.matched) == cap(s.matched) {
		return false, nil
	}
	if !s.Match(name) {
		return true, nil
	}
	s.matched = append(s.matched, name)
	return true, nil
}

// Matched returns the matched metrics.
func (s *Search) Matched() []string {
	return s.matched
}
