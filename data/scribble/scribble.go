// Copyright (C) 2018. See AUTHORS.

package scribble

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/spacemonkeygo/rothko/data"
)

// page keeps track of a mapping of metric name strings to *agg with a time
// that all of the aggs will start at.
type page struct {
	m   sync.Map // map[string]*agg
	now time.Time
}

// newPage creates a new page for the scribbler.
func newPage(now time.Time) *page {
	return &page{
		now: now,
	}
}

// Scribbler keeps track of the distributions of a collection of metrics.
type Scribbler struct {
	page   unsafe.Pointer // contains *page
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

	// skip problematic floating point values
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return
	}

	// load up the page pointer, allocating a fresh page if there isn't one.
	var pi unsafe.Pointer
	for {
		pi = atomic.LoadPointer(&s.page)
		if pi != nil {
			break
		}

		// if we don't have a page, we attempt to compare and swap it with a
		// newly allocated page.
		pi = unsafe.Pointer(newPage(time.Now()))
		if atomic.CompareAndSwapPointer(&s.page, nil, pi) {
			break
		}
	}
	p := (*page)(pi)

	// TODO(jeff): there is a race here where we can lose writes: if someone
	// is calling Capture and that finishes and sets a new page, a call to
	// Scribble may use an agg on a page that will no longer be Captured.
	// Callers may work around this by ensuring no concurrent calls to Scribble
	// and Capture, but the data loss is probably acceptable.

	ai, ok := p.m.Load(metric)
	if !ok {
		// we use LoadOrStore here to avoid a mutex at the cost of wasted
		// allocations for losers during contention.
		ai, _ = p.m.LoadOrStore(metric, newAgg(s.params, p.now))
	}
	a := ai.(*agg)

	a.Observe(value, id)
}

// Capture clears out current set of records for future Scribble calls and
// calls the provided function with every record. You must not hold on to
// any fields of the record after the callback returns.
func (s *Scribbler) Capture(ctx context.Context,
	fn func(metric string, rec data.Record) bool) {

	// read the page out. capture clears out the page so we will be setting
	// it to a new page that we allocate so that the timestamps line up
	// perfectly.
	pi := atomic.LoadPointer(&s.page)
	if pi == nil {
		return
	}
	p := (*page)(pi)
	now := time.Now()

	// swap it out with a new page starting at the capture time. if we are
	// unable to do this, some other call must be ranging on the page, and
	// so we don't want to also range over it.
	new_pi := unsafe.Pointer(newPage(now))
	if !atomic.CompareAndSwapPointer(&s.page, pi, new_pi) {
		return
	}

	// iterate it
	var buf []byte
	p.m.Range(func(key, ai interface{}) (ok bool) {
		var rec data.Record
		buf, rec = ai.(*agg).Finish(buf, now)
		return fn(key.(string), rec)
	})
}

// Iterate calls the provided function with every record. You must not hold on
// to any fields of the record after the callback returns.
func (s *Scribbler) Iterate(ctx context.Context,
	fn func(metric string, rec data.Record) bool) {

	// read the page out. iterate does not clear out the page so we just need
	// to read and if we have no page, we're done.
	pi := atomic.LoadPointer(&s.page)
	if pi == nil {
		return
	}
	p := (*page)(pi)
	now := time.Now()

	// iterate it
	var buf []byte
	p.m.Range(func(key, ai interface{}) (ok bool) {
		var rec data.Record
		buf, rec = ai.(*agg).Finish(buf, now)
		return fn(key.(string), rec)
	})
}
