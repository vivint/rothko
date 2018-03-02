// Copyright (C) 2018. See AUTHORS.

package draw

import (
	"bytes"
	"image"
)

// Canvas is the type of things that can be drawn onto.
type Canvas interface {
	Set(x, y int, c Color)
	Size() (w, h int)
}

// Column represents a column to draw in a context. Data is expected to be
// sorted, non-empty, and contain typical floats (no NaNs/denormals/Inf/etc).
// Obs is the number of observations.
type Column struct {
	X, W int
	Data []float64
	Obs  int64
}

// Color is a simple 8 bits per channel color.
type Color struct {
	R, G, B uint8
}

// RGB is a byte compatabile implementation of image.RGBA, except with much
// less supporting code, and no alpha channel.
type RGB struct {
	Pix    []uint8
	Stride int
	Width  int
	Height int
	X, Y   int
}

// NewRGB contstructs an RGB with space for the width and height.
func NewRGB(w, h int) *RGB {
	return &RGB{
		Pix:    bytes.Repeat([]byte{255}, 4*w*h),
		Stride: 4 * w,
		Width:  w,
		Height: h,
	}
}

// Size returns the width and height of the RGB.
func (m *RGB) Size() (w, h int) {
	return m.Width, m.Height
}

// Set stores the pixel values in the color to the coordinate at x and y. The
// top left corner is (0, 0).
func (m *RGB) Set(x, y int, c Color) {
	i := (y+m.Y)*m.Stride + (x+m.X)*4
	pix := m.Pix[i : i+4]
	pix[0] = c.R
	pix[1] = c.G
	pix[2] = c.B
	pix[3] = 255
}

// Raw returns the raw values at the pixel, including alpha channel. It can
// be mutated.
func (m *RGB) Raw(x, y int) []uint8 {
	i := (y+m.Y)*m.Stride + (x+m.X)*4
	return m.Pix[i : i+4]
}

// View returns a view into the RGB.
func (m RGB) View(x, y, w, h int) *RGB {
	m.X = x
	m.Y = y
	m.Width = w
	m.Height = h
	return &m
}

func (m *RGB) AsImage() *image.RGBA {
	return &image.RGBA{
		Pix:    m.Pix,
		Stride: m.Stride,
		// TODO(jeff): i highly suspect m.Width and m.Height is wrong here.
		// add a test around boundary conditions.
		Rect: image.Rect(-m.X, -m.Y, m.Width, m.Height),
	}
}
