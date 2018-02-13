// Copyright (C) 2018. See AUTHORS.

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/load"
	"github.com/spacemonkeygo/rothko/database/files"
	"github.com/spacemonkeygo/rothko/dist/tdigest"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/spacemonkeygo/rothko/draw/colors"
	"github.com/spacemonkeygo/rothko/draw/graph"
	"github.com/spacemonkeygo/rothko/merge"
	"github.com/zeebo/errs"
)

func runRender(ctx context.Context, di *files.DB, metric string,
	dur time.Duration) error {

	// TODO(jeff): parameterize these constants
	const (
		width       = 1000
		height      = 360
		samples     = 30
		compression = 5
	)

	now := time.Now().UnixNano()
	stop_before := now - dur.Nanoseconds()

	fmt.Println(now, dur.Nanoseconds())

	var measured graph.Measured
	measure_opts := graph.MeasureOptions{
		Now:      now,
		Duration: dur,
		Width:    width,
		Height:   height,
	}

	merger := merge.NewMerger(merge.MergerOptions{
		Samples:  samples,
		Now:      now,
		Duration: dur,
		Params:   tdigest.Params{Compression: compression},
	})

	err := di.Query(ctx, metric, now, nil,
		func(ctx context.Context, start, end int64, buf []byte) (
			bool, error) {

			var rec data.Record
			if err := rec.Unmarshal(buf); err != nil {
				return false, errs.Wrap(err)
			}

			if measure_opts.Earliest == nil {
				dist, err := load.Load(ctx, rec)
				if err != nil {
					return false, errs.Wrap(err)
				}

				fmt.Println(base64.StdEncoding.EncodeToString(buf))

				measure_opts.Earliest = dist
				measured = graph.Measure(ctx, measure_opts)
				merger.SetWidth(measured.Width)
			}

			if err := merger.Push(ctx, rec); err != nil {
				return false, errs.Wrap(err)
			}

			return end < stop_before, nil
		})
	if err != nil {
		return errs.Wrap(err)
	}

	cols, err := merger.Finish(ctx)
	if err != nil {
		return errs.Wrap(err)
	}

	json.NewEncoder(os.Stdout).Encode(cols)

	if measure_opts.Earliest == nil {
		measured = graph.Measure(ctx, measure_opts)
	}

	out := measured.Draw(ctx, graph.DrawOptions{
		Canvas:  nil,
		Columns: cols,
		Colors:  colors.Viridis,
	})

	fh, err := os.OpenFile("test.png", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fh.Close()

	return errs.Wrap(png.Encode(fh, asImage(out)))
}

func asImage(m *draw.RGB) *image.RGBA {
	return &image.RGBA{
		Pix:    m.Pix,
		Stride: m.Stride,
		Rect:   image.Rect(0, 0, m.Width, m.Height),
	}
}
