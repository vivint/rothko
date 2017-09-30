// Copyright (C) 2017. See AUTHORS.

package files

import (
	"context"
	"os"
	"strconv"
	"strings"
)

// metricOptions are all the pieces of data required to work with a metric.
type metricOptions struct {
	fch  *fileCache
	dir  string
	name string
	max  int
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
	if os.IsNotExist(err) {
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

		// find the last record contained in the file
		head, err := lastRecord(ctx, f)
		if err != nil {
			m.opts.fch.releaseFile(path, f)
			return file{}, 0, err
		}

		// if the file contains no records, check a previous file as long as
		// there exists one.
		if head == 0 && m.last > m.first {
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

	// if the last file has a record, ensure monotonicity with it.
	if f.HasRecord(ctx, head-1) {
		last_rec, err := f.Record(ctx, head-1)
		if err != nil {
			return false, err
		}
		if last_rec.start > start || last_rec.end > start {
			return false, nil
		}
	}

	// ensure we have capacity to write the value in the last file and create
	// a new file if necessary.
	nr := numRecords(len(data), f.Size())
	if nr == 0 {
		return false, Error.New("unable to compute number of records")
	}
	if head+nr > f.Capacity() {
		// if this is the first write to an empty file, then we will not be
		// able to write it.
		if m.last == 1 && head == 0 {
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
		head = 0

		path := m.filenameAt(m.last)
		f, err = m.opts.fch.acquireFile(ctx, path, false)
		if err != nil {
			return false, err
		}
		defer m.opts.fch.releaseFile(path, f)

		// update our capacity check and try again. if it fails now, the file
		// is just too large to write. due to fixed size issues with mmap, we
		// can't allow it to write past the capacity.
		nr = numRecords(len(data), f.Size())
		if nr == 0 {
			return false, Error.New("unable to compute number of records")
		}
		if head+nr > f.Capacity() {
			return false, Error.New("value too large for empty file")
		}
	}

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
	meta.Head = head
	meta.End = end
	if meta.Start == 0 {
		meta.Start = start
	}

	err = f.SetMetadata(ctx, meta)
	if err != nil {
		return true, err
	}

	return true, nil
}

// TimeRange returns the start of the first record and the end of the last
// record. if there are no records, it returns 0, 0.
func (m *metric) TimeRange(ctx context.Context) (start, end int64, err error) {
	// load up the last file
	last, head, err := m.acquireLast(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer m.opts.fch.releaseFile(m.filenameAt(m.last), last)

	// if head is 0, then we have no records for this metric
	if head == 0 {
		return 0, 0, nil
	}

	// get the last record (one before head)
	last_rec, err := last.Record(ctx, head-1)
	if err != nil {
		return 0, 0, err
	}

	// load up the first file if it is different than the last file
	var first file
	if m.last != m.first {
		first_path := m.filenameAt(m.first)
		first, err = m.opts.fch.acquireFile(ctx, first_path, true)
		if err != nil {
			return 0, 0, err
		}
		defer m.opts.fch.releaseFile(first_path, first)
	} else {
		first = last
	}

	// get the first record
	first_rec, err := first.Record(ctx, 0)
	if err != nil {
		return 0, 0, err
	}

	return first_rec.start, last_rec.end, nil
}

// startsAfter returns if the file numbered at candidate only contains records
// that start after the start time.
func (m *metric) startsAfter(ctx context.Context, cand int, start int64) (
	ok bool, err error) {

	path := m.filenameAt(cand)
	f, err := m.opts.fch.acquireFile(ctx, path, true)
	if err != nil {
		return false, err
	}
	defer m.opts.fch.releaseFile(path, f)

	rec, err := f.Record(ctx, 0)
	if err != nil {
		return false, err
	}

	return rec.start > start, nil
}

// Search returns the file, file number, and head position of the record
// that is the latest record that starts less than or equal to start. It
// returns 0, 0 if there is no record that starts less than start.
func (m *metric) Search(ctx context.Context, start int64) (num int, head int,
	err error) {

	// first do a bisection on which file we believe the record will be in
	// based on their metadata. we add one to last because we want to include
	// the last file as a possibility.
	first, last := m.first, m.last+1
	for first < last {
		cand := int(uint(first+last) >> 1)

		ok, err := m.startsAfter(ctx, cand, start)
		if err != nil {
			return 0, 0, err
		}

		// if all of the records start after the start time, we reduce the
		// upper bound since all records after it presumably contain records
		// that start after as well.
		if ok {
			last = cand
		} else {
			first = cand + 1
		}
	}

	// first is the smallest file that contains data that is strictly after
	// the start: this means the previous file must contain it, so we start
	// in that file. if that file doesn't exist, we do not have that data.
	first--
	if first < m.first || first > m.last {
		return 0, 0, nil
	}

	// we will do a binary search again inside of the file to find the spot
	// where it transitions from an earlier start to a later start. if we do
	// not find that transition point, we know the next file at index 0
	// contains that transition (but we double check to be sure).

	path := m.filenameAt(first)
	f, err := m.opts.fch.acquireFile(ctx, path, true)
	if err != nil {
		return 0, 0, err
	}
	defer m.opts.fch.releaseFile(path, f)

	last_rec, err := lastRecord(ctx, f)
	if err != nil {
		return 0, 0, err
	}

	begin, end := 0, last_rec
	for begin < end {
		cand := int(uint(begin+end) >> 1)

		rec, err := f.Record(ctx, cand)
		if err != nil {
			return 0, 0, err
		}

		if rec.start > start {
			end = cand
		} else {
			begin = cand + 1
		}
	}

	// if we found a record where it transitions, we're done!
	if begin <= last_rec {
		return first, begin - 1, nil
	}

	// if the index is after last record, we know the next file contains the
	// first record that starts greater than or equal to start. if there is
	// no file, we don't have a record that is greater than or equal to start.

	first++
	if first > m.last {
		return 0, 0, nil
	}

	path = m.filenameAt(first)
	f, err = m.opts.fch.acquireFile(ctx, path, true)
	if err != nil {
		return 0, 0, err
	}
	defer m.opts.fch.releaseFile(path, f)

	// if the next file does not have a record, then there is no record that
	// starts greater than or equal to start.
	if !f.HasRecord(ctx, 0) {
		return 0, 0, nil
	}

	rec, err := f.Record(ctx, 0)
	if err != nil {
		return 0, 0, err
	}

	if rec.start < start {
		return 0, 0, Error.New("data integrity error")
	}

	return first, 0, nil
}

// readRecord reads the n'th record out of the file at index.
func (m *metric) readRecord(ctx context.Context, index, n int) (
	rec record, err error) {

	path := m.filenameAt(index)
	f, err := m.opts.fch.acquireFile(ctx, path, true)
	if err != nil {
		return record{}, err
	}
	defer m.opts.fch.releaseFile(path, f)

	return f.Record(ctx, n)
}

// Read returns all of the writes that have any overlap with start and end. it
// appends the data to the provided buf and runs the provided callback. the
// data slice is reused between callback calls, so callers must ensure they
// do not keep references to the data slice after returning. if the callback
// returns an error, the iteration is stopped and the error is returned.
func (m *metric) Read(ctx context.Context, start, end int64, buf []byte,
	cb func(start, end int64, data []byte) error) error {

	// figure out where we should start reading
	num, head, err := m.Search(ctx, start)
	if err != nil {
		return err
	}

	// if there is no record less than or equal to start, we should start at
	// the first record we have.
	if num == 0 {
		num = m.first
		head = 0
	}

	for {
		done, err := func() (done bool, err error) {
			// load up the file at num so that we can start reading records.
			path := m.filenameAt(num)
			f, err := m.opts.fch.acquireFile(ctx, path, true)
			if err != nil {
				return false, err
			}
			defer m.opts.fch.releaseFile(path, f)

			last, err := lastRecord(ctx, f)
			if err != nil {
				return false, err
			}

			// start reading off values
			for head < last {
				// we know that data values are fully contained in files: if we
				// see a start record, we know we can keep reading inside of
				// this file until the end.
				//
				// also, all of the records involved with a value have the same
				// start and end times in them, so we only need to check the
				// "starting" record for the condition that we're done.

				rec, err := f.Record(ctx, head)
				if err != nil {
					return false, err
				}

				// if the record ends after the requested end, we no longer
				// need to check any more values.
				if rec.end > end {
					return true, nil
				}

				// TODO(jeff): we could be opening these files as read only,
				// but that would complicate the caching semantics. maybe
				// it's worth it to avoid this copy though.
				buf = append(buf[:0], rec.data...)
				head++

				// if we have a complete record, bump the head pointer and
				// move to the next record.
				if rec.kind == recordKind_complete {
					err := cb(rec.start, rec.end, buf)
					if err != nil {
						return false, err
					}
					continue
				}

				// if it's not a begin, we have some data integrity error.
				if rec.kind != recordKind_begin {
					return false, Error.New("data integrity error")
				}

				// read records and append them to the buf while we're getting
				// continues.
				for {
					rec, err := f.Record(ctx, head)
					if err != nil {
						return false, err
					}

					buf = append(buf, rec.data...)
					head++

					if rec.kind == recordKind_continue {
						continue
					}

					// if we see anything other than an end at this point, it's
					// definitely a problem.
					if rec.kind != recordKind_end {
						return false, Error.New("data integrity error")
					}

					// ok we're done with that value, callback and move on to
					// the next value.
					err = cb(rec.start, rec.end, buf)
					if err != nil {
						return false, err
					}

					break
				}
			}

			// we may have to go to the next file, so we're not done yet.
			return false, nil
		}()
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		// go to the next file. if there aren't any more files, we're done.
		num++
		if num > m.last {
			return nil
		}
		head = 0
	}
}
