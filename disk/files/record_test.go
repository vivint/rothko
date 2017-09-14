// Copyright (C) 2017. See AUTHORS.

package files

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestRecordsComplete(t *testing.T) {
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}

	var out []record

	err := records(1234, 5678, data, 1024, func(rec record) bool {
		out = append(out, rec)
		return true
	})
	assert.NoError(t, err)

	assert.Equal(t, len(out), 1)
	assert.DeepEqual(t, out[0], record{
		version: recordVersion,
		kind:    recordKind_complete,
		start:   1234,
		end:     5678,
		size:    100,
		data:    data,
	})
}

func TestRecordsSplit(t *testing.T) {
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}

	var out []record

	err := records(1234, 5678, data, 50, func(rec record) bool {
		out = append(out, rec)
		return true
	})
	assert.NoError(t, err)

	assert.Equal(t, len(out), 4)
	assert.DeepEqual(t, out[0], record{
		version: recordVersion,
		kind:    recordKind_begin,
		start:   1234,
		end:     5678,
		size:    30,
		data:    data[0:30],
	})
	assert.DeepEqual(t, out[1], record{
		version: recordVersion,
		kind:    recordKind_continue,
		start:   1234,
		end:     5678,
		size:    30,
		data:    data[30:60],
	})
	assert.DeepEqual(t, out[2], record{
		version: recordVersion,
		kind:    recordKind_continue,
		start:   1234,
		end:     5678,
		size:    30,
		data:    data[60:90],
	})
	assert.DeepEqual(t, out[3], record{
		version: recordVersion,
		kind:    recordKind_end,
		start:   1234,
		end:     5678,
		size:    10,
		data:    data[90:100],
	})

	assert.Equal(t, len(out[0].Marshal(nil)), 50)
	assert.Equal(t, len(out[1].Marshal(nil)), 50)
	assert.Equal(t, len(out[2].Marshal(nil)), 50)
	assert.That(t, len(out[3].Marshal(nil)) <= 50)
}

func BenchmarkMarshal(b *testing.B) {
	rec := record{
		version: recordVersion,
		kind:    recordKind_complete,
		start:   1234,
		end:     5678,
		size:    100,
		data:    make([]byte, 100),
	}
	out := rec.Marshal(nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec.Marshal(out[:0])
	}
}

func BenchmarkParse(b *testing.B) {
	rec := record{
		version: recordVersion,
		kind:    recordKind_complete,
		start:   1234,
		end:     5678,
		size:    100,
		data:    make([]byte, 100),
	}
	out := rec.Marshal(nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parse(out)
	}
}

func BenchmarkRecordsComplete(b *testing.B) {
	data := make([]byte, 100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		records(1234, 5678, data, 1024, func(rec record) bool {
			return true
		})
	}
}

func BenchmarkRecordsSplit(b *testing.B) {
	data := make([]byte, 100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		records(1234, 5678, data, 50, func(rec record) bool {
			return true
		})
	}
}
