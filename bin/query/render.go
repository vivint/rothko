// Copyright (C) 2017. See AUTHORS.

package main

import (
	"context"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/spacemonkeygo/rothko/data"
	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/data/merge"
	"github.com/spacemonkeygo/rothko/disk/files"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/zeebo/errs"
)

func runRender(ctx context.Context, di *files.DB, metric string,
	dur time.Duration) error {

	// TODO(jeff): parameterize these constants
	const (
		width       = 1000
		height      = 300
		samples     = 64
		compression = 5
	)

	now := time.Now().UnixNano()
	stop_before := now - dur.Nanoseconds()

	merger := NewMerger(MergerOptions{
		Width:    width,
		Samples:  samples,
		Now:      now,
		Duration: dur,
		MergeOptions: merge.MergeOptions{
			Params: tdigest.Params{
				Compression: compression,
			},
		},
	})

	var earliest *data.Record

	err := di.Query(ctx, metric, now, nil,
		func(ctx context.Context, start, end int64, buf []byte) (
			bool, error) {

			var rec data.Record
			if err := rec.Unmarshal(buf); err != nil {
				return false, errs.Wrap(err)
			}

			if earliest == nil {
				earliest = &rec
			}

			if err := merger.Push(rec); err != nil {
				return false, errs.Wrap(err)
			}

			return end < stop_before, nil
		})
	if err != nil {
		return errs.Wrap(err)
	}

	canvas := draw.NewRGB(width, height)

	if earliest != nil {
		dist, err := dists.Load(*earliest)
		if err != nil {
			return errs.Wrap(err)
		}

		cols, err := merger.Finish()
		if err != nil {
			return errs.Wrap(err)
		}

		draw.Context{
			Canvas: canvas,
			Colors: viridis,
			Map:    dist.(interface{ CDF(float64) float64 }).CDF,
		}.Draw(cols)
	}

	fh, err := os.OpenFile("test.png", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fh.Close()

	return errs.Wrap(png.Encode(fh, RGBAsImage(canvas)))
}

func RGBAsImage(m *draw.RGB) *image.RGBA {
	return &image.RGBA{
		Pix:    m.Pix,
		Stride: m.Stride,
		Rect:   image.Rect(0, 0, m.Width, m.Height),
	}
}
