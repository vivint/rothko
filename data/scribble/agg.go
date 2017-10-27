// Copyright (C) 2017. See AUTHORS.

package scribble

import (
	"sync"
	"time"

	"github.com/spacemonkeygo/rothko/data"
)

// agg aggregates observed values into a record.
type agg struct {
	mu     sync.Mutex
	rec    data.Record
	params data.DistParams
	dist   data.Dist
}

// newAgg returns an agg that can observe values and write a record.
func newAgg(params data.DistParams, now time.Time) *agg {
	return &agg{
		params: params,
		rec: data.Record{
			StartTime: now.In(time.UTC).UnixNano(),
			Merged:    1,
		},
	}
}

// Observe adds the value to the aggregated record, recording the id if it
// is larger or smaller than the max and min, respectively. The id is copied
// if it used.
func (a *agg) Observe(val float64, id []byte) {
	a.mu.Lock()

	// add the value into the digest, initializing it if necessary
	if a.dist == nil {
		a.dist = a.params.New()
	}
	a.dist.Observe(val)

	// keep track of min, max and obs to update them after dropping the mutex
	// and bump observations.
	min, max, obs := a.rec.Min, a.rec.Max, a.rec.Observations
	a.rec.Observations++

	a.mu.Unlock()

	// in the common case we don't need to bump min and max, so we do a double
	// check pattern to avoid as much critical section as possible.
	if obs == 0 || val < min || val > max {

		// we only make the copy if there's a good chance we'll be storing it.
		// once again, we do this outside of the mutex to avoid as much
		// critical section as possible.
		id_copy := append([]byte(nil), id...)

		a.mu.Lock()
		if obs == 0 || val < a.rec.Min {
			a.rec.Min = val
			a.rec.MinId = id_copy
		}
		if obs == 0 || val > a.rec.Max {
			a.rec.Max = val
			a.rec.MaxId = id_copy
		}
		a.mu.Unlock()
	}
}

// Finish returns the aggregated record, using the buf to marshal the data
// and returning the buf. Mutating the returned buf invalidates the record.
func (a *agg) Finish(buf []byte, now time.Time) ([]byte, data.Record) {
	a.mu.Lock()
	out := a.rec
	a.mu.Unlock()

	out.EndTime = now.In(time.UTC).UnixNano()
	out.DistributionKind = a.dist.Kind()
	buf = a.dist.Marshal(buf[:0])
	out.Distribution = buf

	return buf, out
}
