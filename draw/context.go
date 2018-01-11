// Copyright (C) 2017. See AUTHORS.

package draw

import "math"

// Column represents a column to draw in a context. Data is expected to be
// sorted, non-empty, and contain typical floats (no NaNs/denormals/Inf/etc).
type Column struct {
	X, W int
	Data []float64
}

type Context struct {
	// Colors is the slice of colors to map the column data on to.
	Colors []Color

	// Canvas to draw on to
	Canvas Canvas

	// Values below and above these get mapped to the first and last color.
	// All other values get mapped to a color based on their percentage
	// difference between based on the scaling.
	Min, Max float64

	// If true, will use logarithms to map between the value scales, where
	// it finds the p where min + (max - min)^p is equal to the value.
	Logrithmic bool
}

func (c Context) Draw(cols []Column) {
	// type assert the canvas to our optimized fast case
	can := c.Canvas
	m, _ := can.(*RGB)

	width, height := can.Size()
	value_delta := c.Max - c.Min
	color_scale := float64(len(c.Colors)-1) / 1

	linear_scale := 0.0
	log_scale := 0.0
	if value_delta > 0 {
		linear_scale = color_scale / value_delta
		log_scale = (math.E - 1) / value_delta
	}

	last_data_len := -1
	index_scale := 0.0

	for _, col := range cols {
		data := col.Data

		// keep track of the last index into the data and what color it had
		last_index := -1
		last_color := c.Colors[0]

		// we want 0 => data[0], and height - 1 => data[len(data)-1]
		// fitting a linear function means we just have a scale factor of
		// m = (len(data) - 1) / (height - 1)
		if len(data) != last_data_len {
			index_scale = float64(len(data)-1) / float64(height-1)
			last_data_len = len(data)
		}

		for y := 0; y < height; y++ {
			// TODO(jeff): truncating might not work. we can also perhaps
			// invert this computation to give us the number of pixels
			// before we get to the next index directly.
			index := int(float64(y) * index_scale)

			// figure out the color if it's different from the last data index
			if index != last_index && value_delta > 0 {
				val := data[index] - c.Min
				if val < 0 {
					val = 0
				}
				if val > value_delta {
					val = value_delta
				}

				var scaled int
				if c.Logrithmic {
					val = (val * log_scale) + 1
					scaled = int(math.Log(val) * color_scale)
				} else {
					scaled = int(val * linear_scale)
				}

				last_color = c.Colors[scaled]
				last_index = index
			}

			if m != nil {
				row := y * m.Stride
				low := row + 4*col.X
				high := low + 4*col.W
				if high > len(m.Pix) {
					high = len(m.Pix)
				}
				pix := m.Pix[low:high]

				for len(pix) >= 4 {
					pix[0] = last_color.R
					pix[1] = last_color.G
					pix[2] = last_color.B
					pix[3] = 255
					pix = pix[4:]
				}

			} else {
				x1 := col.X + col.W
				for x := col.X; x < x1 && x < width; x++ {
					can.Set(x, y, last_color)
				}
			}
		}
	}
}
