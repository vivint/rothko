// Copyright (C) 2018. See AUTHORS.

package observations

import (
	"context"
	"fmt"

	"github.com/vivint/rothko/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	padding = 5
)

type Measured struct {
	// Width is the width in pixels of the observation axis
	Width int

	// Height is the height in pixels of the observation axis
	Height int

	// internal fields
	opts       Options
	label_text string
	bounds     fixed.Rectangle26_6
	max        int64
}

// Options describe the axis rendering options.
type Options struct {
	// Face is the font face to use for rendering the max observations number.
	Face font.Face

	// Width is how long the axis is.
	Width int

	// Height is the height of the bar
	Height int

	// Columns describes the columns values
	Columns []Column
}

// Column represents a column on the heatmap with a number of observations.
type Column struct {
	X, W int
	Obs  int64
}

// Draw renders the axis and returns a canvas allocated for the appopriate
// size. See Measure if you want to control where and how it is drawn.
func Draw(ctx context.Context, opts Options) *draw.RGB {
	return Measure(ctx, opts).Draw(ctx, nil)
}

// Measure measures the axis sizes, and returns some state that can be used
// to draw on to some canvas.
func Measure(ctx context.Context, opts Options) Measured {
	max := int64(0)
	for _, col := range opts.Columns {
		if col.Obs > max {
			max = col.Obs
		}
	}

	label_text := fmt.Sprintf("max: %#.3g", float64(max))
	bounds, _ := font.BoundString(opts.Face, label_text)

	return Measured{
		Width:  opts.Width,
		Height: opts.Height + padding + (b.Max.Y - b.Min.Y).Ceil(),

		opts:       opts,
		label_text: label_text,
		bounds:     bounds,
		max:        max,
	}
}

// Draw performs the drawing of the data on to the canvas. The canvas is
// expected to be large enough to handle the drawing. If the canvas is nil,
// one is allocated. In either case, the canvas is returned.
func (m Measured) Draw(ctx context.Context, canvas *draw.RGB) *draw.RGB {
	// TODO(jeff): implement this
	return nil
}
