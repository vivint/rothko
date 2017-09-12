// Copyright (C) 2017. See AUTHORS.

package assert

import "testing"

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
