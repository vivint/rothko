// Copyright (C) 2017. See AUTHORS.

package scribble

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/rothko/data"
)

// page keeps track of a mapping of metric name strings to *agg with a time
// that all of the aggs will start at.
type page struct {
	m   *sync.Map
	now time.Time
}

// newPage creates a new page for the scribbler.
func newPage() page {
	return page{
		m:   new(sync.Map),
		now: time.Now(),
	}
}

// Scribbler keeps track of the distributions of a collection of metrics.
type Scribbler struct {
	val atomic.Value // contains a page

	params data.DistParams
}

// NewScribbler makes a Scribbler that will return distributions using the
// associated compression.
func NewScribbler(params data.DistParams) *Scribbler {
	return &Scribbler{
		params: params,
	}
}

// Scribble adds the metric value to the current set of records. It will be
// reflected in the distribution of the records returned by Capture. WARNING:
// under some concurrent scenarios, this can lose updates.
func (s *Scribbler) Scribble(ctx context.Context, metric string,
	value float64, id []byte) {

	// warning: there is a race here where values can be lost. two people can
	// create the page and one will use the wrong one.
	pi := s.val.Load()
	if pi == nil {
		pi = newPage()
		s.val.Store(pi)
	}
	p := pi.(page)

	// warning: there is a race here where values can be lost. two people can
	// create the agg and one will use the wrong one.
	ai, ok := p.m.Load(metric)
	if !ok {
		a := newAgg(s.params, p.now)
		ai = &a
		p.m.Store(metric, ai)
	}
	a := ai.(*agg)

	a.Observe(value, id)
}

// Capture clears out current set of records for future Scribble calls and
// calls the provided function with every record.
func (s *Scribbler) Capture(ctx context.Context,
	fn func(metric string, rec data.Record) bool) {

	// warning: there is a race here where values can be lost. a scribbler can
	// have a reference to this map instead of the new map we will be putting
	// into the atomic value. they could then write to that map after the
	// Range call.
	pi := s.val.Load()
	if pi == nil {
		return
	}

	pn := newPage()
	s.val.Store(pn)

	p := pi.(page)
	p.m.Range(func(key, ai interface{}) bool {
		return fn(key.(string), ai.(*agg).Finish(pn.now))
	})
}
