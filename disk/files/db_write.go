// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"
	"sync"
)

// Queue adds the data for the metric and the given start and end times. If
// the start time is before the last end time for the metric, no write will
// happen.
func (db *DB) Queue(ctx context.Context, metric string, start int64, end int64,
	data []byte) (err error) {

	return db.QueueCB(ctx, metric, start, end, data, nil)
}

// QueueCB adds the data for the metric and the given start and end times. If
// the start time is before the last end time for the metric, no write will
// happen. The callback is called with the error value o writing the metric.
func (db *DB) QueueCB(ctx context.Context, metric string, start int64,
	end int64, data []byte, cb func(bool, error)) (err error) {

	buf := db.bufs.Get().([]byte)
	buf = append(buf[:0], data...)

	value := queuedValue{
		metric: metric,
		start:  start,
		end:    end,
		data:   buf,
		done:   cb,
	}

	if db.opts.Drop {
		select {
		case db.queue <- value:
		default:
			db.bufs.Put(value.data)
			if value.done != nil {
				value.done(false, nil)
			}
		}
	} else {
		db.queue <- value
	}

	return nil
}

// Run will read values from the Queue and persist them to disk. It returns
// when the context is done.
func (db *DB) Run(ctx context.Context) (err error) {
	var wg sync.WaitGroup

	wg.Add(db.opts.Workers)
	for i := 0; i < db.opts.Workers; i++ {
		go func() {
			db.worker(ctx)
			wg.Done()
		}()
	}

	// wait for the workers who will exit when the context is done.
	wg.Wait()

	return nil
}

// worker takes data from the queue and writes it into the appropriate metric
// file in the appropriate location.
func (db *DB) worker(ctx context.Context) {
	done := ctx.Done()

	// NOTE(jeff): because there are multiple workers and because we do not
	// allow writes for previous time points, there is a race where a value
	// can be lost: when two workers are working on adding data for the same
	// metric, but scheduled backwards from the insertion into the queue. we
	// accept this risk because typically the queue will be empty before more
	// values are inserted. it may be worth exposing knobs so that consumers
	// can ensure the queue is empty, like perhaps a transactional style api.

	for {
		select {
		case <-done:
			return

		case value := <-db.queue:
			ok, err := db.write(ctx, value)
			db.bufs.Put(value.data)
			if value.done != nil {
				value.done(ok, err)
			}
		}
	}
}

// write puts the queued value into the appropriate file. it can be called
// concurrently with other values, even when they reference the same metric.
func (db *DB) write(ctx context.Context, value queuedValue) (
	ok bool, err error) {

	// lock the metric
	db.locks.Lock(value.metric)
	defer db.locks.Unlock(value.metric)

	// acquire the datastructure encapsulating metric write logic
	met, err := db.newMetric(ctx, value.metric)
	if err != nil {
		return false, err
	}

	// write the value
	ok, err = met.Write(ctx, value.start, value.end, value.data)
	if err != nil {
		return false, err
	}

	return ok, nil
}
