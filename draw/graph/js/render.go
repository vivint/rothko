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

	out, err := runRender(context.Background(), obj.Get("data"))

	self.Call("postMessage", D{
		"error": errorString(err),
		"out":   out,
	})
}

func runRender(ctx context.Context, obj *js.Object) (
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

	parsed := js.Global.Get("JSON").Call("parse", obj.Get("columns").String())
	columns := make([]draw.Column, 0, parsed.Length())
	for i := 0; i < parsed.Length(); i++ {
		col := parsed.Index(i)
		columns = append(columns, draw.Column{
			W:    col.Get("W").Int(),
			X:    col.Get("X").Int(),
			Data: float64s(col.Get("Data")),
		})
	}

	earliest_buf, err := base64.StdEncoding.DecodeString(
		obj.Get("earliest").String())
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

	return graph.Measure(ctx, graph.MeasureOptions{
		Earliest: earliest,
		Now:      obj.Get("now").Int64(),
		Duration: time.Duration(obj.Get("duration").Int64()),
		Width:    obj.Get("width").Int(),
		Height:   obj.Get("height").Int(),
	}).Draw(ctx, graph.DrawOptions{
		Canvas:  nil,
		Columns: columns,
		Colors:  colors.Viridis,
	}), nil
}

func float64s(x *js.Object) []float64 {
	out := make([]float64, 0, x.Length())
	for i := 0; i < x.Length(); i++ {
		out = append(out, x.Index(i).Float())
	}
	return out
}
