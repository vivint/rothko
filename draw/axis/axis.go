// Copyright (C) 2018. See AUTHORS.

package axis

import (
	"context"
	"image"

	"github.com/spacemonkeygo/rothko/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	tickSize          = 10 // px
	axisWidth         = 1  // px
	tickPadding       = 2  // px
	horizLabelSpacing = 10 // px
	vertLabelSpacing  = 2  // px

	textOffset = axisWidth + tickSize + tickPadding // px
)

type Measured struct {
	// Width is the width in pixels of the drawn axis.
	Width int

	// Height is the height in pixels of the drawn axis
	Height int

	// internal fields
	opts      Options
	bounds    []fixed.Rectangle26_6 // same index the labels
	maxHeight int                   // maximum hight of a label
}

// copyLabels makes a copy of the labels to avoid mutation issues.
func copyLabels(labels []Label) (out []Label) {
	return append(out, labels...)
}

// Label represents a tick mark on the axis.
type Label struct {
	// Position is the position of the tick mark as a float in [0, 1].
	Position float64

	// Text is the text of the tick mark.
	Text string
}

// Options describe the axis rendering options.
type Options struct {
	// Face is the font face to use for rendering the label text.
	Face font.Face

	// Labels is the set of labels to draw.
	Labels []Label

	// Vertical is if the axis is vertical.
	Vertical bool

	// Length is how long the axis is.
	Length int

	// If true, vertical axes will be drawn for the left size. Horizontal axes
	// ignore this field.
	Flip bool

	// If true, the label text will not go past the boundaries of Length.
	DontBleed bool
}

// copy returns a deep copy of the Options.
func (o Options) copy() Options {
	// TODO(jeff): font.Face could technically be mutated, but don't worry
	// about it.
	o.Labels = copyLabels(o.Labels)
	return o
}

// Draw renders the axis and returns a canvas allocated for the appopriate
// size. See Measure if you want to control where and how it is drawn.
func Draw(ctx context.Context, opts Options) *draw.RGB {
	return Measure(ctx, opts).Draw(ctx, nil)
}

// Measure measures the axis sizes, and returns some state that can be used
// to draw on to some canvas.
func Measure(ctx context.Context, opts Options) Measured {
	if opts.Vertical {
		return measureVertical(ctx, opts)
	}
	return measureHorizontal(ctx, opts)
}

// Draw performs the drawing of the data on to the canvas. The canvas is
// expected to be large enough to handle the drawing. If the canvas is nil,
// one is allocated. In either case, the canvas is returned.
func (m Measured) Draw(ctx context.Context, canvas *draw.RGB) *draw.RGB {
	if m.opts.Vertical {
		return m.drawVertical(ctx, canvas)
	}
	return m.drawHorizontal(ctx, canvas)
}

func measureVertical(ctx context.Context, opts Options) Measured {
	// TODO(jeff): i know the vertical checking here is off by a pixel or two,
	// but it produces results that are good enough for now.

	// determine the extra space we need to draw the labels
	max_width := 0
	max_height := opts.Length
	occupied := 0
	bounds := make([]fixed.Rectangle26_6, 0, len(opts.Labels))

	for _, label := range opts.Labels {
		b, _ := font.BoundString(opts.Face, label.Text)
		bounds = append(bounds, b)

		y := int(float64(opts.Length-1) * label.Position)
		height := y - b.Min.Y.Ceil()

		if opts.DontBleed && height > opts.Length {
			continue
		}

		if height > max_height {
			max_height = height
		}

		label_width := b.Max.X.Ceil()
		if label_width > max_width {
			max_width = label_width
		}

		if occupied > 0 && y < occupied+vertLabelSpacing {
			continue
		}
		occupied = height
	}

	return Measured{
		Width:  textOffset + max_width,
		Height: max_height,

		opts:   opts.copy(),
		bounds: bounds,
	}
}

func (m Measured) drawVertical(ctx context.Context, canvas *draw.RGB) (
	out *draw.RGB) {

	w, h := 0, 0
	if canvas != nil {
		w, h = canvas.Size()
	}
	if w < m.Width || h < m.Height {
		canvas = draw.NewRGB(m.Width, m.Height)
	}

	// set up the drawer
	d := font.Drawer{
		Dst:  asImage(canvas),
		Src:  image.Black,
		Face: m.opts.Face,
	}

	maybeFlip := func(x int) int {
		if m.opts.Flip {
			return m.Width - 1 - x
		}
		return x
	}

	// first draw the axis
	for y := 0; y < m.opts.Length; y++ {
		for x := 0; x < axisWidth; x++ {
			canvas.Set(maybeFlip(x), y, draw.Color{})
		}
	}

	// render the ticks
	occupied := 0
	for i, label := range m.opts.Labels {
		b := m.bounds[i]
		y := int(float64(m.opts.Length-1) * label.Position)
		height := y - b.Min.Y.Ceil()

		for x := 0; x < tickSize; x++ {
			canvas.Set(maybeFlip(axisWidth+x), y, draw.Color{})
		}

		if m.opts.DontBleed && height > m.opts.Length {
			continue
		}

		if occupied > 0 && y < occupied+vertLabelSpacing {
			continue
		}

		text_size := b.Max.X - b.Min.X
		d.Dot = fixed.Point26_6{
			Y: fixed.I(y - b.Min.Y.Ceil()),
		}
		if m.opts.Flip {
			d.Dot.X = fixed.I(m.Width - textOffset - text_size.Ceil())
		} else {
			d.Dot.X = fixed.I(textOffset)
		}

		occupied = d.Dot.Y.Ceil()
		d.DrawString(label.Text)
	}

	return canvas
}

func measureHorizontal(ctx context.Context, opts Options) Measured {
	max_height := 0
	max_width := opts.Length
	bounds := make([]fixed.Rectangle26_6, 0, len(opts.Labels))
	occupied := 0
	fits := func(x int) bool {
		return occupied == 0 ||
			x < horizLabelSpacing ||
			x > occupied+horizLabelSpacing
	}

	for _, label := range opts.Labels {
		b, _ := font.BoundString(opts.Face, label.Text)
		bounds = append(bounds, b)

		x := int(float64(opts.Length-1) * label.Position)
		label_end := x + (b.Max.X - b.Min.X).Ceil()

		if opts.DontBleed && label_end > opts.Length {
			// as a special case, if we can nudge the x back so that it's just
			// on opts.Length and fits, draw it
			label_end = opts.Length
			x = label_end - (b.Max.X - b.Min.X).Ceil()
		}
		if !fits(x) {
			continue
		}

		if label_end > max_width {
			max_width = label_end
		}

		label_height := (b.Max.Y - b.Min.Y).Ceil() + vertLabelSpacing
		if label_height > max_height {
			max_height = label_height
		}

		occupied = label_end
	}

	return Measured{
		Width:  max_width,
		Height: textOffset + max_height,

		opts:      opts.copy(),
		bounds:    bounds,
		maxHeight: max_height,
	}
}

func (m Measured) drawHorizontal(ctx context.Context, canvas *draw.RGB) (
	out *draw.RGB) {

	w, h := 0, 0
	if canvas != nil {
		w, h = canvas.Size()
	}
	if w < m.Width || h < m.Height {
		canvas = draw.NewRGB(m.Width, m.Height)
	}

	// set up the drawer
	d := font.Drawer{
		Dst:  asImage(canvas),
		Src:  image.Black,
		Face: m.opts.Face,
	}

	// draw the axis
	for x := 0; x < m.opts.Length; x++ {
		for y := 0; y < axisWidth; y++ {
			canvas.Set(x, y, draw.Color{})
		}
	}

	// render the ticks and labels
	occupied := 0
	fits := func(x int) bool {
		return occupied == 0 ||
			x < horizLabelSpacing ||
			x > occupied+horizLabelSpacing
	}

	for i, label := range m.opts.Labels {
		b := m.bounds[i]

		x := int(float64(m.opts.Length-1) * label.Position)
		label_end := x + (b.Max.X - b.Min.X).Ceil()

		for y := 0; y < tickSize; y++ {
			canvas.Set(x, axisWidth+y, draw.Color{})
		}
		tick_x := x

		if m.opts.DontBleed && label_end > m.opts.Length {
			// as a special case, if we can nudge the x back so that it's just
			// on ,.opts.Length and still have the same slot, draw it.
			label_end = m.opts.Length
			x = label_end - (b.Max.X - b.Min.X).Ceil()
		}
		if !fits(x) {
			continue
		}

		for y := 0; y < 3; y++ {
			canvas.Set(tick_x, axisWidth+tickSize+y, draw.Color{})
		}

		d.Dot = fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(textOffset - b.Min.Y.Ceil()),
		}
		d.DrawString(label.Text)

		occupied = d.Dot.X.Ceil()
	}

	return canvas
}
