// Copyright (C) 2017. See AUTHORS.

package files

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/spacemonkeygo/rothko/disk/files/internal/pcg"
	"github.com/spacemonkeygo/rothko/internal/assert"
)

// newTestDB constructs a temporary db.
func newTestDB(t testing.TB, opts Options) (db *DB, cleanup func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "db-")
	assert.NoError(t, err)

	// t.Log("temp dir:", dir)

	return New(dir, opts), func() {
		os.RemoveAll(dir)
	}
}

func testPopulateDB(t testing.TB, db *DB, num int) (
	metrics map[string]struct{}) {

	type res struct {
		b bool
		e error
	}

	ch := make(chan res)
	sendRes := func(ok bool, err error) { ch <- res{ok, err} }
	metrics = make(map[string]struct{}, num)

	// populate uses lowercase names to support case-insensitive filesystems
	// like os x, but the deploy target should definitely be case-sensitive.
	for i := 0; i < num; i++ {
		name := strings.ToLower(randomMetric())
		for {
			if _, ok := metrics[name]; ok {
				name = strings.ToLower(randomMetric())
				continue
			}
			break
		}
		metrics[name] = struct{}{}

		db.QueueCB(ctx, name, 0, 1, make([]byte, 10), sendRes)
		r := <-ch
		assert.NoError(t, r.e)
		assert.That(t, r.b)
	}

	return metrics
}

var metricRNG pcg.PCG

func randomMetric() string {
	components := fastMod(metricRNG.Uint32(), 10) + 1
	parts := make([]string, 0, components)
	for i := 0; i < components; i++ {
		const letters = "" +
			"abcdefghijklmnopqrstuvwxyz" +
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"0123456789" +
			`/.%_-*()@\`

		length := fastMod(metricRNG.Uint32(), 50) + 1
		buf := make([]byte, length)
		for j := range buf {
			buf[j] = letters[fastMod(metricRNG.Uint32(), len(letters))]
		}

		parts = append(parts, string(buf))
	}

	return strings.Join(parts, ".")
}
