// Copyright (C) 2018. See AUTHORS.

package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"time"

	"github.com/vivint/rothko/api/query"
	"github.com/vivint/rothko/data"
	"github.com/vivint/rothko/data/load"
	"github.com/vivint/rothko/database"
	"github.com/vivint/rothko/dist/tdigest"
	"github.com/vivint/rothko/draw/colors"
	"github.com/vivint/rothko/draw/graph"
	"github.com/vivint/rothko/external"
	"github.com/vivint/rothko/merge"
	"github.com/zeebo/errs"
)

// Options for the server.
type Options struct {
	// Origin is sent back in Access-Control-Allow-Origin. If not set, sends
	// back '*'.
	Origin string

	// Username and Password control basic auth to the server. If unset, no
	// basic auth will be required.
	Username string
	Password string
}

// Server is an http.Handler that can serve responses for a frontend.
type Server struct {
	db     database.DB
	static http.Handler
	opts   Options

	username_hash [sha256.Size]byte
	password_hash [sha256.Size]byte
	nonce         string
}

// New returns a new Server.
func New(db database.DB, static http.Handler, opts Options) *Server {
	nonce := make([]byte, 16)
	rand.Read(nonce)
	if opts.Origin == "" {
		opts.Origin = "*"
	}
	return &Server{
		db:     db,
		static: static,
		opts:   opts,

		username_hash: sha256.Sum256([]byte(opts.Username)),
		password_hash: sha256.Sum256([]byte(opts.Password)),
		nonce:         hex.EncodeToString(nonce),
	}
}

// ServeHTTP implements the http.Handler interface for the server. It just
// looks at the method and last path component to route.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	now := time.Now()

	w.Header().Set("Access-Control-Allow-Origin", s.opts.Origin)

	tracker := &respTracker{ResponseWriter: w, wrote: false, code: 0}
	if err := s.serveHTTP(req.Context(), tracker, req); err != nil {
		if !tracker.wrote {
			w.WriteHeader(getStatusCode(err))
			fmt.Fprintf(w, "%+v\n", err)
		}
	}

	// squelch requests for the nonce
	if req.URL.Path == "/api/nonce" {
		return
	}

	external.Infow("http request",
		"status", tracker.code,
		"method", req.Method,
		"duration", time.Since(now).Round(time.Millisecond),
		"path", req.URL.Path,
	)
}

// validAuth returns true if the auth provided is valid.
func (s *Server) validAuth(ctx context.Context,
	username, password string) bool {

	username_hash := sha256.Sum256([]byte(username))
	username_good := subtle.ConstantTimeCompare(
		username_hash[:], s.username_hash[:])

	password_hash := sha256.Sum256([]byte(password))
	password_good := subtle.ConstantTimeCompare(
		password_hash[:], s.password_hash[:])

	return username_good&password_good == 1
}

// serveHTTP is a little wrapper for making error handling easier.
func (s *Server) serveHTTP(ctx context.Context, w http.ResponseWriter,
	req *http.Request) (err error) {

	if s.opts.Username != "" {
		username, password, ok := req.BasicAuth()
		if !ok || !s.validAuth(ctx, username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="rothko"`)
			return errUnauthorized.New("")
		}
	}

	if req.Method != "GET" {
		return errMethodNotAllowed.New("%s", req.Method)
	}

	switch req.URL.Path {
	case "/api/render":
		return s.serveRender(ctx, w, req)

	case "/api/query":
		return s.serveQuery(ctx, w, req)

	case "/api/nonce":
		return s.serveNonce(ctx, w, req)

	default:
		if s.static != nil {
			s.static.ServeHTTP(w, req)
			return nil
		}
		return errNotFound.New("path: %q", req.URL.Path)
	}
}

// serveRender serves either a png of the graph, or a json encoded set of
// columns, and the earliest data so that the frontend can draw the graph.
func (s *Server) serveRender(ctx context.Context, w http.ResponseWriter,
	req *http.Request) (err error) {

	// get the render parameters
	metric := req.FormValue("metric")
	if metric == "" {
		return errBadRequest.New("metric required")
	}

	width := getInt(req.FormValue("width"), 1000)
	height := getInt(req.FormValue("height"), 350)
	padding := getInt(req.FormValue("padding"), 0)
	now := getInt64(req.FormValue("now"), time.Now().UnixNano())
	dur := getDuration(req.FormValue("duration"), 24*time.Hour)
	samples := getInt(req.FormValue("samples"), 30)
	compression := getFloat64(req.FormValue("compression"), 5)
	stop_before := now - dur.Nanoseconds()

	// set up some state for the query
	var measured graph.Measured
	var earliest []byte
	measure_opts := graph.MeasureOptions{
		Now:      now,
		Duration: dur,
		Width:    width,
		Height:   height,
		Padding:  padding,
	}

	merger := merge.NewMerger(merge.MergerOptions{
		Samples:  samples,
		Now:      now,
		Duration: dur,
		Params:   tdigest.Params{Compression: compression},
	})
	var ok bool

	// run the query
	err = s.db.Query(ctx, metric, now, nil,
		func(ctx context.Context, start, end int64, buf []byte) (
			bool, error) {

			// get the record ready
			var rec data.Record
			if err := rec.Unmarshal(buf); err != nil {
				return false, errs.Wrap(err)
			}

			// if we don't have an earliest yet, keep it around and set it up
			if measure_opts.Earliest == nil {
				dist, err := load.Load(ctx, rec)
				if err != nil {
					return false, errs.Wrap(err)
				}

				earliest = append(earliest[:0], buf...)
				measure_opts.Earliest = dist
				measured, ok = graph.Measure(ctx, measure_opts)
				if !ok {
					return false, nil
				}
				merger.SetWidth(measured.Width)
			}

			// push in the record
			if err := merger.Push(ctx, rec); err != nil {
				return false, errs.Wrap(err)
			}

			// keep going until we need to stop based on the duration
			return end >= stop_before, nil
		})
	if err != nil {
		return errs.Wrap(err)
	}
	if !ok {
		return errs.New("too small")
	}

	// grab the columns
	cols, err := merger.Finish(ctx)
	if err != nil {
		return errs.Wrap(err)
	}

	// if it's json, encode it out
	if req.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		type D = map[string]interface{}
		return errs.Wrap(json.NewEncoder(w).Encode(D{
			"metric":   metric,
			"columns":  cols,
			"earliest": earliest,
			"now":      now,
			"duration": dur.Nanoseconds(),
			"width":    width,
			"height":   height,
			"padding":  padding,
		}))
	}

	// if we never got an earliest, we need to measure without it to get the
	// axes ready.
	if measure_opts.Earliest == nil {
		measured, ok = graph.Measure(ctx, measure_opts)
		if !ok {
			return errs.New("too small")
		}
	}

	// draw the graph
	out := measured.Draw(ctx, graph.DrawOptions{
		Canvas:  nil,
		Columns: cols,
		Colors:  colors.Viridis,
	})

	// encode it out as a png
	w.Header().Set("Content-Type", "image/png")
	return errs.Wrap(png.Encode(w, out.AsImage()))
}

// serveQuery returns a set of metrics that match the query as a json list.
func (s *Server) serveQuery(ctx context.Context, w http.ResponseWriter,
	req *http.Request) (err error) {

	// get the render parameters
	_query := req.FormValue("query")
	if _query == "" {
		return errBadRequest.New("query required")
	}
	results := getInt(req.FormValue("results"), 10)

	search := query.New(_query, results)
	if err := s.db.Metrics(ctx, search.Add); err != nil {
		return errs.Wrap(err)
	}

	w.Header().Set("Content-Type", "application/json")
	return errs.Wrap(json.NewEncoder(w).Encode(search.Matched()))
}

// serveNonce returns a nonce associated to the server instance.
func (s *Server) serveNonce(ctx context.Context, w http.ResponseWriter,
	req *http.Request) (err error) {

	io.WriteString(w, s.nonce)
	return nil
}
