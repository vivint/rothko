// Copyright (C) 2017. See AUTHORS.

package draw

import (
	"context"
	"image"
	"image/color"
	"math"
)

// Column represents a column to draw in a context. Data is expected to be
// sorted, non-empty, and contain typical floats (no NaNs/denormals/Inf/etc).
type Column struct {
	X, W int
	Data []float64
}

type Context struct {
	//
	// Required fields
	//

	// Colors is the slice of colors to map the column data on to. It ignores
	// the alpha component.
	Colors []color.RGBA

	// Height of the resulting image.
	Height int

	//
	// Optional fields
	//

	// Uses BlendModeLeft if not specified.
	BlendMode BlendMode

	// if either of these are non-zero, they are used as the minimum and
	// maximum values: values above and below these get mapped to the last and
	// first color.
	Min, Max float64
}

func (c *Context) Draw(ctx context.Context, cols []Column) *image.RGBA {
	// compute the minimum and maximum values
	min, max := c.Min, c.Max
	if min == 0 && max == 0 {
		for i, col := range cols {
			cand_min, cand_max := col.Data[0], col.Data[len(col.Data)-1]
			if i == 0 {
				min, max = cand_min, cand_max
				continue
			}
			if cand_min < min {
				min = cand_max
			}
			if cand_max > max {
				max = cand_max
			}
		}
	}

	// determine the dimensions of the image and create it
	width, height := 0, c.Height
	for i, col := range cols {
		if cand := col.W + col.X; i == 0 || cand > width {
			width = cand
		}
	}

	// TODO(jeff): pool around image buffers?
	m := image.NewRGBA(image.Rect(0, 0, width, height))

	// special case!! if min == max, we want to flood fill the image with the
	// lowest color.
	if min == max {
		color := c.Colors[0]
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				fastSet(m, x, y, color)
			}
		}
		return m
	}

	// start writing columns into the image
	color_scale := float64(len(c.Colors)-1) / 1

	for _, col := range cols {
		data := col.Data
		// keep track of the last index into the data and what color it had
		last_index := -1
		var last_color color.RGBA

		// we want 0 => data[0], and height - 1 => data[len(data)-1]
		// fitting a linear function means we just have a scale factor of
		// m = (len(data) - 1) / (height - 1)

		// TODO(jeff): we should only have to compute this if len(data) changes
		index_scale := float64(len(data)-1) / float64(height-1)

		for y := 0; y < height; y++ {
			// TODO(jeff): truncating might not work. we can also perhaps
			// invert this computation to give us the number of pixels before
			// we get to the next index directly.
			index := int(float64(y) * index_scale)

			// figure out the color if it's different from the last data index
			if index != last_index {
				switch scaled := (data[index] - min) / (max - min); {
				case scaled <= 0:
					last_color = c.Colors[0]
				case scaled >= 1:
					last_color = c.Colors[len(c.Colors)-1]
				case c.BlendMode == nil:
					last_color = c.Colors[fastFloor(scaled*color_scale)]
				default:
					scaled *= color_scale
					left := math.Floor(scaled)
					last_color = c.BlendMode(
						scaled-left,
						c.Colors[int(left)],
						c.Colors[int(left)+1])
				}

				last_index = index
			}

			x1 := col.X + col.W
			for x := col.X; x < x1; x++ {
				fastSet(m, x, y, last_color)
			}
		}
	}

	return m
}
