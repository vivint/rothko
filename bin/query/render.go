// Copyright (C) 2017. See AUTHORS.

package main

import (
	"context"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"time"

	"github.com/spacemonkeygo/rothko/data/dists/tdigest"
	"github.com/spacemonkeygo/rothko/disk/files"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/spacemonkeygo/rothko/draw/graph"
	"github.com/zeebo/errs"
)

func runRender(ctx context.Context, di *files.DB, metric string,
	dur time.Duration) error {

	// TODO(jeff): parameterize these constants
	const (
		width       = 1000
		height      = 300
		samples     = height
		compression = 5
	)

	g := graph.New(graph.Options{
		Duration: dur,
		Samples:  samples,
		Params:   tdigest.Params{Compression: compression},
		Colors:   viridis,
		Width:    width,
		Height:   height,
	})

	err := di.Query(ctx, metric, g.Now(), nil, g.Push)
	if err != nil {
		return errs.Wrap(err)
	}

	out, err := g.Finish(ctx)
	if err != nil {
		return errs.Wrap(err)
	}

	fh, err := os.OpenFile("test.png", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fh.Close()

	return errs.Wrap(png.Encode(ioutil.Discard, asImage(out)))
}

func asImage(m *draw.RGB) *image.RGBA {
	return &image.RGBA{
		Pix:    m.Pix,
		Stride: m.Stride,
		Rect:   image.Rect(0, 0, m.Width, m.Height),
	}
}
