// Copyright (C) 2017. See AUTHORS.

package dump

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/scribble"
	"github.com/spacemonkeygo/rothko/disk"
	"github.com/spacemonkeygo/rothko/external"
)

type Options struct {
	Disk      disk.Disk
	Period    time.Duration
	Resources external.Resources
	Bufsize   int
}

type Dumper struct {
	opts Options

	bufs sync.Pool
}

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

func (d *Dumper) Run(ctx context.Context, scr *scribble.Scribbler) (
	err error) {

	ext := d.opts.Resources
	done := ctx.Done()
	ticker := time.NewTicker(d.opts.Period)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return ctx.Err()

		case <-ticker.C:
			var err error
			var wg sync.WaitGroup
			metrics := int64(0)
			now := time.Now()

			scr.Capture(ctx, func(metric string, rec data.Record) bool {
				// check if we're cancelled
				select {
				case <-done:
					err = ctx.Err()
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
				_, err = rec.MarshalTo(data)
				if err != nil {
					return false
				}

				wg.Add(1)

				err = d.opts.Disk.Queue(ctx,
					metric, rec.StartTime, rec.EndTime, data,
					func(written bool, err error) {
						if !written || err != nil {
							ext.Errorw("metric write problem",
								"written", written,
								"err", err,
							)
						}

						d.bufs.Put(data)
						wg.Done()
						atomic.AddInt64(&metrics, 1)
					})
				if err != nil {
					ext.Errorw("error queuing metric", "err", err)
				}

				return true
			})

			// errors "captured" from the closure are fatal.
			if err != nil {
				return err
			}

			wg.Wait()
			ext.Observe("dumped_metrics_time", time.Since(now).Seconds())
			ext.Observe("dumped_metrics", float64(metrics))
		}
	}
}
