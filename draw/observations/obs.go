// Copyright (C) 2018. See AUTHORS.

package observations

import (
	"context"
	"fmt"
	"image"

	"github.com/vivint/rothko/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	padding = 2
)

// Measured represents a measured observations axis.
type Measured struct {
	// Width is the width in pixels of the observation axis
	Width int

	// Height is the height in pixels of the observation axis
	Height int

	// internal fields
	opts   Options
	bounds fixed.Rectangle26_6
}

// Options describe the axis rendering options.
type Options struct {
	// Face is the font face to use for rendering the max observations number.
	Face font.Face

	// Width is how long the axis is.
	Width int

	// Height is the height of the bar
	Height int
}

// Draw renders the axis and returns a canvas allocated for the appopriate
// size. See Measure if you want to control where and how it is drawn.
func Draw(ctx context.Context, cols []draw.Column, opts Options) *draw.RGB {
	return Measure(ctx, opts).Draw(ctx, cols, nil)
}

// Measure measures the axis sizes, and returns some state that can be used
// to draw on to some canvas.
func Measure(ctx context.Context, opts Options) Measured {
	bounds, _ := font.BoundString(opts.Face, "max: 0.00e-00")
	label_height := (bounds.Max.Y - bounds.Min.Y).Ceil()

	return Measured{
		Width:  opts.Width,
		Height: opts.Height + padding + label_height,

		opts:   opts,
		bounds: bounds,
	}
}

// Draw performs the drawing of the data on to the canvas. The canvas is
// expected to be large enough to handle the drawing. If the canvas is nil,
// one is allocated. In either case, the canvas is returned.
func (m Measured) Draw(ctx context.Context, cols []draw.Column,
	canvas *draw.RGB) *draw.RGB {

	w, h := 0, 0
	if canvas != nil {
		w, h = canvas.Size()
	}
	if w < m.Width || h < m.Height {
		canvas = draw.NewRGB(m.Width, m.Height)
	}

	max := int64(-1)
	x := 0
	for _, col := range cols {
		if col.Obs > max {
			max = col.Obs
			x = col.X
		}
	}

	label_text := fmt.Sprintf("max: %#.3g", float64(max))
	label_height := (m.bounds.Max.Y - m.bounds.Min.Y).Ceil()
	label_width := (m.bounds.Max.X - m.bounds.Min.X).Ceil()

	for _, col := range cols {
		sat := 255 - byte(float64(col.Obs)/float64(max)*255)
		c := draw.Color{sat, sat, sat}
		for y := 0; y < m.opts.Height; y++ {
			for x := 0; x < col.W; x++ {
				canvas.Set(x+col.X, y+padding+label_height, c)
			}
		}
	}

	end := x + label_width
	if end > m.opts.Width {
		x = m.opts.Width - label_width
	}

	(&font.Drawer{
		Dst:  canvas.AsImage(),
		Src:  image.Black,
		Face: m.opts.Face,
		Dot: fixed.Point26_6{
			Y: -m.bounds.Min.Y,
			X: fixed.I(x),
		},
	}).DrawString(label_text)

	return canvas
}
