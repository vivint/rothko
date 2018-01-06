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
	return Wrap(p.NewUnwrapped())
}

// NewUnwrapped returns a new TDigest.
func (p Params) NewUnwrapped() *tdigest.TDigest {
	return tdigest.New(p.Compression)
}

//
// Wrapper
//

type Wrapper struct{ td *tdigest.TDigest }

func Wrap(td *tdigest.TDigest) Wrapper {
	return Wrapper{td: td}
}

func (Wrapper) Kind() data.DistributionKind {
	return data.DistributionKind_TDigest
}

func (w Wrapper) Observe(val float64) {
	w.td.Add(val)
}

func (w Wrapper) Marshal(buf []byte) []byte {
	return w.td.Marshal(buf)
}

func (w Wrapper) Query(x float64) float64 {
	return w.td.Quantile(x)
}

func (w Wrapper) Len() int64 {
	return int64(w.td.Count())
}
