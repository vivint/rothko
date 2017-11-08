// Copyright (C) 2017. See AUTHORS.

package draw

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

func (c *Context) Draw(cols []Column) {
	// type assert the canvas to our optimized fast case
	can := c.Canvas
	m, _ := can.(*RGB)

	// TODO(jeff): this setup is really complicated lol. there's probably a
	// good idea to simplify this.

	width, height := can.Size()
	value_delta := c.Max - c.Min
	color_scale := float64(len(c.Colors)-1) / 1

	last_data_len := -1
	index_scale := 0.0

	// TODO(jeff): right now we do this a column at a time. would it be better
	// cache locality to do it a row at a time since that's the data layout
	// for the image? it's going to be hard to avoid some sort of cache
	// problems due to the data being columns and the image being rows.

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
				val := data[index]
				if val < c.Min {
					val = c.Min
				}
				if val > c.Max {
					val = c.Max
				}

				// TODO(jeff): figure this out

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
