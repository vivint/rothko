// Copyright (C) 2018. See AUTHORS.

package query

import (
	"strings"
)

// spec is a type that contains a set of globs for matching a metric. for
// example, foo.bar.baz will match any metric that contains three dotted
// components that are in a row and individually match the globs foo*,
// bar* and baz* according to the glob function in this package.
type spec struct {
	globs []string
}

// newSpec constructs a spec from the string.
func newSpec(s string) spec {
	globs := strings.Split(s, ".")
	for i := range globs {
		globs[i] = globify(globs[i])
	}
	return spec{
		globs: globs,
	}
}

// globify takes in a query and converts it into a string suitable for globbing
func globify(part string) string {
	var buf strings.Builder
	buf.Grow(len(part) + 6) // some extra for wildcards on special characters
	buf.WriteByte('*')

	for i := 0; i < len(part); i++ {
		c := part[i]

		// ascii lower case it
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}

		buf.WriteByte(c)
	}

	return buf.String()
}

// Match returns true if the spec matches the metric.
func (s spec) Match(metric string) bool {
top:
	for {
		if len(metric) == 0 {
			return false
		}

		var part string
		part, metric = splitMetric(metric)

		tail := metric
		for _, g := range s.globs {
			if !glob(g, part) {
				continue top
			}
			part, tail = splitMetric(tail)
		}

		return true
	}
}

// splitMetric pulls off the first dot separated component of the metric.
func splitMetric(metric string) (prefix, suffix string) {
	index := strings.IndexByte(metric, '.')
	if index == -1 {
		return metric, ""
	}
	return metric[:index], metric[index+1:]
}
