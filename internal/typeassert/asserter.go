// Copyright (C) 2018. See AUTHORS.

package typeassert

import (
	"fmt"

	"github.com/zeebo/errs"
)

// Asserter helps type assertions with an "all-or-nothing" style API.
type Asserter struct {
	x    interface{}
	path string
	err  *error
}

// A wraps the value in an Asserter.
func A(x interface{}) *Asserter {
	return &Asserter{
		x:    x,
		err:  new(error),
		path: "",
	}
}

// Err returns an error if any of the assertions failed. If the error is not
// nil, none of the assertions are valid.
func (a *Asserter) Err() error {
	return *a.err
}

// a wraps the value in an asserter with the same parent error.
func (a *Asserter) a(x interface{}, path string) *Asserter {
	return &Asserter{
		x:    x,
		path: a.path + path,
		err:  a.err,
	}
}

// V returns the current value pointed at by the Asserter.
func (a *Asserter) V() interface{} {
	return a.x
}

// I indexes into a map[string]interface{}.
func (a *Asserter) I(index string) *Asserter {
	if *a.err != nil {
		return a
	}

	path := "." + index

	if a.x == nil {
		return a.a(nil, path)
	}

	m, ok := a.x.(map[string]interface{})
	if !ok {
		*a.err = errs.New("invalid type: map[string]interface{} != %T at %s",
			a.x, a.path)
		return a
	}

	return a.a(m[index], path)
}

// N indexes into a []interface{}.
func (a *Asserter) N(index int) *Asserter {
	if *a.err != nil {
		return a
	}
	path := fmt.Sprintf("[%d]", index)

	if a.x == nil {
		return a.a(nil, path)
	}

	m, ok := a.x.([]interface{})
	if !ok {
		*a.err = errs.New("invalid type: map[string]interface{} != %T at %s",
			a.x, a.path)
		return a
	}
	if index >= len(m) {
		*a.err = errs.New("array out of bounds")
		return a
	}

	return a.a(m[index], path)
}

// Int asserts the value as an int.
func (a *Asserter) Int() int {
	if *a.err != nil || a.x == nil {
		return 0
	}
	m, ok := a.x.(int)
	if !ok {
		*a.err = errs.New("invalid type: int != %T at %s", a.x, a.path)
	}
	return m
}

// Int64 asserts the value as an int64.
func (a *Asserter) Int64() int64 {
	if *a.err != nil || a.x == nil {
		return 0
	}
	m, ok := a.x.(int64)
	if !ok {
		*a.err = errs.New("invalid type: int64 != %T at %s", a.x, a.path)
	}
	return m
}

// String asserts the value as a string.
func (a *Asserter) String() string {
	if *a.err != nil || a.x == nil {
		return ""
	}
	m, ok := a.x.(string)
	if !ok {
		*a.err = errs.New("invalid type: string != %T at %s", a.x, a.path)
	}
	return m
}

// Bool asserts the value as a bool.
func (a *Asserter) Bool() bool {
	if *a.err != nil || a.x == nil {
		return false
	}
	m, ok := a.x.(bool)
	if !ok {
		*a.err = errs.New("invalid type: bool != %T at %s", a.x, a.path)
	}
	return m
}

// Float64 asserts the value as a float64.
func (a *Asserter) Float64() float64 {
	if *a.err != nil || a.x == nil {
		return 0
	}
	m, ok := a.x.(float64)
	if !ok {
		*a.err = errs.New("invalid type: float64 != %T at %s", a.x, a.path)
	}
	return m
}
