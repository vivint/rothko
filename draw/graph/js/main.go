// Copyright (C) 2018. See AUTHORS.

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/spacemonkeygo/rothko/draw/colors"
	"github.com/spacemonkeygo/rothko/draw/graph"
)

func main() {
	self := js.Global.Get("self")
	self.Call("addEventListener", "message", test)
}

var ctx = context.Background()

func log(n *time.Time, name string) {
	n2 := time.Now()
	fmt.Println(name, n2.Sub(*n))
	*n = n2
}

func test() {
	self := js.Global.Get("self")
	self.Call("postMessage", js.Undefined)

	n := time.Now()
	fmt.Println("starting")

	parsed := js.Global.Get("JSON").Call("parse", columnsJson)
	log(&n, "native-json")

	columns := make([]draw.Column, 0, parsed.Length())
	for i := 0; i < parsed.Length(); i++ {
		col := parsed.Index(i)
		columns = append(columns, draw.Column{
			W:    col.Get("W").Int(),
			X:    col.Get("X").Int(),
			Data: float64s(col.Get("Data")),
		})
	}
	log(&n, "loaded")

	var rec data.Record
	var earliest dists.Dist
	earliestData, _ := base64.StdEncoding.DecodeString(earliestB64)
	rec.Unmarshal(earliestData)
	earliest, _ = dists.Load(rec)
	log(&n, "dist")

	out, _ := graph.Draw(ctx, graph.Options{
		Now:      time.Now().UnixNano(),
		Duration: 24 * time.Hour,
		Columns:  columns,
		Colors:   colors.Viridis,
		Earliest: earliest,
		Width:    1000,
		Height:   300,
		// NoAxes:   true,
	})
	log(&n, "draw")

	self.Call("postMessage", out)
}

func float64s(x *js.Object) []float64 {
	out := make([]float64, 0, x.Length())
	for i := 0; i < x.Length(); i++ {
		out = append(out, x.Index(i).Float())
	}
	return out
}
