// Copyright (C) 2018. See AUTHORS.

package files

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
	"github.com/spacemonkeygo/rothko/internal/pcg"
)

func TestRecords(t *testing.T) {
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}

	t.Run("Complete", func(t *testing.T) {
		var out []record

		err := iterateRecords(1234, 5678, data, 1024, func(rec record) error {
			out = append(out, rec)
			return nil
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
	})

	t.Run("Checksum", func(t *testing.T) {
		data := make([]byte, 100)
		rec := record{
			version: recordVersion,
			kind:    recordKind_complete,
			start:   1234,
			end:     5678,
			size:    100,
			data:    data,
		}
		out := rec.Marshal(nil)
		var rng pcg.PCG

		for i := 0; i < 10000; i++ {
			// remarshal, and goof it up!
			out = rec.Marshal(out[:0])

			var goof byte
			for goof == 0 {
				goof = uint8(rng.Uint32())
			}

			out[fastMod(rng.Uint32(), len(out))] += goof

			_, err := parse(out)
			assert.Error(t, err)
		}
	})

	t.Run("Split", func(t *testing.T) {
		var out []record

		err := iterateRecords(1234, 5678, data, 54, func(rec record) error {
			out = append(out, rec)
			return nil
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

		assert.Equal(t, len(out[0].Marshal(nil)), 54)
		assert.Equal(t, len(out[1].Marshal(nil)), 54)
		assert.Equal(t, len(out[2].Marshal(nil)), 54)
		assert.That(t, len(out[3].Marshal(nil)) <= 54)
	})
}

func BenchmarkRecords(b *testing.B) {
	data := make([]byte, 100)
	rec := record{
		version: recordVersion,
		kind:    recordKind_complete,
		start:   1234,
		end:     5678,
		size:    100,
		data:    data,
	}
	out := rec.Marshal(nil)

	b.Run("Marshal", func(b *testing.B) {
		buf := make([]byte, len(out))
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			rec.Marshal(buf[:0])
		}
	})

	b.Run("Parse", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			parse(out)
		}
	})

	b.Run("Complete", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			iterateRecords(1234, 5678, data, 1024, func(rec record) error {
				return nil
			})
		}
	})

	b.Run("Split", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			iterateRecords(1234, 5678, data, 50, func(rec record) error {
				return nil
			})
		}
	})

}
