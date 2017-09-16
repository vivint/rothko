// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	end int64, data []byte, cb func(error)) (err error) {

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
				value.done(nil)
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
			err := db.write(ctx, value)
			db.bufs.Put(value.data)
			if value.done != nil {
				value.done(err)
			}
		}
	}
}

// releaseFile puts the file back into the cache, closing any evicted file.
func (db *DB) releaseFile(path string, f file) {
	db.mu.Lock()
	tok, ev, ok := db.ch.Put(f)
	db.toks[path] = tok
	db.mu.Unlock()

	// TODO(jeff): do we want to call sync or anything?
	if ok {
		ev.Close()
	}
}

// acquireFile opens or creates the file at the path. it is expected to be
// called exclusive to all others that might be interested in the path.
func (db *DB) acquireFile(path string, exists bool) (f file, err error) {
	db.mu.Lock()
	tok, ok := db.toks[path]
	if ok {
		f, ok = db.ch.Take(tok)
	}
	db.mu.Unlock()

	if ok {
		return f, nil
	}

	if exists {
		return openFile(ctx, path)
	}
	return createFile(ctx, path, db.opts.Size, db.opts.Cap)
}

// write puts the queued value into the appropriate file. it can be called
// concurrently with other values, even when they reference the same metric.
func (db *DB) write(ctx context.Context, value queuedValue) (err error) {
	defer mon.Task()(&ctx)(&err)

	// lock the metric
	db.locks.Lock(value.metric)
	defer db.locks.Unlock(value.metric)

	// first, determine which file we're going to write in to.

	// TODO(jeff): perhaps we can have some caching on this? will we need it?
	// it will look good on benchmarks that always write the same metric, but
	// we must not forget the typical access patterns.

	dir := filepath.Join(db.path, string(metricToDir(nil, value.metric)))
	dh, err := os.Open(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return Error.Wrap(err)
		}
		dh, err = os.Open(dir)
	}
	if err != nil {
		return Error.Wrap(err)
	}
	defer dh.Close()

	// figure out what the first and last data files in the directory are.

	names, err := dh.Readdirnames(-1)
	if err != nil {
		return Error.Wrap(err)
	}

	first := 0
	first_name := "0.data"
	first_exists := false

	last := 0
	last_name := "0.data"
	last_exists := false

	for _, name := range names {
		if !strings.HasSuffix(name, ".data") {
			continue
		}
		val, err := strconv.ParseInt(name[:len(name)-5], 10, 0)
		if err != nil {
			continue
		}
		if int(val) > last {
			last = int(val)
			last_name = name
			last_exists = true
		}
		if int(val) < first || !first_exists {
			first = int(val)
			first_name = name
			first_exists = true
		}
	}

	// TODO(jeff): there's a bunch of duplicated logic here. maybe we can do
	// refactor it.

	first_path := filepath.Join(dir, first_name)
	last_path := filepath.Join(dir, last_name)

	f, err := db.acquireFile(last_path, last_exists)
	if err != nil {
		return err
	}
	defer db.releaseFile(last_path, f)

	// figure out where we need to start writing
	head, err := lastRecord(ctx, f)
	if err != nil {
		return err
	}

	// if we have too many records to fit into the file, we need to make a new
	// one.
	nr := numRecords(len(value.data), f.Size())
	if nr == 0 {
		return Error.New("unable to compute number of records")
	}
	if head+nr > f.Capacity() {
		// now that we're allocating a new file, we may need to delete an old
		// one. it might also be in the cache, so we need to acquire it if
		// possible and close the handle.
		if last-first > db.opts.Files {
			var first_f file

			db.mu.Lock()
			tok, ok := db.toks[first_path]
			if ok {
				delete(db.toks, first_path)
				first_f, ok = db.ch.Take(tok)
			}
			db.mu.Unlock()
			if ok {
				first_f.Close()
			}
			os.Remove(first_path)
		}

		last++
		last_name = fmt.Sprintf("%d.data", last)
		last_path = filepath.Join(dir, last_name)

		f, err = db.acquireFile(last_path, false)
		if err != nil {
			return err
		}
		defer db.releaseFile(last_path, f)

		head, err = lastRecord(ctx, f)
		if err != nil {
			return err
		}

		nr = numRecords(len(value.data), f.Size())
		if nr == 0 {
			return Error.New("unable to compute number of records")
		}
		if head+nr > f.Capacity() {
			return Error.New("record too large to fit in new file")
		}
	}

	// TODO(jeff): we have a bug where the monotonicity checking may not work
	// because we're writing the first record of a new file, and the previous
	// file has a newer record.

	// ensure monotonicity of records if applicable.
	if head > 0 {
		last_rec, err := f.Record(ctx, head-1)
		if err != nil {
			return err
		}
		if last_rec.start > value.start || last_rec.end > value.start {
			return Error.New("version monotonicity failure")
		}
	}

	// write the records now that we know we have capacity.
	err = records(value.start, value.end, value.data, f.Size(),
		func(rec record) error {
			err := f.SetRecord(ctx, head, rec)
			head++
			return err
		})
	if err != nil {
		return err
	}

	// update the head pointer.
	m, err := f.Metadata(ctx)
	if err != nil {
		return err
	}
	m.Head = head

	err = f.SetMetadata(ctx, m)
	if err != nil {
		return err
	}

	return f.FullAsync(ctx)
}
