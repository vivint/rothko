// Copyright (C) 2017. See AUTHORS.

package assert

import (
	"reflect"
	"testing"
)

func NoError(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func Equal(t testing.TB, a, b interface{}) {
	t.Helper()

	if a != b {
		t.Fatalf("%#v != %#v", a, b)
	}
}

func DeepEqual(t testing.TB, a, b interface{}) {
	t.Helper()

	if !reflect.DeepEqual(a, b) {
		t.Fatalf("%#v != %#v", a, b)
	}
}

func That(t testing.TB, v bool) {
	t.Helper()

	if !v {
		t.Fatal("expected condition failed")
	}
}
