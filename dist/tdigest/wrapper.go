// Copyright (C) 2018. See AUTHORS.

// package tdigest provides a wrapper around github.com/zeebo/tdigest.
package tdigest

import (
	"github.com/spacemonkeygo/rothko/dist"
	"github.com/zeebo/errs"
	"github.com/zeebo/tdigest"
)

// Params implements dist.Params for a t-digest distribution.
type Params struct {
	Compression float64
}

// Kind returns the TDigest distribution kind.
func (p Params) Kind() string {
	return "tdigest"
}

// New returns a new TDigest as a dist.Dist.
func (p Params) New() (dist.Dist, error) {
	if p.Compression == 0 {
		return nil, errs.New("New called on zero value Params")
	}
	return Wrap(tdigest.New(p.Compression)), nil
}

// Unmarshal loads a dist.Dist out of some bytes.
func (p Params) Unmarshal(data []byte) (dist.Dist, error) {
	dist, err := tdigest.FromBytes(data)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	return Wrap(dist), nil
}

//
// Wrapper
//

// Wrapper implements dist.Dist for a t-digest.
type Wrapper struct {
	td    *tdigest.TDigest
	cache map[float64]float64
}

// Wrap wraps the given t-digest.
func Wrap(td *tdigest.TDigest) *Wrapper {
	return &Wrapper{td: td}
}

// Underlying returns the underlying t-digest.
func (w Wrapper) Underlying() *tdigest.TDigest {
	return w.td
}

// Kind returns the string "tdigest".
func (Wrapper) Kind() string {
	return "tdigest"
}

// Observe adds the value to the t-digest.
func (w Wrapper) Observe(val float64) {
	w.td.Add(val)
}

// Marshal appends a byte form of the t-digest to the provided buffer.
func (w Wrapper) Marshal(buf []byte) []byte {
	return w.td.Marshal(buf)
}

// Query returns the approximate x'th quantile.
func (w Wrapper) Query(x float64) float64 {
	return w.td.Quantile(x)
}

// quan is a wrapper to memoize queried qualtile values for quickly estimating
// the CDF.
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

// CDF returns the estimate CDF at the value x.
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

// Len returns how many items were added to the t-digest.
func (w Wrapper) Len() int64 {
	return int64(w.td.Count())
}
