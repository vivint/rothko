// Copyright (C) 2018. See AUTHORS.

package files

import (
	"bytes"
	"context"
	"syscall"

	"github.com/spacemonkeygo/rothko/database"
	"github.com/spacemonkeygo/rothko/database/files/internal/sset"
	"github.com/spacemonkeygo/rothko/database/files/internal/system"
)

// Query calls the ResultCallback with all of the data slices that end
// strictly before the provided end time in strictly decreasing order by
// their end. It will continue to call the ResultCallback until it exhausts
// all of the records, or the callback returns false.
func (db *DB) Query(ctx context.Context, metric string, end int64,
	buf []byte, cb database.ResultCallback) error {

	db.locks.Lock(metric)
	defer db.locks.Unlock(metric)

	// acquire the datastructure encapsulating metric read logic
	met, err := db.newMetric(ctx, metric, true)
	if err != nil {
		return err
	}

	return met.Read(ctx, end, buf, cb)
}

// QueryLatest returns the latest value stored for the metric. buf is used
// as storage for the data slice if possible.
func (db *DB) QueryLatest(ctx context.Context, metric string, buf []byte) (
	start, end int64, data []byte, err error) {

	db.locks.Lock(metric)
	defer db.locks.Unlock(metric)

	// acquire the datastructure encapsulating metric read logic
	met, err := db.newMetric(ctx, metric, true)
	if err != nil {
		return 0, 0, nil, err
	}

	return met.ReadLast(ctx, buf)
}

// Metrics calls the callback once for every metric stored.
func (db *DB) Metrics(ctx context.Context,
	cb func(name string) (bool, error)) (err error) {

	// load up the readonly names value
	names, ok := db.names.Load().(*sset.Set)
	if !ok {
		names = sset.New(0)
	}

	// merge the worker sets in if we have a flag for it
	copied := false
	locked := false
	for i, names_w := range db.names_w {
		db.names_w_mu[i].Lock()
		len_names_w := names_w.Len()
		db.names_w_mu[i].Unlock()

		// skip if we don't need to merge it in
		if len_names_w == 0 {
			continue
		}

		// lock the populating mutex if required
		if !locked {
			db.names_mu.Lock()
			locked = true
		}

		// lock the worker map again. we drop the worker mutex to avoid
		// deadlocks while taking the db.names_mu.
		db.names_w_mu[i].Lock()

		// lazily copy the names map to ensure it is readonly
		// we reload the names map here after we have the mutex in case
		// some concurrent Metrics call has updated the atomic.Value
		if !copied {
			new_names, ok := db.names.Load().(*sset.Set)
			if ok && new_names.Len() > 0 {
				names = new_names.Copy()
			}
			copied = true
		}

		// merge the worker map in
		names.Merge(names_w)

		// clear out the worker map
		db.names_w[i] = sset.New(0)
		db.names_w_mu[i].Unlock()
	}

	// if we copied, we need to store the value now
	if copied {
		db.names.Store(names)
	}

	// if we locked, we can unlock now as we're done populating and storing
	if locked {
		db.names_mu.Unlock()
	}

	// yay do callbacks
	names.Iter(func(name string) (ok bool) {
		ok, err = cb(name)
		return ok && err == nil
	})
	return err
}

// PopulateMetrics walks the directory tree of the metrics recreating the
// in-memory cache of metric names. It should be called periodically.
func (db *DB) PopulateMetrics(ctx context.Context) (err error) {
	dp := newDBPopulator(db.dir)
	switch err := dp.populate(ctx); {
	case err == context.Canceled:
		return nil
	case err != nil:
		return err
	}

	// all stores to the names map have to be done under the names_mu mutex.
	db.names_mu.Lock()
	db.names.Store(dp.out)
	db.names_mu.Unlock()

	return nil
}

//
// dbPopulator keeps track of some buffers to super efficiently walk the set
// of metric names.
//

type dbPopulator struct {
	dir     string
	out     *sset.Set
	namebuf []byte // buffer for the metric name
	dirbuf  []byte // buffer for the directory name
	pathbuf []byte // buffer for the path
}

func newDBPopulator(dir string) *dbPopulator {
	return &dbPopulator{
		dir: dir,
		out: sset.New(0),
	}
}

func (dp *dbPopulator) populate(ctx context.Context) (err error) {
	// check for a context canceled. if so, start returning some errors to
	// unwind.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// set up path using the pathbuf
	// NOTE: path and pathbuf MUST NOT be used past this point since this
	// function is recursive.
	dp.pathbuf = append(dp.pathbuf[:0], dp.dir...)
	dp.pathbuf = append(dp.pathbuf, '/')
	dp.pathbuf = append(dp.pathbuf, dp.dirbuf...)
	dp.pathbuf = append(dp.pathbuf, 0)
	fd, err := system.Open(dp.pathbuf)
	if err != nil {
		return err
	}
	defer system.Close(fd)

	added := false
	dirents := make([]byte, 4096)
	dirbuf_len := len(dp.dirbuf)

	for {
		n, err := syscall.ReadDirent(int(fd), dirents)
		if err == syscall.Errno(syscall.ENOTDIR) {
			return nil
		}
		if err != nil {
			return Error.Wrap(err)
		}
		if n == 0 {
			return nil
		}

		buf := dirents[:n]
		for len(buf) > 0 {
			var name []byte
			var ok bool

			buf, name, ok = system.NextDirent(buf)
			if !ok {
				continue
			}

			if !bytes.HasSuffix(name, []byte(".data")) {
				// add the name to the dirbuf and recurse since it isn't likely
				// a data file.
				if dirbuf_len > 0 {
					dp.dirbuf = append(dp.dirbuf, '/')
				}
				dp.dirbuf = append(dp.dirbuf, name...)

				if err := dp.populate(ctx); err != nil {
					return err
				}

				// reset dp.dirbuf to be our dir again, while keeping any
				// allocations the above appends may have done
				dp.dirbuf = dp.dirbuf[:dirbuf_len]
				continue
			}
			if added {
				continue
			}

			// we have a metric, so add it to out reusing the dp.namebuf space
			dp.namebuf, err = dirToMetric(dp.namebuf[:0], dp.dirbuf)
			if err != nil {
				return err
			}
			dp.out.Add(string(dp.namebuf))

			// mark it added so we skip over adding other .data files
			added = true
		}
	}
}
