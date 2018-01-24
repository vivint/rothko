// Copyright (C) 2018. See AUTHORS.

package merge

import (
	"context"
	"fmt"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/data/merge"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/zeebo/errs"
	"github.com/zeebo/float16"
)

// TODO(jeff): a slice is absolutely the wrong data structure

const debug = false

func debugPrint(vals ...interface{}) {
	if debug {
		fmt.Println(vals...)
	}
}

// Options are the options the Merger needs to operate.
type Options struct {
	Samples  int
	Now      int64
	Duration time.Duration
	Params   tdigest.Params
}

// Merger allows iterative pushing of records in and constructs a series of
// merged columns. The only requirement is that the end time on the records
// passed to push are decreasing.
type Merger struct {
	opts       Options
	pixel_size int64
	width      int

	completed_px int64
	records      []data.Record
	columns      []draw.Column
}

// New constructs a Merger with the options.
func New(opts Options) *Merger {
	return &Merger{
		opts: opts,
	}
}

// SetWidth sets the width for all of the Push operations. Must be set before
// any Push operations happen.
func (m *Merger) SetWidth(width int) {
	m.width = width
	m.pixel_size = m.opts.Duration.Nanoseconds() / int64(width)
	m.completed_px = int64(width)
}

// timeToPixel maps the time to a pixel.
func (m *Merger) timeToPixel(time int64) int64 {
	delta := m.opts.Now - time + m.pixel_size - 1
	px := int64(m.width) - (delta / m.pixel_size)
	if px < 0 {
		px = 0
	}
	return px
}

// Push adds the record to the Merger. The end time on the records passed to
// Push must be decreasing.
func (m *Merger) Push(ctx context.Context, rec data.Record) error {
	if m.width == 0 {
		return errs.New("invalid: must call SetWidth before Push")
	}

	rec.StartTime = m.timeToPixel(rec.StartTime)
	rec.EndTime = m.timeToPixel(rec.EndTime)
	debugPrint("adding", rec.StartTime, rec.EndTime)
	if err := m.completed(ctx, rec.EndTime+1); err != nil {
		return err
	}
	m.records = append(m.records, rec)
	return nil
}

// Finish returns the set of columns to draw.
func (m *Merger) Finish(ctx context.Context) ([]draw.Column, error) {
	if err := m.completed(ctx, 0); err != nil {
		return nil, err
	}
	return m.columns, nil
}

// completed informs the Merge that the px argument is "completed", meaning
// no more records are going to be pushed that have any overlap with that px.
// it also implies that every px >= the argument is completed.
func (m *Merger) completed(ctx context.Context, completed_px int64) error {
	debugPrint("completed", completed_px)

	// some variables for our state
	var (
		to_emit_end_px int64
		to_emit        []int
		cand_emit      []int
		emit_recs      []data.Record
	)

	for px := m.completed_px - 1; px >= completed_px; px-- {
		// create a slice of records to emit
		cand_emit = cand_emit[:0]
		for i, rec := range m.records {
			if rec.StartTime <= px && px <= rec.EndTime {
				cand_emit = append(cand_emit, i)
			}
		}

		// if the candidate is the same as to_emit, then don't worry about it.
		if intsEq(cand_emit, to_emit) {
			continue
		}

		// if we have something to emit, do it.
		if len(to_emit) > 0 {
			emit_recs = emit_recs[:0]
			for _, v := range to_emit {
				emit_recs = append(emit_recs, m.records[v])
			}
			err := m.emit(ctx, px+1, to_emit_end_px, emit_recs)
			if err != nil {
				return err
			}
		}

		// store the candidate as the next to emit
		to_emit_end_px = px
		to_emit = append(to_emit[:0], cand_emit...)
	}

	// if there's something left to emit, do it.
	if len(to_emit) > 0 {
		emit_recs = emit_recs[:0]
		for _, v := range to_emit {
			emit_recs = append(emit_recs, m.records[v])
		}
		err := m.emit(ctx, completed_px, to_emit_end_px, emit_recs)
		if err != nil {
			return err
		}
	}

	// prune off any prefix records welonger need. first, gather up the indexes
	// of the records we want to remove.
	var to_remove []int
	for i := range m.records {
		if m.records[i].StartTime > completed_px &&
			m.records[i].EndTime > completed_px {

			to_remove = append(to_remove, i)
		}
	}

	// second, filter them out when iterating over the current set of records.
	// since we added them to to_remove in sorted order, we can just iterate
	// over the slice of to_remove.
	current := m.records
	m.records = m.records[:0]
	for i, rec := range current {
		if len(to_remove) > 0 && to_remove[0] == i {
			debugPrint("removing", rec.StartTime, rec.EndTime)
			to_remove = to_remove[1:]
			continue
		}
		m.records = append(m.records, rec)
	}

	// yay we've completed up to the pixel now.
	m.completed_px = completed_px
	return nil
}

// emit constructs a column out of the records for the start and end pixels.
func (m *Merger) emit(ctx context.Context, start, end int64,
	recs []data.Record) error {

	debugPrint("emit", start, end)

	out, err := merge.Merge(ctx, merge.MergeOptions{
		Params:  m.opts.Params,
		Records: recs,
	})
	if err != nil {
		return errs.Wrap(err)
	}
	dist, err := dists.Load(out)
	if err != nil {
		return errs.Wrap(err)
	}
	col := draw.Column{
		X:    int(start),
		W:    int(end - start + 1),
		Data: make([]float64, 0, m.opts.Samples+1),
	}
	f64_samples := float64(m.opts.Samples)
	for i := float64(0); i <= f64_samples; i++ {
		val := dist.Query(i / f64_samples)
		val16, ok := float16.FromFloat64(val)
		if ok {
			val = val16.Float64()
		}
		col.Data = append(col.Data, val)
	}

	m.columns = append(m.columns, col)
	return nil
}

// intsEq returns if the integers are equal.
func intsEq(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
