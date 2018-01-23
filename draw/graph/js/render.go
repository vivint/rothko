// Copyright (C) 2018. See AUTHORS.

package main

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/spacemonkeygo/rothko/draw/colors"
	"github.com/spacemonkeygo/rothko/draw/graph"
	"github.com/zeebo/errs"
)

func render(obj *js.Object) {
	self := js.Global.Get("self")
	self.Call("postMessage", js.Undefined)

	data := obj.Get("data")
	out, err := runRender(ctx, renderData{
		Columns:  data.Get("columns").String(),
		Earliest: data.Get("earliest").String(),
		Now:      data.Get("now").Int64(),
		Duration: time.Duration(data.Get("duration").Int64()),
		Width:    data.Get("width").Int(),
		Height:   data.Get("height").Int(),
	})

	self.Call("postMessage", D{
		"error": errorString(err),
		"out":   out,
	})
}

type renderData struct {
	Columns  string
	Earliest string
	Now      int64
	Duration time.Duration
	Width    int
	Height   int
}

func runRender(ctx context.Context, rdata renderData) (
	out *draw.RGB, err error) {

	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = panicErr.New("%v", r)
			}
		}
	}()

	parsed := js.Global.Get("JSON").Call("parse", rdata.Columns)
	columns := make([]draw.Column, 0, parsed.Length())
	for i := 0; i < parsed.Length(); i++ {
		col := parsed.Index(i)
		columns = append(columns, draw.Column{
			W:    col.Get("W").Int(),
			X:    col.Get("X").Int(),
			Data: float64s(col.Get("Data")),
		})
	}

	earliest_buf, err := base64.StdEncoding.DecodeString(rdata.Earliest)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	var rec data.Record
	if err := rec.Unmarshal(earliest_buf); err != nil {
		return nil, errs.Wrap(err)
	}
	earliest, err := dists.Load(rec)
	if err != nil {
		return nil, errs.Wrap(err)
	}

	return graph.Draw(ctx, graph.Options{
		Now:      rdata.Now,
		Duration: rdata.Duration,
		Columns:  columns,
		Colors:   colors.Viridis,
		Earliest: earliest,
		Width:    rdata.Width,
		Height:   rdata.Height,
	})
}

func float64s(x *js.Object) []float64 {
	out := make([]float64, 0, x.Length())
	for i := 0; i < x.Length(); i++ {
		out = append(out, x.Index(i).Float())
	}
	return out
}
