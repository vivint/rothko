// Copyright (C) 2018. See AUTHORS.

package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/vivint/rothko/external"
)

//
// the metric abstraction keeps track of the number of files backing the metric
// data, and allows quick seeking/reading/writing of the data.
//
// the data format is that each file contains some number of records of some
// size described by the metadata at the start of the file. the records are
// filled in from the last entry to the first in order to make the query
// calls as efficient as possible. we want to cause backward iteration of
// records to be forward reads on disk.
//
// because we are laying out chronoglically later records earlier in the file
// there can be confusion about what directions mean. we always use the term
// "forward" to mean forward in terms of disk layout, and explicitly specify
// "chronologically forward" to mean forward in time.
//

// metricOptions are all the pieces of data required to work with a metric.
type metricOptions struct {
	fch  *fileCache
	dir  string
	name string
	max  int
	ro   bool // read only
}

// filenameBuf is a cache around constructing paths, as it is a significant
// source of allocations.
type filenameBuf []byte

// metricFilenameAt returns the data file for the index.
func (fb *filenameBuf) metricFilenameAt(dir string, index int) string {
	out := []byte(*fb)[:0]
	if initial := len(dir) + 12; cap(out) < initial {
		out = make([]byte, 0, initial)
	}

	out = append(out, dir...)
	if len(out) > 0 && out[len(out)-1] != '/' {
		out = append(out, '/')
	}
	out = strconv.AppendInt(out, int64(index), 10)
	out = append(out, ".data"...)

	*fb = filenameBuf(out)
	return string(out)
}

// metric abstracts logic around reading and writing data for a metric.
type metric struct {
	opts  metricOptions
	dir   string
	first int
	last  int

	// caching around paths because it's a significant source of allocations
	fb       filenameBuf
	interned map[int]string // to avoid reallocating paths from filenamebuf
}

// newMetric constructs a metric for some path. only one metric instance should
// be in use at a time for a given directory.
func newMetric(ctx context.Context, opts metricOptions) (*metric, error) {
	// get the base directory. it's probably going overboard reducing
	// allocations here. hopefully no bugs are ever caused by it :)
	dir_buf := make([]byte, 0, len(opts.dir)+1+len(opts.name)+5)
	dir_buf = append(dir_buf, opts.dir...)
	if len(dir_buf) > 0 && dir_buf[len(dir_buf)-1] != '/' {
		dir_buf = append(dir_buf, '/')
	}
	dir_buf = metricToDir(dir_buf, opts.name)
	dir := string(dir_buf)

	// open it up and read all of the names in it
	dh, err := os.Open(dir)
	if os.IsNotExist(err) && !opts.ro {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, Error.Wrap(err)
		}
		dh, err = os.Open(dir)
	}
	if err != nil {
		return nil, Error.Wrap(err)
	}
	defer dh.Close()

	names, err := dh.Readdirnames(-1)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	// interned keeps track of all of the path strings for a given metric. it
	// is lazily created by filenameAt, but we pre-fill it while walking the
	// directory.
	var fb filenameBuf
	interned := make(map[int]string)

	// compute the first and last metric files for the metric. if first == last
	// then we know there was at most one file. if last == 0, we know there
	// are zero files.
	first, last := 0, 0
	first_set := false
	for _, name := range names {
		if !strings.HasSuffix(name, ".data") {
			continue
		}

		val, err := strconv.ParseInt(name[:len(name)-5], 10, 0)
		if err != nil {
			continue
		}
		iv := int(val)
		interned[iv] = fb.metricFilenameAt(dir, iv)

		if iv > last {
			last = iv
		}
		if iv < first || !first_set {
			first = iv
			first_set = true
		}
	}

	// if we didn't set the first, create an empty file at 1. this simplifies
	// the logic because we can from now on assume that at least one possibly
	// empty file exists.
	if !first_set {
		path := fb.metricFilenameAt(dir, 1)
		f, err := opts.fch.acquireFile(ctx, path, false)
		if err != nil {
			return nil, err
		}
		opts.fch.releaseFile(path, f)

		first, last = 1, 1
	}

	return &metric{
		opts:  opts,
		dir:   dir,
		first: first,
		last:  last,

		fb:       fb,
		interned: interned,
	}, nil
}

// filenameAt returns the filename for the data file at the index.
func (m *metric) filenameAt(index int) string {
	if path, ok := m.interned[index]; ok {
		return path
	}
	path := m.fb.metricFilenameAt(m.dir, index)
	m.interned[index] = path
	return path
}

// acquireLast finds the last file available, what it's head position is and
// which file number it is. release must be called on the returned handle.
// after this call, m.last is set to the last non-empty file, or the first
// empty file if there is no data. f refers to the file opened at m.last.
func (m *metric) acquireLast(ctx context.Context) (f file, head int,
	err error) {

	// while there is a last file
	for m.last >= m.first {
		path := m.filenameAt(m.last)
		f, err := m.opts.fch.acquireFile(ctx, path, true)
		if err != nil {
			return file{}, 0, err
		}

		// get the head pointer for the file.
		head, err := getHeadPointer(ctx, f)
		if err != nil {
			m.opts.fch.releaseFile(path, f)
			return file{}, 0, err
		}

		// if the file contains no records, check a previous file as long as
		// there exists one.
		if head == f.Capacity()-1 && m.last > m.first {
			// since this file is empty and there is an earlier file, we should
			// remove it.
			m.opts.fch.releaseFile(path, f)
			os.Remove(path)
			m.last--

			continue
		}

		// we have successfully found the last file.
		return f, head, nil
	}

	return file{}, 0, Error.New("unable to acquire last file")
}

// Write stores the data in the metric if there is room and the data is
// chronologically later than the last data stored. additionally, it cleans
// any files older than max if it had to allocate a new file. it returns if
// the data was written. this method is not safe to be called concurrently.
func (m *metric) Write(ctx context.Context, start, end int64, data []byte) (
	ok bool, err error) {

	// acquire the last file and determine where the head pointer is for it.
	f, head, err := m.acquireLast(ctx)
	if err != nil {
		return false, err
	}
	defer m.opts.fch.releaseFile(m.filenameAt(m.last), f)

	// point head at the first valid record (or out of bounds at the capacity)
	head++

	// if the last file has a record, ensure monotonicity with it.
	if f.HasRecord(ctx, head) {
		last_rec, err := f.Record(ctx, head)
		if err != nil {
			return false, err
		}
		if last_rec.end >= end {
			return false, nil
		}
	}

	// ensure we have capacity to write the value in the last file and create
	// a new file if necessary.
	nr := numRecords(len(data), f.Size())
	if nr == 0 {
		return false, Error.New("unable to compute number of records")
	}
	if head-nr < 0 {
		last_rec := f.Capacity() - 1

		// if this is the first write to an empty file, then we will not be
		// able to write it.
		if m.last == 1 && head == last_rec {
			return false, Error.New("value too large for empty file")
		}

		// we have to allocate a new file, so remove the first file and bump
		// the first number. since it may be in the cache, we need to evict
		// it from the cache if it is there. bump first once the file is
		// removed.
		if m.last-m.first >= m.opts.max && m.opts.max > 0 {
			first_path := m.filenameAt(m.first)
			m.opts.fch.evictFile(first_path)
			os.Remove(first_path)
			m.first++
		}

		// bump last, open the new handle, and reset the head pointer.
		m.last++

		path := m.filenameAt(m.last)
		f, err = m.opts.fch.acquireFile(ctx, path, false)
		if err != nil {
			return false, err
		}
		defer m.opts.fch.releaseFile(path, f)

		head = f.Capacity()

		// update our capacity check and try again. if it fails now, the file
		// is just too large to write. due to fixed size issues with mmap, we
		// can't allow it to write past the capacity.
		nr = numRecords(len(data), f.Size())
		if nr == 0 {
			return false, Error.New("unable to compute number of records")
		}
		if head-nr < 0 {
			return false, Error.New("value too large for empty file")
		}
	}

	// we are gauranteed that head >= nr, so subtract it to copy in the
	// records.
	head -= nr
	new_head := head - 1

	// write the records into the file
	err = iterateRecords(start, end, data, f.Size(),
		func(rec record) error {
			err := f.SetRecord(ctx, head, rec)
			head++
			return err
		})
	if err != nil {
		return false, err
	}

	// update the metadata to point at the new head, and update the start and
	// end values. this helps for searching and bounds checking.
	meta, err := f.Metadata(ctx)
	if err != nil {
		return true, err
	}
	meta.Head = new_head
	meta.End = end
	if meta.Start == 0 {
		meta.Start = start
	}
	if meta.SmallestEnd == 0 || end < meta.SmallestEnd {
		meta.SmallestEnd = end
	}

	err = f.SetMetadata(ctx, meta)
	if err != nil {
		return true, err
	}

	return true, nil
}

// Read returns all of the writes that are strictly before end. it appends the
// data to the provided buf and runs the provided callback. the data slice is
// reused between callback calls, so callers must ensure they do not keep
// references to the data slice after returning. if the callback returns an
// error, the iteration is stopped and the error is returned. if the callback
// returns false, the iteration is stopped.
func (m *metric) Read(ctx context.Context, end int64, buf []byte,
	cb func(ctx context.Context, start, end int64, data []byte) (
		bool, error)) error {

	// since we expect most queries to be for the most recent data, we do a
	// simple strategy that optimizes for sequential reads: we start at the
	// last file, use the metadata per file to skip ones that are unlikely to
	// contain any data, and then linerally walk until we have records to call
	// back.

	for num := m.last; num >= m.first; num-- {
		ok, err := func() (ok bool, err error) {
			// load up the file at num so that we can start reading records.
			path := m.filenameAt(num)
			f, err := m.opts.fch.acquireFile(ctx, path, true)
			if err != nil {
				return false, err
			}
			defer m.opts.fch.releaseFile(path, f)

			// check if the file would have any data. if not, just skip it.
			meta, err := f.Metadata(ctx)
			if err != nil {
				return false, err
			}
			if meta.SmallestEnd >= end {
				return false, nil
			}

			capacity := f.Capacity()
			head, err := getHeadPointer(ctx, f)
			if err != nil {
				return false, err
			}
			// we know the first record starts at head+1
			head++

			// start reading off values
		new_record:
			for head < capacity {
				// we know that data values are fully contained in files: if we
				// see a start record, we know we can keep reading inside of
				// this file until the end.
				//
				// also, all of the records involved with a value have the same
				// start and end times in them, so we only need to check the
				// "starting" record for the condition that we're done.

				rec, err := f.Record(ctx, head)
				head++
				if err != nil {
					// drop any records we have errors reading (checksums, etc)
					external.Errorw("error reading record",
						"err", err,
					)
					continue new_record
				}

				// if the record ends after the end time, we can skip it. we
				// dont need to check elsewhere because every other record
				// must have the same timestamps.
				if rec.end >= end {
					continue new_record
				}

				// TODO(jeff): we could be opening these files as read only,
				// but that would complicate the caching semantics. maybe
				// it's worth it to avoid this copy though.
				buf = append(buf[:0], rec.data...)

				// if we have a complete record, bump the head pointer and
				// move to the next record.
				if rec.kind == recordKind_complete {
					ok, err := cb(ctx, rec.start, rec.end, buf)
					if err != nil {
						return false, err
					}
					if !ok {
						return false, nil
					}
					select {
					case <-ctx.Done():
						return false, ctx.Err()
					default:
					}
					continue new_record
				}

				// if it's not a begin, we have some data integrity error.
				if rec.kind != recordKind_begin {
					// drop any records we have with errors
					external.Errorw("invalid record kind",
						"kind", rec.kind,
						"expected", recordKind_end,
					)
					continue new_record
				}

				// read records and append them to the buf while we're getting
				// continues.
				for {
					rec, err := f.Record(ctx, head)
					head++
					if err != nil {
						// drop any records we have with errors
						external.Errorw("error reading record",
							"err", err,
						)
						continue new_record
					}

					buf = append(buf, rec.data...)

					if rec.kind == recordKind_continue {
						continue
					}

					// if we see anything other than an end at this point, it's
					// definitely a problem.
					if rec.kind != recordKind_end {
						// drop any records we have with errors
						external.Errorw("invalid record kind",
							"kind", rec.kind,
							"expected", recordKind_end,
						)
						continue new_record
					}

					// ok we're done with that value, callback and move on to
					// the next value.
					ok, err = cb(ctx, rec.start, rec.end, buf)
					if err != nil {
						return false, err
					}
					if !ok {
						return false, nil
					}
					select {
					case <-ctx.Done():
						return false, ctx.Err()
					default:
					}
					continue new_record
				}
			}

			// we may have to go to the next file, so we're not done yet.
			return true, nil
		}()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	return nil
}

// ReadLast reads the last value out of the metric. buf is used as storage for
// the data slice if possible. Returns 0, 0, nil, nil if there is no data.
func (m *metric) ReadLast(ctx context.Context, buf []byte) (
	start, end int64, data []byte, err error) {

	err = m.Read(ctx, 1<<63-1, buf,
		func(ctx context.Context, cb_start, cb_end int64, cb_data []byte) (
			bool, error) {

			start = cb_start
			end = cb_end
			data = cb_data

			return false, nil
		})

	return start, end, data, err
}

//
// debugging helpers
//

// dump outputs a summary of all the records and files for the metric to the
// writer. useful for debugging.
func (m *metric) dump(ctx context.Context, w io.Writer) (err error) {
	for i := m.first; i <= m.last; i++ {
		path := m.filenameAt(i)
		f, err := m.opts.fch.acquireFile(ctx, path, true)
		if err != nil {
			return err
		}

		meta, err := f.Metadata(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, path, fmt.Sprintf("%+v", meta))

		for n := 0; n < f.Capacity(); n++ {
			if f.HasRecord(ctx, n) {
				rec, err := f.Record(ctx, n)
				if err != nil {
					return err
				}
				rec.data = nil
				fmt.Fprintln(w, "\t", n, "\t", fmt.Sprintf("%+v", rec))
			} else {
				fmt.Fprintln(w, "\t", n, "\t", "<no record>")
			}
		}

		m.opts.fch.releaseFile(path, f)
	}

	return nil
}
