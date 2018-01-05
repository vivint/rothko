// Copyright (C) 2017. See AUTHORS.

package files

import (
	"encoding/binary"
	"hash/crc32"
)

// we use castagnoli for the crc checksum for version 2 and above
var castTable = crc32.MakeTable(crc32.Castagnoli)

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
const recordHeaderSize = 1 + 1 + 8 + 8 + 2 + 4

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

// MarshalHeader writes a record header to the provided buf, returning it.
func (r record) MarshalHeader(buf []byte) []byte {
	// resize buf once if necessary
	if size := r.Size(); cap(buf) < size {
		buf = make([]byte, 0, size)
	}

	// help out bounds checking
	buf = buf[:recordHeaderSize]

	buf[0] = uint8(r.version)
	buf[1] = uint8(r.kind)
	binary.BigEndian.PutUint64(buf[2:10], uint64(r.start))
	binary.BigEndian.PutUint64(buf[10:18], uint64(r.end))
	binary.BigEndian.PutUint16(buf[18:20], r.size)

	// the crc is everything but the last 4 bytes of the record header followed
	// by the data.
	var crc uint32
	crc = crc32.Update(crc, castTable, buf[:recordHeaderSize-4])
	crc = crc32.Update(crc, castTable, r.data[:r.size])
	binary.BigEndian.PutUint32(buf[20:24], crc)

	return buf
}

// Marshal appends a record to the provided buf, returning it.
func (r record) Marshal(buf []byte) []byte {
	buf = r.MarshalHeader(buf)
	buf = append(buf, r.data[:r.size]...)
	return buf
}

// parse reads a record out of the byte slice. it returns an error if there is
// not enough data to be a full record.
func parse(buf []byte) (out record, err error) {
	if len(buf) < recordHeaderSize {
		return out, Error.New("record buf not big enough for header")
	}

	out.version = int8(buf[0])
	if out.version != recordVersion {
		return out, Error.New("unknown record header version: %d", out.version)
	}

	out.kind = recordKind(buf[1])
	out.start = int64(binary.BigEndian.Uint64(buf[2:10]))
	out.end = int64(binary.BigEndian.Uint64(buf[10:18]))
	out.size = binary.BigEndian.Uint16(buf[18:20])

	data_end := recordHeaderSize + int(out.size)
	if len(buf) < data_end {
		return out, Error.New("record buf not big enough for data")
	}
	out.data = buf[recordHeaderSize:data_end]

	// the crc is everything but the last 4 bytes of the record header
	// followed by the data.
	var crc uint32
	crc = crc32.Update(crc, castTable, buf[:recordHeaderSize-4])
	crc = crc32.Update(crc, castTable, out.data)
	if disk_crc := binary.BigEndian.Uint32(buf[20:24]); crc != disk_crc {
		return out, Error.New("crc mismatch: %x != disk %x", crc, disk_crc)
	}

	return out, nil
}

// numRecords returns the number of records that will be used. Returns 0 if the
// size is not in a range that is valid.
func numRecords(len, size int) int {
	chunk := size - recordHeaderSize
	if chunk < 0 || int(uint16(chunk)) != chunk {
		return 0
	}
	if len == 0 {
		return 1
	}
	return (len + chunk - 1) / chunk
}

// iterateRecords chunks up the data into individual records whose marshalled
// size is at most size. The records are passed to the callback function.
// If the function returns an error, the iteration stops. Errors if size is not
// inside of a range to produce valid records.
func iterateRecords(start, end int64, data []byte, size int,
	fn func(rec record) error) error {

	chunk := size - recordHeaderSize
	if chunk < 0 || int(uint16(chunk)) != chunk {
		return Error.New("invalid size value")
	}

	complete := len(data) <= chunk
	kind := recordKind_begin

	for len(data) > chunk {
		err := fn(record{
			version: recordVersion,
			kind:    kind,
			start:   start,
			end:     end,
			size:    uint16(chunk),
			data:    data[:chunk],
		})
		if err != nil {
			return err
		}

		data = data[chunk:]
		kind = recordKind_continue
	}

	if complete {
		kind = recordKind_complete
	} else {
		kind = recordKind_end
	}

	return fn(record{
		version: recordVersion,
		kind:    kind,
		start:   start,
		end:     end,
		size:    uint16(len(data)),
		data:    data,
	})
}
