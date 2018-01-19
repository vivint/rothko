// Copyright (C) 2017. See AUTHORS.

package heatmap

import (
	"context"
	"runtime"

	"github.com/spacemonkeygo/rothko/draw"
)

// Options are the things you can specify to control the rendering of a
// heatmap.
type Options struct {
	// Colors is the slice of colors to map the column data on to.
	Colors []draw.Color

	// Canvas to draw on to
	Canvas draw.Canvas

	// Map takes a value from the Data in the column, and expects it to be
	// mapped to a value in [0,1] specifying the color.
	Map func(float64) float64
}

// Heatmap is a struct that draws heatmaps from provided columns.
type Heatmap struct {
	opts Options

	m             *draw.RGB // possibly type asserted
	color_scale   float64
	width, height int
}

// New returns a new Heatmap using the given options.
func New(opts Options) *Heatmap {
	m, _ := opts.Canvas.(*draw.RGB)
	width, height := opts.Canvas.Size()

	return &Heatmap{
		opts: opts,

		m:           m,
		color_scale: float64(len(opts.Colors) - 1),
		width:       width,
		height:      height,
	}
}

// Draw writes the column to the canvas.
func (d *Heatmap) Draw(ctx context.Context, col draw.Column) {
	last_index := -1
	last_color := d.opts.Colors[0]
	index_scale := float64(len(col.Data)-1) / float64(d.height-1)
	m := d.m

	for y := 0; y < d.height; y++ {
		// TODO(jeff): truncating might not work. we can also perhaps
		// invert this computation to give us the number of pixels
		// before we get to the next index directly.
		index := int(float64(y) * index_scale)

		// figure out the color if it's different from the last data index
		if index != last_index {
			color_index := int(d.opts.Map(col.Data[index]) * d.color_scale)
			last_color = d.opts.Colors[color_index]
			last_index = index
		}

		if m != nil {
			row := y * m.Stride
			low := row + 4*col.X
			high := low + 4*col.W
			if high > len(m.Pix) {
				high = len(m.Pix)
			}

			// reslicing in gopherjs always allocates a slice structure. so
			// since it doesn't elide bounds checks, we write it this way even
			// though it'd be slower if compiled by gc.
			if runtime.GOARCH == "js" {
				for offset := low; offset < high; offset += 4 {
					m.Pix[offset+0] = last_color.R
					m.Pix[offset+1] = last_color.G
					m.Pix[offset+2] = last_color.B
					m.Pix[offset+3] = 255
				}
			} else {
				pix := m.Pix[low:high]
				for len(pix) >= 4 {
					pix[0] = last_color.R
					pix[1] = last_color.G
					pix[2] = last_color.B
					pix[3] = 255
					pix = pix[4:]
				}
			}

		} else {
			x1 := col.X + col.W
			for x := col.X; x < x1 && x < d.width; x++ {
				d.opts.Canvas.Set(x, y, last_color)
			}
		}

	}
}
