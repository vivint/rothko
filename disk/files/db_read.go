// Copyright (C) 2018. See AUTHORS.

package files

import (
	"bytes"
	"context"
	"syscall"

	"github.com/spacemonkeygo/rothko/disk"
	"github.com/spacemonkeygo/rothko/disk/files/internal/system"
)

// Query calls the ResultCallback with all of the data slices that end
// strictly before the provided end time in strictly decreasing order by
// their end. It will continue to call the ResultCallback until it exhausts
// all of the records, or the callback returns false.
func (db *DB) Query(ctx context.Context, metric string, end int64,
	buf []byte, cb disk.ResultCallback) error {

	db.locks.Lock(metric)
	defer db.locks.Unlock(metric)

	// acquire the datastructure encapsulating metric read logic
	met, err := db.newMetric(ctx, metric)
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
	met, err := db.newMetric(ctx, metric)
	if err != nil {
		return 0, 0, nil, err
	}

	return met.ReadLast(ctx, buf)
}

// Metrics calls the callback once for every metric stored.
func (db *DB) Metrics(ctx context.Context, cb func(name string) error) (
	err error) {

	// load up the readonly names value
	names, ok := db.names.Load().(map[string]struct{})
	if !ok {
		names = make(map[string]struct{})
	}

	// merge the worker sets in if we have a flag for it
	copied := false
	locked := false
	for i, names_w := range db.names_w {
		db.names_w_mu[i].Lock()
		len_names_w := len(names_w)
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
			new_names, ok := db.names.Load().(map[string]struct{})
			if ok && len(new_names) > 0 {
				names = copyStringSet(new_names)
			}
			copied = true
		}

		// merge the worker map in
		for name := range names_w {
			names[name] = struct{}{}
		}

		// clear out the worker map
		db.names_w[i] = make(map[string]struct{})
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
	for name := range names {
		if err := cb(name); err != nil {
			return err
		}
	}
	return nil
}

// PopulateMetrics walks the directory tree of the metrics recreating the
// in-memory cache of metric names. It should be called periodically.
func (db *DB) PopulateMetrics(ctx context.Context) (err error) {
	dp := newDBPopulator(db.dir)
	if err := dp.populate(ctx); err != nil {
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
	out     map[string]struct{}
	namebuf []byte // buffer for the metric name
	dirbuf  []byte // buffer for the directory name
	pathbuf []byte // buffer for the path
}

func newDBPopulator(dir string) *dbPopulator {
	return &dbPopulator{
		dir: dir,
		out: make(map[string]struct{}),
	}
}

func (dp *dbPopulator) populate(ctx context.Context) (err error) {
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
			dp.out[string(dp.namebuf)] = struct{}{}

			// mark it added so we skip over adding other .data files
			added = true
		}
	}
}
