// Copyright (C) 2018. See AUTHORS.

package query

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestSpec(t *testing.T) {
	// TODO(jeff): write more tests :)

	assert.That(t, newSpec("foo.bar.baz").Match("a.foo.bar.baz.b"))
	assert.That(t, newSpec("foo.bar.baz").Match("a.fool.barl.bazl.b"))
	assert.That(t, !newSpec("foo.bar.baz").Match("a.foe.barl.bazl.b"))
	assert.That(t, !newSpec("foo.bar.baz").Match("a.foo.baz.bazl.b"))
	assert.That(t, !newSpec("foo.bar.baz").Match("a.foo.bar.baf.b"))
	assert.That(t, newSpec("a*p*m*").Match("ActionsPerMinute"))
	assert.That(t, newSpec("_re").Match("success_times_recent"))
	assert.That(t, newSpec("recent").Match("success_times_recent"))
}

var (
	specCreateSink spec
	specMatchSink  bool
)

func BenchmarkSpec(b *testing.B) {
	b.Run("Create", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			specCreateSink = newSpec("foo.bar.baz")
		}
	})

	b.Run("Match", func(b *testing.B) {
		spec := newSpec("foo.bar.baz")

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			specMatchSink = spec.Match("abc.foo.bar.baz.def")
		}
	})
}
