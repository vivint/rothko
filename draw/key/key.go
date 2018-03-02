// Copyright (C) 2018. See AUTHORS.

package key

import (
	"github.com/vivint/rothko/draw"
)

// Options are the things you can specify to control the rendering of a key.
type Options struct {
	// Colors is the slice of colors to map the column data on to.
	Colors []draw.Color

	// Height is how tall the key will be.
	Height int

	// Width is how wide the key will be.
	Width int
}

// Draw draws a key using values from Options on to the provided canvas,
// allocating an output canvas if the input is not large enough.
func Draw(canvas *draw.RGB, opts Options) (out *draw.RGB) {
	w, h := 0, 0
	if canvas != nil {
		w, h = canvas.Size()
	}
	if w < opts.Width || h < opts.Height {
		canvas = draw.NewRGB(opts.Width, opts.Height)
	}

	scale := float64(len(opts.Colors)-1) / float64(opts.Height-1)
	for y := 0; y < opts.Height; y++ {
		index := int(float64(opts.Height-y) * scale)
		color := opts.Colors[index]
		for x := 0; x < opts.Width; x++ {
			canvas.Set(x, y, color)
		}
	}

	return canvas
}
