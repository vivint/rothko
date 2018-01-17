// Copyright (C) 2017. See AUTHORS.

package axis

import (
	"image"

	"github.com/spacemonkeygo/rothko/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

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
}

// Draw renders the axis described by the Options into a *draw.RGB. It computes
// the appropriate sizes.
func Draw(opts Options) *draw.RGB {
	if opts.Vertical {
		return drawVertical(opts)
	}
	return drawHorizontal(opts)
}

const (
	tickSize          = 10 // px
	axisWidth         = 1  // px
	tickPadding       = 2  // px
	horizLabelSpacing = 20 // px
	vertLabelSpacing  = 2  // px

	textOffset = axisWidth + tickSize + tickPadding // px
)

func drawVertical(opts Options) *draw.RGB {
	// TODO(jeff): i know the vertical checking here is off by a pixel or two,
	// but it produces results that are good enough for now.

	// determine the extra space we need to draw the labels
	extra_width := 0
	max_height := 0
	occupied := 0
	for _, label := range opts.Labels {
		bounds, _ := font.BoundString(opts.Face, label.Text)
		if width := bounds.Max.X.Ceil(); width > extra_width {
			extra_width = width
		}

		y := int(float64(opts.Length-1) * label.Position)

		if occupied > 0 && y < occupied+vertLabelSpacing {
			continue
		}

		occupied = y - bounds.Min.Y.Ceil() - bounds.Max.Y.Ceil()
		if occupied > max_height {
			max_height = occupied
		}
	}

	// compute the size and allocate the canvas
	width := textOffset + extra_width
	out := draw.NewRGB(width, max_height)

	// set up the drawer
	d := font.Drawer{
		Dst:  asImage(out),
		Src:  image.Black,
		Face: opts.Face,
	}

	maybeFlip := func(x int) int {
		if opts.Flip {
			return width - 1 - x
		}
		return x
	}

	// first draw the axis
	for y := 0; y < opts.Length; y++ {
		for x := 0; x < axisWidth; x++ {
			out.Set(maybeFlip(x), y, draw.Color{})
		}
	}

	// render the ticks
	occupied = 0
	for _, label := range opts.Labels {
		// TODO(jeff): don't need to recompute this probably.
		bounds, _ := font.BoundString(opts.Face, label.Text)

		y := int(float64(opts.Length-1) * label.Position)

		for x := 0; x < tickSize; x++ {
			out.Set(maybeFlip(axisWidth+x), y, draw.Color{})
		}

		if occupied > 0 && y < occupied+vertLabelSpacing {
			continue
		}

		text_size := bounds.Max.X - bounds.Min.X
		d.Dot = fixed.Point26_6{
			Y: fixed.I(y) - bounds.Min.Y - bounds.Max.Y,
		}
		if opts.Flip {
			d.Dot.X = fixed.I(width-textOffset) - text_size
		} else {
			d.Dot.X = fixed.I(textOffset)
		}

		occupied = d.Dot.Y.Ceil()
		d.DrawString(label.Text)
	}

	return out
}

func drawHorizontal(opts Options) *draw.RGB {
	// TODO(jeff): this assumes all the labels have the same height, or close
	// to it. we could maybe do better.

	// determine the max height of a label. this breaks it up into slots.
	max_height := 0
	for _, label := range opts.Labels {
		bounds, _ := font.BoundString(opts.Face, label.Text)
		label_height := (bounds.Max.Y - bounds.Min.Y).Ceil() + vertLabelSpacing
		if label_height > max_height {
			max_height = label_height
		}
	}

	// determine the largest slot we will occupy so we can allocate a canvas
	max_slot := 0
	max_width := 0
	occupied := make(map[int]int)
	for _, label := range opts.Labels {
		x := int(float64(opts.Length-1) * label.Position)

		slot := 0
		for len(occupied) > 0 &&
			x > horizLabelSpacing &&
			occupied[slot]+horizLabelSpacing > x {

			slot++
		}
		if slot > max_slot {
			max_slot = slot
		}

		advance := font.MeasureString(opts.Face, label.Text)
		width := (fixed.I(x) + advance).Ceil()
		if width > max_width {
			max_width = width
		}
		occupied[slot] = width
	}

	// compute the size of and allocate the canvas
	height := textOffset + (max_slot+1)*max_height
	out := draw.NewRGB(max_width, height)

	// set up the drawer
	d := font.Drawer{
		Dst:  asImage(out),
		Src:  image.Black,
		Face: opts.Face,
	}

	// draw the axis
	for x := 0; x < opts.Length; x++ {
		for y := 0; y < axisWidth; y++ {
			out.Set(x, y, draw.Color{})
		}
	}

	// render the ticks and labels
	occupied = make(map[int]int, len(occupied))
	for _, label := range opts.Labels {
		bounds, _ := font.BoundString(opts.Face, label.Text)
		x := int(float64(opts.Length-1) * label.Position)

		for y := 0; y < tickSize; y++ {
			out.Set(x, axisWidth+y, draw.Color{})
		}

		slot := 0
		for len(occupied) > 0 &&
			x > horizLabelSpacing &&
			occupied[slot]+horizLabelSpacing > x {

			slot++
		}

		offset := fixed.I(textOffset + slot*max_height)
		d.Dot = fixed.Point26_6{
			X: fixed.I(x),
			Y: offset - bounds.Min.Y,
		}

		d.DrawString(label.Text)
		occupied[slot] = d.Dot.X.Ceil()
	}

	return out
}
