// Copyright (C) 2018. See AUTHORS.

package dump

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vivint/rothko/data"
	"github.com/vivint/rothko/database"
	"github.com/vivint/rothko/external"
)

// Options controls the options to the dumper.
type Options struct {
	// The database to dump into.
	DB database.DB

	// How often to dump.
	Period time.Duration

	// How big a buffer to use for records. Defaults to 1024.
	Bufsize int
}

// Dumper is a worker that periodically dumps from a Writer into a database.
type Dumper struct {
	opts Options

	bufs sync.Pool
}

// New constructs a Dumper with the given options.
func New(opts Options) *Dumper {
	if opts.Bufsize == 0 {
		opts.Bufsize = 1024
	}

	return &Dumper{
		opts: opts,

		bufs: sync.Pool{
			New: func() interface{} { return make([]byte, opts.Bufsize) },
		},
	}
}

// Run dumps periodically, until the context is canceled. When the context is
// canceled, it dumps one last time but at most for one minute.
func (d *Dumper) Run(ctx context.Context, w *data.Writer) (err error) {
	done := ctx.Done()
	ticker := time.NewTicker(d.opts.Period)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil

		case <-ticker.C:
			external.Infow("performing dump")
			d.Dump(ctx, w)
		}
	}
}

// Dump writes all of the metrics Captured from the Writer into the DB
// associated with the Dumper.
func (d *Dumper) Dump(ctx context.Context, w *data.Writer) {
	var wg sync.WaitGroup
	writes := int64(0)
	skips := int64(0)
	errors := int64(0)
	now := time.Now()
	done := ctx.Done()

	w.Capture(ctx, func(ctx context.Context, metric string,
		rec data.Record) bool {

		// check if we're cancelled. if so, we're done.
		select {
		case <-done:
			return false
		default:
		}

		// marshal the record, reusing memory if possible
		data := d.bufs.Get().([]byte)
		if size := rec.Size(); cap(data) < size {
			data = make([]byte, size)
		} else {
			data = data[:size]
		}
		_, err := rec.MarshalTo(data)
		if err != nil {
			external.Errorw("record marshal problem",
				"err", err.Error(),
			)
			return true
		}

		// write the database record and wait for it to come back
		wg.Add(1)

		err = d.opts.DB.Queue(ctx, metric, rec.StartTime, rec.EndTime, data,
			func(written bool, err error) {
				if !written || err != nil {
					external.Errorw("metric write problem",
						"written", written,
						"err", safeError(err),
					)
				}

				if written {
					atomic.AddInt64(&writes, 1)
				} else {
					atomic.AddInt64(&skips, 1)
				}
				if err != nil {
					atomic.AddInt64(&errors, 1)
				}

				d.bufs.Put(data)
				wg.Done()

			})
		if err != nil {
			external.Errorw("error queuing metric",
				"err", err.Error(),
			)
			atomic.AddInt64(&skips, 1)
			atomic.AddInt64(&errors, 1)
		}

		return true
	})

	// TODO(jeff): bound how long we will Wait? wait groups and timeouts are
	// tricky to manage :(

	wg.Wait()

	duration := time.Since(now)

	external.Observe("metric_dump_duration", duration.Seconds())
	external.Observe("metric_writes", float64(writes))
	external.Observe("metric_skips", float64(skips))
	external.Observe("metric_errors", float64(errors))

	external.Infow("dump finished",
		"duration", duration,
		"writes", writes,
		"skips", skips,
		"errors", errors,
	)
}

func safeError(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
