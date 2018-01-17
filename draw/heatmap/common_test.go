// Copyright (C) 2017. See AUTHORS.

package heatmap

import (
	"math"

	"github.com/spacemonkeygo/rothko/draw"
)

var grayscale []draw.Color

func init() {
	for i := 0; i < 256; i++ {
		grayscale = append(grayscale, draw.Color{
			R: uint8(i), G: uint8(i), B: uint8(i),
		})
	}
}

func testMakeColumns(cols, height, col_width int, cb func(x, y int) float64) (
	out []draw.Column, linear func(float64) float64) {

	min, max := math.NaN(), math.NaN()
	for i := 0; i < cols; i++ {
		var data []float64
		for j := 0; j < height; j++ {
			val := cb(i, j)
			if math.IsNaN(min) || val < min {
				min = val
			}
			if math.IsNaN(max) || val > max {
				max = val
			}
			data = append(data, val)
		}
		out = append(out, draw.Column{
			X:    i * col_width,
			W:    col_width,
			Data: data,
		})
	}

	return out, func(x float64) float64 {
		return (x - min) / (max - min)
	}
}
