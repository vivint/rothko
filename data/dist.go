// Copyright (C) 2017. See AUTHORS.

package data

// DistParams represents a way to create Dists.
type DistParams interface {
	// Kind returns the kind of the distribution.
	Kind() DistributionKind

	// New creates a new distribution value.
	New() Dist
}

// Dist represents an abstract distribution.
type Dist interface {
	// Kind returns the kind of the distribution.
	Kind() DistributionKind

	// Observe a value.
	Observe(val float64)

	// Marshal by appending to the provided buf.
	Marshal(buf []byte) []byte
}
