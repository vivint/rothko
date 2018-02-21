// Copyright (C) 2018. See AUTHORS.

package sset

import (
	"testing"

	"github.com/vivint/rothko/internal/assert"
)

func collect(s *Set) (out []string) {
	s.Iter(func(key string) bool { out = append(out, key); return true })
	return out
}

func TestSet(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		s := New(0)

		s.Add("z")
		s.Add("y")
		s.Add("x")

		assert.DeepEqual(t, collect(s), []string{"x", "y", "z"})
	})

	t.Run("Len", func(t *testing.T) {
		s := New(0)

		s.Add("z")
		s.Add("y")
		s.Add("x")

		assert.Equal(t, s.Len(), 3)
	})

	t.Run("Has", func(t *testing.T) {
		s := New(0)

		s.Add("x")

		assert.That(t, s.Has("x"))
		assert.That(t, !s.Has("y"))
	})

	t.Run("Merge", func(t *testing.T) {
		a, b := New(0), New(0)

		a.Add("z")
		b.Add("a")

		a.Merge(b)

		assert.DeepEqual(t, collect(a), []string{"a", "z"})
		assert.DeepEqual(t, collect(b), []string{"a"})
	})

	t.Run("Copy", func(t *testing.T) {
		a := New(0)

		a.Add("z")
		b := a.Copy()
		a.Add("a")

		assert.DeepEqual(t, collect(a), []string{"a", "z"})
		assert.DeepEqual(t, collect(b), []string{"z"})
	})
}
