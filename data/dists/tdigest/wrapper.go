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
func (p Params) Kind() data.Kind {
	return data.Kind_TDigest
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

type Wrapper struct {
	td    *tdigest.TDigest
	cache map[float64]float64
}

func Wrap(td *tdigest.TDigest) *Wrapper {
	return &Wrapper{td: td}
}

func (Wrapper) Kind() data.Kind {
	return data.Kind_TDigest
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

func (w *Wrapper) quan(x float64) float64 {
	if w.cache == nil {
		w.cache = make(map[float64]float64)
	}
	if val, ok := w.cache[x]; ok {
		return val
	}
	val := w.td.Quantile(x)
	w.cache[x] = val
	return val
}

func (w *Wrapper) CDF(x float64) float64 {
	// TODO(jeff): CDF actually works, but i think it's way slower and this is
	// basically just as accurate since we only have ~256 colors. essentially,
	// we only have to do ~8 calls to quantile, rather than ~samples calls to
	// CDF.

	// return w.td.CDF(x)

	min, max := w.quan(0), w.quan(1)
	if x <= min {
		return 0
	}
	if x >= max {
		return 1
	}

	minq, maxq := 0.0, 1.0
	for i := 0; i < 6; i++ {
		medq := (minq + maxq) / 2
		val := w.quan(medq)
		if x >= val {
			minq = medq
		} else {
			maxq = medq
		}
	}
	return (minq + maxq) / 2
}

func (w Wrapper) Len() int64 {
	return int64(w.td.Count())
}
