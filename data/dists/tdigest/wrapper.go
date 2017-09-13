// Copyright (C) 2017. See AUTHORS.

// package tdigest provides a wrapper around github.com/zeebo/tdigest.
package tdigest // import "github.com/spacemonkeygo/rothko/data/dists/tdigest"

import (
	"github.com/spacemonkeygo/rothko/data"
	"github.com/zeebo/tdigest"
)

// Params implements data.DistParams for a t-digest distribution.
type Params struct {
	Compression float64
}

// Kind returns the TDigest distribution kind.
func (p Params) Kind() data.DistributionKind {
	return data.DistributionKind_TDigest
}

// New returns a new TDigest as a data.Dist.
func (p Params) New() data.Dist {
	return wrapper{p.NewUnwrapped()}
}

// NewUnwrapped returns a new TDigest.
func (p Params) NewUnwrapped() *tdigest.TDigest {
	return tdigest.New(p.Compression)
}

//
// wrapper
//

type wrapper struct{ td *tdigest.TDigest }

func (wrapper) Kind() data.DistributionKind {
	return data.DistributionKind_TDigest
}

func (w wrapper) Observe(val float64) {
	w.td.Add(val, 1)
}

func (w wrapper) Marshal(buf []byte) []byte {
	return w.td.Marshal(buf)
}
