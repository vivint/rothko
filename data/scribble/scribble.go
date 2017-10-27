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
	mu  sync.Mutex
	m   sync.Map // map[string]*agg
	now time.Time
}

// newPage creates a new page for the scribbler.
func newPage() *page {
	return &page{
		now: time.Now(),
	}
}

// Scribbler keeps track of the distributions of a collection of metrics.
type Scribbler struct {
	val_mu sync.Mutex   // mutex around creating the page
	val    atomic.Value // contains a *page

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

	pi := s.val.Load()
	if pi == nil {
		// if we don't have a page, we need to acquire the mutex and allocate
		// one. after acquiring the mutex, we may have lost a race so we need
		// to double check that we still don't have a page.

		s.val_mu.Lock()
		pi = s.val.Load()
		if pi == nil {
			pi = newPage()
			s.val.Store(pi)
		}
		s.val_mu.Unlock()
	}
	p := pi.(*page)

	// TODO(jeff): there is a race here where we can lose writes: if someone
	// is calling Capture and that finishes and sets a new page, a call to
	// Scribble may use an agg on a page that will no longer be Captured.
	// Callers may work around this by ensuring no concurrent calls to Scribble
	// and Capture, but the data loss is probably acceptable.

	ai, ok := p.m.Load(metric)
	if !ok {
		// if we don't have the agg, we need to acquire the page mutex and
		// allocate one. after acquiring the mutex, we may have lost a race
		// so we need to double check that we still don't have a page.
		//
		// TODO(jeff): we can have sharded by metric name mutexes for creating
		// these aggs.

		p.mu.Lock()
		ai, ok = p.m.Load(metric)
		if !ok {
			ai = newAgg(s.params, p.now)
			p.m.Store(metric, ai)
		}
		p.mu.Unlock()
	}
	a := ai.(*agg)

	a.Observe(value, id)
}

// Capture clears out current set of records for future Scribble calls and
// calls the provided function with every record.
func (s *Scribbler) Capture(ctx context.Context,
	fn func(metric string, rec data.Record)) {

	// call CaptureUnsafe with nil buffers to cause allocations
	s.CaptureUnsafe(ctx, nil, func(metric string, rec data.Record) []byte {
		fn(metric, rec)
		return nil
	})
}

// Iterate calls the provided function with every record.
func (s *Scribbler) Iterate(ctx context.Context,
	fn func(metric string, rec data.Record)) {

	// call IterateUnsafe with nil buffers to cause allocations
	s.IterateUnsafe(ctx, nil, func(metric string, rec data.Record) []byte {
		fn(metric, rec)
		return nil
	})
}

// CaptureUnsafe clears out current set of records for future Scribble calls
// and calls the provided function with every record. The function returns the
// next buffer for the Captrue call to use. The record will be invalidated if
// any passed in buffer for it is modified.
func (s *Scribbler) CaptureUnsafe(ctx context.Context, buf []byte,
	fn func(metric string, rec data.Record) []byte) {

	// acquire the mutex to read and swap out the current page
	s.val_mu.Lock()

	// if we don't have a page yet, we can just be done
	pi := s.val.Load()
	if pi == nil {
		s.val_mu.Unlock()
		return
	}

	// swap it out and unlock
	pn := newPage()
	s.val.Store(pn)
	s.val_mu.Unlock()

	p := pi.(*page)

	// iterate it
	p.m.Range(func(key, ai interface{}) (ok bool) {
		var rec data.Record
		buf, rec = ai.(*agg).Finish(buf[:0], pn.now)
		buf = fn(key.(string), rec)
		return true
	})
}

// Iterate calls the provided function with every record.
func (s *Scribbler) IterateUnsafe(ctx context.Context, buf []byte,
	fn func(metric string, rec data.Record) []byte) {

	// if we don't have a page yet, we can just be done
	pi := s.val.Load()
	if pi == nil {
		return
	}
	p := pi.(*page)
	now := time.Now()

	// iterate it
	p.m.Range(func(key, ai interface{}) (ok bool) {
		var rec data.Record
		buf, rec = ai.(*agg).Finish(buf[:0], now)
		buf = fn(key.(string), rec)
		return true
	})
}
