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

// metric abstracts logic around reading and writing data for a metric.
type metric struct {
	opts  metricOptions
	dir   string
	first int
	last  int
}

// newMetric constructs a metric for some path. only one metric instance should
// be in use at a time for a given directory.
func newMetric(opts metricOptions) (*metric, error) {
	// get the base directory
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
		path := metricFilenameAt(dir, 1)
		f, err := opts.fch.acquireFile(path, false)
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
	}, nil
}

// metricFilenameAt returns the data file for the index.
func metricFilenameAt(dir string, index int) string {
	// careful about allocations!
	out := make([]byte, 0, len(dir)+12)
	out = append(out, dir...)
	if len(out) > 0 && out[len(out)-1] != '/' {
		out = append(out, '/')
	}
	out = strconv.AppendInt(out, int64(index), 10)
	out = append(out, ".data"...)
	return string(out)
}

// acquireLast finds the last file available, what it's head position is and
// which file number it is. release must be called on the returned handle.
// after this call, m.last is set to the last non-empty file, or the first
// empty file if there is no data. f refers to the file opened at m.last.
func (m *metric) acquireLast(ctx context.Context) (f file, head int,
	err error) {
	defer mon.Task()(&ctx)(&err)

	// while there is a last file
	for m.last >= m.first {
		path := metricFilenameAt(m.dir, m.last)
		f, err := m.opts.fch.acquireFile(path, true)
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

// write stores the data in the metric if there is room and the data is
// chronologically later than the last data stored. additionally, it cleans
// any files older than max if it had to allocate a new file. it returns if
// the data was written. this method is not safe to be called concurrently.
func (m *metric) write(ctx context.Context, start, end int64, data []byte) (
	ok bool, err error) {
	defer mon.Task()(&ctx)(&err)

	// acquire the last file and determine where the head pointer is for it.
	f, head, err := m.acquireLast(ctx)
	if err != nil {
		return false, err
	}
	defer m.opts.fch.releaseFile(metricFilenameAt(m.dir, m.last), f)

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
			first_path := metricFilenameAt(m.dir, m.first)
			m.opts.fch.evictFile(first_path)
			os.Remove(first_path)
			m.first++
		}

		// bump last, open the new handle, and reset the head pointer.
		m.last++
		head = 0

		path := metricFilenameAt(m.dir, m.last)
		f, err = m.opts.fch.acquireFile(path, false)
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

	// update the metadata to point at the new head
	meta, err := f.Metadata(ctx)
	if err != nil {
		return true, err
	}
	meta.Head = head

	err = f.SetMetadata(ctx, meta)
	if err != nil {
		return true, err
	}

	return true, nil
}
