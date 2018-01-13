// Copyright (C) 2017. See AUTHORS.

package draw

// Canvas is the type of things that can be drawn onto.
type Canvas interface {
	Set(x, y int, c Color)
	Size() (w, h int)
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
}

// NewRGB contstructs an RGB with space for the width and height.
func NewRGB(w, h int) *RGB {
	return &RGB{
		Pix:    make([]uint8, 4*w*h),
		Stride: 4 * w,
		Width:  w,
		Height: h,
	}
}

// Size returns the width and height of the RGB.
func (r *RGB) Size() (w, h int) {
	return r.Width, r.Height
}

// Set stores the pixel values in the color to the coordinate at x and y. The
// top left corner is (0, 0).
func (r *RGB) Set(x, y int, c Color) {
	i := y*r.Stride + x*4
	pix := r.Pix[i : i+4]
	pix[0] = c.R
	pix[1] = c.G
	pix[2] = c.B
	pix[3] = 255
}
