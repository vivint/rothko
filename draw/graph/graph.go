// Copyright (C) 2018. See AUTHORS.

package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/spacemonkeygo/rothko/dist"
	"github.com/spacemonkeygo/rothko/draw"
	"github.com/spacemonkeygo/rothko/draw/axis"
	"github.com/spacemonkeygo/rothko/draw/heatmap"
	"github.com/spacemonkeygo/rothko/draw/iosevka"
	"github.com/zeebo/float16"
)

const labelGap = 30

type Measured struct {
	// Bottom measured axis.
	Bottom axis.Measured

	// Right measured axis. Only valid if Earliest was passed with the
	// MeasureOptions.
	Right axis.Measured

	// Left measured axis.
	Left axis.Measured

	// Width, Height of the heatmap
	Width, Height int

	// internal state
	opts MeasureOptions
}

type MeasureOptions struct {
	// Earliest is the distribution for the earliest (closest to Now) column.
	Earliest dist.Dist

	// What time the far right of the graph represents.
	Now int64

	// Duration is the amount of history the graph represents from now.
	Duration time.Duration

	// The width of the graph
	Width int

	// The height of the graph
	Height int

	// Padding around the graph
	Padding int
}

// Measure determines the sizes of the graph for the given parameters.
func Measure(ctx context.Context, opts MeasureOptions) Measured {
	var labels []axis.Label

	// measure the bottom axis. this will help us determine the height for the
	// left and right axis.
	labels = labels[:0]

	// determine the "natural" duration unit for 100 px. this is going to
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

	// create the labels from the truncated now
	x := time.Duration(opts.Now).Truncate(natural).Nanoseconds()
	stop_before := opts.Now - opts.Duration.Nanoseconds()
	for x > stop_before {
		labels = append(labels, axis.Label{
			Position: 1 - float64(opts.Now-x)/float64(opts.Duration),
			Text:     time.Unix(0, x).Format("1/2 @ 15:04"),
		})
		x -= natural.Nanoseconds()
	}

	// reverse the labels so that they are in increasing order
	for i := 0; i < len(labels)/2; i++ {
		si := len(labels) - 1 - i
		labels[i], labels[si] = labels[si], labels[i]
	}

	bottom := axis.Measure(ctx, axis.Options{
		Face:      iosevka.Iosevka,
		Labels:    labels,
		Vertical:  false,
		Length:    opts.Width,
		DontBleed: true,
	})

	// calculate the heatmap height and leave a gap for the bottom axis
	height := opts.Height - bottom.Height - labelGap

	// measure the left axis
	labels = labels[:0]
	for y := 0; y <= height-labelGap; y += labelGap {
		pos := float64(y) / float64(height)
		labels = append(labels, axis.Label{
			Position: pos,
			Text:     fmt.Sprintf("%0.2f", 1-pos),
		})
	}
	labels = append(labels, axis.Label{
		Position: 1,
		Text:     "0.00",
	})

	left := axis.Measure(ctx, axis.Options{
		Face:     iosevka.Iosevka,
		Labels:   labels,
		Vertical: true,
		Length:   height,
		Flip:     true,
	})

	// measure the right axis
	var right axis.Measured

	if opts.Earliest != nil {
		labels = labels[:0]
		for y := 0; y <= height-labelGap; y += labelGap {
			pos := float64(y) / float64(height)
			val := opts.Earliest.Query(1 - pos)
			if val16, ok := float16.FromFloat64(val); ok {
				val = val16.Float64()
			}
			labels = append(labels, axis.Label{
				Position: pos,
				Text:     fmt.Sprintf("%#.3g", val),
			})
		}

		val := opts.Earliest.Query(0)
		if val16, ok := float16.FromFloat64(val); ok {
			val = val16.Float64()
		}
		labels = append(labels, axis.Label{
			Position: 1,
			Text:     fmt.Sprintf("%#.3g", val),
		})

		right = axis.Measure(ctx, axis.Options{
			Face:     iosevka.Iosevka,
			Labels:   labels,
			Vertical: true,
			Length:   height,
		})
	}

	return Measured{
		Bottom: bottom,
		Right:  right,
		Left:   left,
		Width:  opts.Width - left.Width - right.Width,
		Height: height,

		opts: opts,
	}
}

// DrawOptions are all the ways you can configure the graph.
type DrawOptions struct {
	// Canvas is where the drawing happens. It is expected to be large enough
	// to handle the drawing. If the canvas is nil or too small, one is
	// allocated.
	Canvas *draw.RGB

	// Columns is the set of columns to draw on the graph.
	Columns []draw.Column

	// Colors used for the heatmap.
	Colors []draw.Color
}

func (m Measured) Draw(ctx context.Context, opts DrawOptions) *draw.RGB {
	cw, ch := m.opts.Width+2*m.opts.Padding, m.opts.Height+2*m.opts.Padding
	w, h := 0, 0
	if opts.Canvas != nil {
		w, h = opts.Canvas.Size()
	}
	if w < cw || h < ch {
		opts.Canvas = draw.NewRGB(cw, ch)
	}

	m.Left.Draw(ctx, opts.Canvas.View(
		m.opts.Padding, m.opts.Padding,
		m.Left.Width, m.Left.Height))

	if m.opts.Earliest != nil {
		hm := heatmap.New(heatmap.Options{
			Canvas: opts.Canvas.View(
				m.Left.Width+m.opts.Padding, m.opts.Padding,
				m.Width, m.Height),
			Colors: opts.Colors,
			Map:    m.opts.Earliest.CDF,
		})
		for _, col := range opts.Columns {
			hm.Draw(ctx, col)
		}

		m.Right.Draw(ctx, opts.Canvas.View(
			m.Left.Width+m.Width+m.opts.Padding, m.opts.Padding,
			m.Right.Width, m.Right.Height))
	}

	m.Bottom.Draw(ctx, opts.Canvas.View(
		m.opts.Padding, m.Height+labelGap+m.opts.Padding,
		m.Bottom.Width, m.Bottom.Height))

	return opts.Canvas
}
