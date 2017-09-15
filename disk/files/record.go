// Copyright (C) 2017. See AUTHORS.

package files

import (
	"encoding/binary"
)

// recordVersion is the version of records this package will write.
const recordVersion = 1

// record represents an individual record inside of a circular buffer file.
type record struct {
	version int8
	kind    recordKind
	start   int64
	end     int64
	size    uint16
	data    []byte
}

// we manually compute this to avoid a dependency on unsafe. it's too bad that
// unsafe.Sizeof (which is fully safe) requires an unsafe import.
const recordHeaderSize = 1 + 1 + 8 + 8 + 2

// recordKind is an enumeration of kinds of records.
type recordKind int8

const (
	// a metadata record. only the first record should be metadata
	recordKind_metadata recordKind = iota + 1

	// a complete record (all data is contained)
	recordKind_complete

	// the beginning of a record (only contains a prefix of the full data)
	recordKind_begin

	// the middle of a record (only contains some interior of the full data)
	recordKind_continue

	// the end of a record (only contains a suffix of the full data)
	recordKind_end
)

// Size returns the marshalled size of the record.
func (r record) Size() int {
	return recordHeaderSize + int(r.size)
}

// Copy copies the data using the backing array of the passed in buf.
func (r *record) Copy(buf []byte) {
	r.data = append(buf[:0], r.data...)
}

// MarshalHeader appends a record header to the provided buf, returning it.
func (r record) MarshalHeader(buf []byte) []byte {
	// resize buf once if necessary
	if size := r.Size(); cap(buf) < size {
		buf = make([]byte, 0, size)
	}

	var scratch [8]byte

	buf = append(buf, uint8(r.version))
	buf = append(buf, uint8(r.kind))

	binary.BigEndian.PutUint64(scratch[:], uint64(r.start))
	buf = append(buf, scratch[:8]...)

	binary.BigEndian.PutUint64(scratch[:], uint64(r.end))
	buf = append(buf, scratch[:8]...)

	binary.BigEndian.PutUint16(scratch[:], r.size)
	buf = append(buf, scratch[:2]...)

	return buf
}

// Marshal appends a record to the provided buf, returning it.
func (r record) Marshal(buf []byte) []byte {
	buf = r.MarshalHeader(buf)
	buf = append(buf, r.data[:r.size]...)
	return buf
}

// consume eats size bytes from the front of the buffer, returning the eaten
// bytes, a slice without the eaten bytes, and an error if there weren't enough
// bytes to eat.
func consume(buf []byte, size int) ([]byte, []byte, error) {
	if len(buf) < size {
		return nil, nil, Error.New(
			"buf not big enough. needed %d. got %d", size, len(buf))
	}
	return buf[:size], buf[size:], nil
}

// parse reads a record out of the byte slice. it returns an error if there is
// not enough data to be a full record.
func parse(buf []byte) (out record, err error) {
	data, buf, err := consume(buf, 1)
	if err != nil {
		return out, err
	}
	if data[0] != recordVersion {
		return out, Error.New("invalid version: %d", buf[0])
	}
	out.version = recordVersion

	data, buf, err = consume(buf, 1)
	if err != nil {
		return out, err
	}
	out.kind = recordKind(data[0])

	data, buf, err = consume(buf, 8)
	if err != nil {
		return out, err
	}
	out.start = int64(binary.BigEndian.Uint64(data))

	data, buf, err = consume(buf, 8)
	if err != nil {
		return out, err
	}
	out.end = int64(binary.BigEndian.Uint64(data))

	data, buf, err = consume(buf, 2)
	if err != nil {
		return out, err
	}
	out.size = binary.BigEndian.Uint16(data)

	data, buf, err = consume(buf, int(out.size))
	if err != nil {
		return out, err
	}
	out.data = data

	return out, nil
}

// records chunks up the data into individual records whose marshalled size is
// at most size. The records are passed to the callback function. If the
// function returns false, the iteration stops. Errors if size is not inside
// of a range to produce valid records.
func records(start, end int64, data []byte, size int,
	fn func(rec record) bool) error {

	chunk := size - recordHeaderSize
	if chunk < 0 || int(uint16(chunk)) != chunk {
		return Error.New("invalid size value")
	}

	complete := len(data) <= chunk
	kind := recordKind_begin

	for len(data) > chunk {
		cont := fn(record{
			version: recordVersion,
			kind:    kind,
			start:   start,
			end:     end,
			size:    uint16(chunk),
			data:    data[:chunk],
		})
		if !cont {
			return nil
		}

		data = data[chunk:]
		kind = recordKind_continue
	}

	if complete {
		kind = recordKind_complete
	} else {
		kind = recordKind_end
	}

	fn(record{
		version: recordVersion,
		kind:    kind,
		start:   start,
		end:     end,
		size:    uint16(len(data)),
		data:    data,
	})

	return nil
}
