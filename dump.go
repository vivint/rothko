// Copyright (C) 2017. See AUTHORS.

package rothko

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/scribble"
	"github.com/spacemonkeygo/rothko/disk"
)

// periodicallyDump is the worker that takes the data from the Sribbler and
// stores it in to the Disk.
func periodicallyDump(ctx context.Context, scr *scribble.Scribbler,
	di disk.Disk) (err error) {

	// TODO(jeff): configs?

	bufs := sync.Pool{
		New: func() interface{} { return make([]byte, 1024) },
	}

	done := ctx.Done()
	ticker := time.NewTicker(10 * time.Minute)
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
				data := bufs.Get().([]byte)
				if size := rec.Size(); cap(data) < size {
					data = make([]byte, size)
				} else {
					data = data[:size]
				}
				_, err = rec.MarshalTo(data)
				if err != nil {
					return false
				}

				// TODO(jeff): log the error that this returns
				wg.Add(1)
				di.Queue(ctx, metric, rec.StartTime, rec.EndTime, data,
					func(written bool, err error) {
						// TODO(jeff): handle the input params appropriately?
						// probably just logging.

						bufs.Put(data)
						wg.Done()
						atomic.AddInt64(&metrics, 1)
					})
				return true
			})

			// errors "captured" from the closure are fatal.
			if err != nil {
				return err
			}

			wg.Wait()
			fmt.Println(metrics, time.Since(now))
		}
	}
}
