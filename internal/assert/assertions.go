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

func Error(t testing.TB, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error")
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

func Nil(t testing.TB, a interface{}) {
	t.Helper()

	if a == nil {
		return
	}

	rv := reflect.ValueOf(a)
	if !canNil(rv) {
		t.Fatal("%#v cannot be nil", a)
	}
	if !rv.IsNil() {
		t.Fatal("%#v != nil", a)
	}
}

func NotNil(t testing.TB, a interface{}) {
	t.Helper()

	if a == nil {
		t.Fatal("expected not nil")
	}

	rv := reflect.ValueOf(a)
	if !canNil(rv) {
		return
	}
	if rv.IsNil() {
		t.Fatal("%#v == nil", a)
	}
}

func canNil(rv reflect.Value) bool {
	if !rv.IsValid() {
		return false
	}
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice:
		return true
	}
	return false
}
