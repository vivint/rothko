// Copyright (C) 2018. See AUTHORS.

package key

import (
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/vivint/rothko/draw"
	"github.com/vivint/rothko/draw/colors"
	"github.com/vivint/rothko/internal/assert"
)

func saveImage(t *testing.T, name string, out *draw.RGB) {
	if false { // set to false to save images
		return
	}

	fh, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer fh.Close()

	assert.NoError(t, png.Encode(fh, asImage(out)))
}

func TestDraw(t *testing.T) {
	saveImage(t, "key.png", Draw(nil, Options{
		Colors: colors.Viridis,
		Width:  10,
		Height: 300,
	}))
}

func asImage(m *draw.RGB) *image.RGBA {
	return &image.RGBA{
		Pix:    m.Pix,
		Stride: m.Stride,
		// TODO(jeff): i highly suspect m.Width and m.Height is wrong here.
		// add a test around boundary conditions.
		Rect: image.Rect(-m.X, -m.Y, m.Width, m.Height),
	}
}
