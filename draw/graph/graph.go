// Copyright (C) 2017. See AUTHORS.

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
	// 1. draw the heatmap (if we have data)
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

	if opts.NoAxes {
		return canvas, nil
	}

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

	left := axis.Draw(ctx, axis.Options{
		Face:     inconsolata.Regular8x16,
		Labels:   labels,
		Vertical: true,
		Length:   opts.Height + 1,
		Flip:     true,
	})

	// 3. the value axis on the right (if we have data)
	var right *draw.RGB

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
		right = axis.Draw(ctx, axis.Options{
			Face:     inconsolata.Regular8x16,
			Labels:   labels,
			Vertical: true,
			Length:   opts.Height + 1,
		})
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

	bottom := axis.Draw(ctx, axis.Options{
		Face:     inconsolata.Regular8x16,
		Labels:   labels,
		Vertical: false,
		Length:   opts.Width,
	})

	// 5. combine all the rendered stuff
	width := left.Width + bottom.Width
	width_canvas := left.Width + canvas.Width
	if right != nil {
		width_canvas += right.Width
	}
	if width_canvas > width {
		width = width_canvas
	}
	height := canvas.Height + bottom.Height

	out := draw.NewRGB(width, height)

	// TODO(jeff): this is insanity!

	// 5a. copy the left axis
	for x := 0; x < left.Width; x++ {
		for y := 0; y < left.Height; y++ {
			copy(out.Raw(x, y), left.Raw(x, y))
		}
	}

	// 5b. copy the canvas
	for x := 0; x < canvas.Width; x++ {
		for y := 0; y < canvas.Height; y++ {
			copy(out.Raw(x+left.Width, y), canvas.Raw(x, y))
		}
	}

	// 5c. copy the bottom axis
	for x := 0; x < bottom.Width; x++ {
		for y := 0; y < bottom.Height; y++ {
			copy(out.Raw(x+left.Width, y+canvas.Height), bottom.Raw(x, y))
		}
	}

	// 5d. copy the right if it exists
	if right != nil {
		offset := left.Width + canvas.Width
		for x := 0; x < right.Width; x++ {
			for y := 0; y < right.Height; y++ {
				copy(out.Raw(x+offset, y), right.Raw(x, y))
			}
		}
	}

	return out, nil
}
