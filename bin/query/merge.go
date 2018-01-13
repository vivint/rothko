// Copyright (C) 2018. See AUTHORS.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/data/merge"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/zeebo/errs"
)

// TODO(jeff): put into a package
// TODO(jeff): a slice is absolutely the wrong data structure

const debug = false

func debugPrint(vals ...interface{}) {
	if debug {
		fmt.Println(vals...)
	}
}

// MergerOptions are the options the Merger needs to operate.
type MergerOptions struct {
	Width        int
	Samples      int
	Now          int64
	Duration     time.Duration
	MergeOptions merge.MergeOptions
}

// Merger allows iterative pushing of records in and constructs a series of
// merged columns. The only requirement is that the end time on the records
// passed to push are decreasing.
type Merger struct {
	opts       MergerOptions
	pixel_size int64

	completed_px int64
	records      []data.Record
	columns      []draw.Column
}

// NewMerger constructs a Merger with the options.
func NewMerger(opts MergerOptions) *Merger {
	return &Merger{
		opts:       opts,
		pixel_size: opts.Duration.Nanoseconds() / int64(opts.Width),

		completed_px: int64(opts.Width),
	}
}

// timeToPixel maps the time to a pixel.
func (m *Merger) timeToPixel(time int64) int64 {
	delta := m.opts.Now - time + m.pixel_size - 1
	px := int64(m.opts.Width) - (delta / m.pixel_size)
	if px < 0 {
		px = 0
	}
	return px
}

// Push adds the record to the merger. The end time on the records passed to
// Push must be decreasing.
func (m *Merger) Push(rec data.Record) error {
	rec.StartTime = m.timeToPixel(rec.StartTime)
	rec.EndTime = m.timeToPixel(rec.EndTime)
	debugPrint("adding", rec.StartTime, rec.EndTime)
	if err := m.completed(rec.EndTime + 1); err != nil {
		return err
	}
	m.records = append(m.records, rec)
	return nil
}

// Finish returns the set of columns to draw.
func (m *Merger) Finish() ([]draw.Column, error) {
	if err := m.completed(0); err != nil {
		return nil, err
	}
	return m.columns, nil
}

// completed informs the Merge that the px argument is "completed", meaning
// no more records are going to be pushed that have any overlap with that px.
// it also implies that every px >= the argument is completed.
func (m *Merger) completed(completed_px int64) error {
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
			if err := m.emit(px+1, to_emit_end_px, emit_recs); err != nil {
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
		if err := m.emit(completed_px, to_emit_end_px, emit_recs); err != nil {
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
func (m *Merger) emit(start, end int64, recs []data.Record) error {
	debugPrint("emit", start, end)

	opts := m.opts.MergeOptions
	opts.Records = recs
	out, err := merge.Merge(context.Background(), opts)
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
		col.Data = append(col.Data, dist.Query(i/f64_samples))
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
