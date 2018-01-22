// Copyright (C) 2018. See AUTHORS.

package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/spacemonkeygo/rothko/data/dists"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/spacemonkeygo/rothko/draw/axis"
	"github.com/spacemonkeygo/rothko/draw/heatmap"
	"golang.org/x/image/font/inconsolata"
)

// Options are all the ways you can configure the graph.
type Options struct {
	// What time the far right of the graph represents.
	Now int64

	// Duration is the amount of history the graph represents from now.
	Duration time.Duration

	// Columns is the set of columns to draw on the graph.
	Columns []draw.Column

	// Colors used for the heatmap.
	Colors []draw.Color

	// Earliest is the distribution for the earliest (closest to Now) column.
	Earliest dists.Dist

	// Width and Height for the heatmap. The axes are extra.
	Width, Height int

	// NoAxes will turn of rendering the axes.
	NoAxes bool
}

// Draw renders a graph with the given options.
func Draw(ctx context.Context, opts Options) (*draw.RGB, error) {
	if opts.NoAxes {
		canvas := draw.NewRGB(opts.Width, opts.Height)
		if opts.Earliest != nil {
			hm := heatmap.New(heatmap.Options{
				Canvas: canvas,
				Colors: opts.Colors,
				Map:    opts.Earliest.CDF,
			})

			for _, col := range opts.Columns {
				hm.Draw(ctx, col)
			}
		}
		return canvas, nil
	}

	copyLabels := func(labels []axis.Label) (out []axis.Label) {
		return append(out, labels...)
	}

	// 1. get the sizes

	// create the axes:
	var labels []axis.Label

	// 2. the percentile axis on the left
	labels = labels[:0]
	for y := 0; y <= opts.Height-30; y += 30 {
		pos := float64(y) / float64(opts.Height)
		labels = append(labels, axis.Label{
			Position: pos,
			Text:     fmt.Sprintf("%0.2f", pos),
		})
	}
	labels = append(labels, axis.Label{
		Position: 1,
		Text:     "1.00",
	})

	left_opts := axis.Options{
		Face:     inconsolata.Regular8x16,
		Labels:   copyLabels(labels),
		Vertical: true,
		Length:   opts.Height + 1,
		Flip:     true,
	}
	left_w, left_h := axis.Draw(ctx, left_opts)

	// 3. the value axis on the right (if we have data)
	var right_opts axis.Options
	right_w, right_h := 0, 0

	if opts.Earliest != nil {
		labels = labels[:0]
		for y := 0; y <= opts.Height-30; y += 30 {
			pos := float64(y) / float64(opts.Height)
			labels = append(labels, axis.Label{
				Position: pos,
				Text:     fmt.Sprintf("%0.4f", opts.Earliest.Query(pos)),
			})
		}
		labels = append(labels, axis.Label{
			Position: 1,
			Text:     fmt.Sprintf("%0.4f", opts.Earliest.Query(1)),
		})
		right_opts = axis.Options{
			Face:     inconsolata.Regular8x16,
			Labels:   copyLabels(labels),
			Vertical: true,
			Length:   opts.Height + 1,
		}
		right_w, right_h = axis.Draw(ctx, right_opts)
	}

	// 4. draw the time axis on the bottom
	labels = labels[:0]

	// 4a. determine the "natural" duration unit for 100 px. this is going to
	// be the largest "natural" unit smaller than the chunk.
	chunk := opts.Duration / time.Duration(opts.Width/100)
	var natural time.Duration
	for _, unit := range naturalUnits {
		if unit < chunk {
			natural = unit
		} else {
			break
		}
	}

	// 4b. create the labels from the truncated now
	x := time.Duration(opts.Now).Truncate(natural).Nanoseconds()
	stop_before := opts.Now - opts.Duration.Nanoseconds()
	for x > stop_before {
		labels = append(labels, axis.Label{
			Position: 1 - float64(opts.Now-x)/float64(opts.Duration),
			Text:     time.Unix(0, x).Format("1/2 @ 15:04"),
		})
		x -= natural.Nanoseconds()
	}

	// 4c. reverse the labels so that they are in increasing order
	for i := 0; i < len(labels)/2; i++ {
		si := len(labels) - 1 - i
		labels[i], labels[si] = labels[si], labels[i]
	}

	bottom_opts := axis.Options{
		Face:     inconsolata.Regular8x16,
		Labels:   copyLabels(labels),
		Vertical: false,
		Length:   opts.Width,
	}
	bottom_w, bottom_h := axis.Draw(ctx, bottom_opts)

	// 5. combine all the rendered stuff sizes
	width := left_w + bottom_w + right_w
	width_canvas := left_w + opts.Width + right_w
	if width_canvas > width {
		width = width_canvas
	}
	height := opts.Height + bottom_h

	// 6. actually draw
	out := draw.NewRGB(width, height)

	// 6a. the left axis
	left_opts.Canvas = out.View(0, 0, left_w, left_h)
	axis.Draw(ctx, left_opts)

	// 6b. the heatmap
	if opts.Earliest != nil {
		hm := heatmap.New(heatmap.Options{
			Canvas: out.View(left_w, 0, opts.Width, opts.Height),
			Colors: opts.Colors,
			Map:    opts.Earliest.CDF,
		})
		for _, col := range opts.Columns {
			hm.Draw(ctx, col)
		}
	}

	// 6c. the bottom axis
	bottom_opts.Canvas = out.View(left_w, opts.Height, bottom_w, bottom_h)
	axis.Draw(ctx, bottom_opts)

	// 6d. the right axis
	if opts.Earliest != nil {
		right_opts.Canvas = out.View(left_w+opts.Width, 0, right_w, right_h)
		axis.Draw(ctx, right_opts)
	}

	return out, nil
}
