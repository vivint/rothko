// Copyright (C) 2018. See AUTHORS.

package dist

// Dist is a representation of a distribution.
type Dist interface {
	// Kind returns the kind of the distribution.
	Kind() string

	// Query returns the value for the x'th percentile. The percentile is
	// represented as a number in [0, 1].
	Query(x float64) float64

	// CDF returns the percentile for the given value. The percentile is
	// represented as a number in [0, 1].
	CDF(x float64) float64

	// Len returns how many observations there were for the distribution.
	Len() int64

	// Observe a value.
	Observe(val float64)

	// Marshal by appending to the provided buf.
	Marshal(buf []byte) []byte
}

// Params represents a way to create Dists. An implementation must cope with
// being created with possibly no configuration if coming from the registry.
// New is allowed to error in this case, but Unmarshal and Kind should not.
type Params interface {
	// Kind returns the kind of the distribution.
	Kind() string

	// New creates a new Dist value.
	New() (Dist, error)

	// Unmarshal should load the Dist from the provided data slice.
	Unmarshal(data []byte) (Dist, error)
}
