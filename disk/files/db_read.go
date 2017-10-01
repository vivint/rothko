// Copyright (C) 2017. See AUTHORS.

package files

import (
	"bytes"
	"context"
	"syscall"

	"github.com/spacemonkeygo/rothko/disk"
	"github.com/spacemonkeygo/rothko/disk/files/internal/system"
)

// Query calls the ResultCallback with all of the data slices that overlap
// their start and end time with the provided values. The buf slice is
// used for storage of the data passed to the ResultCallback if possible.
// The data must not be modified, and no references must be kept after
// the callback returns.
func (db *DB) Query(ctx context.Context, metric string, start, end int64,
	buf []byte, cb disk.ResultCallback) error {

	db.locks.Lock(metric)
	defer db.locks.Unlock(metric)

	// acquire the datastructure encapsulating metric read logic
	met, err := db.newMetric(ctx, metric)
	if err != nil {
		return err
	}

	return met.Read(ctx, start, end, buf, cb)
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

	// do the copy as fast as possible to avoid contention on this mutex.
	db.names_mu.Lock()
	names := make([]string, 0, len(db.names))
	for name := range db.names {
		names = append(names, name)
	}
	db.names_mu.Unlock()

	for _, name := range names {
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

	db.names_mu.Lock()
	db.names = dp.out
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
