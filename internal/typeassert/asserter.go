// Copyright (C) 2018. See AUTHORS.

package typeassert

import "github.com/zeebo/errs"

// Asserter helps type assertions with an "all-or-nothing" style API.
type Asserter struct {
	x   interface{}
	err *error
}

// A wraps the value in an Asserter.
func A(x interface{}) *Asserter {
	return &Asserter{
		x:   x,
		err: new(error),
	}
}

// Err returns an error if any of the assertions failed. If the error is not
// nil, none of the assertions are valid.
func (a *Asserter) Err() error {
	return *a.err
}

// a wraps the value in an asserter with the same parent error.
func (a *Asserter) a(x interface{}) *Asserter {
	return &Asserter{
		x:   x,
		err: a.err,
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

	m, ok := a.x.(map[string]interface{})
	if !ok {
		*a.err = errs.New("invalid type: map[string]interface{} != %T", a.x)
		return a
	}

	return a.a(m[index])
}

// N indexes into a []interface{}.
func (a *Asserter) N(index int) *Asserter {
	if *a.err != nil {
		return a
	}

	m, ok := a.x.([]interface{})
	if !ok {
		*a.err = errs.New("invalid type: map[string]interface{} != %T", a.x)
		return a
	}
	if index >= len(m) {
		*a.err = errs.New("array out of bounds")
		return a
	}

	return a.a(m[index])
}

// Int asserts the value as an int.
func (a *Asserter) Int() int {
	if *a.err != nil {
		return 0
	}
	m, ok := a.x.(int)
	if !ok {
		*a.err = errs.New("invalid type: int != %T", a.x)
	}
	return m
}

// String asserts the value as a string.
func (a *Asserter) String() string {
	if *a.err != nil {
		return ""
	}
	m, ok := a.x.(string)
	if !ok {
		*a.err = errs.New("invalid type: string != %T", a.x)
	}
	return m
}

// Bool asserts the value as a bool.
func (a *Asserter) Bool() bool {
	if *a.err != nil {
		return false
	}
	m, ok := a.x.(bool)
	if !ok {
		*a.err = errs.New("invalid type: bool != %T", a.x)
	}
	return m
}

// Float64 asserts the value as a float64.
func (a *Asserter) Float64() float64 {
	if *a.err != nil {
		return 0
	}
	m, ok := a.x.(float64)
	if !ok {
		*a.err = errs.New("invalid type: float64 != %T", a.x)
	}
	return m
}
