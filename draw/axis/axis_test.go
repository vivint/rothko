// Copyright (C) 2018. See AUTHORS.

package axis

import (
	"context"
	"image/png"
	"os"
	"testing"

	"github.com/spacemonkeygo/rothko/draw"
	"github.com/spacemonkeygo/rothko/internal/assert"
	"golang.org/x/image/font/inconsolata"
)

var (
	ctx = context.Background()

	vopts = Options{
		Canvas: draw.NewRGB(5000, 5000),
		Face:   inconsolata.Regular8x16,
		Labels: []Label{
			{0.0, "0.0"},
			{0.1, "0.1"},
			{0.2, "0.2"},
			{0.3, "0.3"},
			{0.4, "0.4"},
			{0.5, "0.5"},
			{0.6, "0.6"},
			{0.7, "0.7"},
			{0.8, "0.8"},
			{0.9, "0.9"},
			{1.0, "1.0"},
		},
		Vertical: true,
		Length:   300,
	}

	hopts = Options{
		Canvas: draw.NewRGB(5000, 5000),
		Face:   inconsolata.Regular8x16,
		Labels: []Label{
			{0.0, "1/16 @ 00:00"},
			{0.1, "1/16 @ 01:00"},
			{0.2, "1/16 @ 02:00"},
			{0.3, "1/16 @ 03:00"},
			{0.4, "1/16 @ 04:00"},
			{0.5, "1/16 @ 05:00"},
			{0.6, "1/16 @ 06:00"},
			{0.7, "1/16 @ 07:00"},
			{0.8, "1/16 @ 08:00"},
			{0.9, "1/16 @ 09:00"},
			{1.0, "1/16 @ 10:00"},
		},
		Vertical: false,
		Length:   1000,
	}
)

func saveImage(t *testing.T, name string, out *draw.RGB) {
	if true { // set to false to save images
		return
	}

	fh, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer fh.Close()

	assert.NoError(t, png.Encode(fh, asImage(out)))
}

func TestDraw(t *testing.T) {
	t.Run("Vertical", func(t *testing.T) {
		width, height := Draw(ctx, vopts)
		saveImage(t, "testv.png", vopts.Canvas.View(0, 0, width, height))
	})

	t.Run("Horizontal", func(t *testing.T) {
		width, height := Draw(ctx, hopts)
		saveImage(t, "testh.png", hopts.Canvas.View(0, 0, width, height))
	})
}

func BenchmarkDraw(b *testing.B) {
	b.Run("Vertical", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			Draw(ctx, vopts)
		}
	})

	b.Run("Horizontal", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			Draw(ctx, hopts)
		}
	})
}
