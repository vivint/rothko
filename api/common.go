// Copyright (C) 2018. See AUTHORS.

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/zeebo/errs"
	"github.com/zeebo/errs/errdata"
)

var (
	errNotFound         = errs.Class("not found")
	errMethodNotAllowed = errs.Class("method not allowed")
	errBadRequest       = errs.Class("bad request")
)

type statusCode struct{}

func init() {
	errdata.Set(&errNotFound, statusCode{}, http.StatusNotFound)
	errdata.Set(&errMethodNotAllowed, statusCode{}, http.StatusMethodNotAllowed)
	errdata.Set(&errBadRequest, statusCode{}, http.StatusBadRequest)
}

func getStatusCode(err error) int {
	if code, ok := errdata.Get(err, statusCode{}).(int); ok {
		return code
	}
	return http.StatusInternalServerError
}

type respTracker struct {
	http.ResponseWriter
	wrote bool
	code  int
}

func (r *respTracker) Write(p []byte) (n int, err error) {
	if !r.wrote {
		r.code = 200
	}
	r.wrote = true
	return r.ResponseWriter.Write(p)
}

func (r *respTracker) WriteHeader(code int) {
	if !r.wrote {
		r.code = code
	}
	r.wrote = true
	r.ResponseWriter.WriteHeader(code)
}

func getInt64(x string, def int64) int64 {
	if val, err := strconv.ParseInt(x, 10, 64); err == nil {
		return val
	}
	return def
}

func getInt(x string, def int) int {
	if val, err := strconv.ParseInt(x, 10, 0); err == nil {
		return int(val)
	}
	return def
}

func getDuration(x string, def time.Duration) time.Duration {
	if val, err := time.ParseDuration(x); err == nil {
		return val
	}
	return def
}

func getFloat64(x string, def float64) float64 {
	if val, err := strconv.ParseFloat(x, 64); err == nil {
		return val
	}
	return def
}
