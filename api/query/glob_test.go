// Copyright (C) 2018. See AUTHORS.

package query

import (
	"testing"

	"github.com/spacemonkeygo/rothko/internal/assert"
)

func TestGlob(t *testing.T) {
	// TODO(jeff): write more tests :)

	assert.That(t, glob("abc*", "abcdefg"))
	assert.That(t, glob("abc", "abcdefg"))
	assert.That(t, glob("a*bc", "afffbcdefg"))
	assert.That(t, !glob("abc", "aabc"))
	assert.That(t, !glob("abcd", "abc"))
	assert.That(t, glob("abc", "ABC"))
}

func BenchmarkGlob(b *testing.B) {
	for i := 0; i < b.N; i++ {
		glob("asdf*asdf*asdf", "asdf1234asdf1234asdf1234")
	}
}
