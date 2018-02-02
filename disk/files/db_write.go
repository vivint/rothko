// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"

	"github.com/spacemonkeygo/rothko/disk/files/internal/sset"
)

// Queue adds the data for the metric and the given start and end times. If
// the start time is before the last end time for the metric, no write will
// happen. The callback is called with the error value of writing the metric.
func (db *DB) Queue(ctx context.Context, metric string, start int64,
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

// worker takes data from the queue and writes it into the appropriate metric
// file in the appropriate location.
func (db *DB) worker(ctx context.Context, num int) {
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
			ok, err := db.write(ctx, num, value)
			db.bufs.Put(value.data)
			if value.done != nil {
				value.done(ok, err)
			}
		}
	}
}

// write puts the queued value into the appropriate file. it can be called
// concurrently with other values, even when they reference the same metric.
func (db *DB) write(ctx context.Context, num int, value queuedValue) (
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

	// update the names map for this worker if not already present in the
	// central set of names. we avoid contention as much as possible through
	// a readonly names set and per worker mutexes and names.
	if ok {
		should_store := false

		names, found := db.names.Load().(*sset.Set)
		if !found {
			should_store = true
		} else if found := names.Has(value.metric); !found {
			should_store = true
		}

		if should_store {
			db.names_w_mu[num].Lock()
			db.names_w[num].Add(value.metric)
			db.names_w_mu[num].Unlock()
		}
	}

	return ok, nil
}
