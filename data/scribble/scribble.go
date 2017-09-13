// Copyright (C) 2017. See AUTHORS.

package scribble

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/rothko/data"
)

// Scribbler keeps track of the distributions of a collection of metrics.
type Scribbler struct {
	val atomic.Value // contains sync.Map of string -> *agg

	params data.DistParams
	hooks  struct {
		now func() time.Time
	}
}

// NewScribbler makes a Scribbler that will return distributions using the
// associated compression.
func NewScribbler(params data.DistParams) *Scribbler {
	s := &Scribbler{
		params: params,
	}

	// set the hooks
	s.hooks.now = time.Now

	return s
}

// Scribble adds the metric value to the current set of records. It will be
// reflected in the distribution of the records returned by Capture. WARNING:
// under some concurrent scenarios, this can lose updates.
func (s *Scribbler) Scribble(ctx context.Context, metric string,
	value float64, id []byte) {

	// warning: there is a race here where values can be lost. two people can
	// create the map and one will use the wrong one.
	mi := s.val.Load()
	if mi == nil {
		mi = new(sync.Map)
		s.val.Store(mi)
	}
	m := mi.(*sync.Map)

	// warning: there is a race here where values can be lost. two people can
	// create the agg and one will use the wrong one.
	ai, ok := m.Load(metric)
	if !ok {
		a := newAgg(s.params, s.hooks.now())
		ai = &a
		m.Store(metric, ai)
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
	// into the atomic value.
	mi := s.val.Load()
	if mi == nil {
		return
	}
	s.val.Store(new(sync.Map))

	m := mi.(*sync.Map)
	now := s.hooks.now()

	m.Range(func(key, ai interface{}) bool {
		return fn(key.(string), ai.(*agg).Finish(now))
	})
}
